package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/phenrril/tienda3d/internal/domain"
)

type ProductUC struct {
	Products domain.ProductRepo
}

func (uc *ProductUC) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	if f.PageSize == 0 {
		f.PageSize = 20
	}
	return uc.Products.List(ctx, f)
}

func (uc *ProductUC) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	if slug == "" {
		return nil, errors.New("slug vacío")
	}
	return uc.Products.FindBySlug(ctx, slug)
}

func (uc *ProductUC) Create(ctx context.Context, p *domain.Product) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.Slug = strings.ToLower(strings.ReplaceAll(p.Name, " ", "-"))
	return uc.Products.Save(ctx, p)
}

func (uc *ProductUC) Update(ctx context.Context, p *domain.Product) error {
	// No regeneramos el slug en actualizaciones
	return uc.Products.Save(ctx, p)
}

func (uc *ProductUC) AddImages(ctx context.Context, productID uuid.UUID, imgs []domain.Image) error {
	return uc.Products.AddImages(ctx, productID, imgs)
}

func (uc *ProductUC) GetImageByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	if id == uuid.Nil {
		return nil, errors.New("id vacío")
	}
	if repo, ok := uc.Products.(interface {
		FindImageByID(context.Context, uuid.UUID) (*domain.Image, error)
	}); ok {
		return repo.FindImageByID(ctx, id)
	}
	return nil, errors.New("repo no soporta find image")
}

func (uc *ProductUC) DeleteImageByID(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("id vacío")
	}
	if repo, ok := uc.Products.(interface {
		DeleteImageByID(context.Context, uuid.UUID) error
	}); ok {
		return repo.DeleteImageByID(ctx, id)
	}
	return errors.New("repo no soporta delete image")
}

func (uc *ProductUC) DeleteBySlug(ctx context.Context, slug string) error {
	if slug == "" {
		return errors.New("slug vacío")
	}

	if repo, ok := uc.Products.(interface {
		DeleteBySlug(context.Context, string) error
	}); ok {
		return repo.DeleteBySlug(ctx, slug)
	}
	return errors.New("repo no soporta delete")
}

func (uc *ProductUC) DeleteFullBySlug(ctx context.Context, slug string) ([]string, error) {
	if slug == "" {
		return nil, errors.New("slug vacío")
	}
	if repo, ok := uc.Products.(interface {
		DeleteFullBySlug(context.Context, string) ([]string, error)
	}); ok {
		return repo.DeleteFullBySlug(ctx, slug)
	}

	return nil, uc.DeleteBySlug(ctx, slug)
}

func (uc *ProductUC) BulkUpdatePrices(ctx context.Context, updates []domain.PriceUpdate) error {
	if len(updates) == 0 {
		return errors.New("sin cambios")
	}
	for _, u := range updates {
		if u.Slug == "" {
			return errors.New("slug vacío en bulk update")
		}
		if u.BasePrice != nil && *u.BasePrice < 0 {
			return errors.New("precio base negativo")
		}
		if u.GrossPrice != nil && *u.GrossPrice < 0 {
			return errors.New("precio bruto negativo")
		}
		if u.Profit != nil && *u.Profit < 0 {
			return errors.New("ganancia negativa")
		}
	}
	return uc.Products.BulkUpdatePrices(ctx, updates)
}

func (uc *ProductUC) Categories(ctx context.Context) ([]string, error) {
	if repo, ok := uc.Products.(interface {
		DistinctCategories(context.Context) ([]string, error)
	}); ok {
		return repo.DistinctCategories(ctx)
	}
	return []string{}, nil
}
