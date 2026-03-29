package service

import (
	"context"
	"ticktok-service/api/user/v1"
	"ticktok-service/internal/user/model"
	"ticktok-service/internal/user/repository"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/util"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	userRepo *repository.UserRepo
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepo(mysql.DB),
	}
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	// Check if user exists
	count, err := s.userRepo.CountByUsername(req.Username)
	if err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  errno.ErrDatabase.Message,
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
		Username: req.Username,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return &user.RegisterResponse{
			Code: int32(errno.ErrDatabase.Code),
			Msg:  errno.ErrDatabase.Message,
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
	u, err := s.userRepo.FindByID(uint(req.UserId))
	if err != nil {
		return &user.GetUserInfoResponse{
			Code: int32(errno.ErrUserNotFound.Code),
			Msg:  errno.ErrUserNotFound.Message,
		}, nil
	}

	userInfo := &user.User{
		Id:              int64(u.ID),
		Name:            u.Username,
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
		pbUsers = append(pbUsers, &user.User{
			Id:              int64(u.ID),
			Name:            u.Username,
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			// TODO: Implement other fields (follow count, etc.)
		})
	}

	return &user.MGetUserInfoResponse{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Message,
		Users: pbUsers,
	}, nil
}
