# Python script to implement complete audit user population across all models, usecases, handlers, and main.go

# 1. Update backend/models/base.go
base_content = """package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Base Model for Audit Fields
type BaseModel struct {
	CreatedAt   *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	CreatedUser *string    `gorm:"column:created_user" json:"created_user"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedUser *string    `gorm:"column:updated_user" json:"updated_user"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if b.CreatedAt == nil {
		b.CreatedAt = &now
	}
	if b.UpdatedAt == nil {
		b.UpdatedAt = &now
	}
	if b.CreatedUser == nil || strings.TrimSpace(*b.CreatedUser) == "" {
		defUser := "SYSTEM_AUTO"
		b.CreatedUser = &defUser
	}
	if b.UpdatedUser == nil || strings.TrimSpace(*b.UpdatedUser) == "" {
		b.UpdatedUser = b.CreatedUser
	}
	return nil
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	b.UpdatedAt = &now
	if b.UpdatedUser == nil || strings.TrimSpace(*b.UpdatedUser) == "" {
		if b.CreatedUser != nil && strings.TrimSpace(*b.CreatedUser) != "" {
			b.UpdatedUser = b.CreatedUser
		} else {
			defUser := "SYSTEM_AUTO"
			b.UpdatedUser = &defUser
		}
	}
	return nil
}

// BaseModel with Soft Delete for Master Tables & Audit Protection
type MasterBaseModel struct {
	BaseModel
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedUser *string        `gorm:"column:deleted_user" json:"deleted_user"`
}

func (m *MasterBaseModel) BeforeDelete(tx *gorm.DB) error {
	if m.DeletedUser == nil || strings.TrimSpace(*m.DeletedUser) == "" {
		if m.UpdatedUser != nil && strings.TrimSpace(*m.UpdatedUser) != "" {
			m.DeletedUser = m.UpdatedUser
		} else if m.CreatedUser != nil && strings.TrimSpace(*m.CreatedUser) != "" {
			m.DeletedUser = m.CreatedUser
		} else {
			defUser := "SYSTEM_AUTO"
			m.DeletedUser = &defUser
		}
	}
	return nil
}
"""

with open('backend/models/base.go', 'w', encoding='utf-8') as f:
    f.write(base_content)

print("Successfully updated backend/models/base.go with GORM hooks!")

# 2. Update backend/usecases/application_usecase.go to set CreatedUser and UpdatedUser explicitly
with open('backend/usecases/application_usecase.go', 'r', encoding='utf-8') as f:
    uc_code = f.read()

# Update SubmitApplication app initialization
old_submit_app = """	var createdUserPtr *string
	if strings.TrimSpace(req.CreatedUser) != "" {
		cu := strings.TrimSpace(req.CreatedUser)
		createdUserPtr = &cu
	}

	app := &models.LoanApplication{
		ApplicationNo:     appNo,
		MemberNo:          member.MemberNo,
		ProductID:         req.ProductID,
		SubmissionDate:    now,
		RequestedAmount:   req.RequestedAmount,
		Tenor:             req.Tenor,
		EligibilityResult: simResult.EligibilityResult,
		Status:            "SUBMITTED",
		PrincipalPerMonth: simResult.PrincipalPerMonth,
		InterestPerMonth:  simResult.InterestPerMonth,
		AdminFee:          simResult.AdminFee,
		TotalInstallment:  simResult.TotalInstallment,
		TotalLoanCost:     simResult.TotalLoanCost,
		InterestRate:      simResult.InterestRate,
		CreditLimit:       simResult.CreditLimit,
		CreatedUser:       createdUserPtr,
		UpdatedUser:       createdUserPtr,
	}"""

new_submit_app = """	userStr := strings.TrimSpace(req.CreatedUser)
	if userStr == "" {
		userStr = fmt.Sprintf("%d", member.MemberNo)
	}

	app := &models.LoanApplication{
		ApplicationNo:     appNo,
		MemberNo:          member.MemberNo,
		ProductID:         req.ProductID,
		SubmissionDate:    now,
		RequestedAmount:   req.RequestedAmount,
		Tenor:             req.Tenor,
		EligibilityResult: simResult.EligibilityResult,
		Status:            "SUBMITTED",
		PrincipalPerMonth: simResult.PrincipalPerMonth,
		InterestPerMonth:  simResult.InterestPerMonth,
		AdminFee:          simResult.AdminFee,
		TotalInstallment:  simResult.TotalInstallment,
		TotalLoanCost:     simResult.TotalLoanCost,
		InterestRate:      simResult.InterestRate,
		CreditLimit:       simResult.CreditLimit,
		BaseModel: models.BaseModel{
			CreatedAt:   &now,
			CreatedUser: &userStr,
			UpdatedAt:   &now,
			UpdatedUser: &userStr,
		},
	}"""

uc_code = uc_code.replace(old_submit_app, new_submit_app)

# Update Submit LoanTracking initialization
old_submit_tracking = """	tracking := &models.LoanTracking{
		ApplicationNo: appNo,
		Status:        "SUBMITTED",
		UserID:        subUserId,
		UserName:      subUserName,
		ActionDate:    now,
		Notes:         "Pengajuan pinjaman baru disubmit",
		SLADuration:   "-",
		IPAddress:     "127.0.0.1",
		UserAgent:     "Browser/Kopkara LMS",
	}"""

new_submit_tracking = """	tracking := &models.LoanTracking{
		ApplicationNo: appNo,
		Status:        "SUBMITTED",
		UserID:        subUserId,
		UserName:      subUserName,
		ActionDate:    now,
		Notes:         "Pengajuan pinjaman baru disubmit",
		SLADuration:   "-",
		IPAddress:     "127.0.0.1",
		UserAgent:     "Browser/Kopkara LMS",
		BaseModel: models.BaseModel{
			CreatedAt:   &now,
			CreatedUser: &userStr,
			UpdatedAt:   &now,
			UpdatedUser: &userStr,
		},
	}"""

uc_code = uc_code.replace(old_submit_tracking, new_submit_tracking)

# Update ProcessApproval Contract initialization
old_proc_contract = """		// Create Loan Contract for APPROVED status
		contract := models.LoanContract{
			ApplicationNo:     app.ApplicationNo,
			MemberNo:          app.MemberNo,
			ProductID:         app.ProductID,
			ApprovedAmount:    *app.ApprovedAmount,
			Tenor:             app.Tenor,
			PrincipalPerMonth: app.PrincipalPerMonth,
			InterestPerMonth:  app.InterestPerMonth,
			AdminFee:          app.AdminFee,
			TotalInstallment:  app.TotalInstallment,
			TotalLoanCost:     app.TotalLoanCost,
			ContractDate:      time.Now(),
			Status:            "ACTIVE",
		}"""

new_proc_contract = """		upUserStr := strings.TrimSpace(updatedUser)
		if upUserStr == "" {
			upUserStr = "SYSTEM_AUTO"
		}
		// Create Loan Contract for APPROVED status
		contract := models.LoanContract{
			ApplicationNo:     app.ApplicationNo,
			MemberNo:          app.MemberNo,
			ProductID:         app.ProductID,
			ApprovedAmount:    *app.ApprovedAmount,
			Tenor:             app.Tenor,
			PrincipalPerMonth: app.PrincipalPerMonth,
			InterestPerMonth:  app.InterestPerMonth,
			AdminFee:          app.AdminFee,
			TotalInstallment:  app.TotalInstallment,
			TotalLoanCost:     app.TotalLoanCost,
			ContractDate:      time.Now(),
			Status:            "ACTIVE",
			BaseModel: models.BaseModel{
				CreatedAt:   &now,
				CreatedUser: &upUserStr,
				UpdatedAt:   &now,
				UpdatedUser: &upUserStr,
			},
		}"""

uc_code = uc_code.replace(old_proc_contract, new_proc_contract)

# Update ProcessApproval Tracking initialization
old_proc_tracking = """	appUserId, appUserName := u.resolveUserInfo(updatedUser)
	tracking := &models.LoanTracking{
		ApplicationNo: app.ApplicationNo,
		Status:        app.Status,
		UserID:        appUserId,
		UserName:      appUserName,
		ActionDate:    now,
		Notes:         notes,
		SLADuration:   slaStr,
		IPAddress:     "127.0.0.1",
		UserAgent:     "Browser/Kopkara LMS",
	}"""

new_proc_tracking = """	upUserStr := strings.TrimSpace(updatedUser)
	if upUserStr == "" {
		upUserStr = "SYSTEM_AUTO"
	}
	app.UpdatedAt = &now
	app.UpdatedUser = &upUserStr

	appUserId, appUserName := u.resolveUserInfo(updatedUser)
	tracking := &models.LoanTracking{
		ApplicationNo: app.ApplicationNo,
		Status:        app.Status,
		UserID:        appUserId,
		UserName:      appUserName,
		ActionDate:    now,
		Notes:         notes,
		SLADuration:   slaStr,
		IPAddress:     "127.0.0.1",
		UserAgent:     "Browser/Kopkara LMS",
		BaseModel: models.BaseModel{
			CreatedAt:   &now,
			CreatedUser: &upUserStr,
			UpdatedAt:   &now,
			UpdatedUser: &upUserStr,
		},
	}"""

uc_code = uc_code.replace(old_proc_tracking, new_proc_tracking)

# Update DisburseApplication Loan initialization
old_disb_loan = """	loan := models.Loan{
		LoanNo:            loanNo,
		ApplicationNo:     app.ApplicationNo,
		MemberNo:          app.MemberNo,
		PrincipalAmount:   principalAmt,
		InterestAmount:    app.InterestPerMonth * float64(app.Tenor),
		AdminFee:          adminFee,
		TotalLoanCost:     app.TotalLoanCost,
		OutstandingAmount: app.TotalLoanCost,
		Tenor:             app.Tenor,
		DisbursementDate:  now,
		Status:            "ACTIVE",
	}"""

new_disb_loan = """	disbUserStr := strings.TrimSpace(updatedUser)
	if disbUserStr == "" {
		disbUserStr = "SYSTEM_AUTO"
	}
	loan := models.Loan{
		LoanNo:            loanNo,
		ApplicationNo:     app.ApplicationNo,
		MemberNo:          app.MemberNo,
		PrincipalAmount:   principalAmt,
		InterestAmount:    app.InterestPerMonth * float64(app.Tenor),
		AdminFee:          adminFee,
		TotalLoanCost:     app.TotalLoanCost,
		OutstandingAmount: app.TotalLoanCost,
		Tenor:             app.Tenor,
		DisbursementDate:  now,
		Status:            "ACTIVE",
		BaseModel: models.BaseModel{
			CreatedAt:   &now,
			CreatedUser: &disbUserStr,
			UpdatedAt:   &now,
			UpdatedUser: &disbUserStr,
		},
	}"""

uc_code = uc_code.replace(old_disb_loan, new_disb_loan)

# Update DisburseApplication LoanSchedule initialization
old_disb_sched = """		sched := models.LoanSchedule{
			LoanNo:               loanNo,
			InstallmentNo:        i,
			DueDate:              dueDate,
			PrincipalAmount:      app.PrincipalPerMonth,
			InterestAmount:       app.InterestPerMonth,
			TotalInstallment:     app.TotalInstallment,
			RemainingInstallment: app.TotalInstallment,
			Status:               "UNPAID",
		}"""

new_disb_sched = """		sched := models.LoanSchedule{
			LoanNo:               loanNo,
			InstallmentNo:        i,
			DueDate:              dueDate,
			PrincipalAmount:      app.PrincipalPerMonth,
			InterestAmount:       app.InterestPerMonth,
			TotalInstallment:     app.TotalInstallment,
			RemainingInstallment: app.TotalInstallment,
			Status:               "UNPAID",
			BaseModel: models.BaseModel{
				CreatedAt:   &now,
				CreatedUser: &disbUserStr,
				UpdatedAt:   &now,
				UpdatedUser: &disbUserStr,
			},
		}"""

uc_code = uc_code.replace(old_disb_sched, new_disb_sched)

# Update DisburseApplication app.UpdatedUser
old_disb_app_update = """	// 4. Update Application Status to DISBURSED
	app.Status = "DISBURSED"
	if err := u.appRepo.Update(&app); err != nil {"""

new_disb_app_update = """	// 4. Update Application Status to DISBURSED
	app.Status = "DISBURSED"
	app.UpdatedAt = &now
	app.UpdatedUser = &disbUserStr
	if err := u.appRepo.Update(&app); err != nil {"""

uc_code = uc_code.replace(old_disb_app_update, new_disb_app_update)

with open('backend/usecases/application_usecase.go', 'w', encoding='utf-8') as f:
    f.write(uc_code)

print("Successfully updated backend/usecases/application_usecase.go with explicit CreatedUser and UpdatedUser!")
