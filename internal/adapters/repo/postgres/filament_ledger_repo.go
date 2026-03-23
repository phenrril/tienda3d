package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type FilamentLedgerRepo struct{ db *gorm.DB }

func NewFilamentLedgerRepo(db *gorm.DB) *FilamentLedgerRepo {
	return &FilamentLedgerRepo{db: db}
}

func (r *FilamentLedgerRepo) AddPurchase(ctx context.Context, colorSlug string, grams int, unitCost float64, note string) error {
	if grams <= 0 {
		grams = 1000
	}
	cost := unitCost
	e := domain.FilamentLedgerEntry{
		ID:         uuid.New(),
		ColorSlug:  colorSlug,
		DeltaGrams: grams,
		EntryType:  domain.FilamentEntryPurchase,
		UnitCost:   &cost,
		Note:       note,
		CreatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(&e).Error
}

func (r *FilamentLedgerRepo) ListInRange(ctx context.Context, from, to time.Time) ([]domain.FilamentLedgerEntry, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), to.Location())
	var list []domain.FilamentLedgerEntry
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", from, to).
		Order("created_at asc").
		Find(&list).Error
	return list, err
}

func (r *FilamentLedgerRepo) PurchaseTotalsByColor(ctx context.Context) (map[string]struct{ Grams int64; Cost float64 }, error) {
	type row struct {
		Color string
		Grams int64
		Cost  float64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&domain.FilamentLedgerEntry{}).
		Select("color_slug as color, SUM(delta_grams) as grams, COALESCE(SUM(unit_cost),0) as cost").
		Where("entry_type = ?", domain.FilamentEntryPurchase).
		Group("color_slug").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{ Grams int64; Cost float64 })
	for _, rw := range rows {
		out[rw.Color] = struct{ Grams int64; Cost float64 }{Grams: rw.Grams, Cost: rw.Cost}
	}
	return out, nil
}

func (r *FilamentLedgerRepo) StockByColor(ctx context.Context) (map[string]int, error) {
	type row struct {
		Color string
		Sum   int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&domain.FilamentLedgerEntry{}).
		Select("color_slug as color, SUM(delta_grams) as sum").
		Group("color_slug").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int)
	for _, rw := range rows {
		out[rw.Color] = int(rw.Sum)
	}
	return out, nil
}

func (r *FilamentLedgerRepo) TotalPurchasesInRange(ctx context.Context, from, to time.Time) (float64, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), to.Location())
	var sum *float64
	err := r.db.WithContext(ctx).Model(&domain.FilamentLedgerEntry{}).
		Where("entry_type = ? AND created_at BETWEEN ? AND ?", domain.FilamentEntryPurchase, from, to).
		Select("COALESCE(SUM(unit_cost),0)").
		Scan(&sum).Error
	if err != nil || sum == nil {
		return 0, err
	}
	return *sum, nil
}
