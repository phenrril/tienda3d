package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phenrril/tienda3d/internal/domain"
)

type CouponUseCase struct {
	repo      domain.CouponRepo
	orderRepo domain.OrderRepo
}

func NewCouponUseCase(repo domain.CouponRepo, orderRepo domain.OrderRepo) *CouponUseCase {
	return &CouponUseCase{
		repo:      repo,
		orderRepo: orderRepo,
	}
}

// ValidateCoupon valida que un cupón pueda ser usado por un usuario específico
func (uc *CouponUseCase) ValidateCoupon(ctx context.Context, code, email string, subtotal float64) (*domain.Coupon, error) {
	// Normalizar código y email
	code = strings.ToUpper(strings.TrimSpace(code))
	email = strings.ToLower(strings.TrimSpace(email))

	if code == "" {
		return nil, errors.New("código de cupón vacío")
	}

	if email == "" {
		return nil, errors.New("email requerido para validar cupón")
	}

	// 1. Buscar cupón por código
	coupon, err := uc.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, errors.New("cupón no encontrado")
		}
		return nil, fmt.Errorf("error al buscar cupón: %w", err)
	}

	// 2. Validar que está activo
	if !coupon.Active {
		return nil, errors.New("cupón desactivado")
	}

	// 3. Validar que no ha expirado
	if coupon.ExpiresAt != nil && time.Now().After(*coupon.ExpiresAt) {
		return nil, errors.New("cupón expirado")
	}

	// 4. Validar que no se alcanzó el límite de usos
	if coupon.MaxUses != nil && coupon.CurrentUses >= *coupon.MaxUses {
		return nil, errors.New("cupón alcanzó el límite de usos")
	}

	// 5. Validar monto mínimo de compra
	if subtotal < coupon.MinPurchaseAmount {
		return nil, fmt.Errorf("monto mínimo de compra no alcanzado (requerido: $%.2f)", coupon.MinPurchaseAmount)
	}

	// 6. Validar si el usuario ya lo usó (usos confirmados)
	usages, err := uc.repo.FindUsagesByEmail(ctx, email, coupon.ID)
	if err != nil {
		return nil, fmt.Errorf("error al verificar usos previos: %w", err)
	}

	if len(usages) > 0 {
		return nil, errors.New("ya has usado este cupón anteriormente")
	}

	// 7. Validar si el usuario tiene órdenes pendientes con este cupón
	pendingOrders, err := uc.orderRepo.FindPendingByEmailAndCoupon(ctx, email, code)
	if err != nil {
		return nil, fmt.Errorf("error al verificar órdenes pendientes: %w", err)
	}

	if len(pendingOrders) > 0 {
		return nil, fmt.Errorf("ya tienes %d orden(es) pendiente(s) con este cupón. Completa o cancela esa orden antes de usar el cupón nuevamente", len(pendingOrders))
	}

	return coupon, nil
}

// CalculateDiscount calcula el descuento a aplicar basado en el cupón y el subtotal
func (uc *CouponUseCase) CalculateDiscount(coupon *domain.Coupon, subtotal float64) float64 {
	if coupon == nil {
		return 0
	}

	var discount float64

	switch coupon.DiscountType {
	case domain.DiscountTypePercentage:
		discount = subtotal * (coupon.DiscountValue / 100.0)
	case domain.DiscountTypeFixedAmount:
		discount = coupon.DiscountValue
	default:
		return 0
	}

	// Limitar descuento al subtotal (no puede ser mayor)
	if discount > subtotal {
		discount = subtotal
	}

	// No puede ser negativo
	if discount < 0 {
		discount = 0
	}

	return discount
}

// ApplyCoupon registra el uso de un cupón después de que la orden fue creada exitosamente
func (uc *CouponUseCase) ApplyCoupon(ctx context.Context, couponID, orderID uuid.UUID, email string, discountApplied, orderTotal float64) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. Incrementar contador de usos
	if err := uc.repo.IncrementUses(ctx, couponID); err != nil {
		return fmt.Errorf("error al incrementar usos del cupón: %w", err)
	}

	// 2. Guardar registro de uso
	usage := &domain.CouponUsage{
		ID:              uuid.New(),
		CouponID:        couponID,
		OrderID:         orderID,
		Email:           email,
		DiscountApplied: discountApplied,
		OrderTotal:      orderTotal,
		UsedAt:          time.Now(),
	}

	if err := uc.repo.SaveUsage(ctx, usage); err != nil {
		return fmt.Errorf("error al guardar uso del cupón: %w", err)
	}

	return nil
}

// GetCouponStats obtiene estadísticas de uso de un cupón
func (uc *CouponUseCase) GetCouponStats(ctx context.Context, couponID uuid.UUID) (totalUses int64, totalDiscount float64, err error) {
	return uc.repo.GetUsageStats(ctx, couponID)
}

// CreateCoupon crea un nuevo cupón
func (uc *CouponUseCase) CreateCoupon(ctx context.Context, coupon *domain.Coupon) error {
	if coupon == nil {
		return errors.New("cupón nil")
	}

	// Validaciones básicas
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	if coupon.Code == "" {
		return errors.New("código de cupón requerido")
	}

	if coupon.DiscountValue <= 0 {
		return errors.New("valor de descuento debe ser mayor a 0")
	}

	if coupon.DiscountType == domain.DiscountTypePercentage && coupon.DiscountValue > 100 {
		return errors.New("porcentaje de descuento no puede ser mayor a 100")
	}

	// Verificar que el código no exista
	existing, err := uc.repo.FindByCode(ctx, coupon.Code)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("error al verificar código: %w", err)
	}
	if existing != nil {
		return errors.New("ya existe un cupón con ese código")
	}

	if coupon.ID == uuid.Nil {
		coupon.ID = uuid.New()
	}

	return uc.repo.Save(ctx, coupon)
}

// UpdateCoupon actualiza un cupón existente
func (uc *CouponUseCase) UpdateCoupon(ctx context.Context, coupon *domain.Coupon) error {
	if coupon == nil {
		return errors.New("cupón nil")
	}

	if coupon.ID == uuid.Nil {
		return errors.New("ID de cupón requerido")
	}

	// Validaciones básicas
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	if coupon.Code == "" {
		return errors.New("código de cupón requerido")
	}

	if coupon.DiscountValue <= 0 {
		return errors.New("valor de descuento debe ser mayor a 0")
	}

	if coupon.DiscountType == domain.DiscountTypePercentage && coupon.DiscountValue > 100 {
		return errors.New("porcentaje de descuento no puede ser mayor a 100")
	}

	// Verificar que el cupón existe
	existing, err := uc.repo.FindByID(ctx, coupon.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return errors.New("cupón no encontrado")
		}
		return fmt.Errorf("error al buscar cupón: %w", err)
	}

	// Verificar que el código no esté en uso por otro cupón
	if existing.Code != coupon.Code {
		codeCheck, err := uc.repo.FindByCode(ctx, coupon.Code)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("error al verificar código: %w", err)
		}
		if codeCheck != nil && codeCheck.ID != coupon.ID {
			return errors.New("ya existe otro cupón con ese código")
		}
	}

	return uc.repo.Save(ctx, coupon)
}

// GetCoupon obtiene un cupón por ID
func (uc *CouponUseCase) GetCoupon(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	return uc.repo.FindByID(ctx, id)
}

// ListCoupons lista cupones con paginación
func (uc *CouponUseCase) ListCoupons(ctx context.Context, activeOnly bool, page, pageSize int) ([]domain.Coupon, int64, error) {
	return uc.repo.List(ctx, activeOnly, page, pageSize)
}

// ToggleCoupon activa o desactiva un cupón
func (uc *CouponUseCase) ToggleCoupon(ctx context.Context, id uuid.UUID) error {
	coupon, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	coupon.Active = !coupon.Active
	return uc.repo.Save(ctx, coupon)
}

// DeleteCoupon elimina un cupón
func (uc *CouponUseCase) DeleteCoupon(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}
