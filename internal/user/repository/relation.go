package repository

import (
	"ticktok-service/internal/user/model"
	"ticktok-service/pkg/common"

	"gorm.io/gorm"
)

type RelationRepo struct {
	*common.BaseRepo[model.Relation]
}

func NewRelationRepo(db *gorm.DB) *RelationRepo {
	return &RelationRepo{
		BaseRepo: common.NewBaseRepo[model.Relation](db),
	}
}

func (r *RelationRepo) GetRelation(userID, toUserID int64) (*model.Relation, error) {
	var relation model.Relation
	err := r.DB.Where("user_id = ? AND to_user_id = ?", userID, toUserID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

func (r *RelationRepo) Upsert(relation *model.Relation) error {
	return r.DB.Save(relation).Error
}

func (r *RelationRepo) GetFollowList(userID int64) ([]*model.Relation, error) {
	var relations []*model.Relation
	err := r.DB.Where("user_id = ? AND status = 1", userID).Order("created_at desc").Find(&relations).Error
	return relations, err
}

func (r *RelationRepo) GetFollowerList(toUserID int64) ([]*model.Relation, error) {
	var relations []*model.Relation
	err := r.DB.Where("to_user_id = ? AND status = 1", toUserID).Order("created_at desc").Find(&relations).Error
	return relations, err
}
