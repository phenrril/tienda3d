package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type CouponRepo struct{ db *gorm.DB }

func NewCouponRepo(db *gorm.DB) *CouponRepo { return &CouponRepo{db: db} }

func (r *CouponRepo) Save(ctx context.Context, c *domain.Coupon) error {
	if c == nil {
		return errors.New("coupon nil")
	}

	// Normalizar el código a mayúsculas
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))

	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Coupon{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		// Crear nuevo cupón
		return r.db.WithContext(ctx).Create(c).Error
	}

	// Actualizar cupón existente
	return r.db.WithContext(ctx).Model(&domain.Coupon{}).Where("id = ?", c.ID).Updates(map[string]any{
		"code":                c.Code,
		"discount_type":       c.DiscountType,
		"discount_value":      c.DiscountValue,
		"min_purchase_amount": c.MinPurchaseAmount,
		"max_uses":            c.MaxUses,
		"current_uses":        c.CurrentUses,
		"expires_at":          c.ExpiresAt,
		"active":              c.Active,
		"description":         c.Description,
	}).Error
}

func (r *CouponRepo) FindByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	var c domain.Coupon
	if err := r.db.WithContext(ctx).First(&c, "UPPER(code) = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	var c domain.Coupon
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepo) List(ctx context.Context, activeOnly bool, page, pageSize int) ([]domain.Coupon, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Coupon{})
	if activeOnly {
		q = q.Where("active = ?", true)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []domain.Coupon
	if err := q.Order("created_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *CouponRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Coupon{}, "id = ?", id).Error
}

func (r *CouponRepo) IncrementUses(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Coupon{}).Where("id = ?", id).
		UpdateColumn("current_uses", gorm.Expr("current_uses + ?", 1)).Error
}

func (r *CouponRepo) SaveUsage(ctx context.Context, usage *domain.CouponUsage) error {
	if usage == nil {
		return errors.New("usage nil")
	}

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}

	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *CouponRepo) FindUsagesByEmail(ctx context.Context, email string, couponID uuid.UUID) ([]domain.CouponUsage, error) {
	var usages []domain.CouponUsage
	if err := r.db.WithContext(ctx).
		Where("email = ? AND coupon_id = ?", strings.ToLower(strings.TrimSpace(email)), couponID).
		Order("used_at desc").
		Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

func (r *CouponRepo) GetUsageStats(ctx context.Context, couponID uuid.UUID) (totalUses int64, totalDiscount float64, err error) {
	type Result struct {
		TotalUses     int64
		TotalDiscount float64
	}

	var result Result
	if err := r.db.WithContext(ctx).Model(&domain.CouponUsage{}).
		Select("COUNT(*) as total_uses, COALESCE(SUM(discount_applied), 0) as total_discount").
		Where("coupon_id = ?", couponID).
		Scan(&result).Error; err != nil {
		return 0, 0, err
	}

	return result.TotalUses, result.TotalDiscount, nil
}
