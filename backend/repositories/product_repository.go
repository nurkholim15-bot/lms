package repositories

import (
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
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db}
}

func (r *productRepository) FindAll() ([]models.LoanProduct, error) {
	var products []models.LoanProduct
	err := r.db.Where("deleted_at IS NULL").Order("id ASC").Find(&products).Error
	return products, err
}

func (r *productRepository) FindByID(id int64) (models.LoanProduct, error) {
	var product models.LoanProduct
	err := r.db.Where("deleted_at IS NULL AND id = ?", id).First(&product).Error
	return product, err
}

func (r *productRepository) Create(product *models.LoanProduct) error {
	return r.db.Create(product).Error
}

func (r *productRepository) Save(product *models.LoanProduct) error {
	if product.ID > 0 {
		return r.db.Save(product).Error
	}
	return r.db.Create(product).Error
}

func (r *productRepository) Delete(id int64) error {
	return r.db.Model(&models.LoanProduct{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
