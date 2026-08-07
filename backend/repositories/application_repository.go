package repositories

import (
	"fmt"
	"strings"
	"time"

	"lms-backend/models"

	"gorm.io/gorm"
)

type ApplicationRepository interface {
	FindAll() ([]models.LoanApplication, error)
	FindByPeriod(period string) ([]models.LoanApplication, error)
	FindByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error)
	FindByPeriodStatusAndMember(period string, status string, memberNo int64) ([]models.LoanApplication, error)
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

func (r *applicationRepository) FindByPeriod(period string) ([]models.LoanApplication, error) {
	return r.FindByPeriodStatusAndMember(period, "", 0)
}

func (r *applicationRepository) FindByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error) {
	return r.FindByPeriodStatusAndMember(period, status, 0)
}

func (r *applicationRepository) FindByPeriodStatusAndMember(period string, status string, memberNo int64) ([]models.LoanApplication, error) {
	var apps []models.LoanApplication
	cleanPeriod := strings.TrimSpace(strings.ReplaceAll(period, "-", ""))
	if cleanPeriod == "" {
		cleanPeriod = time.Now().Format("200601")
	}

	if len(cleanPeriod) == 6 {
		partitionTableName := fmt.Sprintf("loan_applications_%s", cleanPeriod)
		partitionTableFull := fmt.Sprintf("lms_sch.%s", partitionTableName)

		var tableExists bool
		r.db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'lms_sch' AND table_name = ?)", partitionTableName).Scan(&tableExists)

		if tableExists {
			query := r.db.Table(partitionTableFull)
			cleanStatus := strings.TrimSpace(status)
			if cleanStatus != "" && cleanStatus != "ALL" {
				statuses := strings.Split(cleanStatus, ",")
				for i := range statuses {
					statuses[i] = strings.TrimSpace(statuses[i])
				}
				if len(statuses) == 1 {
					query = query.Where("status = ?", statuses[0])
				} else {
					query = query.Where("status IN ?", statuses)
				}
			}

			if memberNo > 0 {
				query = query.Where("member_no = ? OR member_no IN (SELECT member_no FROM lms_sch.members WHERE employee_id = ?)", memberNo, memberNo)
			}

			err := query.Order("member_no ASC, created_at DESC").Find(&apps).Error
			return apps, err
		}

		// Partition table does not exist: return empty array immediately (NO parent table query)
		return []models.LoanApplication{}, nil
	}

	query := r.db.Model(&models.LoanApplication{})
	if strings.TrimSpace(status) != "" && strings.TrimSpace(status) != "ALL" {
		statuses := strings.Split(status, ",")
		for i := range statuses {
			statuses[i] = strings.TrimSpace(statuses[i])
		}
		if len(statuses) == 1 {
			query = query.Where("status = ?", statuses[0])
		} else {
			query = query.Where("status IN ?", statuses)
		}
	}

	if memberNo > 0 {
		query = query.Where("member_no = ? OR member_no IN (SELECT member_no FROM lms_sch.members WHERE employee_id = ?)", memberNo, memberNo)
	}

	err := query.Order("member_no ASC, created_at DESC").Find(&apps).Error
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
