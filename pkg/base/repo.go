package base

import (
	"gorm.io/gorm"
)

type BaseRepo[T any] struct {
	DB *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{DB: db}
}

func (r *BaseRepo[T]) Create(entity *T) error {
	return r.DB.Create(entity).Error
}

func (r *BaseRepo[T]) Update(entity *T) error {
	return r.DB.Save(entity).Error
}

func (r *BaseRepo[T]) Delete(entity *T) error {
	return r.DB.Delete(entity).Error
}

func (r *BaseRepo[T]) FindByID(id uint) (*T, error) {
	var entity T
	err := r.DB.First(&entity, id).Error
	return &entity, err
}
