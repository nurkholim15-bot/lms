package repositories

import (
	"lms-backend/models"

	"gorm.io/gorm"
)

type ParameterRepository interface {
	FindAll() ([]models.GlobalParameter, error)
	FindByKey(keyName string) (models.GlobalParameter, error)
	Create(param *models.GlobalParameter) error
	Update(param *models.GlobalParameter) error
	Delete(id int64) error
}

type parameterRepository struct {
	db *gorm.DB
}

func NewParameterRepository(db *gorm.DB) ParameterRepository {
	return &parameterRepository{db}
}

func (r *parameterRepository) FindAll() ([]models.GlobalParameter, error) {
	var params []models.GlobalParameter
	err := r.db.Where("deleted_at IS NULL").Find(&params).Error
	return params, err
}

func (r *parameterRepository) FindByKey(keyName string) (models.GlobalParameter, error) {
	var param models.GlobalParameter
	err := r.db.Where("deleted_at IS NULL AND key_name = ?", keyName).First(&param).Error
	return param, err
}

func (r *parameterRepository) Create(param *models.GlobalParameter) error {
	return r.db.Create(param).Error
}

func (r *parameterRepository) Update(param *models.GlobalParameter) error {
	// Updates all fields for simplicity
	return r.db.Save(param).Error
}

func (r *parameterRepository) Delete(id int64) error {
	// Soft delete logic for GORM, or simple update if manual
	return r.db.Model(&models.GlobalParameter{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
