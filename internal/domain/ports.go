package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProductRepo interface {
	Save(ctx context.Context, p *Product) error
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	AddImages(ctx context.Context, productID uuid.UUID, imgs []Image) error
	FindImageByID(ctx context.Context, id uuid.UUID) (*Image, error)
	DeleteImageByID(ctx context.Context, id uuid.UUID) error
	DistinctCategories(ctx context.Context) ([]string, error)
}

type CustomerRepo interface {
	FindByEmail(ctx context.Context, email string) (*Customer, error)
	Save(ctx context.Context, c *Customer) error
}

type ProductFilter struct {
	Category          string
	ReadyToShip       *bool
	Sort              string
	Page              int
	PageSize          int
	Query             string
	ExcludeCategories []string
}

type OrderRepo interface {
	Save(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindByPreferenceID(ctx context.Context, prefID string) (*Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, st OrderStatus) error
	List(ctx context.Context, status *OrderStatus, mpStatus *string, page, pageSize int) ([]Order, int64, error)
	ListInRange(ctx context.Context, from, to time.Time) ([]Order, error)
	FindPendingByEmailAndCoupon(ctx context.Context, email, couponCode string) ([]Order, error)
}

type QuoteRepo interface {
	Save(ctx context.Context, q *Quote) error
	FindByID(ctx context.Context, id uuid.UUID) (*Quote, error)
}

type UploadedModelRepo interface {
	Save(ctx context.Context, m *UploadedModel) error
	FindByID(ctx context.Context, id uuid.UUID) (*UploadedModel, error)
}

type PageRepo interface {
	FindBySlug(ctx context.Context, slug string) (*Page, error)
	Save(ctx context.Context, p *Page) error
}

type FeaturedProductRepo interface {
	Save(ctx context.Context, fp *FeaturedProduct) error
	FindByProductID(ctx context.Context, productID uuid.UUID) (*FeaturedProduct, error)
	FindAll(ctx context.Context) ([]FeaturedProduct, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

type HiddenCategoryRepo interface {
	FindAll(ctx context.Context) ([]HiddenCategory, error)
	ReplaceAll(ctx context.Context, categories []string) error
}

type CouponRepo interface {
	Save(ctx context.Context, c *Coupon) error
	FindByCode(ctx context.Context, code string) (*Coupon, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Coupon, error)
	List(ctx context.Context, activeOnly bool, page, pageSize int) ([]Coupon, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUses(ctx context.Context, id uuid.UUID) error

	// Tracking de uso
	SaveUsage(ctx context.Context, usage *CouponUsage) error
	FindUsagesByEmail(ctx context.Context, email string, couponID uuid.UUID) ([]CouponUsage, error)
	GetUsageStats(ctx context.Context, couponID uuid.UUID) (totalUses int64, totalDiscount float64, err error)
}

type QuoteService interface {
	EstimateFromModel(ctx context.Context, modelID uuid.UUID, cfg QuoteConfig) (*Quote, error)
}

type PricingService interface {
	Price(volumeCM3 float64, timeMin int, material Material, quality PrintQuality, infillPct int, layerMM float64) (float64, map[string]float64)
}

type PaymentGateway interface {
	CreatePreference(ctx context.Context, o *Order) (initPoint string, err error)
	VerifyWebhook(signature string, body []byte) (event interface{}, err error)
	PaymentInfo(ctx context.Context, paymentID string) (status string, externalRef string, err error)
}

type FileStorage interface {
	SaveModel(ctx context.Context, filename string, data []byte) (string, error)
	SaveImage(ctx context.Context, filename string, data []byte) (string, error)
}

type EmailService interface {
	SendOrderConfirmation(ctx context.Context, order *Order) error
}

type Clock interface{ Now() time.Time }

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
