package postgres

import (
    "context"

    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/phenrril/tienda3d/internal/domain"
)

type ExpenseRepo struct{ db *gorm.DB }

func NewExpenseRepo(db *gorm.DB) *ExpenseRepo { return &ExpenseRepo{db: db} }

func (r *ExpenseRepo) Save(ctx context.Context, e *domain.Expense) error {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    return r.db.WithContext(ctx).Save(e).Error
}

func (r *ExpenseRepo) Recent(ctx context.Context, limit int) ([]domain.Expense, error) {
    if limit <= 0 || limit > 200 {
        limit = 50
    }
    var list []domain.Expense
    err := r.db.WithContext(ctx).Order("date desc, created_at desc").Limit(limit).Find(&list).Error
    return list, err
}
