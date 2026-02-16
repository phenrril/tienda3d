package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/domain"
)

type ProductRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) Save(ctx context.Context, p *domain.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *ProductRepo) AddImages(ctx context.Context, productID uuid.UUID, imgs []domain.Image) error {
	if len(imgs) == 0 {
		return nil
	}
	for i := range imgs {
		if imgs[i].ID == uuid.Nil {
			imgs[i].ID = uuid.New()
		}
		imgs[i].ProductID = productID
		if imgs[i].CreatedAt.IsZero() {
			imgs[i].CreatedAt = time.Now()
		}
	}
	return r.db.WithContext(ctx).Create(&imgs).Error
}

func (r *ProductRepo) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.WithContext(ctx).Preload("Images").Preload("Variants").First(&p, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) FindImageByID(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	var img domain.Image
	if err := r.db.WithContext(ctx).First(&img, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &img, nil
}

func (r *ProductRepo) DeleteImageByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Image{}, "id = ?", id).Error
}

func (r *ProductRepo) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	var list []domain.Product
	q := r.db.WithContext(ctx).Model(&domain.Product{})
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if len(f.ExcludeCategories) > 0 {
		q = q.Where("category NOT IN ?", f.ExcludeCategories)
	}
	if f.ReadyToShip != nil {
		q = q.Where("ready_to_ship = ?", *f.ReadyToShip)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("LOWER(name) LIKE LOWER(?) OR LOWER(category) LIKE LOWER(?)", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	switch f.Sort {
	case "price_desc":
		q = q.Order("base_price desc")
	case "price_asc":
		q = q.Order("base_price asc")
	case "newest":
		q = q.Order("created_at desc")
	default:
		q = q.Order("name asc")
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize
	if err := q.Offset(offset).Limit(f.PageSize).Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("created_at asc") }).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ProductRepo) DeleteBySlug(ctx context.Context, slug string) error {
	return r.db.WithContext(ctx).Where("slug = ?", slug).Delete(&domain.Product{}).Error
}

func (r *ProductRepo) DeleteFullBySlug(ctx context.Context, slug string) ([]string, error) {
	if slug == "" {
		return nil, errors.New("slug vacío")
	}
	var p domain.Product
	if err := r.db.WithContext(ctx).Preload("Images").Preload("Variants").First(&p, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	imgPaths := []string{}
	for _, im := range p.Images {
		imgPaths = append(imgPaths, im.URL)
	}
	return imgPaths, r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", p.ID).Delete(&domain.Image{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", p.ID).Delete(&domain.Variant{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&domain.Product{}, "id = ?", p.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ProductRepo) DistinctCategories(ctx context.Context) ([]string, error) {
	cats := []string{}
	if err := r.db.WithContext(ctx).Model(&domain.Product{}).
		Distinct("category").Where("category <> ''").Order("category asc").Pluck("category", &cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}
