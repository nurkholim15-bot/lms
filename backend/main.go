package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"lms-backend/config"
	"lms-backend/handlers"
	"lms-backend/models"
	"lms-backend/repositories"
	"lms-backend/usecases"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// CORSMiddleware enables cross-origin resource sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error reading it. Using OS environment variables.")
	}

	// Initialize Database Connection
	config.ConnectDatabase()

	r := gin.Default()
	r.Use(CORSMiddleware()) // Enable CORS for frontend

	// Custom Request Logger Middleware for complete Backend Investigation
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[LMS-API-LOG] %s | %s | %3d | %13v | %s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
		)
	}))

	// Dependencies Injection (CRUD Master & Transactions)
	productRepo := repositories.NewProductRepository(config.DB)
	productUseCase := usecases.NewProductUseCase(productRepo)
	productHandler := handlers.NewProductHandler(productUseCase)

	paramRepo := repositories.NewParameterRepository(config.DB)
	paramUseCase := usecases.NewParameterUseCase(paramRepo)
	paramHandler := handlers.NewParameterHandler(paramUseCase)

	appRepo := repositories.NewApplicationRepository(config.DB)
	appUseCase := usecases.NewApplicationUseCase(appRepo, productRepo, paramRepo)
	appHandler := handlers.NewApplicationHandler(appUseCase)

	masterHandler := handlers.NewMasterDataHandler(config.DB)

	// LMS Core Endpoints
	api := r.Group("/api")
	{
		// Built-in Karisma Authentication Routes with detailed investigation logs
		karisma := api.Group("/karisma")
		{
			karisma.POST("/login", func(c *gin.Context) {
				var creds struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := c.ShouldBindJSON(&creds); err != nil {
					log.Printf("[KARISMA-AUTH-ERROR] Invalid login payload: %v", err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Payload login tidak valid"})
					return
				}

				username := strings.TrimSpace(creds.Username)
				password := strings.TrimSpace(creds.Password)
				log.Printf("[KARISMA-AUTH-REQUEST] Login attempt for Username: '%s'", username)

				if username == "admin" {
					if password != "admin123" {
						log.Printf("[KARISMA-AUTH-FAILED] Admin password incorrect for '%s'", username)
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Password Admin salah! (Gunakan 'admin123')"})
						return
					}
					token := "mock-token-admin"
					log.Printf("[KARISMA-AUTH-SUCCESS] ✅ Admin '%s' logged in successfully. Token: %s", username, token)
					c.JSON(http.StatusOK, gin.H{"token": token})
					return
				}

				if username == "hrd" {
					if password != "hrd123" {
						log.Printf("[KARISMA-AUTH-FAILED] HRD password incorrect for '%s'", username)
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Password HRD salah! (Gunakan 'hrd123')"})
						return
					}
					token := "mock-token-hrd"
					log.Printf("[KARISMA-AUTH-SUCCESS] ✅ HRD '%s' logged in successfully. Token: %s", username, token)
					c.JSON(http.StatusOK, gin.H{"token": token})
					return
				}

				if empID, err := strconv.Atoi(username); err == nil {
					var empName string
					config.DB.Raw("SELECT name FROM lms_sch.employees WHERE employee_id = ? LIMIT 1", empID).Scan(&empName)
					if empName == "" {
						config.DB.Raw("SELECT bank_account_name FROM lms_sch.members WHERE member_no = ? LIMIT 1", empID).Scan(&empName)
					}
					if empName == "" {
						empName = "Karyawan #" + username
					}

					if password != "password123" && password != "123456" && password != "admin123" {
						log.Printf("[KARISMA-AUTH-FAILED] Incorrect password for Anggota ID %d ('%s')", empID, empName)
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Password Anggota/Karyawan salah! (Gunakan 'password123')"})
						return
					}

					token := "mock-token-" + username
					log.Printf("[KARISMA-AUTH-SUCCESS] ✅ Anggota ID %d ('%s') logged in successfully. Token: %s", empID, empName, token)
					c.JSON(http.StatusOK, gin.H{"token": token})
					return
				}

				log.Printf("[KARISMA-AUTH-FAILED] Username '%s' not recognized in system", username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
			})

			karisma.POST("/verify", func(c *gin.Context) {
				authHeader := c.GetHeader("Authorization")
				if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
					log.Printf("[KARISMA-VERIFY-ERROR] Invalid token format in header: '%s'", authHeader)
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token Authorization tidak valid"})
					return
				}

				token := authHeader[7:]
				username := strings.TrimPrefix(token, "mock-token-")
				log.Printf("[KARISMA-VERIFY-ATTEMPT] Verifying token for user: '%s'", username)

				role := "anggota"
				name := "Karyawan #" + username
				empID := 1001

				if username == "admin" {
					role = "admin"
					name = "Administrator Koperasi"
					empID = 9999
				} else if username == "hrd" {
					role = "hrd"
					name = "Tim HRD Adira"
					empID = 8888
				} else if idVal, err := strconv.Atoi(username); err == nil {
					empID = idVal
					var dbName string
					config.DB.Raw("SELECT name FROM lms_sch.employees WHERE employee_id = ? LIMIT 1", empID).Scan(&dbName)
					if dbName != "" {
						name = dbName
					}
				}

				log.Printf("[KARISMA-VERIFY-SUCCESS] ✅ Token verified for '%s' (ID: %d, Role: %s)", name, empID, role)
				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"user": gin.H{
						"employee_id": empID,
						"name":        name,
						"eligible":    true,
						"role":        role,
					},
				})
			})
		}

		// Master Data Generic Endpoints
		master := api.Group("/master")
		{
			master.GET("/:table", masterHandler.GetAll)
			master.POST("/:table", masterHandler.Save)
			master.DELETE("/:table/:id", masterHandler.Delete)
		}

		api.GET("/user-info/:employee_id", masterHandler.GetUserInfo)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "LMS Backend is running smoothly."})
		})
		
		// Master Data: Products
		products := api.Group("/products")
		{
			products.GET("", productHandler.GetAll)
			products.GET("/:id", productHandler.GetByID)
			products.POST("", productHandler.Save)
			products.PUT("/:id", productHandler.Save)
			products.DELETE("/:id", productHandler.Delete)
		}

		// Transaction: Loan Applications (Origination)
		applications := api.Group("/applications")
		{
			applications.GET("", appHandler.GetAll)
			applications.POST("/simulate", appHandler.Simulate)
			applications.POST("", appHandler.Submit)
			applications.POST("/:id/approve", appHandler.Approve)
			applications.POST("/:id/disburse", appHandler.Disburse)
			applications.GET("/:id/trackings", appHandler.GetTrackings)
		}

		// Master Data: Parameters
		parameters := api.Group("/parameters")
		{
			parameters.GET("", paramHandler.GetAll)
			parameters.POST("", paramHandler.Save)
		}

		// Get All Members Endpoint for Manual Repayment Dropdown
		api.GET("/members/all", func(c *gin.Context) {
			type MemberDTO struct {
				MemberNo   int64  `json:"member_no"`
				EmployeeID int64  `json:"employee_id"`
				Name       string `json:"name"`
			}
			var dtos []MemberDTO
			query := `
				SELECT 
					m.member_no,
					COALESCE(e.employee_id, m.employee_id) AS employee_id,
					COALESCE(e.name, m.bank_account_name, CONCAT('Anggota #', m.member_no)) AS name
				FROM lms_sch.members m
				LEFT JOIN lms_sch.employees e ON m.employee_id = e.employee_id OR m.member_no = e.employee_id
				ORDER BY m.member_no ASC;
			`
			if err := config.DB.Raw(query).Scan(&dtos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": dtos})
		})

		// Real Payroll Schedules Endpoint (PostgreSQL DB)
		api.GET("/payroll/schedules", func(c *gin.Context) {
			type PayrollScheduleDTO struct {
				ID               int64     `json:"id"`
				LoanNo           int64     `json:"loan_no"`
				ApplicationNo    int64     `json:"application_no"`
				MemberNo         int64     `json:"member_no"`
				EmployeeName     string    `json:"employee_name"`
				DeptNo           string    `json:"dept_no"`
				NIK              string    `json:"nik"`
				Period           string    `json:"period"`
				InstallmentNo    int       `json:"installment_no"`
				Principal        float64   `json:"principal"`
				Interest         float64   `json:"interest"`
				TotalInstallment float64   `json:"total_installment"`
				Status           string    `json:"status"`
				DueDate          time.Time `json:"due_date"`
				RefNo            string    `json:"ref_no"`
			}

			var dtos []PayrollScheduleDTO
			query := `
				SELECT 
					ls.id,
					ls.loan_no,
					l.application_no,
					l.member_no,
					COALESCE(e.name, m.bank_account_name, CONCAT('Karyawan #', l.member_no)) AS employee_name,
					COALESCE(e.deptno, 'IT-01') AS dept_no,
					COALESCE(m.bank_account_no, CAST(l.member_no AS VARCHAR)) AS nik,
					ls.period,
					ls.installment_no,
					ls.principal,
					ls.interest,
					ls.total_installment,
					ls.status,
					ls.due_date,
					CONCAT('LMS-PAY-', REPLACE(ls.period, '-', ''), '-', l.member_no) AS ref_no
				FROM lms_sch.loan_schedules ls
				JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
				LEFT JOIN lms_sch.members m ON l.member_no = m.member_no
				LEFT JOIN lms_sch.employees e ON e.employee_id = l.member_no OR e.employee_id = m.employee_id
				WHERE ls.status != 'PAID' AND (l.status = 'ACTIVE' OR l.status IS NULL)
				ORDER BY ls.due_date ASC, ls.id ASC;
			`
			if err := config.DB.Raw(query).Scan(&dtos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, dtos)
		})

		// Direct Export File to FOLDER_BILL_EXPORT Directory Endpoint
		api.POST("/payroll/export", func(c *gin.Context) {
			type ExportRequest struct {
				CustomFolder string `json:"custom_folder"`
				CutoffDate   string `json:"cutoff_date"`
			}
			var req ExportRequest
			_ = c.ShouldBindJSON(&req)

			cutoffDate := strings.TrimSpace(req.CutoffDate)
			if cutoffDate == "" {
				now := time.Now()
				lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
				cutoffDate = lastDay.Format("2006-01-02")
			}
			cutoffPeriod := cutoffDate
			if len(cutoffPeriod) >= 7 {
				cutoffPeriod = cutoffPeriod[0:7]
			}

			targetFolder := req.CustomFolder
			if targetFolder == "" {
				var folderParam string
				config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'FOLDER_BILL_EXPORT' LIMIT 1").Scan(&folderParam)
				if folderParam != "" {
					targetFolder = folderParam
				} else {
					targetFolder = "D:\\Data_NK\\Project5\\LMS\\export_payroll"
				}
			}

			type RowDTO struct {
				Nik             string  `gorm:"column:nik_adira"`
				EmployeeID      int64   `gorm:"column:employee_id"`
				LoanNo          int64   `gorm:"column:loan_no"`
				NamaKaryawan    string  `gorm:"column:nama_karyawan"`
				DeptNo          string  `gorm:"column:dept_no"`
				KodePotongan    string  `gorm:"column:kode_potongan"`
				NamaPotongan    string  `gorm:"column:nama_potongan"`
				Periode         string  `gorm:"column:periode"`
				NominalPotongan float64 `gorm:"column:nominal_potongan"`
				NoReferensi     string  `gorm:"column:no_referensi"`
			}

			var rows []RowDTO
			query := `
				SELECT 
					COALESCE(m.bank_account_no, CAST(l.member_no AS VARCHAR)) AS nik_adira,
					l.member_no AS employee_id,
					l.loan_no AS loan_no,
					COALESCE(e.name, m.bank_account_name, CONCAT('Anggota #', l.member_no)) AS nama_karyawan,
					COALESCE(e.deptno, 'IT-01') AS dept_no,
					'POT_KOPKARA' AS kode_potongan,
					'Potongan Angsuran Kopkara' AS nama_potongan,
					ls.period AS periode,
					ls.total_installment AS nominal_potongan,
					CONCAT('LMS-PAY-', REPLACE(ls.period, '-', ''), '-', l.member_no) AS no_referensi
				FROM lms_sch.loan_schedules ls
				JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
				LEFT JOIN lms_sch.members m ON l.member_no = m.member_no
				LEFT JOIN lms_sch.employees e ON e.employee_id = l.member_no OR e.employee_id = m.employee_id
				WHERE ls.status IN ('UNPAID', 'PARTIAL') 
				  AND (l.status = 'ACTIVE' OR l.status IS NULL)
				  AND (ls.due_date <= CAST(? AS DATE) OR ls.period <= ?)
				ORDER BY ls.due_date ASC, ls.id ASC;
			`
			if err := config.DB.Raw(query, cutoffDate, cutoffPeriod).Scan(&rows).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed: " + err.Error()})
				return
			}

			csvContent := "NIK_ADIRA,EMPLOYEE_ID,LOAN_NO,NAMA_KARYAWAN,DEPT_NO,KODE_POTONGAN,NAMA_POTONGAN,PERIODE,NOMINAL_TAGIHAN,NOMINAL_TERPOTONG,STATUS_POTONGAN,KETERANGAN,NO_REFERENSI\n"
			for _, r := range rows {
				csvContent += fmt.Sprintf("%s,%d,%d,%s,%s,%s,%s,%s,%.2f,,,%s\n",
					r.Nik, r.EmployeeID, r.LoanNo, r.NamaKaryawan, r.DeptNo, r.KodePotongan, r.NamaPotongan, r.Periode, r.NominalPotongan, r.NoReferensi)
			}

			// Smart Cross-Platform Path Resolver (Windows / WSL / Linux)
			rawPath := strings.TrimRight(targetFolder, "/\\")
			rawPath = strings.ReplaceAll(rawPath, "/", "\\") // Normalize to Windows style for display

			// For file system operations on Linux/WSL if path starts with D:\ or D:/
			diskPath := rawPath
			if strings.HasPrefix(strings.ToLower(diskPath), "d:\\") || strings.HasPrefix(strings.ToLower(diskPath), "d:/") {
				if os.PathSeparator == '/' {
					// Running inside Linux / WSL environment
					diskPath = "/mnt/d/" + strings.ReplaceAll(diskPath[3:], "\\", "/")
				}
			}

			// Ensure target folder exists on server/disk
			if err := os.MkdirAll(diskPath, 0755); err != nil {
				log.Printf("Error creating export directory %s: %v", diskPath, err)
			}

			fileName := fmt.Sprintf("ADIRA_PAYROLL_KOPKARA_OUTGOING_%s.csv", time.Now().Format("200601"))

			// Disk write path
			diskFullPath := filepath.Join(diskPath, fileName)
			if err := os.WriteFile(diskFullPath, []byte(csvContent), 0644); err != nil {
				log.Printf("Failed writing file to disk %s: %v", diskFullPath, err)
			}

			// Clean User Display Path
			userDisplayPath := rawPath + "\\" + fileName

			c.JSON(http.StatusOK, gin.H{
				"status":      "success",
				"message":     fmt.Sprintf("File CSV berhasil digenerate dan disimpan langsung ke folder: %s", userDisplayPath),
				"file_path":   userDisplayPath,
				"file_name":   fileName,
				"total_rows":  len(rows),
				"csv_content": csvContent,
			})
		})

		// Import Payroll Reconciliation & Update Database Endpoint
		api.POST("/payroll/import", func(c *gin.Context) {
			type ImportRow struct {
				RefNo           string  `json:"ref_no"`
				EmployeeID      int64   `json:"employee_id"`
				LoanNo          int64   `json:"loan_no"`
				Period          string  `json:"period"`
				NominalOriginal float64 `json:"nominal_original"`
				Deducted        float64 `json:"deducted"`
				Status          string  `json:"status"`
				Keterangan      string  `json:"keterangan"`
			}
			type ImportRequest struct {
				FileName    string      `json:"file_name"`
				UpdatedUser string      `json:"updated_user"`
				Rows        []ImportRow `json:"rows"`
			}
			var req ImportRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Printf("[IMPORT-API] Error binding JSON: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid import payload: " + err.Error()})
				return
			}

			fileName := strings.TrimSpace(req.FileName)
			if fileName != "" {
				var count int64
				config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loan_payroll_import_logs WHERE file_name = ?", fileName).Scan(&count)
				if count > 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "File import sudah pernah diupdate"})
					return
				}
			} else {
				fileName = fmt.Sprintf("payroll_import_%s.csv", time.Now().Format("20060102_150405"))
			}

			updaterName := strings.TrimSpace(req.UpdatedUser)
			if updaterName == "" {
				updaterName = "10101"
			}

			log.Printf("[IMPORT-API] Processing import for file '%s' by user '%s' with %d rows", fileName, updaterName, len(req.Rows))

			totalRecords := len(req.Rows)
			var totalAmount float64
			var totalSuccessRecords int
			var totalSuccessAmount float64
			var totalFailedRecords int
			var totalFailedAmount float64

			for _, row := range req.Rows {
				statusUpper := strings.ToUpper(strings.TrimSpace(row.Status))
				memberNo := row.EmployeeID
				periodStr := strings.TrimSpace(row.Period)
				if periodStr == "" {
					periodStr = "2026-08"
				}

				totalAmount += row.NominalOriginal

				// Find corresponding loan_no
				var loanNoPtr *int64
				var foundLoanNo int64
				if row.LoanNo > 0 {
					foundLoanNo = row.LoanNo
					loanNoPtr = &foundLoanNo
				} else {
					config.DB.Raw("SELECT loan_no FROM lms_sch.loans WHERE member_no = ? OR loan_no = ? LIMIT 1", memberNo, memberNo).Scan(&foundLoanNo)
					if foundLoanNo > 0 {
						loanNoPtr = &foundLoanNo
					}
				}

				// Insert log to lms_sch.payroll_deductions
				insertLogSQL := `
					INSERT INTO lms_sch.payroll_deductions (loan_no, period, nominal_original, nominal_deducted, process_date, status, failure_reason, created_at, created_user)
					VALUES (?, ?, ?, ?, NOW(), ?, ?, NOW(), ?);
				`
				config.DB.Exec(insertLogSQL, loanNoPtr, periodStr, row.NominalOriginal, row.Deducted, statusUpper, row.Keterangan, updaterName)

				deductedVal := row.Deducted
				if deductedVal <= 0 && row.NominalOriginal > 0 && statusUpper != "FAILED" && statusUpper != "REJECTED" {
					deductedVal = row.NominalOriginal
				}

				if deductedVal > 0 {
					totalSuccessRecords++
					totalSuccessAmount += deductedVal
					if deductedVal < row.NominalOriginal {
						totalFailedAmount += (row.NominalOriginal - deductedVal)
					}
				} else {
					totalFailedRecords++
					totalFailedAmount += row.NominalOriginal
				}

				// 1. Direct LOAN_NO Processing Branch (when CSV row explicitly specifies loan_no)
				if row.LoanNo > 0 {
					type ScheduleRow struct {
						ID                   int64
						LoanNo               int64
						ApplicationNo        int64
						TotalInstallment     float64
						AmountPaid           float64
						RemainingInstallment float64
						Status               string
					}
					var s ScheduleRow
					config.DB.Raw(`
						SELECT ls.id, ls.loan_no, l.application_no, ls.total_installment, 
						       COALESCE(ls.amount_paid, 0) AS amount_paid, 
						       COALESCE(ls.remaining_installment, 0) AS remaining_installment, 
						       ls.status
						FROM lms_sch.loan_schedules ls
						JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
						WHERE ls.loan_no = ?
						  AND (ls.period = ? OR ? = '')
						ORDER BY ls.due_date ASC, ls.id ASC
						LIMIT 1
					`, row.LoanNo, periodStr, periodStr).Scan(&s)

					if s.ID > 0 {
						payThis := deductedVal
						newAmountPaid := payThis
						var newRemaining float64

						if newAmountPaid == s.TotalInstallment {
							newRemaining = 0
						} else if newAmountPaid > s.TotalInstallment {
							newRemaining = s.TotalInstallment - newAmountPaid // Negative (overpaid credit, e.g. 609k - 650k = -41k)
						} else {
							newRemaining = newAmountPaid - s.TotalInstallment // Negative (underpaid shortfall, e.g. 41k - 307.5k = -266.5k)
						}

						newStatus := "UNPAID"
						if newAmountPaid >= s.TotalInstallment {
							newStatus = "PAID"
						} else if newAmountPaid > 0 {
							newStatus = "PARTIAL"
						}

						// Update loan_schedules strictly for this loan_no without accumulating to other loans
						config.DB.Exec(`
							UPDATE lms_sch.loan_schedules
							SET amount_paid = ?, remaining_installment = ?, status = ?, updated_at = NOW(), updated_user = ?
							WHERE id = ?
						`, newAmountPaid, newRemaining, newStatus, updaterName, s.ID)

						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET outstanding_amount = GREATEST(0, outstanding_amount - ?), updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?
						`, payThis, updaterName, s.LoanNo)

						var unpaidCount int64
						config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID'", s.LoanNo).Scan(&unpaidCount)
						if unpaidCount == 0 {
							config.DB.Exec(`
								UPDATE lms_sch.loans
								SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
								WHERE loan_no = ?
							`, updaterName, s.LoanNo)

							config.DB.Exec(`
								UPDATE lms_sch.loan_applications
								SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
								WHERE application_no = ?
							`, updaterName, s.ApplicationNo)
						}
					}
				} else if deductedVal > 0 {
					// 2. Fallback Waterfall Payment Allocation Branch (when CSV row does NOT specify loan_no)
					availPayment := deductedVal

					type ScheduleRow struct {
						ID                   int64
						LoanNo               int64
						ApplicationNo        int64
						TotalInstallment     float64
						AmountPaid           float64
						RemainingInstallment float64
						Status               string
					}
					var scheds []ScheduleRow

					config.DB.Raw(`
						SELECT ls.id, ls.loan_no, l.application_no, ls.total_installment, 
						       COALESCE(ls.amount_paid, 0) AS amount_paid, 
						       COALESCE(ls.remaining_installment, 0) AS remaining_installment, 
						       ls.status
						FROM lms_sch.loan_schedules ls
						JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
						WHERE (l.member_no = ? OR ls.loan_no = ?)
						  AND ls.status IN ('UNPAID', 'PARTIAL')
						  AND (ls.period = ? OR ? = '')
						ORDER BY ls.due_date ASC, ls.id ASC
					`, memberNo, memberNo, periodStr, periodStr).Scan(&scheds)

					if len(scheds) == 0 {
						config.DB.Raw(`
							SELECT ls.id, ls.loan_no, l.application_no, ls.total_installment, 
							       COALESCE(ls.amount_paid, 0) AS amount_paid, 
							       COALESCE(ls.remaining_installment, 0) AS remaining_installment, 
							       ls.status
							FROM lms_sch.loan_schedules ls
							JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
							WHERE (l.member_no = ? OR ls.loan_no = ?)
							  AND ls.status IN ('UNPAID', 'PARTIAL')
							ORDER BY ls.due_date ASC, ls.id ASC
						`, memberNo, memberNo).Scan(&scheds)
					}

					for idx, s := range scheds {
						if availPayment <= 0 {
							break
						}

						isLastSched := (idx == len(scheds)-1)

						payThis := availPayment
						if !isLastSched && payThis > s.TotalInstallment && s.TotalInstallment > 0 {
							payThis = s.TotalInstallment
						}

						newAmountPaid := payThis
						var newRemaining float64
						if newAmountPaid == s.TotalInstallment {
							newRemaining = 0
						} else if newAmountPaid > s.TotalInstallment {
							newRemaining = s.TotalInstallment - newAmountPaid
						} else {
							newRemaining = newAmountPaid - s.TotalInstallment
						}

						availPayment -= payThis

						newStatus := "UNPAID"
						if newAmountPaid >= s.TotalInstallment {
							newStatus = "PAID"
						} else if newAmountPaid > 0 {
							newStatus = "PARTIAL"
						}

						// 1. Update loan_schedules
						config.DB.Exec(`
							UPDATE lms_sch.loan_schedules
							SET amount_paid = ?, remaining_installment = ?, status = ?, updated_at = NOW(), updated_user = ?
							WHERE id = ?
						`, newAmountPaid, newRemaining, newStatus, updaterName, s.ID)

						// 2. Update loans outstanding_amount
						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET outstanding_amount = GREATEST(0, outstanding_amount - ?), updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?
						`, payThis, updaterName, s.LoanNo)

						// 3. Check if all schedules for loan are PAID -> CLOSE loan and loan_application
						var unpaidCount int64
						config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID'", s.LoanNo).Scan(&unpaidCount)
						if unpaidCount == 0 {
							config.DB.Exec(`
								UPDATE lms_sch.loans
								SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
								WHERE loan_no = ?
							`, updaterName, s.LoanNo)

							config.DB.Exec(`
								UPDATE lms_sch.loan_applications
								SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
								WHERE application_no = ?
							`, updaterName, s.ApplicationNo)
						}
					}
				}
			}

			// Auto Backup / Move CSV File to FOLDER_BILL_EXPORT_BCK or FOLDER_BILL_IMPORT_BCK
			backupMsg := ""
			if req.FileName != "" {
				var bckFolderParam, importFolderParam string
				config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'FOLDER_BILL_IMPORT_BCK' LIMIT 1").Scan(&bckFolderParam)
				if bckFolderParam == "" {
					config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'FOLDER_BILL_EXPORT_BCK' LIMIT 1").Scan(&bckFolderParam)
				}
				config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'FOLDER_BILL_IMPORT' LIMIT 1").Scan(&importFolderParam)

				if bckFolderParam == "" {
					bckFolderParam = "D:\\Data_NK\\Project5\\LMS\\Billing\\Backup\\"
				}
				if importFolderParam == "" {
					importFolderParam = "D:\\Data_NK\\Project5\\LMS\\Billing\\Import\\"
				}

				resolvePath := func(p string) string {
					clean := strings.TrimRight(p, "/\\")
					clean = strings.ReplaceAll(clean, "/", "\\")
					if os.PathSeparator == '/' && (strings.HasPrefix(strings.ToLower(clean), "d:\\") || strings.HasPrefix(strings.ToLower(clean), "d:/")) {
						return "/mnt/d/" + strings.ReplaceAll(clean[3:], "\\", "/")
					}
					return clean
				}

				diskImportDir := resolvePath(importFolderParam)
				diskBckDir := resolvePath(bckFolderParam)
				_ = os.MkdirAll(diskBckDir, 0755)

				srcFilePath := filepath.Join(diskImportDir, req.FileName)
				if _, err := os.Stat(srcFilePath); err == nil {
					ext := filepath.Ext(req.FileName)
					baseName := strings.TrimSuffix(req.FileName, ext)
					bckFileName := fmt.Sprintf("%s_PROCESSED_%s%s", baseName, time.Now().Format("20060102_150405"), ext)
					dstFilePath := filepath.Join(diskBckDir, bckFileName)

					if err := os.Rename(srcFilePath, dstFilePath); err == nil {
						log.Printf("[IMPORT-API] File successfully backed up: %s -> %s", srcFilePath, dstFilePath)
						backupMsg = fmt.Sprintf("\n\n📦 File CSV [%s] telah otomatis dipindahkan ke folder Backup:\n📁 %s", req.FileName, dstFilePath)
					} else {
						inputData, readErr := os.ReadFile(srcFilePath)
						if readErr == nil {
							_ = os.WriteFile(dstFilePath, inputData, 0644)
							_ = os.Remove(srcFilePath)
							log.Printf("[IMPORT-API] File copied and deleted: %s -> %s", srcFilePath, dstFilePath)
							backupMsg = fmt.Sprintf("\n\n📦 File CSV [%s] telah otomatis dipindahkan ke folder Backup:\n📁 %s", req.FileName, dstFilePath)
						}
					}
				}
			}

			// Insert summary record into lms_sch.loan_payroll_import_logs
			config.DB.Exec(`
				INSERT INTO lms_sch.loan_payroll_import_logs 
				(file_name, total_records, total_amount, total_success_records, total_success_amount, total_failed_records, total_failed_amount, created_at, created_user, updated_at, updated_user)
				VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), ?)
			`, fileName, totalRecords, totalAmount, totalSuccessRecords, totalSuccessAmount, totalFailedRecords, totalFailedAmount, updaterName, updaterName)

			msgInfo := fmt.Sprintf("Import file '%s' berhasil diproses! Total %d tagihan ter-update.%s", fileName, totalSuccessRecords, backupMsg)
			if totalSuccessRecords == 0 {
				msgInfo = fmt.Sprintf("Import file '%s' diproses! Total 0 tagihan ter-update (Catatan: seluruh tagihan karyawan pada file ini sudah berstatus LUNAS / PAID sebelumnya di database).%s", fileName, backupMsg)
			}

			c.JSON(http.StatusOK, gin.H{
				"status":                "success",
				"message":               msgInfo,
				"file_name":             fileName,
				"total_records":         totalRecords,
				"total_amount":          totalAmount,
				"total_success_records": totalSuccessRecords,
				"total_success_amount":  totalSuccessAmount,
				"total_failed_records":  totalFailedRecords,
				"total_failed_amount":   totalFailedAmount,
			})
		})

		// Manual Repayment Endpoint (For Resigned Employees / Direct Transfer / Pesangon Offset)
		api.POST("/payroll/manual-repayment", func(c *gin.Context) {
			type ManualRequest struct {
				LoanNo       int64   `json:"loan_no"`
				MemberNo     int64   `json:"member_no"`
				Period       string  `json:"period"`
				PaymentType  string  `json:"payment_type"` // 'TRANSFER_BANK', 'POTONG_PESANGON', 'KOMPENSASI_SIMPANAN'
				Nominal      float64 `json:"nominal"`
				ReferenceNo  string  `json:"reference_no"`
				Notes        string  `json:"notes"`
				UpdatedUser  string  `json:"updated_user"`
				IsFullSettle bool    `json:"is_full_settlement"`
			}
			var req ManualRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Payload pelunasan manual tidak valid: " + err.Error()})
				return
			}

			updaterName := strings.TrimSpace(req.UpdatedUser)
			if updaterName == "" {
				updaterName = "10101"
			}

			// 1. Log audit to lms_sch.payroll_deductions
			reasonLog := fmt.Sprintf("Pelunasan Manual [%s] - Ref: %s - %s", req.PaymentType, req.ReferenceNo, req.Notes)
			config.DB.Exec(`
				INSERT INTO lms_sch.payroll_deductions (loan_no, period, nominal_original, nominal_deducted, process_date, status, failure_reason, created_at, created_user)
				VALUES (?, ?, ?, ?, NOW(), 'SUCCESS_MANUAL', ?, NOW(), ?);
			`, req.LoanNo, req.Period, req.Nominal, req.Nominal, reasonLog, updaterName)

			// 2. Update loan_schedules
			if req.IsFullSettle {
				// Pay all remaining unpaid schedules for this loan
				config.DB.Exec(`
					UPDATE lms_sch.loan_schedules
					SET amount_paid = total_installment, remaining_installment = 0, status = 'PAID', updated_at = NOW(), updated_user = ?
					WHERE loan_no = ? AND status != 'PAID';
				`, updaterName, req.LoanNo)

				// Close loan in lms_sch.loans and lms_sch.loan_applications
				config.DB.Exec(`
					UPDATE lms_sch.loans
					SET outstanding_amount = 0, status = 'CLOSED', updated_at = NOW(), updated_user = ?
					WHERE loan_no = ?;
				`, updaterName, req.LoanNo)

				config.DB.Exec(`
					UPDATE lms_sch.loan_applications
					SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
					WHERE application_no = (SELECT application_no FROM lms_sch.loans WHERE loan_no = ?);
				`, updaterName, req.LoanNo)
			} else {
				// Pay single installment
				var schedID int64
				var totInst float64
				config.DB.Raw(`
					SELECT id, total_installment FROM lms_sch.loan_schedules
					WHERE loan_no = ? AND status != 'PAID'
					ORDER BY due_date ASC, id ASC LIMIT 1
				`, req.LoanNo).Row().Scan(&schedID, &totInst)

				if schedID > 0 {
					paidVal := req.Nominal
					if paidVal <= 0 {
						paidVal = totInst
					}
					remVal := paidVal - totInst
					newStat := "PARTIAL"
					if paidVal >= totInst {
						newStat = "PAID"
					}
					config.DB.Exec(`
						UPDATE lms_sch.loan_schedules
						SET amount_paid = ?, remaining_installment = ?, status = ?, updated_at = NOW(), updated_user = ?
						WHERE id = ?;
					`, paidVal, remVal, newStat, updaterName, schedID)

					config.DB.Exec(`
						UPDATE lms_sch.loans
						SET outstanding_amount = GREATEST(0, outstanding_amount - ?), updated_at = NOW(), updated_user = ?
						WHERE loan_no = ?;
					`, paidVal, updaterName, req.LoanNo)

					// Check if all schedules are PAID
					var unpaidCount int64
					config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID'", req.LoanNo).Scan(&unpaidCount)
					if unpaidCount == 0 {
						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?;
						`, updaterName, req.LoanNo)

						config.DB.Exec(`
							UPDATE lms_sch.loan_applications
							SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
							WHERE application_no = (SELECT application_no FROM lms_sch.loans WHERE loan_no = ?);
						`, updaterName, req.LoanNo)
					}
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": fmt.Sprintf("Berhasil memproses Pelunasan Manual [%s] sebesar Rp %.2f untuk Pinjaman #%d oleh User ID: %s", req.PaymentType, req.Nominal, req.LoanNo, updaterName),
			})
		})

		// GET Payroll Deductions Log Endpoint
		api.GET("/payroll/deductions", func(c *gin.Context) {
			var logs []map[string]interface{}
			query := `
				SELECT 
					pd.id,
					pd.loan_no,
					COALESCE(l.member_no, 0) AS member_no,
					COALESCE(e.name, m.bank_account_name, 'Karyawan') AS employee_name,
					pd.period,
					pd.nominal_original,
					pd.nominal_deducted,
					pd.process_date,
					pd.status,
					pd.failure_reason,
					pd.created_user
				FROM lms_sch.payroll_deductions pd
				LEFT JOIN lms_sch.loans l ON pd.loan_no = l.loan_no
				LEFT JOIN lms_sch.employees e ON e.employee_id = l.member_no
				LEFT JOIN lms_sch.members m ON m.member_no = l.member_no
				ORDER BY pd.id DESC
				LIMIT 100;
			`
			config.DB.Raw(query).Scan(&logs)
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": logs})
		})

		// GET Payroll Import Logs Endpoint
		api.GET("/payroll/import-logs", func(c *gin.Context) {
			var logs []models.LoanPayrollImportLog
			config.DB.Raw("SELECT * FROM lms_sch.loan_payroll_import_logs ORDER BY id DESC LIMIT 100").Scan(&logs)
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": logs})
		})
		
		parameters.DELETE("/:id", paramHandler.Delete)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
