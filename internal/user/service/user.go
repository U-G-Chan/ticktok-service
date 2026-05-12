package service

import (
	"context"
	"ticktok-service/api/user/v1"
	"ticktok-service/internal/user/model"
	"ticktok-service/internal/user/repository"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/snowflake"
	"ticktok-service/pkg/util"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	userRepo     *repository.UserRepo
	relationRepo *repository.RelationRepo
}

func NewUserService() *UserService {
	return &UserService{
		userRepo:     repository.NewUserRepo(mysql.DB),
		relationRepo: repository.NewRelationRepo(mysql.DB),
	}
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	// Check if user exists
	count, err := s.userRepo.CountByUsername(req.Username)
	if err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  "CountByUsername err: " + err.Error(),
		}, nil
	}
	if count > 0 {
		return &user.RegisterResponse{
			Code: int32(errno.ErrUserNotFound.Code), // TODO: Add ErrUserAlreadyExists
			Msg:  "User already exists",
		}, nil
	}

	// Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.InternalServerError.Code),
			Msg:  err.Error(),
		}, nil
	}

	newUser := &model.User{
		ID:       snowflake.GenerateMsgID(),
		Username: req.Username,
		Nickname: req.Username,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  "Create user err: " + err.Error(),
		}, nil
	}

	// Generate tokens
	accessToken, refreshToken, err := util.GenerateToken(int64(newUser.ID))
	if err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.ErrToken.Code),
			Msg:  errno.ErrToken.Message,
		}, nil
	}

	return &user.RegisterResponse{
		Code:         int32(errno.Success.Code),
		Msg:          errno.Success.Message,
		UserId:       int64(newUser.ID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	u, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return &user.LoginResponse{
			Code: int32(errno.ErrUserNotFound.Code),
			Msg:  errno.ErrUserNotFound.Message,
		}, nil
	}

	if !util.CheckPasswordHash(req.Password, u.Password) {
		return &user.LoginResponse{
			Code: int32(errno.ErrPasswordIncorrect.Code),
			Msg:  errno.ErrPasswordIncorrect.Message,
		}, nil
	}

	accessToken, refreshToken, err := util.GenerateToken(int64(u.ID))
	if err != nil {
		return &user.LoginResponse{
			Code: int32(errno.ErrToken.Code),
			Msg:  errno.ErrToken.Message,
		}, nil
	}

	return &user.LoginResponse{
		Code:         int32(errno.Success.Code),
		Msg:          errno.Success.Message,
		UserId:       int64(u.ID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *user.GetUserInfoRequest) (*user.GetUserInfoResponse, error) {
	u, err := s.userRepo.FindByID(req.UserId)
	if err != nil {
		return &user.GetUserInfoResponse{
			Code: int32(errno.ErrUserNotFound.Code),
			Msg:  errno.ErrUserNotFound.Message,
		}, nil
	}

	name := u.Username
	if u.Nickname != "" {
		name = u.Nickname
	}
	userInfo := &user.User{
		Id:              int64(u.ID),
		Name:            name,
		Avatar:          u.Avatar,
		BackgroundImage: u.BackgroundImage,
		Signature:       u.Signature,
		// TODO: Implement other fields
	}

	return &user.GetUserInfoResponse{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Message,
		User: userInfo,
	}, nil
}

func (s *UserService) MGetUserInfo(ctx context.Context, req *user.MGetUserInfoRequest) (*user.MGetUserInfoResponse, error) {
	if len(req.UserIds) == 0 {
		return &user.MGetUserInfoResponse{
			Code:  int32(errno.Success.Code),
			Msg:   errno.Success.Message,
			Users: []*user.User{},
		}, nil
	}

	// Remove duplicate user IDs
	uniqueIDs := make([]int64, 0, len(req.UserIds))
	seen := make(map[int64]struct{})
	for _, id := range req.UserIds {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	users, err := s.userRepo.FindByIDs(uniqueIDs)
	if err != nil {
		return &user.MGetUserInfoResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  err.Error(),
		}, nil
	}

	var pbUsers []*user.User
	for _, u := range users {
		name := u.Username
		if u.Nickname != "" {
			name = u.Nickname
		}
		isFollow := false
		if req.TokenUserId > 0 {
			rel, err := s.relationRepo.GetRelation(req.TokenUserId, int64(u.ID))
			if err == nil && rel.Status == 1 {
				isFollow = true
			}
		}

		pbUsers = append(pbUsers, &user.User{
			Id:              int64(u.ID),
			Name:            name,
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			IsFollow:        isFollow,
			// TODO: Implement other fields (follow count, etc.)
		})
	}

	return &user.MGetUserInfoResponse{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Message,
		Users: pbUsers,
	}, nil
}

func (s *UserService) RelationAction(ctx context.Context, req *user.RelationActionRequest) (*user.RelationActionResponse, error) {
	if req.UserId == 0 || req.ToUserId == 0 || req.UserId == req.ToUserId {
		return &user.RelationActionResponse{
			Code: int32(errno.ErrValidation.Code),
			Msg:  "invalid action",
		}, nil
	}

	// 1-follow, 2-unfollow
	status := int8(0)
	if req.ActionType == 1 {
		status = 1
	} else if req.ActionType == 2 {
		status = 0
	} else {
		return &user.RelationActionResponse{
			Code: int32(errno.ErrValidation.Code),
			Msg:  "invalid action type",
		}, nil
	}

	rel, err := s.relationRepo.GetRelation(req.UserId, req.ToUserId)
	if err != nil {
		// Create new relation
		rel = &model.Relation{
			ID:       snowflake.GenerateMsgID(),
			UserID:   req.UserId,
			ToUserID: req.ToUserId,
			Status:   status,
		}
	} else {
		rel.Status = status
	}

	if err := s.relationRepo.Upsert(rel); err != nil {
		return &user.RelationActionResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  err.Error(),
		}, nil
	}

	return &user.RelationActionResponse{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Message,
	}, nil
}

func (s *UserService) GetFollowList(ctx context.Context, req *user.GetFollowListRequest) (*user.GetFollowListResponse, error) {
	relations, err := s.relationRepo.GetFollowList(req.UserId)
	if err != nil {
		return &user.GetFollowListResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  err.Error(),
		}, nil
	}

	var userIDs []int64
	for _, r := range relations {
		userIDs = append(userIDs, r.ToUserID)
	}

	mgetResp, err := s.MGetUserInfo(ctx, &user.MGetUserInfoRequest{
		UserIds:     userIDs,
		TokenUserId: req.TokenUserId,
	})

	if err != nil {
		return &user.GetFollowListResponse{
			Code: int32(errno.InternalServerError.Code),
			Msg:  err.Error(),
		}, nil
	}

	return &user.GetFollowListResponse{
		Code:     int32(errno.Success.Code),
		Msg:      errno.Success.Message,
		UserList: mgetResp.Users,
	}, nil
}

func (s *UserService) GetFollowerList(ctx context.Context, req *user.GetFollowerListRequest) (*user.GetFollowerListResponse, error) {
	relations, err := s.relationRepo.GetFollowerList(req.UserId)
	if err != nil {
		return &user.GetFollowerListResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  err.Error(),
		}, nil
	}

	var userIDs []int64
	for _, r := range relations {
		userIDs = append(userIDs, r.UserID)
	}

	mgetResp, err := s.MGetUserInfo(ctx, &user.MGetUserInfoRequest{
		UserIds:     userIDs,
		TokenUserId: req.TokenUserId,
	})

	if err != nil {
		return &user.GetFollowerListResponse{
			Code: int32(errno.InternalServerError.Code),
			Msg:  err.Error(),
		}, nil
	}

	return &user.GetFollowerListResponse{
		Code:     int32(errno.Success.Code),
		Msg:      errno.Success.Message,
		UserList: mgetResp.Users,
	}, nil
}
