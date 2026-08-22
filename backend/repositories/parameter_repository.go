package repositories

import (
	"strings"
	"sync"

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
	db       *gorm.DB
	mu       sync.RWMutex
	cache    []models.GlobalParameter
	cacheMap map[string]models.GlobalParameter
	isLoaded bool
}

func NewParameterRepository(db *gorm.DB) ParameterRepository {
	repo := &parameterRepository{
		db:       db,
		cacheMap: make(map[string]models.GlobalParameter),
	}
	repo.refreshCache()
	return repo
}

func (r *parameterRepository) refreshCache() {
	var params []models.GlobalParameter
	if err := r.db.Where("deleted_at IS NULL").Order("id ASC").Find(&params).Error; err == nil {
		r.mu.Lock()
		r.cache = params
		r.cacheMap = make(map[string]models.GlobalParameter)
		for _, p := range params {
			r.cacheMap[strings.ToUpper(p.KeyName)] = p
		}
		r.isLoaded = true
		r.mu.Unlock()
	}
}

func (r *parameterRepository) FindAll() ([]models.GlobalParameter, error) {
	r.refreshCache()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache, nil
}

func (r *parameterRepository) FindByKey(keyName string) (models.GlobalParameter, error) {
	r.mu.RLock()
	if r.isLoaded {
		param, exists := r.cacheMap[strings.ToUpper(keyName)]
		r.mu.RUnlock()
		if exists {
			return param, nil
		}
		// Jika cache sudah dimuat (isLoaded = true), key yang tidak ada di cacheMap memang tidak ada di DB.
		// Kembalikan ErrRecordNotFound langsung dari RAM tanpa query DB redundan!
		return models.GlobalParameter{}, gorm.ErrRecordNotFound
	}
	r.mu.RUnlock()

	r.refreshCache()
	r.mu.RLock()
	defer r.mu.RUnlock()
	param, exists := r.cacheMap[strings.ToUpper(keyName)]
	if exists {
		return param, nil
	}
	return models.GlobalParameter{}, gorm.ErrRecordNotFound
}

func (r *parameterRepository) Create(param *models.GlobalParameter) error {
	err := r.db.Create(param).Error
	if err == nil {
		r.refreshCache()
	}
	return err
}

func (r *parameterRepository) Update(param *models.GlobalParameter) error {
	err := r.db.Save(param).Error
	if err == nil {
		r.refreshCache()
	}
	return err
}

func (r *parameterRepository) Delete(id int64) error {
	err := r.db.Model(&models.GlobalParameter{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
	if err == nil {
		r.refreshCache()
	}
	return err
}
