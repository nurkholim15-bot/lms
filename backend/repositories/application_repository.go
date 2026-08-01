package repositories

import (
	"lms-backend/models"

	"gorm.io/gorm"
)

type ApplicationRepository interface {
	FindAll() ([]models.LoanApplication, error)
	FindByID(applicationNo int64) (models.LoanApplication, error)
	Create(app *models.LoanApplication) error
	Update(app *models.LoanApplication) error
	GetDB() *gorm.DB
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db}
}

func (r *applicationRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *applicationRepository) FindAll() ([]models.LoanApplication, error) {
	var apps []models.LoanApplication
	err := r.db.Find(&apps).Error
	return apps, err
}

func (r *applicationRepository) FindByID(applicationNo int64) (models.LoanApplication, error) {
	var app models.LoanApplication
	err := r.db.Where("application_no = ?", applicationNo).First(&app).Error
	return app, err
}

func (r *applicationRepository) Create(app *models.LoanApplication) error {
	return r.db.Create(app).Error
}

func (r *applicationRepository) Update(app *models.LoanApplication) error {
	return r.db.Save(app).Error
}
