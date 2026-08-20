package repositories

import (
	"log"
	"time"

	"lms-backend/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll() ([]models.LoanProduct, error)
	FindByID(id int64) (models.LoanProduct, error)
	Create(product *models.LoanProduct) error
	Save(product *models.LoanProduct) error
	Delete(id int64) error
	WarmCache() error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

// WarmCache (No-Op: Membaca langsung dari PostgreSQL tanpa caching di RAM).
func (r *productRepository) WarmCache() error {
	log.Println("[PRODUCT-REPO] Caching dinonaktifkan. Data produk pinjaman dibaca real-time dari database PostgreSQL.")
	return nil
}

// FindAll mengembalikan semua produk langsung dari database PostgreSQL secara real-time.
func (r *productRepository) FindAll() ([]models.LoanProduct, error) {
	log.Println("[PRODUCT-REPO] Reading loan products real-time from PostgreSQL database...")
	var products []models.LoanProduct
	err := r.db.Order("id ASC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// FindByID mengembalikan produk berdasarkan ID langsung dari database PostgreSQL.
func (r *productRepository) FindByID(id int64) (models.LoanProduct, error) {
	var product models.LoanProduct
	err := r.db.Where("id = ?", id).First(&product).Error
	return product, err
}

// Create menyimpan produk baru langsung ke database PostgreSQL.
func (r *productRepository) Create(product *models.LoanProduct) error {
	return r.db.Create(product).Error
}

// Save menyimpan (insert atau update) produk langsung ke database PostgreSQL.
func (r *productRepository) Save(product *models.LoanProduct) error {
	if product.ID > 0 {
		return r.db.Save(product).Error
	}
	return r.db.Create(product).Error
}

// Delete melakukan soft-delete produk di database PostgreSQL.
func (r *productRepository) Delete(id int64) error {
	return r.db.Model(&models.LoanProduct{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
