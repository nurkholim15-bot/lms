package models

import (
	"time"
)

// LoanApplication represents lms_sch.loan_applications
type LoanApplication struct {
	ApplicationNo     int64      `gorm:"primaryKey;column:application_no" json:"application_no,string"`
	MemberNo          int64      `gorm:"column:member_no" json:"member_no,string"`
	ProductID         int64      `gorm:"column:product_id" json:"product_id"`
	SubmissionDate    time.Time  `gorm:"column:submission_date" json:"submission_date"`
	RequestedAmount   float64    `gorm:"column:requested_amount" json:"requested_amount"`
	ApprovedAmount    *float64   `gorm:"column:approved_amount" json:"approved_amount"`
	ApprovedAt        *time.Time `gorm:"column:approved_at" json:"approved_at"`
	Tenor             int        `gorm:"column:tenor" json:"tenor"`
	EligibilityResult string     `gorm:"column:eligibility_result" json:"eligibility_result"`
	Status            string     `gorm:"column:status" json:"status"`
	ApprovalNotes     string     `gorm:"column:approval_notes" json:"approval_notes"`
	PrincipalPerMonth float64    `gorm:"column:principal_per_month" json:"principal_per_month"`
	InterestPerMonth  float64    `gorm:"column:interest_per_month" json:"interest_per_month"`
	AdminFee          float64    `gorm:"column:admin_fee" json:"admin_fee"`
	TotalInstallment  float64    `gorm:"column:total_installment" json:"total_installment"`
	TotalLoanCost     float64    `gorm:"column:total_loan_cost" json:"total_loan_cost"`
	InterestRate      float64    `gorm:"column:interest_rate" json:"interest_rate"`
	CreditLimit       float64    `gorm:"column:credit_limit" json:"credit_limit"`
	BaseModel
}

func (LoanApplication) TableName() string {
	return "lms_sch.loan_applications"
}

// LoanTracking represents lms_sch.loan_trackings
type LoanTracking struct {
	ID            int64     `gorm:"primaryKey;column:id"`
	ApplicationNo int64     `gorm:"column:application_no;index"`
	Status        string    `gorm:"column:status"`
	UserID        string    `gorm:"column:user_id;index"`
	UserName      string    `gorm:"column:user_name"`
	ActionDate    time.Time `gorm:"column:action_date"`
	Notes         string    `gorm:"column:notes"`
	SLADuration   string    `gorm:"column:sla_duration"`
	IPAddress     string    `gorm:"column:ip_address"`
	UserAgent     string    `gorm:"column:user_agent"`
	BaseModel
}

func (LoanTracking) TableName() string {
	return "lms_sch.loan_trackings"
}

// LoanContract represents lms_sch.loan_contracts
type LoanContract struct {
	ContractNo        int64     `gorm:"primaryKey;column:contract_no"`
	ApplicationNo     int64     `gorm:"column:application_no;uniqueIndex"`
	MemberNo          int64     `gorm:"column:member_no;index"`
	ProductID         int64     `gorm:"column:product_id"`
	ApprovedAmount    float64   `gorm:"column:approved_amount"`
	Tenor             int       `gorm:"column:tenor"`
	PrincipalPerMonth float64   `gorm:"column:principal_per_month"`
	InterestPerMonth  float64   `gorm:"column:interest_per_month"`
	AdminFee          float64   `gorm:"column:admin_fee"`
	TotalInstallment  float64   `gorm:"column:total_installment"`
	TotalLoanCost     float64   `gorm:"column:total_loan_cost"`
	ContractDate      time.Time `gorm:"column:contract_date"`
	Status            string    `gorm:"column:status"`
	BaseModel
}

func (LoanContract) TableName() string {
	return "lms_sch.loan_contracts"
}

// Loan represents lms_sch.loans
type Loan struct {
	LoanNo             int64   `gorm:"primaryKey;column:loan_no"`
	ApplicationNo      int64   `gorm:"column:application_no"`
	MemberNo           int64   `gorm:"column:member_no"`
	PrincipalAmount    float64 `gorm:"column:principal_amount"`
	AdminFee           float64 `gorm:"column:admin_fee"`
	DisbursementAmount float64 `gorm:"column:disbursement_amount"`
	OutstandingAmount  float64 `gorm:"column:outstanding_amount"`
	Status             string  `gorm:"column:status"`
	BaseModel
}

func (Loan) TableName() string {
	return "lms_sch.loans"
}

// LoanSchedule represents lms_sch.loan_schedules
type LoanSchedule struct {
	ID                   int64     `gorm:"primaryKey;column:id" json:"id"`
	LoanNo               int64     `gorm:"column:loan_no" json:"loan_no"`
	Period               string    `gorm:"column:period" json:"period"`
	InstallmentNo        int       `gorm:"column:installment_no" json:"installment_no"`
	Principal            float64   `gorm:"column:principal" json:"principal"`
	Interest             float64   `gorm:"column:interest" json:"interest"`
	TotalInstallment     float64   `gorm:"column:total_installment" json:"total_installment"`
	AmountPaid           float64   `gorm:"column:amount_paid" json:"amount_paid"`
	RemainingInstallment float64   `gorm:"column:remaining_installment" json:"remaining_installment"`
	Status               string    `gorm:"column:status" json:"status"`
	DueDate              time.Time `gorm:"column:due_date" json:"due_date"`
	BaseModel
}

func (LoanSchedule) TableName() string {
	return "lms_sch.loan_schedules"
}

// PayrollDeduction represents lms_sch.payroll_deductions
type PayrollDeduction struct {
	ID              int64      `gorm:"primaryKey;column:id" json:"id"`
	LoanNo          *int64     `gorm:"column:loan_no" json:"loan_no"`
	Period          string     `gorm:"column:period" json:"period"`
	NominalOriginal float64    `gorm:"column:nominal_original" json:"nominal_original"`
	NominalDeducted float64    `gorm:"column:nominal_deducted" json:"nominal_deducted"`
	ProcessDate     *time.Time `gorm:"column:process_date" json:"process_date"`
	Status          string     `gorm:"column:status" json:"status"`
	FailureReason   string     `gorm:"column:failure_reason" json:"failure_reason"`
	BaseModel
}

func (PayrollDeduction) TableName() string {
	return "lms_sch.payroll_deductions"
}

// LoanPayrollImportLog represents lms_sch.loan_payroll_import_logs
type LoanPayrollImportLog struct {
	ID                  int64   `gorm:"primaryKey;column:id;autoIncrement:true" json:"id"`
	FileName            string  `gorm:"column:file_name" json:"file_name"`
	TotalRecords        int     `gorm:"column:total_records" json:"total_records"`
	TotalAmount         float64 `gorm:"column:total_amount" json:"total_amount"`
	TotalSuccessRecords int     `gorm:"column:total_success_records" json:"total_success_records"`
	TotalSuccessAmount  float64 `gorm:"column:total_success_amount" json:"total_success_amount"`
	TotalFailedRecords  int     `gorm:"column:total_failed_records" json:"total_failed_records"`
	TotalFailedAmount   float64 `gorm:"column:total_failed_amount" json:"total_failed_amount"`
	BaseModel
}

func (LoanPayrollImportLog) TableName() string {
	return "lms_sch.loan_payroll_import_logs"
}
