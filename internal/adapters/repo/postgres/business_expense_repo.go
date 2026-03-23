package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type BusinessExpenseRepo struct{ db *gorm.DB }

func NewBusinessExpenseRepo(db *gorm.DB) *BusinessExpenseRepo {
	return &BusinessExpenseRepo{db: db}
}

func (r *BusinessExpenseRepo) Save(ctx context.Context, e *domain.BusinessExpense) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *BusinessExpenseRepo) ListInRange(ctx context.Context, from, to time.Time) ([]domain.BusinessExpense, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), to.Location())
	var list []domain.BusinessExpense
	err := r.db.WithContext(ctx).
		Where("spent_at BETWEEN ? AND ?", from, to).
		Order("spent_at desc, created_at desc").
		Find(&list).Error
	return list, err
}

func (r *BusinessExpenseRepo) SumInRange(ctx context.Context, from, to time.Time) (float64, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), to.Location())
	var sum *float64
	err := r.db.WithContext(ctx).Model(&domain.BusinessExpense{}).
		Where("spent_at BETWEEN ? AND ?", from, to).
		Select("COALESCE(SUM(amount),0)").
		Scan(&sum).Error
	if err != nil || sum == nil {
		return 0, err
	}
	return *sum, nil
}
