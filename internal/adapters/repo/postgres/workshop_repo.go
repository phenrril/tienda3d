package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type WorkshopRepo struct{ db *gorm.DB }

func NewWorkshopRepo(db *gorm.DB) *WorkshopRepo { return &WorkshopRepo{db: db} }

func (r *WorkshopRepo) List(ctx context.Context) ([]domain.WorkshopOrder, error) {
	var list []domain.WorkshopOrder
	err := r.db.WithContext(ctx).
		Preload("Deposits", func(db *gorm.DB) *gorm.DB { return db.Order("paid_at asc, created_at asc") }).
		Preload("Filaments").
		Order("delivery_date asc NULLS LAST, created_at desc").
		Find(&list).Error
	return list, err
}

func (r *WorkshopRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.WorkshopOrder, error) {
	var o domain.WorkshopOrder
	err := r.db.WithContext(ctx).
		Preload("Deposits", func(db *gorm.DB) *gorm.DB { return db.Order("paid_at asc, created_at asc") }).
		Preload("Filaments").
		First(&o, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *WorkshopRepo) Save(ctx context.Context, o *domain.WorkshopOrder) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.WorkshopOrder{}).Where("id = ?", o.ID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return r.db.WithContext(ctx).Create(o).Error
	}
	return r.db.WithContext(ctx).Model(&domain.WorkshopOrder{}).Where("id = ?", o.ID).Updates(map[string]any{
		"client_slug":    o.ClientSlug,
		"requested_at":   o.RequestedAt,
		"delivery_date":  o.DeliveryDate,
		"detail":         o.Detail,
		"total_amount":   o.TotalAmount,
		"is_barter":      o.IsBarter,
		"status":         o.Status,
		"delivered_at":   o.DeliveredAt,
		"updated_at":     time.Now(),
	}).Error
}

func (r *WorkshopRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ref_workshop_order_id = ?", id).Delete(&domain.FilamentLedgerEntry{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workshop_order_id = ?", id).Delete(&domain.WorkshopDeposit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workshop_order_id = ?", id).Delete(&domain.WorkshopOrderFilament{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.WorkshopOrder{}, "id = ?", id).Error
	})
}

func (r *WorkshopRepo) AddDeposit(ctx context.Context, d *domain.WorkshopDeposit) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *WorkshopRepo) ListDeposits(ctx context.Context, orderID uuid.UUID) ([]domain.WorkshopDeposit, error) {
	var list []domain.WorkshopDeposit
	err := r.db.WithContext(ctx).Where("workshop_order_id = ?", orderID).Order("paid_at asc, created_at asc").Find(&list).Error
	return list, err
}

func (r *WorkshopRepo) SumDeposits(ctx context.Context, orderID uuid.UUID) (float64, error) {
	var sum *float64
	err := r.db.WithContext(ctx).Model(&domain.WorkshopDeposit{}).
		Where("workshop_order_id = ?", orderID).
		Select("COALESCE(SUM(amount),0)").Scan(&sum).Error
	if err != nil || sum == nil {
		return 0, err
	}
	return *sum, nil
}

func (r *WorkshopRepo) ListUpcomingForDigest(ctx context.Context, from, toDate time.Time) ([]domain.WorkshopOrder, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	toD := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, toDate.Location())
	var list []domain.WorkshopOrder
	err := r.db.WithContext(ctx).
		Where("delivery_date >= ? AND delivery_date <= ? AND status <> ?", fromD, toD, domain.WorkshopEntregado).
		Order("delivery_date asc, client_slug asc").
		Find(&list).Error
	return list, err
}

func (r *WorkshopRepo) FindUndeliveredByClientSlug(ctx context.Context, slug string) ([]domain.WorkshopOrder, error) {
	var list []domain.WorkshopOrder
	err := r.db.WithContext(ctx).
		Where("LOWER(client_slug) = LOWER(?) AND status <> ?", slug, domain.WorkshopEntregado).
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (r *WorkshopRepo) ListDeliveredInRange(ctx context.Context, from, to time.Time) ([]domain.WorkshopOrder, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), to.Location())
	var list []domain.WorkshopOrder
	err := r.db.WithContext(ctx).
		Where("status = ? AND delivered_at IS NOT NULL AND delivered_at BETWEEN ? AND ?", domain.WorkshopEntregado, from, to).
		Find(&list).Error
	return list, err
}

func (r *WorkshopRepo) UpdateStatus(ctx context.Context, id uuid.UUID, st domain.WorkshopOrderStatus) error {
	updates := map[string]any{"status": st, "updated_at": time.Now()}
	if st == domain.WorkshopEntregado {
		now := time.Now()
		updates["delivered_at"] = &now
	}
	return r.db.WithContext(ctx).Model(&domain.WorkshopOrder{}).Where("id = ?", id).Updates(updates).Error
}

func (r *WorkshopRepo) SaveOrderWithFilaments(ctx context.Context, o *domain.WorkshopOrder, filaments []domain.WorkshopOrderFilament) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&domain.WorkshopOrder{}).Where("id = ?", o.ID).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			if err := tx.Where("ref_workshop_order_id = ? AND entry_type = ?", o.ID, domain.FilamentEntryConsumption).
				Delete(&domain.FilamentLedgerEntry{}).Error; err != nil {
				return err
			}
			if err := tx.Where("workshop_order_id = ?", o.ID).Delete(&domain.WorkshopOrderFilament{}).Error; err != nil {
				return err
			}
		}

		if count == 0 {
			if err := tx.Create(o).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&domain.WorkshopOrder{}).Where("id = ?", o.ID).Updates(map[string]any{
				"client_slug":   o.ClientSlug,
				"requested_at":  o.RequestedAt,
				"delivery_date": o.DeliveryDate,
				"detail":        o.Detail,
				"total_amount":  o.TotalAmount,
				"is_barter":     o.IsBarter,
				"status":        o.Status,
				"delivered_at":  o.DeliveredAt,
				"updated_at":    time.Now(),
			}).Error; err != nil {
				return err
			}
		}

		for i := range filaments {
			if filaments[i].Grams <= 0 {
				continue
			}
			if filaments[i].ID == uuid.Nil {
				filaments[i].ID = uuid.New()
			}
			filaments[i].WorkshopOrderID = o.ID
			if err := tx.Create(&filaments[i]).Error; err != nil {
				return err
			}
			oid := o.ID
			leg := domain.FilamentLedgerEntry{
				ID:                 uuid.New(),
				ColorSlug:          filaments[i].ColorSlug,
				DeltaGrams:         -filaments[i].Grams,
				EntryType:          domain.FilamentEntryConsumption,
				RefWorkshopOrderID: &oid,
				CreatedAt:          time.Now(),
			}
			if err := tx.Create(&leg).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
