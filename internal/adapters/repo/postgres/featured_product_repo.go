package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type FeaturedProductRepo struct{ db *gorm.DB }

func NewFeaturedProductRepo(db *gorm.DB) *FeaturedProductRepo {
	return &FeaturedProductRepo{db: db}
}

func (r *FeaturedProductRepo) Save(ctx context.Context, fp *domain.FeaturedProduct) error {
	if fp.ID == uuid.Nil {
		fp.ID = uuid.New()
	}
	if fp.CreatedAt.IsZero() {
		fp.CreatedAt = time.Now()
	}
	fp.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(fp).Error
}

func (r *FeaturedProductRepo) FindByProductID(ctx context.Context, productID uuid.UUID) (*domain.FeaturedProduct, error) {
	var fp domain.FeaturedProduct
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Product.Images").
		First(&fp, "product_id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &fp, nil
}

func (r *FeaturedProductRepo) FindAll(ctx context.Context) ([]domain.FeaturedProduct, error) {
	var list []domain.FeaturedProduct
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Product.Images").
		Where("active = ?", true).
		Order("display_order asc, created_at asc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *FeaturedProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.FeaturedProduct{}, "id = ?", id).Error
}

func (r *FeaturedProductRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.FeaturedProduct{}).
		Where("active = ?", true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
