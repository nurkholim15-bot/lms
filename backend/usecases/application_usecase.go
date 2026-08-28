package usecases

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"gorm.io/gorm"

	"lms-backend/cache"
	"lms-backend/models"
	"lms-backend/repositories"
)

type ApplicationUseCase interface {
	GetAllApplications() ([]models.LoanApplication, error)
	GetApplicationsByPeriod(period string) ([]models.LoanApplication, error)
	GetApplicationsByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error)
	GetApplicationsFiltered(period string, status string, memberNo int64, roleId string, userEmployeeId int64, limit int, offset int) ([]models.LoanApplication, error)
	IsHighPrivilegeRole(roleId string) bool
	SimulateApplication(req SubmitApplicationRequest) (SimulationResult, error)
	SubmitApplication(req SubmitApplicationRequest) (*models.LoanApplication, error)
	ApproveApplication(applicationNo int64, approvedAmount float64, notes string, updatedUser ...string) error
	ProcessApproval(applicationNo int64, action string, notes string, updatedUser string) error
	GetLoanTrackings(applicationNo int64) ([]models.LoanTracking, error)
	DisburseApplication(applicationNo int64, req DisburseRequest, updatedUser string) error
}

type applicationUseCase struct {
	appRepo      repositories.ApplicationRepository
	productRepo  repositories.ProductRepository
	paramRepo    repositories.ParameterRepository
	paramCache   sync.Map
	productCache sync.Map
}

func FormatIDR(val float64) string {
	in := int64(val + 0.5)
	if in == 0 {
		return "Rp 0"
	}
	sign := ""
	if in < 0 {
		sign = "-"
		in = -in
	}
	str := strconv.FormatInt(in, 10)
	n := len(str)
	if n <= 3 {
		return fmt.Sprintf("Rp %s%s", sign, str)
	}
	var out []byte
	rem := n % 3
	if rem > 0 {
		out = append(out, str[:rem]...)
		out = append(out, '.')
	}
	for i := rem; i < n; i += 3 {
		out = append(out, str[i:i+3]...)
		if i+3 < n {
			out = append(out, '.')
		}
	}
	return "Rp " + sign + string(out)
}

func NewApplicationUseCase(ar repositories.ApplicationRepository, pr repositories.ProductRepository, pRepo repositories.ParameterRepository) ApplicationUseCase {
	return &applicationUseCase{appRepo: ar, productRepo: pr, paramRepo: pRepo}
}

type SubmitApplicationRequest struct {
	MemberNo        int64   `json:"member_no"`
	ProductID       int64   `json:"product_id"`
	RequestedAmount float64 `json:"requested_amount"`
	Tenor           int     `json:"tenor"`
	CreatedUser     string  `json:"created_user"`
}

type SimulationResult struct {
	PrincipalPerMonth  float64 `json:"principal_per_month"`
	InterestPerMonth   float64 `json:"interest_per_month"`
	AdminFee           float64 `json:"admin_fee"`
	TotalInstallment   float64 `json:"total_installment"`
	TotalLoanCost      float64 `json:"total_loan_cost"`
	Tenor              int     `json:"tenor"`
	InterestRate       float64 `json:"interest_rate"`
	CreditLimit        float64 `json:"credit_limit"`
	TotalPreviousLoans float64 `json:"total_previous_loans"`
	RemainingLimit     float64 `json:"remaining_limit"`
}

// Helper to resolve employee_id and name dynamically from lms_sch.employees
// Helper to resolve real users.username from DB if createdUser/updatedUser is not provided
func (u *applicationUseCase) resolveUsername(providedUser string, memberNo int64, empID int64) string {
	providedUser = strings.TrimSpace(providedUser)
	if strings.EqualFold(providedUser, "SYSTEM_AUTO") || strings.EqualFold(providedUser, "SYSTEM") {
		return "SYSTEM_AUTO"
	}
	if providedUser != "" && providedUser != "10101" {
		return providedUser
	}

	db := u.appRepo.GetDB()
	if db != nil {
		var usr models.User
		if memberNo > 0 {
			if err := db.Where("member_no = ? AND deleted_at IS NULL", memberNo).First(&usr).Error; err == nil && strings.TrimSpace(usr.Username) != "" {
				return strings.TrimSpace(usr.Username)
			}
		}
		if empID > 0 {
			if err := db.Where("(id = ? OR member_no = ?) AND deleted_at IS NULL", empID, empID).First(&usr).Error; err == nil && strings.TrimSpace(usr.Username) != "" {
				return strings.TrimSpace(usr.Username)
			}
		}
	}

	if memberNo > 0 {
		return fmt.Sprintf("%d", memberNo)
	}
	if empID > 0 {
		return fmt.Sprintf("%d", empID)
	}
	return "SYSTEM_AUTO"
}

func (u *applicationUseCase) resolveUserInfo(userKey string) (string, string) {
	userKey = strings.TrimSpace(userKey)
	if userKey == "" {
		userKey = "10101"
	}

	if strings.EqualFold(userKey, "SYSTEM_AUTO") || strings.EqualFold(userKey, "SYSTEM") {
		return "SYSTEM_AUTO", "System Automation"
	}

	var emp models.Employee
	empIdInt, errConv := strconv.ParseInt(userKey, 10, 64)
	if errConv == nil && empIdInt > 0 {
		if err := u.appRepo.GetDB().Where("employee_id = ?", empIdInt).First(&emp).Error; err == nil && emp.Name != "" {
			return strconv.FormatInt(emp.EmployeeID, 10), emp.Name
		}
	}

	if err := u.appRepo.GetDB().Where("name = ?", userKey).First(&emp).Error; err == nil && emp.Name != "" {
		return strconv.FormatInt(emp.EmployeeID, 10), emp.Name
	}

	return userKey, "User " + userKey
}

// Helper to get parameter value safely with RAM caching
func (u *applicationUseCase) getParamVal(key string, defaultVal string) string {
	return cache.ParameterCache.Get(key, defaultVal, u.appRepo.GetDB())
}

// Helper to get product safely with RAM caching
func (u *applicationUseCase) getProductCached(productID int64) (models.LoanProduct, error) {
	if val, ok := u.productCache.Load(productID); ok {
		return val.(models.LoanProduct), nil
	}
	product, err := u.productRepo.FindByID(productID)
	if err != nil {
		return models.LoanProduct{}, err
	}
	u.productCache.Store(productID, product)
	return product, nil
}

func (u *applicationUseCase) runSimulation(req SubmitApplicationRequest, product models.LoanProduct, existingMember ...*models.Member) (SimulationResult, error) {
	db := u.appRepo.GetDB()

	var member models.Member
	if len(existingMember) > 0 && existingMember[0] != nil && existingMember[0].MemberNo > 0 {
		member = *existingMember[0]
	} else {
		if mPtr, errM := cache.IdentityCache.GetMemberByNo(db, req.MemberNo); errM == nil && mPtr != nil {
			member = *mPtr
		} else {
			member = models.Member{MemberNo: req.MemberNo, EmployeeID: req.MemberNo}
		}
	}

	var employee models.Employee
	if ePtr, errE := cache.IdentityCache.GetEmployeeByID(db, member.EmployeeID); errE == nil && ePtr != nil {
		employee = *ePtr
	} else {
		employee = models.Employee{EmployeeID: member.EmployeeID, Salary: 8000000, CategoryCode: "PERMANENT"}
	}

	var category models.EmployeeCategory
	if catPtr, errCat := cache.IdentityCache.GetCategoryByCode(db, employee.CategoryCode); errCat == nil && catPtr != nil {
		category = *catPtr
	} else {
		category = models.EmployeeCategory{CategoryCode: employee.CategoryCode, MaxLimit: 100000000}
	}

	// 1. Fetch Global Parameters
	minAdminFeeStr := u.getParamVal("LOAN_MIN_ADMIN_FEE", "0")
	adminFormulaStr := u.getParamVal("LOAN_ADMIN_FORMULA", "0")
	limitFormulaStr := u.getParamVal("LOAN_LIMIT_FORMULA", "REQUESTED_AMOUNT") // Default: no limit

	// Evaluate Limits (Real Salary)
	salary := employee.Salary
	currentDay := float64(time.Now().Day())
	
	params := make(map[string]interface{})
	params["REQUESTED_AMOUNT"] = req.RequestedAmount
	params["SALARY"] = salary
	params["salary"] = salary
	params["TENOR"] = req.Tenor
	params["DAY"] = currentDay
	params["Day"] = currentDay
	params["day"] = currentDay

	// Check Limit
	limitVal := 0.0
	limitExpr, err := govaluate.NewEvaluableExpression(limitFormulaStr)
	if err == nil {
		limitResult, err := limitExpr.Evaluate(params)
		if err == nil {
			if val, ok := limitResult.(float64); ok {
				limitVal = val
				
				// Apply Category Max Limit Constraint
				if category.MaxLimit > 0 && limitVal > category.MaxLimit {
					limitVal = category.MaxLimit
				}
				
				// Multi-Loan Credit Limit Check: total_loan = Pokok + Bunga
				interestRate := product.InterestRate
				newAppTotalLoan := req.RequestedAmount + (req.RequestedAmount * (interestRate / 100.0) * float64(req.Tenor))
				cumulativeExposure := employee.TotalLoan + newAppTotalLoan

				if cumulativeExposure > limitVal {
					availableLimit := limitVal - employee.TotalLoan
					if availableLimit < 0 {
						availableLimit = 0
					}
					return SimulationResult{}, fmt.Errorf("Pengajuan ditolak. Total kewajiban pinjaman Anda (%s) + pengajuan baru (%s) = %s melebihi Credit Limit (%s). Sisa limit yang dapat diajukan (Pokok + Bunga): %s", 
						FormatIDR(employee.TotalLoan), FormatIDR(newAppTotalLoan), FormatIDR(cumulativeExposure), FormatIDR(limitVal), FormatIDR(availableLimit))
				}
			}
		}
	}

	// Calculate Admin Fee
	adminFee := 0.0
	minAdminFee, _ := strconv.ParseFloat(minAdminFeeStr, 64)
	adminExpr, err := govaluate.NewEvaluableExpression(adminFormulaStr)
	if err == nil {
		adminResult, err := adminExpr.Evaluate(params)
		if err == nil {
			if val, ok := adminResult.(float64); ok {
				adminFee = val
			}
		}
	}
	if adminFee < minAdminFee {
		adminFee = minAdminFee
	}

	// 2. Simple Simulation Logic (Pokok + Bunga Flat)
	principalPerMonth := req.RequestedAmount / float64(req.Tenor)
	interestPerMonth := (req.RequestedAmount * (product.InterestRate / 100.0))
	totalInstallment := principalPerMonth + interestPerMonth

	remLimit := limitVal - employee.TotalLoan
	if remLimit < 0 {
		remLimit = 0
	}

	return SimulationResult{
		PrincipalPerMonth:  principalPerMonth,
		InterestPerMonth:   interestPerMonth,
		AdminFee:           adminFee,
		TotalInstallment:   totalInstallment,
		TotalLoanCost:      totalInstallment*float64(req.Tenor) + adminFee,
		Tenor:              req.Tenor,
		InterestRate:       product.InterestRate,
		CreditLimit:        limitVal,
		TotalPreviousLoans: employee.TotalLoan,
		RemainingLimit:     remLimit,
	}, nil
}

func (u *applicationUseCase) SimulateApplication(req SubmitApplicationRequest) (SimulationResult, error) {
	product, err := u.getProductCached(req.ProductID)
	if err != nil {
		return SimulationResult{}, errors.New("product not found")
	}
	return u.runSimulation(req, product)
}

func (u *applicationUseCase) IsHighPrivilegeRole(roleId string) bool {
	cleanRole := strings.TrimSpace(strings.ToLower(roleId))
	if cleanRole == "" {
		return false
	}

	highPrivConfig := strings.TrimSpace(os.Getenv("HIGH_PRIVILEGE_ROLES"))
	if param, err := u.paramRepo.FindByKey("HIGH_PRIVILEGE_ROLES"); err == nil && strings.TrimSpace(param.KeyValue) != "" {
		highPrivConfig = strings.TrimSpace(param.KeyValue)
	}

	if highPrivConfig == "" {
		highPrivConfig = "1,3,admin,hrd"
	}

	allowedRoles := strings.Split(highPrivConfig, ",")
	for _, r := range allowedRoles {
		if strings.TrimSpace(strings.ToLower(r)) == cleanRole {
			return true
		}
	}
	return false
}

func (u *applicationUseCase) GetApplicationsFiltered(period string, status string, memberNo int64, roleId string, userEmployeeId int64, limit int, offset int) ([]models.LoanApplication, error) {
	isHighPriv := u.IsHighPrivilegeRole(roleId)

	targetMemberNo := memberNo
	if !isHighPriv {
		if userEmployeeId > 0 {
			targetMemberNo = userEmployeeId
		} else if memberNo > 0 {
			targetMemberNo = memberNo
		}
	}

	// If limit is not explicitly passed (limit <= 0), fetch all matching records for client-side pagination
	if limit <= 0 {
		limit = 0
	}

	return u.appRepo.FindByPeriodStatusAndMember(period, status, targetMemberNo, limit, offset)
}

func (u *applicationUseCase) GetAllApplications() ([]models.LoanApplication, error) {
	return u.appRepo.FindAll()
}

func (u *applicationUseCase) GetApplicationsByPeriod(period string) ([]models.LoanApplication, error) {
	return u.appRepo.FindByPeriod(period)
}

func (u *applicationUseCase) GetApplicationsByPeriodAndStatus(period string, status string) ([]models.LoanApplication, error) {
	return u.appRepo.FindByPeriodAndStatus(period, status)
}

func (u *applicationUseCase) writeSubmitLoanLog(rcParam string, defaultRC string, empID int64, productID int64, submissionDate time.Time, requestedAmount float64, tenor int, status string, appNo int64, createdOrUpdatedUser string) {
	rc := u.getParamVal(rcParam, defaultRC)
	logPath := u.getParamVal("LOG_LOAN_TRANSACTION_PATH", "./logs/loans.log")

	dir := filepath.Dir(logPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	appNoStr := "NULL"
	if (rc == "00" || rc == "200") && appNo > 0 {
		appNoStr = fmt.Sprintf("%d", appNo)
	}

	userStr := strings.TrimSpace(createdOrUpdatedUser)
	if userStr == "" {
		userStr = "SYSTEM_AUTO"
	}

	roleName := "User"
	var roleID int64 = 0
	userLower := strings.ToLower(userStr)
	if userLower == "system_auto" || userLower == "system" {
		roleName = "System Automation"
		roleID = 0
	} else if userLower == "admin" || userLower == "9999" {
		roleName = "Administrator"
		roleID = 1
	} else if userLower == "hrd" {
		roleName = "HRD"
		roleID = 3
	} else {
		db := u.appRepo.GetDB()
		if db != nil {
			var usr models.User
			if errU := db.Where("username = ? OR CAST(id AS TEXT) = ? OR CAST(member_no AS TEXT) = ?", userStr, userStr, userStr).First(&usr).Error; errU == nil {
				roleID = usr.RoleID
				if usr.RoleID == 1 {
					roleName = "Administrator"
				} else if usr.RoleID == 3 {
					roleName = "HRD"
				} else if usr.RoleID == 2 {
					roleName = "Anggota"
				}
			}
		}
	}

	highPrivWarning := ""
	if strings.EqualFold(status, "SUBMITTED") {
		highPrivRoles := u.getParamVal("HIGH_PRIVILEGE_ROLES", "1,3,Admin,HRD,Administrator")
		isHighPriv := false
		if roleID == 1 || roleID == 3 || strings.EqualFold(roleName, "Administrator") || strings.EqualFold(roleName, "HRD") {
			isHighPriv = true
		} else {
			tokens := strings.Split(highPrivRoles, ",")
			for _, t := range tokens {
				t = strings.TrimSpace(t)
				if strings.EqualFold(t, fmt.Sprintf("%d", roleID)) || strings.EqualFold(t, roleName) {
					isHighPriv = true
					break
				}
			}
		}
		if isHighPriv {
			highPrivWarning = u.getParamVal("LOG_WARNING_HIGH_PRIV", "HIGH_PRIVILEGE_ACTION")
		}
	}

	nowStr := time.Now().Format("2006-01-02 15:04:05")
	subDateStr := submissionDate.Format("2006-01-02")
	logLine := fmt.Sprintf("%s, %s, %s, %d, %d, %s, %.0f, %d, %s, %s, %s, %s\n",
		nowStr, appNoStr, rc, empID, productID, subDateStr, requestedAmount, tenor, status, userStr, roleName, highPrivWarning)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		_, _ = f.WriteString(logLine)
	}
}

func (u *applicationUseCase) SubmitApplication(req SubmitApplicationRequest) (*models.LoanApplication, error) {
	now := time.Now()
	// 1. Get Product Details for validation
	product, err := u.getProductCached(req.ProductID)
	if err != nil {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_OTHERS", "99", req.MemberNo, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_PRODUCT_NOT_FOUND", 0, req.CreatedUser)
		return nil, errors.New("product not found")
	}

	db := u.appRepo.GetDB()
	var member models.Member
	if err := db.Where("member_no = ? OR employee_id = ?", req.MemberNo, req.MemberNo).First(&member).Error; err != nil {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_NON_ADIRA", "11", req.MemberNo, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_NON_ADIRA", 0, req.CreatedUser)
		return nil, fmt.Errorf("ditolak bukan karyawan Adira")
	}
	req.MemberNo = member.MemberNo

	// Validate Kopkara Member Keaktifan
	statusUpper := strings.ToUpper(strings.TrimSpace(member.KopkaraStatus))
	if member.MemberNo == 0 || (statusUpper != "ACTIVE" && statusUpper != "AKTIF" && statusUpper != "") {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_NON_KOPKARA", "401", member.EmployeeID, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_NON_KOPKARA", 0, req.CreatedUser)
		return nil, fmt.Errorf("ditolak bukan anggota Kopkara")
	}

	// 2. Validate Submission Date (LOAN_START_PERIOD to LOAN_END_PERIOD)
	startPeriodStr := u.getParamVal("LOAN_START_PERIOD", "1")
	endPeriodStr := u.getParamVal("LOAN_END_PERIOD", "31")
	startPeriod, _ := strconv.Atoi(startPeriodStr)
	endPeriod, _ := strconv.Atoi(endPeriodStr)
	
	currentDay := now.Day()
	if currentDay < startPeriod || currentDay > endPeriod {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_PERIOD", "14", member.EmployeeID, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_PERIOD", 0, req.CreatedUser)
		return nil, fmt.Errorf("ditolak Di Luar Periode Tanggal Pengajuan (hanya diizinkan antara tanggal %d sampai %d)", startPeriod, endPeriod)
	}

	// 3. Validate Tenor (Override product with Global Max Tenor if smaller)
	maxTenorStr := u.getParamVal("LOAN_MAX_TENOR", strconv.Itoa(product.MaxTenorMonths))
	maxTenor, _ := strconv.Atoi(maxTenorStr)
	if product.MaxTenorMonths < maxTenor {
		maxTenor = product.MaxTenorMonths
	}
	
	if req.Tenor > maxTenor {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_TENOR", "12", member.EmployeeID, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_TENOR", 0, req.CreatedUser)
		return nil, fmt.Errorf("ditolak Tenor Melebihi Batas Maksimum (%d bulan)", maxTenor)
	}

	// 4. Run Simulation & get Breakdown (Credit Limit Validation)
	simResult, err := u.runSimulation(req, product, &member)
	if err != nil {
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_CREDIT_LIMIT", "13", member.EmployeeID, req.ProductID, now, req.RequestedAmount, req.Tenor, "REJECTED_CREDIT_LIMIT", 0, req.CreatedUser)
		return nil, err
	}

	// 5. Create Application Record
	userStr := u.resolveUsername(req.CreatedUser, member.MemberNo, member.EmployeeID)
	// Gunakan UnixNano + offset mikrodetik unik untuk mencegah bentrok primary key saat high-concurrency (500 VUs)
	appNo := now.UnixNano() + time.Now().UnixNano()%100000 + int64(now.Nanosecond()%999) 
	
	app := &models.LoanApplication{
		ApplicationNo:     appNo,
		MemberNo:          member.MemberNo,
		ProductID:         req.ProductID,
		SubmissionDate:    now,
		RequestedAmount:   req.RequestedAmount,
		Tenor:             req.Tenor,
		EligibilityResult: "ELIGIBLE",
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
	}

	err = u.appRepo.Create(app)
	if err != nil {
		return nil, err
	}

	// Write successful submit transaction log
	u.writeSubmitLoanLog("RC_SUBMIT_LOAN_SUCCESS", "00", member.EmployeeID, req.ProductID, now, req.RequestedAmount, req.Tenor, "SUBMITTED", appNo, userStr)

	// Accumulate employees.total_loan (+ Pokok + Bunga) upon successful application submission
	interestRate := product.InterestRate
	newAppTotalLoan := req.RequestedAmount + (req.RequestedAmount * (interestRate / 100.0) * float64(req.Tenor))
	_ = db.Model(&models.Employee{}).Where("employee_id = ?", member.EmployeeID).
		Updates(map[string]interface{}{
			"total_loan":   gorm.Expr("COALESCE(total_loan, 0) + ?", newAppTotalLoan),
			"updated_user": userStr,
			"updated_at":   now,
		}).Error

	// Save initial Tracking entry
	subUserId, subUserName := u.resolveUserInfo(fmt.Sprintf("%d", member.EmployeeID))
	appUserStr := strings.TrimSpace(req.CreatedUser)
	if appUserStr == "" {
		appUserStr = fmt.Sprintf("%d", member.MemberNo)
	}
	tracking := &models.LoanTracking{
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
			CreatedUser: &appUserStr,
			UpdatedAt:   &now,
			UpdatedUser: &appUserStr,
		},
	}
	_ = u.appRepo.GetDB().Create(tracking).Error

	// 6. Check LOAN_APPROVAL_AUTOMATIC parameter (default: "true")
	autoApproveStr := u.getParamVal("LOAN_APPROVAL_AUTOMATIC", "true")
	if strings.EqualFold(strings.TrimSpace(autoApproveStr), "true") {
		if errApp := u.ProcessApproval(appNo, "APPROVED", "Auto-approved by System Automation", "SYSTEM_AUTO"); errApp == nil {
			// 7. Check LOAN_DISBURSE_AUTOMATIC parameter (default: "true")
			autoDisburseStr := u.getParamVal("LOAN_DISBURSE_AUTOMATIC", "true")
			if strings.EqualFold(strings.TrimSpace(autoDisburseStr), "true") {
				var emp models.Employee
				_ = db.Where("employee_id = ?", member.EmployeeID).First(&emp).Error
				disburseReq := DisburseRequest{
					BankAccountNo: emp.BankAccountNo,
					BankName:      emp.BankName,
					Notes:         "Auto-disbursed by System Automation",
					UpdatedUser:   "SYSTEM_AUTO",
				}
				if errDisb := u.DisburseApplication(appNo, disburseReq, "SYSTEM_AUTO"); errDisb == nil {
					if refreshedApp, errRef := u.appRepo.FindByID(appNo); errRef == nil {
						return &refreshedApp, nil
					}
				}
			} else {
				if refreshedApp, errRef := u.appRepo.FindByID(appNo); errRef == nil {
					return &refreshedApp, nil
				}
			}
		}
	}

	return app, nil
}

func (u *applicationUseCase) ApproveApplication(applicationNo int64, approvedAmount float64, notes string, updatedUser ...string) error {
	userStr := "admin"
	if len(updatedUser) > 0 && strings.TrimSpace(updatedUser[0]) != "" {
		userStr = strings.TrimSpace(updatedUser[0])
	}
	return u.ProcessApproval(applicationNo, "APPROVED", notes, userStr)
}

func (u *applicationUseCase) ProcessApproval(applicationNo int64, action string, notes string, updatedUser string) error {
	app, err := u.appRepo.FindByID(applicationNo)
	if err != nil {
		return errors.New("application not found")
	}

	if notes == "" {
		return errors.New("catatan approval wajib diisi")
	}

	now := time.Now()
	app.ApprovedAt = &now

	// Calculate SLA duration directly from app.CreatedAt in memory (0 DB queries)
	createdTime := now
	if app.CreatedAt != nil {
		createdTime = *app.CreatedAt
	}
	duration := now.Sub(createdTime)
	hours := int(duration.Hours())
	mins := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	slaStr := fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)

	switch action {
	case "APPROVED", "SETUJU":
		app.Status = "APPROVED"
		if app.ApprovedAmount == nil || *app.ApprovedAmount == 0 {
			app.ApprovedAmount = &app.RequestedAmount
		}
		app.ApprovalNotes = notes

		upUserStr := u.resolveUsername(updatedUser, app.MemberNo, 0)
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
		}
		_ = u.appRepo.GetDB().Create(&contract).Error

		var empIdVal int64 = app.MemberNo
		var m models.Member
		if errM := u.appRepo.GetDB().Where("member_no = ?", app.MemberNo).First(&m).Error; errM == nil && m.EmployeeID > 0 {
			empIdVal = m.EmployeeID
		}
		u.writeSubmitLoanLog("RC_SUBMIT_LOAN_SUCCESS", "00", empIdVal, app.ProductID, now, *app.ApprovedAmount, app.Tenor, "APPROVED", app.ApplicationNo, updatedUser)

	case "REJECTED", "TOLAK":
		app.Status = "REJECTED"
		app.ApprovalNotes = notes
	case "REVISION_REQUIRED", "REVISI":
		app.Status = "REVISION_REQUIRED"
		app.ApprovalNotes = notes
	default:
		return errors.New("action invalid: gunakan APPROVED, REJECTED, atau REVISION_REQUIRED")
	}

	// Create LoanTracking record dynamically from lms_sch.employees
	upUserStr := u.resolveUsername(updatedUser, app.MemberNo, 0)
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
	}
	_ = u.appRepo.GetDB().Create(tracking).Error

	return u.appRepo.Update(&app)
}

func (u *applicationUseCase) GetLoanTrackings(applicationNo int64) ([]models.LoanTracking, error) {
	var trackings []models.LoanTracking
	err := u.appRepo.GetDB().Where("application_no = ?", applicationNo).Order("action_date desc").Find(&trackings).Error
	return trackings, err
}

type DisburseRequest struct {
	BankAccountNo string `json:"bank_account_no"`
	BankName      string `json:"bank_name"`
	Notes         string `json:"notes"`
	UpdatedUser   string `json:"updated_user"`
}

func (u *applicationUseCase) DisburseApplication(applicationNo int64, req DisburseRequest, updatedUser string) error {
	app, err := u.appRepo.FindByID(applicationNo)
	if err != nil {
		return errors.New("application not found")
	}

	disbUserStr := u.resolveUsername(updatedUser, app.MemberNo, 0)

	if app.Status != "APPROVED" {
		return fmt.Errorf("hanya pengajuan berstatus APPROVED yang dapat dicairkan (status saat ini: %s)", app.Status)
	}

	now := time.Now()

	principalAmt := app.RequestedAmount
	if app.ApprovedAmount != nil && *app.ApprovedAmount > 0 {
		principalAmt = *app.ApprovedAmount
	}
	adminFee := app.AdminFee
	disbursementAmt := principalAmt - adminFee
	if disbursementAmt < 0 {
		disbursementAmt = principalAmt
	}

	// Generate loan_no
	loanNo := time.Now().UnixNano() / int64(time.Millisecond)

	// 2. Create Record in lms_sch.loans
	loanRecord := &models.Loan{
		LoanNo:             loanNo,
		ApplicationNo:      applicationNo,
		MemberNo:           app.MemberNo,
		PrincipalAmount:    principalAmt,
		AdminFee:           adminFee,
		DisbursementAmount: disbursementAmt,
		OutstandingAmount:  principalAmt,
		Status:             "ACTIVE",
	}
	if err := u.appRepo.GetDB().Create(loanRecord).Error; err != nil {
		return fmt.Errorf("gagal membuat rekening pinjaman (loans): %v", err)
	}

	// 3. Generate Loan Schedules in lms_sch.loan_schedules
	tenor := app.Tenor
	if tenor <= 0 {
		tenor = 1
	}

	dueDay := 25
	if dueParam, err := u.paramRepo.FindByKey("LOAN_DUEDATE"); err == nil && dueParam.KeyValue != "" {
		if d, errConv := strconv.Atoi(dueParam.KeyValue); errConv == nil && d >= 1 && d <= 28 {
			dueDay = d
		}
	}

	dueMonthOffset := 0
	if dueMonthParam, err := u.paramRepo.FindByKey("LOAN_DUEMONTH"); err == nil && dueMonthParam.KeyValue != "" {
		if m, errConv := strconv.Atoi(strings.TrimSpace(dueMonthParam.KeyValue)); errConv == nil {
			dueMonthOffset = m
		}
	}

	for i := 1; i <= tenor; i++ {
		monthAdd := dueMonthOffset + (i - 1)
		dueDate := now.AddDate(0, monthAdd, 0)
		dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDay, 0, 0, 0, 0, dueDate.Location())
		periodStr := dueDate.Format("2006-01")

		schedule := &models.LoanSchedule{
			LoanNo:           loanNo,
			Period:           periodStr,
			InstallmentNo:    i,
			Principal:        app.PrincipalPerMonth,
			Interest:         app.InterestPerMonth,
			TotalInstallment: app.TotalInstallment,
			Status:           "UNPAID",
			DueDate:          dueDate,
		}
		_ = u.appRepo.GetDB().Create(schedule).Error
	}

	// 4. Update Application Status to DISBURSED
	app.Status = "DISBURSED"
	app.UpdatedAt = &now
	app.UpdatedUser = &disbUserStr
	if err := u.appRepo.Update(&app); err != nil {
		return err
	}

	var empIdVal int64 = app.MemberNo
	var m models.Member
	if errM := u.appRepo.GetDB().Where("member_no = ?", app.MemberNo).First(&m).Error; errM == nil && m.EmployeeID > 0 {
		empIdVal = m.EmployeeID
	}
	u.writeSubmitLoanLog("RC_SUBMIT_LOAN_SUCCESS", "00", empIdVal, app.ProductID, now, principalAmt, app.Tenor, "DISBURSED", app.ApplicationNo, updatedUser)

	// 5. Update Contract Status to DISBURSED
	u.appRepo.GetDB().Model(&models.LoanContract{}).Where("application_no = ?", applicationNo).Updates(map[string]interface{}{"status": "DISBURSED", "updated_user": disbUserStr, "updated_at": now})

	// 6. Calculate SLA & Record LoanTracking
	slaStr := "-"
	var lastTracking models.LoanTracking
	if err := u.appRepo.GetDB().Where("application_no = ?", applicationNo).Order("action_date desc").First(&lastTracking).Error; err == nil {
		duration := now.Sub(lastTracking.ActionDate)
		hours := int(duration.Hours())
		mins := int(duration.Minutes()) % 60
		secs := int(duration.Seconds()) % 60
		slaStr = fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}

	disburseNotes := req.Notes
	if disburseNotes == "" {
		disburseNotes = fmt.Sprintf("Dana bersih Rp %s telah dicairkan via %s (No. Rek: %s)",
			fmt.Sprintf("%.0f", disbursementAmt), req.BankName, req.BankAccountNo)
	}

	execUserKey := updatedUser
	if req.UpdatedUser != "" {
		execUserKey = req.UpdatedUser
	}
	disbUserId, disbUserName := u.resolveUserInfo(execUserKey)

	tracking := &models.LoanTracking{
		ApplicationNo: applicationNo,
		Status:        "DISBURSED",
		UserID:        disbUserId,
		UserName:      disbUserName,
		ActionDate:    now,
		Notes:         disburseNotes,
		SLADuration:   slaStr,
		IPAddress:     "127.0.0.1",
		UserAgent:     "Browser/Kopkara LMS",
	}
	_ = u.appRepo.GetDB().Create(tracking).Error

	return nil
}
