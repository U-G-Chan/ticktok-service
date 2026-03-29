package repository

import (
	"ticktok-service/internal/user/model"
	"ticktok-service/pkg/common"

	"gorm.io/gorm"
)

type UserRepo struct {
	*common.BaseRepo[model.User]
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		BaseRepo: common.NewBaseRepo[model.User](db),
	}
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByIDs(ids []int64) ([]*model.User, error) {
	var users []*model.User
	if len(ids) == 0 {
		return users, nil
	}
	err := r.DB.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *UserRepo) CountByUsername(username string) (int64, error) {
	var count int64
	err := r.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count, err
}
