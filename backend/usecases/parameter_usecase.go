package usecases

import (
	"errors"
	"lms-backend/models"
	"lms-backend/repositories"
)

type ParameterUseCase interface {
	GetAllParameters() ([]models.GlobalParameter, error)
	GetParameterByKey(key string) (models.GlobalParameter, error)
	SaveParameter(req SaveParameterRequest) (*models.GlobalParameter, error)
	DeleteParameter(id int64) error
}

type parameterUseCase struct {
	repo repositories.ParameterRepository
}

func NewParameterUseCase(r repositories.ParameterRepository) ParameterUseCase {
	return &parameterUseCase{repo: r}
}

type SaveParameterRequest struct {
	ID          int64  `json:"id"`
	KeyName     string `json:"key_name"`
	KeyValue    string `json:"key_value"`
	Description string `json:"description"`
}

func (u *parameterUseCase) GetAllParameters() ([]models.GlobalParameter, error) {
	return u.repo.FindAll()
}

func (u *parameterUseCase) GetParameterByKey(key string) (models.GlobalParameter, error) {
	return u.repo.FindByKey(key)
}

func (u *parameterUseCase) SaveParameter(req SaveParameterRequest) (*models.GlobalParameter, error) {
	if req.KeyName == "" || req.KeyValue == "" {
		return nil, errors.New("key_name and key_value are required")
	}

	var param models.GlobalParameter
	// Update existing if ID is provided
	if req.ID > 0 {
		var err error
		param, err = u.repo.FindByKey(req.KeyName)
		if err != nil {
			return nil, errors.New("parameter not found for update")
		}
		param.KeyValue = req.KeyValue
		param.Description = req.Description
		err = u.repo.Update(&param)
		if err != nil {
			return nil, err
		}
	} else {
		// Create new
		param = models.GlobalParameter{
			KeyName:     req.KeyName,
			KeyValue:    req.KeyValue,
			Description: req.Description,
		}
		err := u.repo.Create(&param)
		if err != nil {
			return nil, err
		}
	}
	return &param, nil
}

func (u *parameterUseCase) DeleteParameter(id int64) error {
	return u.repo.Delete(id)
}
