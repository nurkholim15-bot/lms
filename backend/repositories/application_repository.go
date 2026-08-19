package repositories

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"lms-backend/models"

	"gorm.io/gorm"
)

type ApplicationRepository interface {
	FindAll() ([]models.LoanApplication, error)
	FindByPeriod(period string) ([]models.LoanApplication, error)
	FindByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error)
	FindByPeriodStatusAndMember(period string, status string, memberNo int64, limit int, offset int) ([]models.LoanApplication, error)
	FindByID(applicationNo int64) (models.LoanApplication, error)
	Create(app *models.LoanApplication) error
	Update(app *models.LoanApplication) error
	GetDB() *gorm.DB
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
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
	return r.FindByPeriodStatusAndMember(period, "", 0, 0, 0)
}

func (r *applicationRepository) FindByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error) {
	return r.FindByPeriodStatusAndMember(period, status, 0, 0, 0)
}

func (r *applicationRepository) FindByPeriodStatusAndMember(period string, status string, memberNo int64, limit int, offset int) ([]models.LoanApplication, error) {
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

			if limit > 0 {
				query = query.Limit(limit).Offset(offset)
			}

			err := query.Order("created_at DESC, application_no DESC").Find(&apps).Error
			return apps, err
		}
	}

	// Dynamic fallback for main table lms_sch.loan_applications with submission_date period filter
	query := r.db.Model(&models.LoanApplication{})
	if cleanPeriod != "" && len(cleanPeriod) == 6 {
		year, _ := strconv.Atoi(cleanPeriod[:4])
		month, _ := strconv.Atoi(cleanPeriod[4:])
		if year > 2000 && month >= 1 && month <= 12 {
			startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
			endDate := startDate.AddDate(0, 1, 0)
			query = query.Where("submission_date >= ? AND submission_date < ?", startDate, endDate)
		}
	}

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

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Order("submission_date DESC, application_no DESC").Find(&apps).Error
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
