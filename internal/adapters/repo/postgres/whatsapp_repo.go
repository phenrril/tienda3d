package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/phenrril/tienda3d/internal/domain"
	"gorm.io/gorm"
)

type WhatsAppRepo struct {
	db *gorm.DB
}

func NewWhatsAppRepo(db *gorm.DB) domain.WhatsAppRepo {
	return &WhatsAppRepo{db: db}
}

// SaveOrder guarda una orden de WhatsApp
func (r *WhatsAppRepo) SaveOrder(ctx context.Context, order *domain.WhatsAppOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// FindOrderByWhatsAppID busca una orden por su ID de WhatsApp
func (r *WhatsAppRepo) FindOrderByWhatsAppID(ctx context.Context, whatsappID string) (*domain.WhatsAppOrder, error) {
	var order domain.WhatsAppOrder
	err := r.db.WithContext(ctx).Where("whatsapp_id = ?", whatsappID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

// ListPendingOrders lista órdenes pendientes de procesar
func (r *WhatsAppRepo) ListPendingOrders(ctx context.Context) ([]domain.WhatsAppOrder, error) {
	var orders []domain.WhatsAppOrder
	err := r.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at ASC").Find(&orders).Error
	return orders, err
}

// UpdateOrderStatus actualiza el estado de una orden
func (r *WhatsAppRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, orderID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": sql.NullTime{Valid: true, Time: time.Now()},
	}

	if orderID != nil {
		updates["order_id"] = *orderID
	}

	return r.db.WithContext(ctx).Model(&domain.WhatsAppOrder{}).Where("id = ?", id).Updates(updates).Error
}

// SaveProductSync guarda información de sincronización de producto
func (r *WhatsAppRepo) SaveProductSync(ctx context.Context, sync *domain.WhatsAppProductSync) error {
	return r.db.WithContext(ctx).Create(sync).Error
}

// FindProductSyncByWhatsAppID busca sincronización por ID de WhatsApp
func (r *WhatsAppRepo) FindProductSyncByWhatsAppID(ctx context.Context, whatsappID string) (*domain.WhatsAppProductSync, error) {
	var sync domain.WhatsAppProductSync
	err := r.db.WithContext(ctx).Where("whatsapp_product_id = ?", whatsappID).First(&sync).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &sync, nil
}

// ListProductsToSync lista productos que necesitan sincronización
func (r *WhatsAppRepo) ListProductsToSync(ctx context.Context) ([]domain.WhatsAppProductSync, error) {
	var syncs []domain.WhatsAppProductSync
	err := r.db.WithContext(ctx).Where("sync_status IN ?", []string{"pending", "failed"}).Order("created_at ASC").Find(&syncs).Error
	return syncs, err
}

// UpdateSyncStatus actualiza el estado de sincronización
func (r *WhatsAppRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"sync_status": status,
		"last_synced": time.Now(),
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return r.db.WithContext(ctx).Model(&domain.WhatsAppProductSync{}).Where("id = ?", id).Updates(updates).Error
}
