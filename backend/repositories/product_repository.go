package repositories

import (
	"time"

	"lms-backend/cache"
	"lms-backend/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll() ([]models.LoanProduct, error)
	FindByID(id int64) (models.LoanProduct, error)
	Create(product *models.LoanProduct) error
	Save(product *models.LoanProduct) error
	Delete(id int64) error
	// WarmCache memuat semua produk dari database ke cache.
	// Dipanggil saat backend startup.
	WarmCache() error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

// WarmCache memuat semua produk dari database ke cache saat startup.
func (r *productRepository) WarmCache() error {
	var products []models.LoanProduct
	err := r.db.Where("deleted_at IS NULL").Order("id ASC").Find(&products).Error
	if err != nil {
		return err
	}
	cache.ProductCache.Set(products)
	return nil
}

// FindAll mengembalikan semua produk. Membaca dari cache jika tersedia,
// jika tidak akan query ke database dan simpan hasilnya ke cache.
func (r *productRepository) FindAll() ([]models.LoanProduct, error) {
	if products, ok := cache.ProductCache.Get(); ok {
		return products, nil
	}
	// Cache miss: load dari database
	var products []models.LoanProduct
	err := r.db.Where("deleted_at IS NULL").Order("id ASC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	cache.ProductCache.Set(products)
	return products, nil
}

// FindByID mengembalikan produk berdasarkan ID. Mencari dari cache terlebih dahulu.
func (r *productRepository) FindByID(id int64) (models.LoanProduct, error) {
	if product, ok := cache.ProductCache.GetByID(id); ok {
		return product, nil
	}
	// Cache miss: query langsung ke database
	var product models.LoanProduct
	err := r.db.Where("deleted_at IS NULL AND id = ?", id).First(&product).Error
	return product, err
}

// Create menyimpan produk baru ke database dan invalidate cache.
func (r *productRepository) Create(product *models.LoanProduct) error {
	err := r.db.Create(product).Error
	if err == nil {
		cache.ProductCache.Invalidate()
	}
	return err
}

// Save menyimpan (insert atau update) produk dan invalidate cache.
func (r *productRepository) Save(product *models.LoanProduct) error {
	var err error
	if product.ID > 0 {
		err = r.db.Save(product).Error
	} else {
		err = r.db.Create(product).Error
	}
	if err == nil {
		cache.ProductCache.Invalidate()
	}
	return err
}

// Delete melakukan soft-delete produk dan invalidate cache.
func (r *productRepository) Delete(id int64) error {
	err := r.db.Model(&models.LoanProduct{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
	if err == nil {
		cache.ProductCache.Invalidate()
	}
	return err
}
