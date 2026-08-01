package usecases

import (
	"lms-backend/models"
	"lms-backend/repositories"
)

type ProductUseCase interface {
	GetAllProducts() ([]models.LoanProduct, error)
	GetProductByID(id int64) (models.LoanProduct, error)
	CreateProduct(product *models.LoanProduct) error
	SaveProduct(product *models.LoanProduct) error
	DeleteProduct(id int64) error
}

type productUseCase struct {
	repo repositories.ProductRepository
}

func NewProductUseCase(repo repositories.ProductRepository) ProductUseCase {
	return &productUseCase{repo}
}

func (u *productUseCase) GetAllProducts() ([]models.LoanProduct, error) {
	return u.repo.FindAll()
}

func (u *productUseCase) GetProductByID(id int64) (models.LoanProduct, error) {
	return u.repo.FindByID(id)
}

func (u *productUseCase) CreateProduct(product *models.LoanProduct) error {
	return u.repo.Create(product)
}

func (u *productUseCase) SaveProduct(product *models.LoanProduct) error {
	return u.repo.Save(product)
}

func (u *productUseCase) DeleteProduct(id int64) error {
	return u.repo.Delete(id)
}
