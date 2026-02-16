package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type HiddenCategoryRepo struct{ db *gorm.DB }

func NewHiddenCategoryRepo(db *gorm.DB) *HiddenCategoryRepo {
	return &HiddenCategoryRepo{db: db}
}

func (r *HiddenCategoryRepo) FindAll(ctx context.Context) ([]domain.HiddenCategory, error) {
	var list []domain.HiddenCategory
	if err := r.db.WithContext(ctx).Order("category asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *HiddenCategoryRepo) ReplaceAll(ctx context.Context, categories []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&domain.HiddenCategory{}).Error; err != nil {
			return err
		}
		for _, cat := range categories {
			hc := domain.HiddenCategory{
				ID:       uuid.New(),
				Category: cat,
			}
			if err := tx.Create(&hc).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
