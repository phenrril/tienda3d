package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/phenrril/tienda3d/internal/domain"
)

type AppSettingRepo struct{ db *gorm.DB }

func NewAppSettingRepo(db *gorm.DB) *AppSettingRepo { return &AppSettingRepo{db: db} }

func (r *AppSettingRepo) Get(ctx context.Context, key string) (string, error) {
	var s domain.AppSetting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return s.Value, nil
}

func (r *AppSettingRepo) Set(ctx context.Context, key, value string) error {
	s := domain.AppSetting{Key: key, Value: value}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&s).Error
}
