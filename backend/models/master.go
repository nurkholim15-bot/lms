package models

// GlobalParameters represents lms_sch.global_parameters
type GlobalParameter struct {
	ID          int64  `gorm:"primaryKey;column:id;autoIncrement:true" json:"id"`
	KeyName     string `gorm:"column:key_name" json:"key_name"`
	KeyValue    string `gorm:"column:key_value" json:"key_value"`
	Description string `gorm:"column:description" json:"description"`
	MasterBaseModel
}

func (GlobalParameter) TableName() string {
	return "lms_sch.global_parameters"
}

// EmployeeCategory represents lms_sch.employee_categories
type EmployeeCategory struct {
	CategoryCode string  `gorm:"primaryKey;column:category_code" json:"category_code"`
	Description  string  `gorm:"column:description" json:"description"`
	MaxLimit     float64 `gorm:"column:max_limit" json:"max_limit"`
	IsEligible   bool    `gorm:"column:is_eligible" json:"is_eligible"`
	MasterBaseModel
}

func (EmployeeCategory) TableName() string {
	return "lms_sch.employee_categories"
}

// Employee represents lms_sch.employees
type Employee struct {
	EmployeeID     int64  `gorm:"primaryKey;column:employee_id;autoIncrement:false" json:"employee_id"`
	Name           string `gorm:"column:name" json:"name"`
	EmployeeStatus string  `gorm:"column:employee_status" json:"employee_status"` // renamed from work_status
	DeptNo         string `gorm:"column:deptno" json:"deptno"`
	CategoryCode   string  `gorm:"column:category_code" json:"category_code"`
	RoleID         int64  `gorm:"column:role_id" json:"role_id"`
	Salary         float64 `gorm:"column:salary" json:"salary"`
	MasterBaseModel
}

func (Employee) TableName() string {
	return "lms_sch.employees"
}

// Member represents lms_sch.members
type Member struct {
	MemberNo        int64  `gorm:"primaryKey;column:member_no;autoIncrement:false" json:"member_no"`
	EmployeeID      int64  `gorm:"column:employee_id;index:idx_member_employee" json:"employee_id"`
	KopkaraStatus   string `gorm:"column:kopkara_status" json:"kopkara_status"` // renamed from status
	JoinDate        string `gorm:"column:join_date" json:"join_date"`
	BankName        string `gorm:"column:bank_name" json:"bank_name"`
	BankAccountNo   string `gorm:"column:bank_account_no" json:"bank_account_no"`
	BankAccountName string `gorm:"column:bank_account_name" json:"bank_account_name"`
	MasterBaseModel
}

func (Member) TableName() string {
	return "lms_sch.members"
}

// EmployeeStatus represents lms_sch.employee_statuses
type EmployeeStatus struct {
	StatusCode  string `gorm:"primaryKey;column:status_code" json:"status_code"`
	Description string `gorm:"column:description" json:"description"`
	MasterBaseModel
}

func (EmployeeStatus) TableName() string {
	return "lms_sch.employee_statuses"
}

// Department represents lms_sch.departments
type Department struct {
	DeptNo   string `gorm:"primaryKey;column:deptno" json:"deptno"`
	DeptName string `gorm:"column:dept_name" json:"dept_name"`
	MasterBaseModel
}

func (Department) TableName() string {
	return "lms_sch.departments"
}

// KopkaraStatus represents lms_sch.kopkara_statuses
type KopkaraStatus struct {
	StatusCode  string `gorm:"primaryKey;column:status_code" json:"status_code"`
	Description string `gorm:"column:description" json:"description"`
	MasterBaseModel
}

func (KopkaraStatus) TableName() string {
	return "lms_sch.kopkara_statuses"
}

// LoanProduct represents lms_sch.loan_products
type LoanProduct struct {
	ID                    int64   `gorm:"primaryKey;column:id;autoIncrement:true" json:"id"`
	Name                  string  `gorm:"column:name" json:"name"`
	LoanType              string  `gorm:"column:loan_type" json:"loan_type"`
	MaxTenorMonths        int     `gorm:"column:max_tenor_months" json:"max_tenor_months"`
	SubmissionPeriodStart int     `gorm:"column:submission_period_start" json:"submission_period_start"`
	SubmissionPeriodEnd   int     `gorm:"column:submission_period_end" json:"submission_period_end"`
	MaxPercentageSalary   float64 `gorm:"column:max_percentage_salary" json:"max_percentage_salary"`
	InterestRate          float64 `gorm:"column:interest_rate" json:"interest_rate"`
	Status                string  `gorm:"column:status" json:"status"`
	MasterBaseModel
}

func (LoanProduct) TableName() string {
	return "lms_sch.loan_products"
}

// AdminFee represents lms_sch.admin_fees
type AdminFee struct {
	ID        int64   `gorm:"primaryKey;column:id"`
	FeeType   string  `gorm:"column:fee_type"`
	FeeValue  float64 `gorm:"column:fee_value"`
	MinFee    float64 `gorm:"column:min_fee"`
	ValidFrom string  `gorm:"column:valid_from"`
	ValidTo   string  `gorm:"column:valid_to"`
	MasterBaseModel
}

func (AdminFee) TableName() string {
	return "lms_sch.admin_fees"
}

// Role represents lms_sch.roles
type Role struct {
	RoleID      int64  `gorm:"primaryKey;column:role_id;autoIncrement:true" json:"role_id"`
	RoleName    string `gorm:"column:role_name" json:"role_name"`
	Description string `gorm:"column:description" json:"description"`
	MasterBaseModel
}

func (Role) TableName() string {
	return "lms_sch.roles"
}

// Menu represents lms_sch.menus
type Menu struct {
	MenuID   int64  `gorm:"primaryKey;column:menu_id;autoIncrement:true" json:"menu_id"`
	ParentID *int64 `gorm:"column:parent_id" json:"parent_id"` // Nullable for top level menus
	Title    string `gorm:"column:title" json:"title"`
	Icon     string `gorm:"column:icon" json:"icon"`
	Path     string `gorm:"column:path" json:"path"`
	Order    int    `gorm:"column:order_seq" json:"order"`
	MasterBaseModel
}

func (Menu) TableName() string {
	return "lms_sch.menus"
}

// RoleMenu represents lms_sch.role_menus
type RoleMenu struct {
	RoleID int64 `gorm:"primaryKey;column:role_id;autoIncrement:false" json:"role_id"`
	MenuID int64 `gorm:"primaryKey;column:menu_id;autoIncrement:false" json:"menu_id"`
	MasterBaseModel
}

func (RoleMenu) TableName() string {
	return "lms_sch.role_menus"
}
