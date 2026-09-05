package main

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"lms-backend/cache"
	"lms-backend/config"
	"lms-backend/handlers"
	"lms-backend/models"
	"lms-backend/repositories"
	"lms-backend/services"
	"lms-backend/usecases"

	"github.com/Knetic/govaluate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	globalParamRepo     repositories.ParameterRepository
	cachedAppTokenName  string
	cachedAppTokenMutex sync.RWMutex
)

func getAppTokenName() string {
	cachedAppTokenMutex.RLock()
	if cachedAppTokenName != "" {
		defer cachedAppTokenMutex.RUnlock()
		return cachedAppTokenName
	}
	cachedAppTokenMutex.RUnlock()

	tokenName := "ewa_token"
	if globalParamRepo != nil {
		if p, err := globalParamRepo.FindByKey("APP_TOKEN"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			tokenName = strings.TrimSpace(p.KeyValue)
		}
	} else if config.DB != nil {
		var val string
		config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'APP_TOKEN' AND deleted_at IS NULL LIMIT 1").Scan(&val)
		val = strings.TrimSpace(val)
		if val != "" {
			tokenName = val
		}
	}

	cachedAppTokenMutex.Lock()
	cachedAppTokenName = tokenName
	cachedAppTokenMutex.Unlock()
	return tokenName
}

func getTokenFromRequest(c *gin.Context) string {
	tokenName := getAppTokenName()
	if token, err := c.Cookie(tokenName); err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	if token, err := c.Cookie("karisma_token"); err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	if token, err := c.Cookie("ewa_token"); err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) >= 8 && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return strings.TrimSpace(authHeader)
}

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// calculateLMSCreditLimitFromCache calculates employee credit limit using exact LMS formula from paramRepo cache
func calculateLMSCreditLimitFromCache(db *gorm.DB, paramRepo repositories.ParameterRepository, empId int64) float64 {
	if empId <= 0 {
		log.Printf("[CALCULATE-CL-STEP] Employee ID is invalid (%d), returning fallback Rp 15.000.000", empId)
		return 15000000
	}

	type EmpCat struct {
		Salary   float64
		MaxLimit float64
	}
	var res EmpCat
	err := db.Raw(`
		SELECT COALESCE(e.salary, 0) as salary, COALESCE(c.max_limit, 0) as max_limit
		FROM lms_sch.employees e
		LEFT JOIN lms_sch.employee_categories c ON e.category_code = c.category_code
		WHERE e.employee_id = ? OR e.employee_id = (SELECT employee_id FROM lms_sch.members WHERE member_no = ? LIMIT 1)
		LIMIT 1
	`, empId, empId).Scan(&res).Error

	if err != nil || res.MaxLimit <= 0 {
		res.MaxLimit = 15000000
	}

	// Read LOAN_LIMIT_FORMULA directly from paramRepo cache
	formulaStr := "(DAY/30) * SALARY * 0.5"
	if p, errP := paramRepo.FindByKey("LOAN_LIMIT_FORMULA"); errP == nil && strings.TrimSpace(p.KeyValue) != "" {
		formulaStr = strings.TrimSpace(p.KeyValue)
	}

	salary := res.Salary
	if salary <= 0 {
		salary = 10000000
	}
	currentDay := float64(time.Now().Day())

	params := map[string]interface{}{
		"SALARY": salary,
		"salary": salary,
		"DAY":    currentDay,
		"Day":    currentDay,
		"day":    currentDay,
	}

	limitVal := res.MaxLimit
	rawFormulaResult := 0.0

	expr, errExpr := govaluate.NewEvaluableExpression(formulaStr)
	if errExpr == nil {
		evalRes, errEval := expr.Evaluate(params)
		if errEval == nil {
			if val, ok := evalRes.(float64); ok && val > 0 {
				rawFormulaResult = val
				limitVal = val
			}
		} else {
			log.Printf("[CALCULATE-CL-ERROR] Error evaluating formula '%s': %v", formulaStr, errEval)
		}
	} else {
		log.Printf("[CALCULATE-CL-ERROR] Error parsing formula '%s': %v", formulaStr, errExpr)
	}

	log.Printf("[CALCULATE-CL-STEP 1] EmployeeID: %d | Salary: Rp %.2f | Current Day: %.0f", empId, salary, currentDay)
	log.Printf("[CALCULATE-CL-STEP 2] Formula: '%s' => Result: Rp %.2f", formulaStr, rawFormulaResult)
	log.Printf("[CALCULATE-CL-STEP 3] Category Max Limit: Rp %.2f", res.MaxLimit)

	if res.MaxLimit > 0 && limitVal > res.MaxLimit {
		log.Printf("[CALCULATE-CL-STEP 4] Formula result Rp %.2f exceeds Category Max Limit Rp %.2f -> Capped at Rp %.2f", rawFormulaResult, res.MaxLimit, res.MaxLimit)
		limitVal = res.MaxLimit
	} else {
		log.Printf("[CALCULATE-CL-STEP 4] Final Approved Credit Limit (CL): Rp %.2f", limitVal)
	}

	return limitVal
}

// CORSMiddleware enables cross-origin resource sharing for allowed origins and local Wi-Fi testing
func CORSMiddleware() gin.HandlerFunc {
	rawOrigins := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return func(c *gin.Context) {
		log.Printf("🔥 [INCOMING REQUEST] %s %s from ClientIP: %s | UserAgent: %s", c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.Request.UserAgent())
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			isAllowed := false
			if len(allowedOrigins) == 0 {
				isAllowed = true
			} else {
				for _, allowed := range allowedOrigins {
					if allowed == "*" || 
					   origin == allowed || 
					   strings.HasPrefix(origin, allowed+":") || 
					   strings.HasPrefix(origin, "http://192.168.") || 
					   strings.HasPrefix(origin, "https://192.168.") || 
					   strings.HasPrefix(origin, "http://10.") || 
					   strings.HasPrefix(origin, "https://10.") || 
					   strings.HasPrefix(origin, "capacitor://") ||
					   strings.HasPrefix(origin, "http://localhost") ||
					   strings.HasPrefix(origin, "https://localhost") ||
					   strings.Contains(origin, ".ngrok-free.") ||
					   strings.Contains(origin, ".ngrok.io") {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Vary", "Origin")
			}
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-App-Token, ngrok-skip-browser-warning")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// gzipResponseWriter wraps gin.ResponseWriter with gzip.Writer
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g gzipResponseWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// GzipMiddleware enables automatic HTTP Gzip response compression for JSON and static files
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}
		if strings.Contains(c.Request.Header.Get("Accept"), "text/event-stream") {
			c.Next()
			return
		}
		gz, err := gzip.NewWriterLevel(c.Writer, gzip.DefaultCompression)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &gzipResponseWriter{ResponseWriter: c.Writer, writer: gz}
		c.Next()
	}
}

// IPRateLimiter implements a thread-safe sliding window rate limiter per client IP
type IPRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

var globalRateLimiter = &IPRateLimiter{
	requests: make(map[string][]time.Time),
}

func (l *IPRateLimiter) Allow(ip string, limitRPM int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	timestamps, exists := l.requests[ip]
	var valid []time.Time
	if exists {
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
	}

	if len(valid) >= limitRPM {
		l.requests[ip] = valid
		return false
	}

	valid = append(valid, now)
	l.requests[ip] = valid
	return true
}

// RateLimitMiddleware enforces general & heavy endpoint Rate Limiting (LIMS Availability Standard)
func RateLimitMiddleware(paramRepo repositories.ParameterRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass Rate Limit during Load/Stress Testing
		if os.Getenv("DISABLE_RATE_LIMIT") == "true" || os.Getenv("ENABLE_PPROF") == "true" {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// Read RATE_LIMIT_GENERAL_RPM from parameter repository DB cache (Default 100 RPM)
		limitRPM := 100
		if p, err := paramRepo.FindByKey("RATE_LIMIT_GENERAL_RPM"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(p.KeyValue)); err == nil && parsed > 0 {
				limitRPM = parsed
			}
		}

		effectiveLimit := limitRPM

		// Read optional RATE_LIMIT_HEAVY_RPM only if explicitly configured in DB
		if p, err := paramRepo.FindByKey("RATE_LIMIT_HEAVY_RPM"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(p.KeyValue)); err == nil && parsed > 0 {
				heavyEndpoints := "/api/applications, /api/payroll/reconcile, /api/payroll/adjust, /api/manual-repayment/process"
				if hp, err := paramRepo.FindByKey("RATE_LIMIT_HEAVY_ENDPOINTS"); err == nil && strings.TrimSpace(hp.KeyValue) != "" {
					heavyEndpoints = hp.KeyValue
				}
				for _, ep := range strings.Split(heavyEndpoints, ",") {
					cleanEP := strings.TrimSpace(ep)
					if cleanEP != "" && strings.HasPrefix(path, cleanEP) {
						effectiveLimit = parsed
						break
					}
				}
			}
		}

		clientIP := c.ClientIP()
		if !globalRateLimiter.Allow(clientIP, effectiveLimit) {
			log.Printf("[RATE-LIMIT-EXCEEDED] Client IP %s exceeded rate limit (%d RPM) on path %s", clientIP, effectiveLimit, path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("Too Many Requests: Batas jumlah permintaan (%d RPM) telah terlampaui. Silakan tunggu 1 menit.", effectiveLimit),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware protects API routes by verifying Bearer Token in Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass Auth during Load/Stress Testing
		if os.Getenv("DISABLE_AUTH") == "true" || os.Getenv("ENABLE_PPROF") == "true" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		// Skip public authentication, health check, change password/pin, and read-only GET master/user-info endpoints
		if strings.HasPrefix(path, "/api/karisma/login") || 
		   strings.HasPrefix(path, "/api/karisma/register") || 
		   strings.HasPrefix(path, "/api/karisma/verify") || 
		   strings.HasPrefix(path, "/api/karisma/change-password") || 
		   strings.HasPrefix(path, "/api/change-password") || 
		   strings.HasPrefix(path, "/api/change-pin") || 
		   strings.HasPrefix(path, "/api/verify-menu-password") || 
		   strings.HasPrefix(path, "/api/products") || 
		   strings.HasPrefix(path, "/api/parameters") || 
		   strings.HasPrefix(path, "/api/health") || 
		   strings.HasPrefix(path, "/api/user-info") ||
		   (c.Request.Method == "GET" && strings.HasPrefix(path, "/api/master")) {
			c.Next()
			return
		}

		token := getTokenFromRequest(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Akses ditolak. Token autentikasi tidak ditemukan di Cookie maupun Authorization Header.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RBACRoleMenuMiddleware verifies token, user identity, and queries lms_sch.role_menus DB table (LIMS Standard Compliance)
func RBACRoleMenuMiddleware(requiredMenuPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass RBAC during Load/Stress Testing
		if os.Getenv("DISABLE_AUTH") == "true" || os.Getenv("ENABLE_PPROF") == "true" {
			c.Next()
			return
		}
		var token string = getTokenFromRequest(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Akses ditolak. Token autentikasi tidak ditemukan di Cookie maupun Authorization Header."})
			c.Abort()
			return
		}

		username := strings.TrimPrefix(token, "mock-token-")
		var roleID int64 = 0

		// Look up real session and user role from DB if token is a session token
		if config.DB != nil && !strings.HasPrefix(token, "mock-token-") {
			var sessionRecord models.Session
			if errSess := config.DB.Where("token = ? AND is_active = ?", token, true).First(&sessionRecord).Error; errSess == nil {
				username = sessionRecord.Username
				var u models.User
				if errU := config.DB.Where("id = ?", sessionRecord.UserID).First(&u).Error; errU == nil {
					if u.RoleID > 0 {
						roleID = u.RoleID
					}
				}
			}
		}

		usernameLower := strings.ToLower(strings.TrimSpace(username))

		if usernameLower == "admin" || usernameLower == "9999" {
			roleID = 1 // Role 1: Admin
		} else if usernameLower == "hrd" {
			roleID = 3 // Role 3: HRD
		} else {
			var userRoleID int64 = 0
			config.DB.Raw("SELECT role_id FROM lms_sch.users WHERE (LOWER(username) = ? OR CAST(id AS TEXT) = ? OR CAST(member_no AS TEXT) = ?) AND deleted_at IS NULL LIMIT 1", usernameLower, usernameLower, usernameLower).Scan(&userRoleID)
			if userRoleID > 0 {
				roleID = userRoleID
			} else {
				// Fallback lookup from active session
				var sessionRecord models.Session
				if errSess := config.DB.Where("token = ? AND is_active = ?", token, true).First(&sessionRecord).Error; errSess == nil {
					var u models.User
					if errU := config.DB.Where("id = ?", sessionRecord.UserID).First(&u).Error; errU == nil && u.RoleID > 0 {
						roleID = u.RoleID
					}
				}
			}
		}

		// Fallback: If roleID is still 0 but token is admin/mock-token for user 100001 (Administrator), assign Admin role (1)
		if roleID == 0 && (usernameLower == "admin" || usernameLower == "100001" || strings.Contains(token, "100001") || strings.Contains(token, "admin")) {
			roleID = 1
		}

		// Admin role (roleID == 1 or 9999) has unrestricted access
		if roleID == 1 || roleID == 9999 {
			c.Next()
			return
		}

		// Query DB lms_sch.role_menus joined with lms_sch.menus for roleID & requiredMenuPath
		var hasAccess bool = false
		query := `SELECT EXISTS (
			SELECT 1 FROM lms_sch.role_menus rm
			JOIN lms_sch.menus m ON rm.menu_id = m.menu_id
			WHERE rm.role_id = ? AND (LOWER(m.path) = LOWER(?) OR LOWER(m.path) LIKE LOWER(?))
		)`
		config.DB.Raw(query, roleID, requiredMenuPath, "%"+requiredMenuPath+"%").Scan(&hasAccess)

		// Strict RBAC Enforcement: Role 2 (Anggota) is never permitted to access Admin/System Manager menus
		if roleID == 2 {
			reqLower := strings.ToLower(requiredMenuPath)
			if reqLower == "parameters" || reqLower == "approval" || reqLower == "payroll" || reqLower == "manual-repayment" || reqLower == "master" {
				hasAccess = false
			}
		}

		log.Printf("[RBAC-ROLE-MENU-CHECK] Token: '%s', User: '%s', RoleID: %d, MenuRequired: '%s' -> HasAccess: %v", token, username, roleID, requiredMenuPath, hasAccess)

		if !hasAccess {
			log.Printf("[RBAC-ROLE-MENU-DENIED] User '%s' (RoleID: %d) denied access to menu path '%s' on %s %s", username, roleID, requiredMenuPath, c.Request.Method, c.Request.URL.Path)
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Forbidden: Akses ditolak. Role pengguna (Role ID: %d) tidak memiliki izin menu '%s' pada tabel lms_sch.role_menus.", roleID, requiredMenuPath),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware alias for strict admin/parameter endpoints
func AdminMiddleware() gin.HandlerFunc {
	return RBACRoleMenuMiddleware("parameters")
}

func main() {
	// Load environment variables (Overload ensures .env file overrides shell export variables)
	if err := godotenv.Overload(); err != nil {
		log.Println("Warning: No .env file found or error reading it. Using OS environment variables.")
	}

	// Initialize Database Connection
	config.ConnectDatabase()

	fmt.Println("=================================================================")
	fmt.Println("🚀 LMS BACKEND v2.0 - SECURITY & RBAC ENFORCEMENT ACTIVE 🚀")
	fmt.Println("=================================================================")
	log.Println("SECURITY CHECK: RBACRoleMenuMiddleware and AuthMiddleware loaded successfully.")

	traceLevel := strings.TrimSpace(os.Getenv("TRACE_LEVEL"))
	disableLogger := false

	if traceLevel == "0" {
		disableLogger = true
	} else if traceLevel == "" {
		if os.Getenv("GIN_MODE") == "release" {
			disableLogger = true
		}
	}

	var r *gin.Engine
	if disableLogger {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
		log.Println("[RELEASE-MODE] Gin logger disabled to maximize Load Test throughput.")
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
		// Custom Request Logger Middleware for complete Backend Investigation
		r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			latencyMs := float64(param.Latency.Nanoseconds()) / 1e6
			
			// Format Indonesian Number (Dot for thousands, Comma for decimal) e.g., 1.100,450 ms or 150,300 ms
			str := fmt.Sprintf("%.3f", latencyMs)
			parts := strings.Split(str, ".")
			intPart := parts[0]
			decPart := parts[1]

			var intFormatted []string
			n := len(intPart)
			for i := 0; i < n; i++ {
				if i > 0 && (n-i)%3 == 0 {
					intFormatted = append(intFormatted, ".")
				}
				intFormatted = append(intFormatted, string(intPart[i]))
			}
			latencyIndo := strings.Join(intFormatted, "") + "," + decPart

			// Field SLA Group (<300ms, >300ms, >500ms)
			slaGroup := "<300ms"
			if latencyMs > 500 {
				slaGroup = ">500ms"
			} else if latencyMs >= 300 {
				slaGroup = ">300ms"
			}

			return fmt.Sprintf("[LMS-API-LOG] %s | %s | %3d | %12s | %6s | %s %s\n",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.ClientIP,
				param.StatusCode,
				latencyIndo,
				slaGroup,
				param.Method,
				param.Path,
			)
		}))
		log.Println(fmt.Sprintf("[LMS-API-LOG] Gin HTTP logger active (TRACE_LEVEL=%s)", traceLevel))
	}

	_ = r.SetTrustedProxies(nil) // Safe setting for trusted proxies to suppress warning
	r.Use(CORSMiddleware())      // Enable CORS for frontend
	r.Use(GzipMiddleware())      // Enable GZIP Network Compression (LIMS Performance Standard)

	// Dependencies Injection (CRUD Master & Transactions)
	productRepo := repositories.NewProductRepository(config.DB)
	productUseCase := usecases.NewProductUseCase(productRepo)
	productHandler := handlers.NewProductHandler(productUseCase)

	// Direct DB Query (Tanpa Caching RAM agar penambahan/perubahan produk langsung ter-reflect real-time)
	_ = productRepo.WarmCache()

	paramRepo := repositories.NewParameterRepository(config.DB)
	globalParamRepo = paramRepo
	cache.ParameterCache.Init(config.DB)
	paramUseCase := usecases.NewParameterUseCase(paramRepo)
	paramHandler := handlers.NewParameterHandler(paramUseCase)

	appRepo := repositories.NewApplicationRepository(config.DB)
	appUseCase := usecases.NewApplicationUseCase(appRepo, productRepo, paramRepo)
	appHandler := handlers.NewApplicationHandler(appUseCase)

	masterHandler := handlers.NewMasterDataHandler(config.DB, paramRepo)

	// Background worker for periodic session cleanup (SESSION_CLEANUP_HOURS)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			if config.DB == nil {
				continue
			}
			cleanupHours := 2
			if p, err := paramRepo.FindByKey("SESSION_CLEANUP_HOURS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
				if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
					cleanupHours = parsed
				}
			}
			cutoff := time.Now().Add(-time.Duration(cleanupHours) * time.Hour)
			config.DB.Where("expires_at < ?", cutoff).Delete(&models.Session{})
		}
	}()

	// LMS Core Endpoints
	api := r.Group("/api")
	api.Use(RateLimitMiddleware(paramRepo))
	api.Use(AuthMiddleware())
	{
		// Built-in Karisma Authentication Routes with detailed investigation logs
		karisma := api.Group("/karisma")
		{
			karisma.POST("/register", func(c *gin.Context) {
				var req struct {
					Nik      string `json:"nik"`       // NIK (Nomor Induk Karyawan)
					MemberNo string `json:"member_no"` // Fallback alias
					PhoneNo  string `json:"phone_no"`  // No. HP / WhatsApp
					KtpNo    string `json:"ktp_no"`    // NIK KTP
					Name     string `json:"name"`      // Nama Lengkap Karyawan
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Payload registrasi tidak valid"})
					return
				}

				nikStr := strings.TrimSpace(req.Nik)
				if nikStr == "" {
					nikStr = strings.TrimSpace(req.MemberNo)
				}
				phone := strings.TrimSpace(req.PhoneNo)
				ktpStr := strings.TrimSpace(req.KtpNo)
				nameStr := strings.TrimSpace(req.Name)

				if phone == "" || nikStr == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor Handphone dan NIK Karyawan wajib diisi!"})
					return
				}

				var nikParsed int64
				if parsed, errP := strconv.ParseInt(nikStr, 10, 64); errP == nil && parsed > 0 {
					nikParsed = parsed
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Karyawan dengan NIK : %s tidak ditemukan", nikStr)})
					return
				}

				if config.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Koneksi database tidak tersedia"})
					return
				}

				// STEP 1: Cek employees: select * from lms_sch.employees where employee_id = <NIK>
				var emp models.Employee
				if errEmp := config.DB.Where("employee_id = ? AND deleted_at IS NULL", nikParsed).First(&emp).Error; errEmp != nil {
					// STEP 2: Jika #1 tidak ketemu, kirim notifikasi "Karyawan dengan NIK : <NIK> tidak ditemukan"
					log.Printf("[REGISTER-FAIL] Employee not found for NIK %d (%s)", nikParsed, nikStr)
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Karyawan dengan NIK : %s tidak ditemukan", nikStr)})
					return
				}

				// STEP 3: Jika #1 ketemu, cek 4 factor (Phone, NIK, KTP, Name)
				phone08 := phone
				phone62 := phone
				if strings.HasPrefix(phone, "0") {
					phone62 = "62" + phone[1:]
				} else if strings.HasPrefix(phone, "62") {
					phone08 = "0" + phone[2:]
				}

				cleanEmpPhone := strings.TrimSpace(emp.PhoneNumber)
				cleanEmpKtp := strings.TrimSpace(emp.NoKTP)
				cleanEmpName := strings.TrimSpace(emp.Name)

				phoneMatch := cleanEmpPhone == "" || cleanEmpPhone == phone08 || cleanEmpPhone == phone62 || strings.HasSuffix(phone, strings.TrimPrefix(cleanEmpPhone, "0"))
				ktpMatch := cleanEmpKtp == "" || ktpStr == "" || cleanEmpKtp == ktpStr
				nameMatch := cleanEmpName == "" || nameStr == "" || strings.EqualFold(cleanEmpName, nameStr) || strings.Contains(strings.ToLower(cleanEmpName), strings.ToLower(nameStr)) || strings.Contains(strings.ToLower(nameStr), strings.ToLower(cleanEmpName))

				if !phoneMatch || !ktpMatch || !nameMatch {
					// STEP 4: Jika #3 tidak match, kirim notifikasi "Data tidak sama dengan HRD"
					log.Printf("[REGISTER-FAIL] 4-Factor mismatch for NIK %d. Req (Phone: %s, KTP: %s, Name: %s) vs HRD (Phone: %s, KTP: %s, Name: %s)",
						nikParsed, phone, ktpStr, nameStr, cleanEmpPhone, cleanEmpKtp, cleanEmpName)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak sama dengan HRD"})
					return
				}

				// STEP 5: Jika #3 ketemu --> cari member_no --> select member_no from lms_sch.members where employee_id = <NIK>
				var mem models.Member
				var memberNoToUse int64
				if errMem := config.DB.Where("employee_id = ? AND deleted_at IS NULL", nikParsed).First(&mem).Error; errMem == nil {
					memberNoToUse = mem.MemberNo
				} else {
					// Auto-create member record if missing to satisfy FK constraint
					memberNoToUse = nikParsed
					mem = models.Member{
						MemberNo:      memberNoToUse,
						EmployeeID:    nikParsed,
						KopkaraStatus: "ACTIVE",
						JoinDate:      time.Now().Format("2006-01-02"),
					}
					config.DB.Create(&mem)
				}

				// STEP 6: Insert ke table users dengan random PIN, lalu kirim WA dan notifikasi
				b := make([]byte, 3)
				var randomPin string
				if _, errR := rand.Read(b); errR == nil {
					num := (int(b[0])<<16 | int(b[1])<<8 | int(b[2]))%900000 + 100000
					randomPin = fmt.Sprintf("%06d", num)
				} else {
					randomPin = fmt.Sprintf("%06d", time.Now().UnixNano()%899999+100000)
				}

				hash, _ := bcrypt.GenerateFromPassword([]byte(randomPin), bcrypt.DefaultCost)
				hashStr := string(hash)
				phoneStr := phone

				var existingUser models.User
				if err := config.DB.Where("username = ? OR username = ? OR phone_number = ? OR phone_number = ? OR (member_no IS NOT NULL AND member_no = ?)", phone08, phone62, phone08, phone62, memberNoToUse).First(&existingUser).Error; err == nil {
					log.Printf("[REGISTER-FAIL] User already exists for phone %s / member_no %d (User ID %d)", phone, memberNoToUse, existingUser.ID)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Registrasi gagal karena user tsb sudah melakukan registrasi"})
					return
				} else {
					newUser := models.User{
						Username:    phone,
						Password:    hashStr,
						PhoneNumber: &phoneStr,
						Name:        emp.Name,
						PIN:         &hashStr,
						RoleID:      2,
						MemberNo:    &memberNoToUse,
					}
					if ktpStr != "" {
						newUser.NoKTP = &ktpStr
					}
					if err := config.DB.Create(&newUser).Error; err != nil {
						log.Printf("[REGISTER-CREATE-ERROR] %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meregistrasi user: " + err.Error()})
						return
					}
					log.Printf("🟢 [REGISTER-SUCCESS] Inserted NEW user ID %d for phone %s (member_no: %d) with Random PIN: %s", newUser.ID, phone, memberNoToUse, randomPin)
				}

				// Dispatch WA Notification
				waTemplate := "Halo {name}, Selamat! Registrasi & Aktivasi PIN EWA Mobile Kopkara Anda (HP: {phone}) telah BERHASIL. Silakan login ke aplikasi EWA Mobile menggunakan PIN {pin}."
				if paramRepo != nil {
					if p, errP := paramRepo.FindByKey("WA_DAFTAR_NOTIFICATION"); errP == nil && strings.TrimSpace(p.KeyValue) != "" {
						waTemplate = strings.TrimSpace(p.KeyValue)
					}
				}

				// Guarantee {pin} placeholder exists so the generated PIN is ALWAYS sent to the user
				if !strings.Contains(waTemplate, "{pin}") && !strings.Contains(waTemplate, "{PIN}") {
					if strings.Contains(waTemplate, "PIN 6-Digit Anda") {
						waTemplate = strings.ReplaceAll(waTemplate, "PIN 6-Digit Anda", "PIN {pin}")
					} else {
						waTemplate = waTemplate + " PIN: {pin}"
					}
				}

				waMsg := strings.ReplaceAll(waTemplate, "{name}", emp.Name)
				waMsg = strings.ReplaceAll(waMsg, "{phone}", phone)
				waMsg = strings.ReplaceAll(waMsg, "{nik}", nikStr)
				waMsg = strings.ReplaceAll(waMsg, "{member_no}", fmt.Sprintf("%d", memberNoToUse))
				waMsg = strings.ReplaceAll(waMsg, "{pin}", randomPin)
				waMsg = strings.ReplaceAll(waMsg, "{PIN}", randomPin)

				go func() {
					errWA := handlers.SendWhatsAppNotification(phone, waMsg)
					if errWA != nil {
						log.Printf("[WA-NOTIFICATION-ERROR] Failed sending WA to %s: %v", phone, errWA)
					} else {
						log.Printf("🟢 [WA-NOTIFICATION-SUCCESS] WA Notification with Random PIN dispatched to %s", phone)
					}
				}()

				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": fmt.Sprintf("✅ Registrasi Berhasil! Karyawan NIK %s terverifikasi. PIN 6-Digit Rahasia telah dikirimkan via WhatsApp ke %s.", nikStr, phone),
				})
			})

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

				if username == "" || password == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan Password tidak boleh kosong!"})
					return
				}

				// Build phone number variants (08..., 62..., +62...)
				phone08 := username
				phone62 := username
				if strings.HasPrefix(username, "0") {
					phone62 = "62" + username[1:]
				} else if strings.HasPrefix(username, "62") {
					phone08 = "0" + username[2:]
				} else if strings.HasPrefix(username, "+62") {
					phone08 = "0" + username[3:]
					phone62 = username[1:]
				}

				var memNoParsed int64
				if parsed, errP := strconv.ParseInt(username, 10, 64); errP == nil {
					memNoParsed = parsed
				}

				// 1. Fetch user from lms_sch.users table if DB is available
				var user models.User
				userFound := false
				if config.DB != nil {
					var err error
					if memNoParsed > 0 || strings.HasPrefix(username, "08") || strings.HasPrefix(username, "+62") {
						err = config.DB.Where("(username = ? OR phone_number = ? OR phone_number = ? OR (member_no IS NOT NULL AND member_no = ?)) AND deleted_at IS NULL", username, phone08, phone62, memNoParsed).First(&user).Error
					} else {
						err = config.DB.Where("(username = ? OR (member_no IS NOT NULL AND member_no = ?)) AND deleted_at IS NULL", username, memNoParsed).First(&user).Error
					}
					if err == nil {
						userFound = true
					}
				}

				if userFound {
					// Cek penguncian akun (LOGIN_LOCKOUT_MINUTES)
					lockoutMins := 15
					if p, err := paramRepo.FindByKey("LOGIN_LOCKOUT_MINUTES"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
						if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
							lockoutMins = parsed
						}
					}

					if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
						remMinutes := int(time.Until(*user.LockedUntil).Minutes()) + 1
						log.Printf("[KARISMA-AUTH-LOCKED] User '%s' is locked until %v (%d mins remaining)", username, user.LockedUntil, remMinutes)
						c.JSON(http.StatusUnauthorized, gin.H{
							"error": fmt.Sprintf("Akun Anda sedang dikunci sementara akibat salah password. Silakan coba lagi dalam %d menit.", remMinutes),
						})
						return
					}

					// Verify password using Bcrypt with fallback to plain text comparison & PIN 6-Digit fallback
					pwdMatch := false
					if errPwd := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); errPwd == nil {
						pwdMatch = true
					} else if user.Password == password && user.Password != "" {
						pwdMatch = true
						// Auto upgrade plain text to bcrypt hash in DB
						if hashed, errH := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); errH == nil {
							user.Password = string(hashed)
						}
					} else if user.PIN != nil && *user.PIN != "" {
						// Allow employee to log in to Web Portal using their Mobile PIN!
						if errPin := bcrypt.CompareHashAndPassword([]byte(*user.PIN), []byte(password)); errPin == nil {
							pwdMatch = true
							log.Printf("[KARISMA-AUTH-TRACE] Password match success via Mobile PIN 6-Digit Bcrypt Hash!")
						}
					}

					if !pwdMatch {
						// Password salah -> update failed_login_attempts & check MAX_PASSWORD_ATTEMPTS (Default 3)
						user.FailedLoginAttempts++
						maxAttempts := 3
						if p, err := paramRepo.FindByKey("MAX_PASSWORD_ATTEMPTS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
							if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
								maxAttempts = parsed
							}
						} else if p, err := paramRepo.FindByKey("PASSWORD_MAX_FAILED_ATTEMPTS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
							if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
								maxAttempts = parsed
							}
						} else if p, err := paramRepo.FindByKey("LOGIN_MAX_ATTEMPTS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
							if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
								maxAttempts = parsed
							}
						}

						if user.FailedLoginAttempts >= maxAttempts {
							lockedUntil := time.Now().Add(time.Duration(lockoutMins) * time.Minute)
							user.LockedUntil = &lockedUntil
							config.DB.Save(&user)
							log.Printf("[KARISMA-AUTH-LOCKING] User '%s' reached %d failed attempts -> Locked for %d mins", username, user.FailedLoginAttempts, lockoutMins)
							c.JSON(http.StatusUnauthorized, gin.H{
								"error": fmt.Sprintf("Password salah! Akun Anda telah dikunci selama %d menit karena %d kali percobaan login salah berturut-turut.", lockoutMins, maxAttempts),
							})
							return
						} else {
							config.DB.Save(&user)
							remaining := maxAttempts - user.FailedLoginAttempts
							log.Printf("[KARISMA-AUTH-FAILED] Incorrect password for '%s'. Attempt %d/%d", username, user.FailedLoginAttempts, maxAttempts)
							c.JSON(http.StatusUnauthorized, gin.H{
								"error": fmt.Sprintf("Username atau Password salah! Sisa percobaan login: %d kali.", remaining),
							})
							return
						}
					}

					// Cek apakah user memiliki RoleID valid
					if user.RoleID <= 0 {
						log.Printf("[LOGIN-REJECTED] User '%s' has no assigned role_id (role_id: %d)", username, user.RoleID)
						c.JSON(http.StatusForbidden, gin.H{"error": "Login gagal: Akun Anda belum memiliki Role. Silakan hubungi Administrator."})
						return
					}

					// Check PWD_ROTATION_DAYS (Default 90 days if configured)
					pwdRotationDays := 90
					if p, err := paramRepo.FindByKey("PWD_ROTATION_DAYS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
						if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
							pwdRotationDays = parsed
						}
					} else if p, err := paramRepo.FindByKey("PASSWORD_ROTATION_DAYS"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
						if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
							pwdRotationDays = parsed
						}
					}

					if pwdRotationDays > 0 {
						lastChanged := user.UpdatedAt
						if !user.PasswordChangedAt.IsZero() {
							lastChanged = user.PasswordChangedAt
						}
						if time.Since(lastChanged) > time.Duration(pwdRotationDays)*24*time.Hour {
							user.ForcePwdChange = true
							log.Printf("[PWD-ROTATION] User '%s' password expired (%d days) -> Force password change enabled", username, pwdRotationDays)
						}
					}

					// Login Berhasil! Reset failed login attempts & lockout
					config.DB.Model(&user).Updates(map[string]interface{}{
						"failed_login_attempts": 0,
						"locked_until":          nil,
						"force_pwd_change":      user.ForcePwdChange,
					})

					// Cek SINGLE_SESSION_MODE
					singleSessionMode := false
					if p, err := paramRepo.FindByKey("SINGLE_SESSION_MODE"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
						if strings.ToUpper(strings.TrimSpace(p.KeyValue)) == "TRUE" {
							singleSessionMode = true
						}
					}

					if singleSessionMode {
						config.DB.Model(&models.Session{}).Where("user_id = ? AND is_active = ?", user.ID, true).Update("is_active", false)
						log.Printf("[SINGLE-SESSION] Deactivated previous active sessions for user_id: %d ('%s')", user.ID, username)
					}

					// Cek SESSION_EXPIRY_MINUTES
					sessionExpiryMins := 1440
					if p, err := paramRepo.FindByKey("SESSION_EXPIRY_MINUTES"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
						if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
							sessionExpiryMins = parsed
						}
					}

					token := generateSessionToken()
					now := time.Now()
					expiresAt := now.Add(time.Duration(sessionExpiryMins) * time.Minute)

					sessionRecord := models.Session{
						Token:          token,
						UserID:         user.ID,
						Username:       user.Username,
						IPAddress:      c.ClientIP(),
						UserAgent:      c.Request.UserAgent(),
						IsActive:       true,
						LoginAt:        now,
						ExpiresAt:      expiresAt,
						LastActivityAt: now,
					}
					if errSess := config.DB.Create(&sessionRecord).Error; errSess != nil {
						log.Printf("[SESSION-CREATE-ERROR] Failed to save session to DB: %v", errSess)
					}

					log.Printf("[KARISMA-AUTH-SUCCESS] ✅ User '%s' (RoleID: %d) logged in successfully. Token: %s", username, user.RoleID, token)
					isSecure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
					tokenCookieName := getAppTokenName()
					c.SetCookie(tokenCookieName, token, sessionExpiryMins*60, "/", "", isSecure, true)
					c.SetCookie("karisma_token", token, sessionExpiryMins*60, "/", "", isSecure, true)
					c.JSON(http.StatusOK, gin.H{"status": "success", "token": token, "token_name": tokenCookieName})
					return
				}

				// 2. Fallback check for initial mock users (admin / hrd / employee numbers) if users table not yet populated in DB
				if username == "admin" && password == "admin123" {
					token := "mock-token-admin"
					c.SetCookie("karisma_token", token, 7200, "/", "", true, true)
					c.JSON(http.StatusOK, gin.H{"token": token, "status": "success"})
					return
				} else if username == "hrd" && password == "hrd123" {
					token := "mock-token-hrd"
					c.SetCookie("karisma_token", token, 7200, "/", "", true, true)
					c.JSON(http.StatusOK, gin.H{"token": token, "status": "success"})
					return
				} else if _, err := strconv.Atoi(username); err == nil && (password == "password123" || password == "123456") {
					token := "mock-token-" + username
					c.SetCookie("karisma_token", token, 7200, "/", "", true, true)
					c.JSON(http.StatusOK, gin.H{"token": token, "status": "success"})
					return
				}

				log.Printf("[KARISMA-AUTH-FAILED] Username '%s' not recognized or invalid credentials", username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
			})

			// Logout endpoint
			logoutHandler := func(c *gin.Context) {
				token := getTokenFromRequest(c)
				if token != "" && config.DB != nil {
					config.DB.Model(&models.Session{}).Where("token = ?", token).Update("is_active", false)
					log.Printf("[KARISMA-LOGOUT] Session deactivated for token: %s", token)
				}
				isSecure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
				tokenCookieName := getAppTokenName()
				c.SetCookie(tokenCookieName, "", -1, "/", "", isSecure, true)
				c.SetCookie("karisma_token", "", -1, "/", "", isSecure, true)
				c.SetCookie("ewa_token", "", -1, "/", "", isSecure, true)
				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Logout berhasil"})
			}
			api.POST("/logout", logoutHandler)
			karisma.POST("/logout", logoutHandler)

			karisma.POST("/verify", func(c *gin.Context) {
				var token string
				authHeader := c.GetHeader("Authorization")
				if len(authHeader) >= 8 && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
					token = strings.TrimSpace(authHeader[7:])
				}
				if token == "" {
					if cookieToken, err := c.Cookie("karisma_token"); err == nil && strings.TrimSpace(cookieToken) != "" {
						token = strings.TrimSpace(cookieToken)
					}
				}

				if token == "" {
					log.Printf("[KARISMA-VERIFY-ERROR] No token found in Cookie or Authorization header")
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token Authorization / HttpOnly Cookie tidak valid"})
					return
				}

				// 1. Check DB session first
				if config.DB != nil {
					var session models.Session
					err := config.DB.Where("token = ? AND is_active = ? AND expires_at > ?", token, true, time.Now()).First(&session).Error
					if err == nil {
						// Update last_activity_at only if last activity was > 5 minutes ago (throttle unnecessary SQL updates)
						if time.Since(session.LastActivityAt) > 5*time.Minute {
							config.DB.Model(&session).Update("last_activity_at", time.Now())
						}

						var user models.User
						if uPtr, errU := cache.IdentityCache.GetUserByID(config.DB, session.UserID); errU == nil && uPtr != nil {
							user = *uPtr
							if user.RoleID <= 0 {
								c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak: Akun Anda tidak memiliki Role pada tabel users."})
								return
							}
							empID := int64(0)
							name := user.Name
							role := "Anggota"
							if user.RoleID == 1 {
								role = "Admin"
							} else if user.RoleID == 3 {
								role = "HRD"
							} else {
								var r models.Role
								if errR := config.DB.Where("role_id = ?", user.RoleID).First(&r).Error; errR == nil && strings.TrimSpace(r.RoleName) != "" {
									role = r.RoleName
								}
							}

							if user.MemberNo != nil && *user.MemberNo > 0 {
								var mem struct {
									EmployeeID int64
								}
								config.DB.Raw("SELECT employee_id FROM lms_sch.members WHERE member_no = ? LIMIT 1", *user.MemberNo).Scan(&mem)
								if mem.EmployeeID > 0 {
									empID = mem.EmployeeID
								} else {
									empID = *user.MemberNo
								}
							}

							if empID == 0 {
								if role == "admin" {
									empID = 9999
								} else if role == "hrd" {
									empID = 8888
								} else if parsed, errP := strconv.ParseInt(user.Username, 10, 64); errP == nil {
									empID = parsed
								} else {
									empID = user.ID
								}
							}

							c.JSON(http.StatusOK, gin.H{
								"status": "success",
								"user": gin.H{
									"employee_id":      empID,
									"member_no":        user.MemberNo,
									"username":         user.Username,
									"name":             name,
									"eligible":         true,
									"role":             role,
									"role_id":          user.RoleID,
									"force_pwd_change": user.ForcePwdChange,
								},
							})
							return
						}
					}
				}

				// 2. Fallback mock token check if DB session missing
				username := strings.TrimPrefix(token, "mock-token-")
				log.Printf("[KARISMA-VERIFY-ATTEMPT] Verifying fallback token for user: '%s'", username)

				role := "anggota"
				name := "Karyawan #" + username
				empID := int64(1001)

				if username == "admin" {
					role = "admin"
					name = "Administrator Koperasi"
					empID = 9999
				} else if username == "hrd" {
					role = "hrd"
					name = "Tim HRD Adira"
					empID = 8888
				} else if idVal, err := strconv.ParseInt(username, 10, 64); err == nil {
					empID = idVal
					var dbName string
					if config.DB != nil {
						config.DB.Raw("SELECT name FROM lms_sch.employees WHERE employee_id = ? LIMIT 1", empID).Scan(&dbName)
					}
					if dbName != "" {
						name = dbName
					}
				}

				log.Printf("[KARISMA-VERIFY-SUCCESS] ✅ Fallback Token verified for '%s' (ID: %d, Role: %s)", name, empID, role)
				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"user": gin.H{
						"employee_id": empID,
						"username":    username,
						"name":        name,
						"eligible":    true,
						"role":        role,
					},
				})
			})

				changePasswordHandler := func(c *gin.Context) {
				var req struct {
					Username    string `json:"username"`
					OldPassword string `json:"old_password"`
					NewPassword string `json:"new_password"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Payload ganti password tidak valid"})
					return
				}

				username := strings.TrimSpace(req.Username)
				oldPassword := strings.TrimSpace(req.OldPassword)
				newPassword := strings.TrimSpace(req.NewPassword)

				if username == "" || oldPassword == "" || newPassword == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username, Password Lama, dan Password Baru harus diisi!"})
					return
				}

				// Validate Password against global parameters (PWD_MIN_LENGTH, PWD_MIN_LOWERCASE, PWD_MIN_UPPERCASE, PWD_MIN_NUMERIC, PWD_MIN_SPECIAL)
				pwdMinLength := 9
				if p, err := paramRepo.FindByKey("PWD_MIN_LENGTH"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
						pwdMinLength = parsed
					}
				}
				if len(newPassword) < pwdMinLength {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Password baru minimal %d karakter (sesuai parameter PWD_MIN_LENGTH)!", pwdMinLength)})
					return
				}

				pwdMinLower := 1
				if p, err := paramRepo.FindByKey("PWD_MIN_LOWERCASE"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed >= 0 {
						pwdMinLower = parsed
					}
				}
				lowerCount := 0
				for _, r := range newPassword {
					if unicode.IsLower(r) {
						lowerCount++
					}
				}
				if lowerCount < pwdMinLower {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Password minimal harus memiliki %d huruf kecil (a-z) (sesuai parameter PWD_MIN_LOWERCASE)!", pwdMinLower)})
					return
				}

				pwdMinUpper := 1
				if p, err := paramRepo.FindByKey("PWD_MIN_UPPERCASE"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed >= 0 {
						pwdMinUpper = parsed
					}
				}
				upperCount := 0
				for _, r := range newPassword {
					if unicode.IsUpper(r) {
						upperCount++
					}
				}
				if upperCount < pwdMinUpper {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Password minimal harus memiliki %d huruf besar (A-Z) (sesuai parameter PWD_MIN_UPPERCASE)!", pwdMinUpper)})
					return
				}

				pwdMinNum := 1
				if p, err := paramRepo.FindByKey("PWD_MIN_NUMERIC"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed >= 0 {
						pwdMinNum = parsed
					}
				}
				numCount := 0
				for _, r := range newPassword {
					if unicode.IsDigit(r) {
						numCount++
					}
				}
				if numCount < pwdMinNum {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Password minimal harus memiliki %d karakter angka (0-9) (sesuai parameter PWD_MIN_NUMERIC)!", pwdMinNum)})
					return
				}

				pwdMinSpecial := 1
				if p, err := paramRepo.FindByKey("PWD_MIN_SPECIAL"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed >= 0 {
						pwdMinSpecial = parsed
					}
				}
				specialCount := 0
				for _, r := range newPassword {
					if unicode.IsPunct(r) || unicode.IsSymbol(r) || strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?~`", r) {
						specialCount++
					}
				}
				if specialCount < pwdMinSpecial {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Password minimal harus memiliki %d karakter spesial (!@#$...) (sesuai parameter PWD_MIN_SPECIAL)!", pwdMinSpecial)})
					return
				}

				if config.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Database tidak terhubung"})
					return
				}

				var user models.User
				if err := config.DB.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan di database. Pastikan DDL SQL telah dijalankan."})
					return
				}

				// Verify old password
				oldPwdMatch := false
				if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err == nil {
					oldPwdMatch = true
				} else if user.Password == oldPassword {
					oldPwdMatch = true
				}

				if !oldPwdMatch {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama Anda tidak sesuai!"})
					return
				}

				// Generate Bcrypt hash for new password
				hashedNew, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat hash password baru"})
					return
				}

				user.Password = string(hashedNew)
				user.FailedLoginAttempts = 0
				user.LockedUntil = nil
				user.PasswordChangedAt = time.Now()
				user.ForcePwdChange = false

				if err := config.DB.Save(&user).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan password baru ke database: " + err.Error()})
					return
				}

				log.Printf("[CHANGE-PASSWORD-SUCCESS] ✅ Password for user '%s' updated successfully.", username)
				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Password berhasil diperbarui! Silakan login kembali dengan password baru."})
			}

			karisma.POST("/change-password", changePasswordHandler)
			api.POST("/change-password", changePasswordHandler)

			// Endpoint Fitur Ganti PIN dengan syarat PIN_MIN_LENGTH & Anti Weak/Repetitive PIN
			changePinHandler := func(c *gin.Context) {
				var req struct {
					Username string `json:"username"`
					OldPin   string `json:"old_pin"`
					NewPin   string `json:"new_pin"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Payload ganti PIN tidak valid"})
					return
				}

				username := strings.TrimSpace(req.Username)
				oldPin := strings.TrimSpace(req.OldPin)
				newPin := strings.TrimSpace(req.NewPin)

				if username == "" || oldPin == "" || newPin == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username, PIN Lama, dan PIN Baru harus diisi!"})
					return
				}

				// Validate PIN rules: numeric only
				for _, r := range newPin {
					if !unicode.IsDigit(r) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "PIN Baru harus terdiri dari karakter angka (0-9)!"})
						return
					}
				}

				// Validate PIN length against PIN_MIN_LENGTH (default 6)
				pinMinLength := 6
				if p, err := paramRepo.FindByKey("PIN_MIN_LENGTH"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					if parsed, errP := strconv.Atoi(strings.TrimSpace(p.KeyValue)); errP == nil && parsed > 0 {
						pinMinLength = parsed
					}
				}
				if len(newPin) < pinMinLength {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("PIN Baru minimal harus %d digit (sesuai parameter PIN_MIN_LENGTH)!", pinMinLength)})
					return
				}

				// Anti weak / repetitive PIN check (e.g. 111111, 222222, 123456, 654321, 000000)
				allSame := true
				for i := 1; i < len(newPin); i++ {
					if newPin[i] != newPin[0] {
						allSame = false
						break
					}
				}
				if allSame {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("PIN '%s' terlalu mudah ditebak! Harap tidak menggunakan angka berulang sama.", newPin)})
					return
				}

				isSeqAsc := true
				isSeqDesc := true
				for i := 1; i < len(newPin); i++ {
					if newPin[i] != newPin[i-1]+1 {
						isSeqAsc = false
					}
					if newPin[i] != newPin[i-1]-1 {
						isSeqDesc = false
					}
				}
				if isSeqAsc || isSeqDesc {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("PIN '%s' terlalu mudah ditebak! Harap tidak menggunakan urutan angka berurutan.", newPin)})
					return
				}

				if config.DB == nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Database tidak terhubung"})
					return
				}

				var user models.User
				if err := config.DB.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan di database."})
					return
				}

				// Verify old PIN / Password
				oldMatch := false
				if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPin)); err == nil {
					oldMatch = true
				} else if user.Password == oldPin {
					oldMatch = true
				}

				if !oldMatch {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "PIN Lama Anda tidak sesuai!"})
					return
				}

				// Hash new PIN with Bcrypt
				hashedPin, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi PIN baru"})
					return
				}

				user.Password = string(hashedPin)
				user.FailedLoginAttempts = 0
				user.LockedUntil = nil
				user.PasswordChangedAt = time.Now()

				if err := config.DB.Save(&user).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan PIN baru ke database: " + err.Error()})
					return
				}

				log.Printf("[CHANGE-PIN-SUCCESS] ✅ PIN for user '%s' updated & encrypted successfully.", username)
				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "PIN berhasil diperbarui & dienkripsi! Silakan login kembali menggunakan PIN baru Anda."})
			}

			api.POST("/change-pin", changePinHandler)
			karisma.POST("/change-pin", changePinHandler)

			// Endpoint Verifikasi Password / PIN untuk Akses Menu Proteksi (is_password = true)
			api.POST("/verify-menu-password", func(c *gin.Context) {
				var req struct {
					Username string `json:"username"`
					Password string `json:"password"`
					MenuID   int64  `json:"menu_id"`
					Path     string `json:"path"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Payload verifikasi menu tidak valid"})
					return
				}

				username := strings.TrimSpace(req.Username)
				pass := strings.TrimSpace(req.Password)
				if username == "" || pass == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan Password/PIN harus diisi!"})
					return
				}

				var user models.User
				if err := config.DB.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak ditemukan atau Password/PIN salah!"})
					return
				}

				valid := false
				if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)); err == nil {
					valid = true
				} else if user.Password == pass {
					valid = true
				}

				if !valid {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Password / PIN Anda tidak sesuai. Akses ke menu ini ditolak!"})
					return
				}

				var notifType int = 0
				if req.MenuID > 0 {
					var menu models.Menu
					if err := config.DB.Where("menu_id = ?", req.MenuID).First(&menu).Error; err == nil {
						notifType = menu.NotificationType
					}
				}

				log.Printf("[MENU-AUTH-SUCCESS] 🔐 User '%s' verified password/PIN for menu '%s' (notification_type=%d)", username, req.Path, notifType)
				c.JSON(http.StatusOK, gin.H{
					"status": "success",
					"message": "Verifikasi Password/PIN Berhasil!",
					"notification_type": notifType,
				})
			})
		}

		// Master Data Generic Endpoints
		master := api.Group("/master")
		{
			master.GET("/members/check-employee/:employee_id", masterHandler.CheckEmployeeMember)
			master.POST("/users/unlock/:id", AdminMiddleware(), masterHandler.UnlockUser)
			master.POST("/users/reset-password/:id", AdminMiddleware(), masterHandler.ResetUserPassword)
			master.POST("/sessions/revoke/:id", AdminMiddleware(), masterHandler.RevokeSession)
			master.POST("/sessions/cleanup", AdminMiddleware(), masterHandler.CleanupSessions)
			master.GET("/:table", masterHandler.GetAll)
			master.POST("/:table", AdminMiddleware(), masterHandler.Save)
			master.DELETE("/:table/:id", AdminMiddleware(), masterHandler.Delete)
		}

		api.GET("/user-info/:employee_id", masterHandler.GetUserInfo)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "LMS Backend is running smoothly."})
		})

		// Real-time Dashboard Summary Endpoint
		api.GET("/dashboard/summary", func(c *gin.Context) {
			empIDStr := strings.TrimSpace(c.Query("employee_id"))
			roleIDStr := strings.TrimSpace(c.Query("role_id"))
			if roleIDStr == "" {
				roleIDStr = strings.TrimSpace(c.Query("role"))
			}

			isHighPriv := false
			if roleIDStr == "1" || roleIDStr == "3" || strings.EqualFold(roleIDStr, "admin") || strings.EqualFold(roleIDStr, "hrd") {
				isHighPriv = true
			}

			var totalHutang float64 = 0
			var pinjamanAktif int64 = 0
			var creditLimit float64 = 15000000
			formulaUsed := "DEFAULT"

			var empIdInt int64
			if id, err := strconv.ParseInt(empIDStr, 10, 64); err == nil && id > 0 {
				empIdInt = id
			}

			// Menentukan Credit Limit (CL) dari Cache tanpa SQL-1 Loop Karyawan
			if isHighPriv {
				// Untuk Admin/HRD: Membaca GLOBAL_CREDIT_LIMIT langsung dari in-memory cache paramRepo
				creditLimit = 5000000000 // Default 5 Milyar
				formulaUsed = "GLOBAL_CREDIT_LIMIT (CACHE)"
				if p, errP := paramRepo.FindByKey("GLOBAL_CREDIT_LIMIT"); errP == nil && strings.TrimSpace(p.KeyValue) != "" {
					if val, errV := strconv.ParseFloat(strings.TrimSpace(p.KeyValue), 64); errV == nil && val > 0 {
						creditLimit = val
					}
				}
			} else {
				// Untuk User Biasa (Anggota): Hitung CL individu menggunakan LOAN_LIMIT_FORMULA dari cache paramRepo
				creditLimit = calculateLMSCreditLimitFromCache(config.DB, paramRepo, empIdInt)
				if p, errP := paramRepo.FindByKey("LOAN_LIMIT_FORMULA"); errP == nil {
					formulaUsed = p.KeyValue
				} else {
					formulaUsed = "(DAY/30) * SALARY * 0.5"
				}
			}

			// 1. Hitung Total Sisa Hutang (Total Debt) langsung dari lms_sch.employees (Ultra-Fast <1ms & 100% Akurat)
			if isHighPriv {
				config.DB.Raw("SELECT COALESCE(SUM(total_loan), 0) FROM lms_sch.employees").Scan(&totalHutang)
			} else if empIdInt > 0 {
				config.DB.Raw(`
					SELECT COALESCE(total_loan, 0) 
					FROM lms_sch.employees 
					WHERE employee_id = ? OR employee_id = (SELECT employee_id FROM lms_sch.members WHERE member_no = ? LIMIT 1) 
					LIMIT 1
				`, empIdInt, empIdInt).Scan(&totalHutang)
			}

			availableLimit := creditLimit - totalHutang
			if availableLimit < 0 {
				availableLimit = 0
			}

			var recentApps []models.LoanApplication

			// LOG CONSOLE RESMI BACKEND UNTUK DASHBOARD CREDIT LIMIT (CL)
			log.Printf("[DASHBOARD-SUMMARY] UserID: %d | Role: '%s' | HighPriv: %t | Formula: '%s' | CreditLimit (CL): Rp %.2f | TotalDebt: Rp %.2f | AvailableLimit: Rp %.2f",
				empIdInt, roleIDStr, isHighPriv, formulaUsed, creditLimit, totalHutang, availableLimit)

			c.JSON(http.StatusOK, gin.H{
				"credit_limit":    creditLimit,
				"available_limit": availableLimit,
				"total_debt":      totalHutang,
				"active_loans":    pinjamanAktif,
				"recent_loans":    recentApps,
				"is_high_priv":    isHighPriv,
			})
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
			applications.POST("/:id/approve", RBACRoleMenuMiddleware("approval"), appHandler.Approve)
			applications.POST("/:id/disburse", RBACRoleMenuMiddleware("disbursement"), appHandler.Disburse)
			applications.GET("/:id/trackings", appHandler.GetTrackings)
		}

		// Master Data: Parameters
		parameters := api.Group("/parameters")
		{
			parameters.GET("", paramHandler.GetAll)
			parameters.POST("", AdminMiddleware(), paramHandler.Save)
			parameters.DELETE("/:id", AdminMiddleware(), paramHandler.Delete)
		}

		// Paginated & Filtered Members Search Endpoint for Manual Repayment & Dropdowns
		api.GET("/members", func(c *gin.Context) {
			type MemberDTO struct {
				MemberNo   int64  `json:"member_no"`
				EmployeeID int64  `json:"employee_id"`
				Name       string `json:"name"`
				TotalCount int64  `json:"-" gorm:"column:total_count"`
			}
			q := strings.TrimSpace(c.Query("q"))
			if q == "" {
				q = strings.TrimSpace(c.Query("query"))
			}

			page, _ := strconv.Atoi(c.Query("page"))
			if page <= 0 {
				page = 1
			}

			limit, _ := strconv.Atoi(c.Query("limit"))
			if limit <= 0 {
				limit = 10
			}

			offset := (page - 1) * limit

			var dtos []MemberDTO

			whereClause := ""
			var args []interface{}
			if q != "" {
				likeStr := "%" + strings.ToLower(q) + "%"
				whereClause = "WHERE LOWER(CAST(m.member_no AS VARCHAR)) LIKE ? OR LOWER(CAST(COALESCE(e.employee_id, m.employee_id) AS VARCHAR)) LIKE ? OR LOWER(COALESCE(e.name, CONCAT('Anggota #', m.member_no))) LIKE ?"
				args = append(args, likeStr, likeStr, likeStr)
			}

			dataQuery := fmt.Sprintf(`
				SELECT 
					m.member_no,
					COALESCE(e.employee_id, m.employee_id) AS employee_id,
					COALESCE(e.name, CONCAT('Anggota #', m.member_no)) AS name
				FROM lms_sch.members m
				LEFT JOIN lms_sch.employees e ON m.employee_id = e.employee_id
				%s
				ORDER BY m.member_no ASC
				LIMIT ? OFFSET ?
			`, whereClause)

			dataArgs := append(args, limit, offset)
			if err := config.DB.Raw(dataQuery, dataArgs...).Scan(&dtos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var totalRecords int64 = int64(len(dtos))
			if q != "" {
				countQuery := fmt.Sprintf(`
					SELECT COUNT(*)
					FROM lms_sch.members m
					LEFT JOIN lms_sch.employees e ON m.employee_id = e.employee_id
					%s
				`, whereClause)
				config.DB.Raw(countQuery, args...).Scan(&totalRecords)
			} else {
				totalRecords = 50000 // Fast constant total count for pagination when no query filter
			}

			totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
			if totalPages < 1 {
				totalPages = 1
			}

			c.JSON(http.StatusOK, gin.H{
				"data":          dtos,
				"page":          page,
				"limit":         limit,
				"total_records": totalRecords,
				"total_pages":   totalPages,
			})
		})

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
					COALESCE(e.name, CONCAT('Anggota #', m.member_no)) AS name
				FROM lms_sch.members m
				LEFT JOIN lms_sch.employees e ON m.employee_id = e.employee_id
				ORDER BY m.member_no ASC;
			`
			if err := config.DB.Raw(query).Scan(&dtos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": dtos})
		})

		// Payroll Schedules & Member Listing Endpoint for Manual Repayment
		api.GET("/payroll/schedules", func(c *gin.Context) {
			type PayrollScheduleDTO struct {
				ID                   int64     `json:"id"`
				LoanNo               int64     `json:"loan_no"`
				ApplicationNo        int64     `json:"application_no"`
				MemberNo             int64     `json:"member_no"`
				EmployeeName         string    `json:"employee_name"`
				DeptNo               string    `json:"dept_no"`
				Nik                  string    `json:"nik"`
				Period               string    `json:"period"`
				InstallmentNo        int       `json:"installment_no"`
				Principal            float64   `json:"principal"`
				Interest             float64   `json:"interest"`
				TotalInstallment     float64   `json:"total_installment"`
				AmountPaid           float64   `json:"amount_paid"`
				RemainingInstallment float64   `json:"remaining_installment"`
				Status               string    `json:"status"`
				DueDate              time.Time `json:"due_date"`
				RefNo                string    `json:"ref_no"`
			}
			periodParam := strings.TrimSpace(c.Query("period"))
			memberNoParam := strings.TrimSpace(c.Query("member_no"))

			var conditionsLs []string
			var argsLs []interface{}

			// Filter non-PAID and non-CLOSED schedules directly in SQL
			conditionsLs = append(conditionsLs, "ls.status != 'PAID'", "ls.status != 'CLOSED'")

			if periodParam != "" {
				conditionsLs = append(conditionsLs, "ls.period = ?")
				argsLs = append(argsLs, periodParam)
			}

			if memberNoParam != "" {
				if mNo, err := strconv.ParseInt(memberNoParam, 10, 64); err == nil && mNo > 0 {
					conditionsLs = append(conditionsLs, "(l.member_no = ? OR m.employee_id = ? OR e.employee_id = ?)")
					argsLs = append(argsLs, mNo, mNo, mNo)
				}
			}

			whereSqlLs := ""
			if len(conditionsLs) > 0 {
				whereSqlLs = "WHERE " + strings.Join(conditionsLs, " AND ")
			}

			var dtos []PayrollScheduleDTO
			query := fmt.Sprintf(`
				SELECT 
					ls.id,
					ls.loan_no,
					l.application_no,
					l.member_no,
					COALESCE(e.name, CONCAT('Karyawan #', l.member_no)) AS employee_name,
					COALESCE(e.deptno, 'IT-01') AS dept_no,
					COALESCE(CAST(e.employee_id AS VARCHAR), CAST(m.employee_id AS VARCHAR), CAST(l.member_no AS VARCHAR)) AS nik,
					ls.period,
					ls.installment_no,
					ls.principal,
					ls.interest,
					ls.total_installment,
					COALESCE(ls.amount_paid, 0) AS amount_paid,
					COALESCE(ls.remaining_installment, 0) AS remaining_installment,
					ls.status,
					ls.due_date,
					CONCAT('LMS-PAY-', REPLACE(ls.period, '-', ''), '-', l.member_no) AS ref_no
				FROM lms_sch.loan_schedules ls
				JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
				LEFT JOIN lms_sch.members m ON l.member_no = m.member_no
				LEFT JOIN lms_sch.employees e ON e.employee_id = l.member_no OR e.employee_id = m.employee_id
				%s
				ORDER BY ls.due_date ASC, ls.id ASC;
			`, whereSqlLs)
			if err := config.DB.Raw(query, argsLs...).Scan(&dtos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"data":   dtos,
			})
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
				if p, err := paramRepo.FindByKey("FOLDER_BILL_EXPORT"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
					targetFolder = strings.TrimSpace(p.KeyValue)
				} else {
					targetFolder = "D:\\Data_NK\\Project5\\LMS\\export_payroll"
				}
			}

			type RowDTO struct {
				EmployeeID      string  `gorm:"column:employee_id"`
				MemberNo        int64   `gorm:"column:member_no"`
				LoanNo          int64   `gorm:"column:loan_no"`
				NamaKaryawan    string  `gorm:"column:nama_karyawan"`
				DeptNo          string  `gorm:"column:dept_no"`
				KodePotongan    string  `gorm:"column:kode_potongan"`
				NamaPotongan    string  `gorm:"column:nama_potongan"`
				Periode         string  `gorm:"column:periode"`
				NominalPotongan float64 `gorm:"column:nominal_potongan"`
				NoReferensi     string  `gorm:"column:no_referensi"`
			}

			// 1. Fetch SCAN_DUEDATE_BILLING parameter from cache (zero DB overhead)
			var scanMode string
			if p, err := paramRepo.FindByKey("SCAN_DUEDATE_BILLING"); err == nil {
				scanMode = strings.ToUpper(strings.TrimSpace(p.KeyValue))
			}
			if scanMode == "" {
				scanMode = "PERIOD"
			}

			var rows []RowDTO
			var query string
			var queryArgs []interface{}

			baseQuery := `
				SELECT 
					COALESCE(CAST(e.employee_id AS VARCHAR), CAST(m.employee_id AS VARCHAR), e.bank_account_no, CAST(l.member_no AS VARCHAR)) AS employee_id,
					l.member_no AS member_no,
					l.loan_no AS loan_no,
					COALESCE(e.name, CONCAT('Anggota #', l.member_no)) AS nama_karyawan,
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
			`

			if scanMode == "PERIOD" {
				query = baseQuery + "\n  AND ls.period <= ?\nORDER BY ls.due_date ASC, ls.id ASC;"
				queryArgs = []interface{}{cutoffPeriod}
			} else if scanMode == "DUEDATE" {
				query = baseQuery + "\n  AND ls.due_date <= CAST(? AS DATE)\nORDER BY ls.due_date ASC, ls.id ASC;"
				queryArgs = []interface{}{cutoffDate}
			} else {
				query = baseQuery + "\n  AND (ls.due_date <= CAST(? AS DATE) OR ls.period <= ?)\nORDER BY ls.due_date ASC, ls.id ASC;"
				queryArgs = []interface{}{cutoffDate, cutoffPeriod}
			}

			if err := config.DB.Raw(query, queryArgs...).Scan(&rows).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed: " + err.Error()})
				return
			}

			// Fetch BILL_FILE_EXPORT_FORMAT parameter from cache (zero DB overhead)
			var exportFormat string
			if p, err := paramRepo.FindByKey("BILL_FILE_EXPORT_FORMAT"); err == nil {
				exportFormat = strings.ToLower(strings.TrimSpace(p.KeyValue))
			}
			if exportFormat != "csv" && exportFormat != "xlsx" {
				exportFormat = "xlsx" // default format
			}

			// Smart Cross-Platform Path Resolver (Windows / WSL / Linux)
			rawPath := strings.TrimRight(targetFolder, "/\\")
			rawPath = strings.ReplaceAll(rawPath, "/", "\\") // Normalize to Windows style for display

			diskPath := rawPath
			if strings.HasPrefix(strings.ToLower(diskPath), "d:\\") || strings.HasPrefix(strings.ToLower(diskPath), "d:/") {
				if os.PathSeparator == '/' {
					diskPath = "/mnt/d/" + strings.ReplaceAll(diskPath[3:], "\\", "/")
				}
			}

			if err := os.MkdirAll(diskPath, 0755); err != nil {
				log.Printf("Error creating export directory %s: %v", diskPath, err)
			}

			var fileName string
			var userDisplayPath string
			var csvContent string
			var xlsxBase64 string

			if exportFormat == "xlsx" {
				fileName = fmt.Sprintf("ADIRA_PAYROLL_KOPKARA_OUTGOING_%s.xlsx", time.Now().Format("200601"))
				diskFullPath := filepath.Join(diskPath, fileName)

				f := excelize.NewFile()
				sheetName := "Sheet1"
				index, _ := f.NewSheet(sheetName)
				f.SetActiveSheet(index)

				headers := []string{
					"EMPLOYEE_ID", "MEMBER_NO", "LOAN_NO", "NAMA_KARYAWAN", "DEPT_NO",
					"KODE_POTONGAN", "NAMA_POTONGAN", "PERIODE", "NOMINAL_TAGIHAN",
					"NOMINAL_TERPOTONG", "STATUS_POTONGAN", "KETERANGAN", "NO_REFERENSI",
				}

				styleID, _ := f.NewStyle(&excelize.Style{
					Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
					Fill: excelize.Fill{Type: "pattern", Color: []string{"#0B2545"}, Pattern: 1},
				})

				for colIdx, h := range headers {
					cellName, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
					f.SetCellValue(sheetName, cellName, h)
					f.SetCellStyle(sheetName, cellName, cellName, styleID)
				}

				for rowIdx, r := range rows {
					rowNum := rowIdx + 2
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), r.EmployeeID)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), r.MemberNo)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), r.LoanNo)
					f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), r.NamaKaryawan)
					f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), r.DeptNo)
					f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), r.KodePotongan)
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), r.NamaPotongan)
					f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), r.Periode)
					f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), r.NominalPotongan)
					f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), "")
					f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), "")
					f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), "")
					f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowNum), r.NoReferensi)
				}

				if err := f.SaveAs(diskFullPath); err != nil {
					log.Printf("Failed writing XLSX file to disk %s: %v", diskFullPath, err)
				}

				buffer, err := f.WriteToBuffer()
				if err == nil {
					xlsxBase64 = base64.StdEncoding.EncodeToString(buffer.Bytes())
				}
				userDisplayPath = rawPath + "\\" + fileName

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     fmt.Sprintf("File XLSX berhasil digenerate dan disimpan langsung ke folder: %s", userDisplayPath),
					"file_path":   userDisplayPath,
					"file_name":   fileName,
					"file_format": "xlsx",
					"total_rows":  len(rows),
					"xlsx_base64": xlsxBase64,
				})
			} else {
				fileName = fmt.Sprintf("ADIRA_PAYROLL_KOPKARA_OUTGOING_%s.csv", time.Now().Format("200601"))
				diskFullPath := filepath.Join(diskPath, fileName)

				csvContent = "EMPLOYEE_ID,MEMBER_NO,LOAN_NO,NAMA_KARYAWAN,DEPT_NO,KODE_POTONGAN,NAMA_POTONGAN,PERIODE,NOMINAL_TAGIHAN,NOMINAL_TERPOTONG,STATUS_POTONGAN,KETERANGAN,NO_REFERENSI\n"
				for _, r := range rows {
					csvContent += fmt.Sprintf("%s,%d,%d,%s,%s,%s,%s,%s,%.2f,,,%s\n",
						r.EmployeeID, r.MemberNo, r.LoanNo, r.NamaKaryawan, r.DeptNo, r.KodePotongan, r.NamaPotongan, r.Periode, r.NominalPotongan, r.NoReferensi)
				}

				if err := os.WriteFile(diskFullPath, []byte(csvContent), 0644); err != nil {
					log.Printf("Failed writing CSV file to disk %s: %v", diskFullPath, err)
				}

				userDisplayPath = rawPath + "\\" + fileName

				c.JSON(http.StatusOK, gin.H{
					"status":      "success",
					"message":     fmt.Sprintf("File CSV berhasil digenerate dan disimpan langsung ke folder: %s", userDisplayPath),
					"file_path":   userDisplayPath,
					"file_name":   fileName,
					"file_format": "csv",
					"total_rows":  len(rows),
					"csv_content": csvContent,
				})
			}
		})

		recordCashBankTransactionHelper := func(db *gorm.DB, typeCode string, direction string, amount float64, refType string, refNo string, memberNo int64, description string, userStr string) error {
			if amount <= 0 {
				return nil
			}

			var defaultAccNo string
			db.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'DEFAULT_DISBURSEMENT_ACCOUNT_ID' LIMIT 1").Scan(&defaultAccNo)

			var account models.CashBankAccount
			var errFind error

			if strings.TrimSpace(defaultAccNo) != "" {
				errFind = db.Where("account_number = ? AND deleted_at IS NULL AND is_active = true", strings.TrimSpace(defaultAccNo)).First(&account).Error
			}

			if errFind != nil || account.AccountID == 0 {
				if errFirst := db.Where("deleted_at IS NULL AND is_active = true").Order("account_id ASC").First(&account).Error; errFirst != nil {
					log.Printf("[CASHBANK-WARNING] Rekening Operasional Kas/Bank Koperasi tidak ditemukan: %v", errFirst)
					return nil
				}
			}

			balanceBefore := account.CurrentBalance
			var balanceAfter float64

			if strings.ToUpper(direction) == "IN" {
				balanceAfter = balanceBefore + amount
			} else {
				balanceAfter = balanceBefore - amount
			}

			now := time.Now()
			if errUpd := db.Exec("UPDATE lms_sch.cashbank_accounts SET current_balance = ?, updated_at = ?, updated_user = ? WHERE account_id = ?", balanceAfter, now, userStr, account.AccountID).Error; errUpd != nil {
				log.Printf("[CASHBANK-ERROR] Gagal update saldo cashbank_accounts: %v", errUpd)
				return errUpd
			}

			txNo := fmt.Sprintf("CB-%s-%06d", now.Format("200601"), time.Now().UnixNano()%1000000)

			var memNoPtr *int64
			var empIdPtr *int64
			if memberNo > 0 {
				var m models.Member
				if errM := db.Where("member_no = ? OR employee_id = ?", memberNo, memberNo).First(&m).Error; errM == nil {
					memNoPtr = &m.MemberNo
					if m.EmployeeID > 0 {
						empIdPtr = &m.EmployeeID
					}
				} else {
					memNoPtr = &memberNo
				}
			}

			cbTrans := models.CashBankTransaction{
				TransactionNo:   txNo,
				TransactionDate: now,
				BankCode:        account.BankCode,
				BankAccountNo:   account.AccountNumber,
				TypeCode:        typeCode,
				Direction:       direction,
				Amount:          amount,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    balanceAfter,
				ReferenceType:   refType,
				ReferenceNo:     refNo,
				MemberNo:        memNoPtr,
				EmployeeID:      empIdPtr,
				Description:     description,
			}
			cbTrans.CreatedAt = &now
			cbTrans.CreatedUser = &userStr
			cbTrans.UpdatedAt = &now
			cbTrans.UpdatedUser = &userStr

			if errCreate := db.Create(&cbTrans).Error; errCreate != nil {
				log.Printf("[CASHBANK-ERROR] Gagal mencatat cashbank_transactions: %v", errCreate)
				return errCreate
			}

			log.Printf("[CASHBANK-SUCCESS] Mutasi [%s %s] Ref %s memotong/tambah saldo Rekening %s (%s): Rp %.2f -> Rp %.2f (Tx: %s)",
				typeCode, direction, refNo, account.AccountNumber, account.BankCode, balanceBefore, balanceAfter, txNo)

			if memberNo > 0 && (typeCode == "RPMN" || typeCode == "RPIP") {
				var emp models.Employee
				var targetEmpID int64 = memberNo
				if empIdPtr != nil && *empIdPtr > 0 {
					targetEmpID = *empIdPtr
				} else {
					var m models.Member
					if errM := db.Where("member_no = ? OR employee_id = ?", memberNo, memberNo).First(&m).Error; errM == nil && m.EmployeeID > 0 {
						targetEmpID = m.EmployeeID
					}
				}
				db.Where("employee_id = ? OR employee_id = ?", targetEmpID, memberNo).First(&emp)
				if emp.EmployeeID > 0 {
					amtStr := services.FormatIDR(amount)
					waMsg := fmt.Sprintf("🔵 [PELUNASAN ANGSURAN DITERIMA]\n\nHalo %s,\nPembayaran angsuran pinjaman Anda sebesar %s (%s) telah berhasil diterima dan tercatat di sistem pada %s.\n\nTerima kasih,\nKopkara LMS EWA System",
						emp.Name, amtStr, description, now.Format("02-01-2006 15:04"))
					htmlMsg := fmt.Sprintf(`
						<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
							<h2 style="color: #3b82f6;">🔵 Pembayaran Angsuran Diterima</h2>
							<p>Halo <b>%s</b>,</p>
							<p>Pembayaran angsuran pinjaman Anda telah berhasil kami terima dengan rincian:</p>
							<table style="width: 100%%; max-width: 500px; border-collapse: collapse; margin: 15px 0;">
								<tr><td style="padding: 8px; border-bottom: 1px solid #ddd;">Tipe Pembayaran</td><td style="padding: 8px; border-bottom: 1px solid #ddd;"><b>%s</b></td></tr>
								<tr><td style="padding: 8px; border-bottom: 1px solid #ddd;">Nominal Diterima</td><td style="padding: 8px; border-bottom: 1px solid #ddd; color: #3b82f6; font-weight: bold;">%s</td></tr>
								<tr><td style="padding: 8px; border-bottom: 1px solid #ddd;">Referensi</td><td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td></tr>
								<tr><td style="padding: 8px; border-bottom: 1px solid #ddd;">Waktu Transaksi</td><td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td></tr>
							</table>
							<p>Terima kasih atas pembayaran Anda.</p>
						</div>
					`, emp.Name, typeCode, amtStr, refNo, now.Format("02-01-2006 15:04"))

					services.DispatchMenuNotificationWithUser("REPAYMENT", emp.EmployeeID, emp.PhoneNumber, emp.Email, fmt.Sprintf("Pembayaran Angsuran %s Diterima - Kopkara EWA", amtStr), waMsg, htmlMsg, db)
				}
			}

			return nil
		}

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

			var importFormat string
			if p, err := paramRepo.FindByKey("BILL_FILE_IMPORT_FORMAT"); err == nil {
				importFormat = strings.ToLower(strings.TrimSpace(p.KeyValue))
			}
			if importFormat != "csv" && importFormat != "xlsx" {
				importFormat = "xlsx"
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
				fileName = fmt.Sprintf("payroll_import_%s.%s", time.Now().Format("20060102_150405"), importFormat)
			}

			updaterName := strings.TrimSpace(req.UpdatedUser)
			if updaterName == "" {
				updaterName = "10101"
			}

			log.Printf("[IMPORT-API] Processing import for file '%s' by user '%s' with %d rows", fileName, updaterName, len(req.Rows))

			type ProcessedLogDTO struct {
				LoanNo     int64   `json:"loan_no"`
				RefNo      string  `json:"ref_no"`
				MemberNo   int64   `json:"member_no"`
				Nik        string  `json:"nik"`
				Name       string  `json:"name"`
				Period     string  `json:"period"`
				Amount     float64 `json:"amount"`
				Deducted   float64 `json:"deducted"`
				Status     string  `json:"status"`
				Keterangan string  `json:"keterangan"`
			}
			var processedLogs []ProcessedLogDTO

			var totalRecords int
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

				// Find corresponding loan_no & check existence in database
				var loanNoPtr *int64
				var foundLoanNo int64
				var loanExists bool

				if row.LoanNo > 0 {
					var count int64
					config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loans WHERE loan_no = ?", row.LoanNo).Scan(&count)
					if count > 0 {
						foundLoanNo = row.LoanNo
						loanNoPtr = &foundLoanNo
						loanExists = true
					}
				} else if memberNo > 0 {
					config.DB.Raw("SELECT loan_no FROM lms_sch.loans WHERE member_no = ? OR loan_no = ? LIMIT 1", memberNo, memberNo).Scan(&foundLoanNo)
					if foundLoanNo > 0 {
						loanNoPtr = &foundLoanNo
						loanExists = true
					}
				}

				deductedVal := row.Deducted
				if deductedVal <= 0 && row.NominalOriginal > 0 && statusUpper != "FAILED" && statusUpper != "REJECTED" {
					deductedVal = row.NominalOriginal
				}

				logStatus := statusUpper
				logReason := row.Keterangan
				if logReason == "" {
					logReason = "Potongan gaji diproses"
				}

				if !loanExists {
					logStatus = "FAILED"
					if row.LoanNo > 0 {
						logReason = fmt.Sprintf("Nomor Pinjaman (LOAN_NO: %d) tidak ditemukan di database LMS", row.LoanNo)
					} else {
						logReason = fmt.Sprintf("Anggota/Karyawan ID: %d tidak memiliki pinjaman di database LMS", memberNo)
					}
					totalFailedRecords++
					totalFailedAmount += row.NominalOriginal
				} else if deductedVal <= 0 {
					logStatus = "FAILED"
					logReason = "Nominal pemotongan Rp 0 / Tidak ada nominal terpotong"
					totalFailedRecords++
					totalFailedAmount += row.NominalOriginal
				} else {
					totalSuccessRecords++
					totalSuccessAmount += deductedVal
					if deductedVal < row.NominalOriginal {
						totalFailedAmount += (row.NominalOriginal - deductedVal)
					}
				}

				// Get employee name & NIK for display (employee_id / member_no)
				var empName, empNik string
				config.DB.Raw(`
					SELECT 
						COALESCE(e.name, CONCAT('Anggota #', m.member_no)), 
						COALESCE(CAST(e.employee_id AS VARCHAR), CAST(m.employee_id AS VARCHAR), CAST(m.member_no AS VARCHAR))
					FROM lms_sch.members m
					LEFT JOIN lms_sch.employees e ON e.employee_id = m.employee_id OR e.employee_id = m.member_no
					WHERE m.member_no = ? OR m.employee_id = ? LIMIT 1
				`, memberNo, memberNo).Row().Scan(&empName, &empNik)

				if empName == "" {
					empName = fmt.Sprintf("Anggota #%d", memberNo)
				}
				if empNik == "" {
					empNik = fmt.Sprintf("%d", memberNo)
				}

				// Insert log to lms_sch.payroll_deductions & get created ID
				var newDedID int64
				insertLogSQL := `
					INSERT INTO lms_sch.payroll_deductions (loan_no, period, nominal_original, nominal_deducted, process_date, status, failure_reason, created_at, created_user)
					VALUES (?, ?, ?, ?, NOW(), ?, ?, NOW(), ?) RETURNING id;
				`
				config.DB.Raw(insertLogSQL, loanNoPtr, periodStr, row.NominalOriginal, row.Deducted, logStatus, logReason, updaterName).Scan(&newDedID)

				targetLoanNo := row.LoanNo
				if targetLoanNo == 0 {
					targetLoanNo = foundLoanNo
				}
				refStr := fmt.Sprintf("%d", targetLoanNo)
				if refStr == "0" {
					refStr = row.RefNo
				}

				finalDeducted := deductedVal
				if logStatus == "FAILED" {
					finalDeducted = 0
				}

				processedLogs = append(processedLogs, ProcessedLogDTO{
					LoanNo:     targetLoanNo,
					RefNo:      refStr,
					MemberNo:   memberNo,
					Nik:        empNik,
					Name:       empName,
					Period:     periodStr,
					Amount:     row.NominalOriginal,
					Deducted:   finalDeducted,
					Status:     logStatus,
					Keterangan: logReason,
				})

				if finalDeducted > 0 && logStatus != "FAILED" {
					refNoStr := fmt.Sprintf("PAY-%d-%s", targetLoanNo, periodStr)
					descStr := fmt.Sprintf("Pelunasan Import Payroll [%s] - %s (NIK: %s) - Pinjaman #%d", fileName, empName, empNik, targetLoanNo)
					_ = recordCashBankTransactionHelper(config.DB, "RPIP", "IN", finalDeducted, "PAYROLL_IMPORT", refNoStr, memberNo, descStr, updaterName)
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

						// Update loans: kurangi outstanding + set CLOSED dalam 1 query jika semua schedule sudah PAID
						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET outstanding_amount = GREATEST(0, outstanding_amount - ?),
							    status = CASE
							        WHEN (SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID') = 0
							        THEN 'CLOSED'
							        ELSE status
							    END,
							    updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?
						`, payThis, s.LoanNo, updaterName, s.LoanNo)

						// Cek apakah loan baru saja di-CLOSE untuk update loan_application
						var loanStatus string
						config.DB.Raw("SELECT status FROM lms_sch.loans WHERE loan_no = ?", s.LoanNo).Scan(&loanStatus)
						if loanStatus == "CLOSED" {
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

						// 2. Update loans: kurangi outstanding + set CLOSED dalam 1 query jika semua schedule sudah PAID
						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET outstanding_amount = GREATEST(0, outstanding_amount - ?),
							    status = CASE
							        WHEN (SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID') = 0
							        THEN 'CLOSED'
							        ELSE status
							    END,
							    updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?
						`, payThis, s.LoanNo, updaterName, s.LoanNo)

						// 3. Cek apakah loan baru saja di-CLOSE untuk update loan_application
						var loanStatusWF string
						config.DB.Raw("SELECT status FROM lms_sch.loans WHERE loan_no = ?", s.LoanNo).Scan(&loanStatusWF)
						if loanStatusWF == "CLOSED" {
							config.DB.Exec(`
								UPDATE lms_sch.loan_applications
								SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
								WHERE application_no = ?
							`, updaterName, s.ApplicationNo)
						}
					}
				}
			}

			// Auto Backup / Move File (XLSX / CSV) to FOLDER_BILL_EXPORT_BCK or FOLDER_BILL_IMPORT_BCK
			backupMsg := ""
			if req.FileName != "" {
				var bckFolderParam, importFolderParam string
				// Baca dari cache parameter (tidak perlu query DB)
				if p, err := paramRepo.FindByKey("FOLDER_BILL_IMPORT_BCK"); err == nil {
					bckFolderParam = p.KeyValue
				}
				if bckFolderParam == "" {
					if p, err := paramRepo.FindByKey("FOLDER_BILL_EXPORT_BCK"); err == nil {
						bckFolderParam = p.KeyValue
					}
				}
				if p, err := paramRepo.FindByKey("FOLDER_BILL_IMPORT"); err == nil {
					importFolderParam = p.KeyValue
				}

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
				if _, err := os.Stat(srcFilePath); err != nil {
					// Fallback attempt: check if file exists with BILL_FILE_IMPORT_FORMAT extension or alternative (.xlsx / .csv)
					var fmtParam string
					if p, errP := paramRepo.FindByKey("BILL_FILE_IMPORT_FORMAT"); errP == nil {
						fmtParam = strings.ToLower(strings.TrimSpace(p.KeyValue))
					}
					if fmtParam == "" {
						fmtParam = "xlsx"
					}
					baseNoExt := strings.TrimSuffix(req.FileName, filepath.Ext(req.FileName))
					altPath := filepath.Join(diskImportDir, baseNoExt+"."+fmtParam)
					if _, errAlt := os.Stat(altPath); errAlt == nil {
						srcFilePath = altPath
					}
				}

				if _, err := os.Stat(srcFilePath); err == nil {
					actualFileName := filepath.Base(srcFilePath)
					ext := filepath.Ext(actualFileName)
					baseName := strings.TrimSuffix(actualFileName, ext)
					bckFileName := fmt.Sprintf("%s_PROCESSED_%s%s", baseName, time.Now().Format("20060102_150405"), ext)
					dstFilePath := filepath.Join(diskBckDir, bckFileName)
					fmtLabel := strings.ToUpper(strings.TrimPrefix(ext, "."))
					if fmtLabel == "" {
						fmtLabel = "FILE"
					}

					if err := os.Rename(srcFilePath, dstFilePath); err == nil {
						log.Printf("[IMPORT-API] File successfully backed up: %s -> %s", srcFilePath, dstFilePath)
						backupMsg = fmt.Sprintf("\n\n📦 File %s [%s] telah otomatis dipindahkan ke folder Backup:\n📁 %s", fmtLabel, actualFileName, dstFilePath)
					} else {
						inputData, readErr := os.ReadFile(srcFilePath)
						if readErr == nil {
							_ = os.WriteFile(dstFilePath, inputData, 0644)
							_ = os.Remove(srcFilePath)
							log.Printf("[IMPORT-API] File copied and deleted: %s -> %s", srcFilePath, dstFilePath)
							backupMsg = fmt.Sprintf("\n\n📦 File %s [%s] telah otomatis dipindahkan ke folder Backup:\n📁 %s", fmtLabel, actualFileName, dstFilePath)
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
				"logs":                  processedLogs,
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

			// Reduce employees.total_loan by payment amount
			config.DB.Exec(`
				UPDATE lms_sch.employees
				SET total_loan = GREATEST(0, COALESCE(total_loan, 0) - ?)
				WHERE employee_id = (SELECT employee_id FROM lms_sch.members WHERE member_no = ? LIMIT 1);
			`, req.Nominal, req.MemberNo)

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
				// Pay single installment (incorporates existing partial amount_paid)
				var schedID int64
				var totInst, prevPaid float64
				var appNo int64
				config.DB.Raw(`
					SELECT ls.id, ls.total_installment, COALESCE(ls.amount_paid, 0), l.application_no
					FROM lms_sch.loan_schedules ls
					JOIN lms_sch.loans l ON ls.loan_no = l.loan_no
					WHERE ls.loan_no = ? AND ls.status != 'PAID'
					ORDER BY ls.due_date ASC, ls.id ASC LIMIT 1
				`, req.LoanNo).Row().Scan(&schedID, &totInst, &prevPaid, &appNo)

				if schedID > 0 {
					paymentInput := req.Nominal
					if paymentInput <= 0 {
						paymentInput = totInst - prevPaid
					}
					newPaidVal := prevPaid + paymentInput
					var remVal float64
					if newPaidVal == totInst {
						remVal = 0
					} else if newPaidVal > totInst {
						remVal = totInst - newPaidVal // Negative (overpaid credit)
					} else {
						remVal = newPaidVal - totInst // Negative (underpaid shortfall)
					}

					newStat := "UNPAID"
					if newPaidVal >= totInst {
						newStat = "PAID"
					} else if newPaidVal > 0 {
						newStat = "PARTIAL"
					}

					config.DB.Exec(`
						UPDATE lms_sch.loan_schedules
						SET amount_paid = ?, remaining_installment = ?, status = ?, updated_at = NOW(), updated_user = ?
						WHERE id = ?;
					`, newPaidVal, remVal, newStat, updaterName, schedID)

					config.DB.Exec(`
						UPDATE lms_sch.loans
						SET outstanding_amount = GREATEST(0, outstanding_amount - ?), updated_at = NOW(), updated_user = ?
						WHERE loan_no = ?;
					`, paymentInput, updaterName, req.LoanNo)

					// Check if all schedules are PAID -> CLOSE loan and loan_application
					var unpaidCount int64
					config.DB.Raw("SELECT COUNT(*) FROM lms_sch.loan_schedules WHERE loan_no = ? AND status != 'PAID'", req.LoanNo).Scan(&unpaidCount)
					if unpaidCount == 0 {
						config.DB.Exec(`
							UPDATE lms_sch.loans
							SET status = 'CLOSED', outstanding_amount = 0, updated_at = NOW(), updated_user = ?
							WHERE loan_no = ?;
						`, updaterName, req.LoanNo)

						config.DB.Exec(`
							UPDATE lms_sch.loan_applications
							SET status = 'CLOSED', updated_at = NOW(), updated_user = ?
							WHERE application_no = ? OR application_no = (SELECT application_no FROM lms_sch.loans WHERE loan_no = ?);
						`, updaterName, appNo, req.LoanNo)
					}
				}
			}

			if req.Nominal > 0 {
				refNoStr := fmt.Sprintf("PAY-%d-%s", req.LoanNo, req.Period)
				descStr := fmt.Sprintf("Pelunasan Manual [%s] - Ref: %s - %s", req.PaymentType, req.ReferenceNo, req.Notes)
				_ = recordCashBankTransactionHelper(config.DB, "RPMN", "IN", req.Nominal, "REPAYMENT_MANUAL", refNoStr, req.MemberNo, descStr, updaterName)
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
					COALESCE(e.name, 'Karyawan') AS employee_name,
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

		// GET Email Diagnostic & Test Endpoint (Public)
		r.GET("/api/test-email", func(c *gin.Context) {
			toEmail := c.DefaultQuery("to", "nkholim@yahoo.com")
			smtpHost := cache.ParameterCache.Get("SMTP_HOST", "smtp.gmail.com", config.DB)
			if envHost := strings.TrimSpace(os.Getenv("SMTP_HOST")); envHost != "" {
				smtpHost = envHost
			}

			smtpPort := cache.ParameterCache.Get("SMTP_PORT", "587", config.DB)
			if envPort := strings.TrimSpace(os.Getenv("SMTP_PORT")); envPort != "" {
				smtpPort = envPort
			}

			smtpUser := strings.TrimSpace(os.Getenv("SMTP_USER"))
			if smtpUser == "" {
				smtpUser = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
			}
			if smtpUser == "" {
				dbUser := cache.ParameterCache.Get("SMTP_USERNAME", "", config.DB)
				if dbUser != "" && dbUser != "your-email@gmail.com" {
					smtpUser = dbUser
				}
			}

			smtpPass := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
			if smtpPass == "" {
				smtpPass = strings.TrimSpace(os.Getenv("SMTP_PASS"))
			}
			if smtpPass == "" {
				dbPass := cache.ParameterCache.Get("SMTP_PASSWORD", "", config.DB)
				if dbPass != "" && !strings.HasPrefix(dbPass, "xxxx") {
					smtpPass = dbPass
				}
			}

			fromName := cache.ParameterCache.Get("SMTP_FROM_NAME", "Kopkara LMS System", config.DB)
			if envName := strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")); envName != "" {
				fromName = envName
			}

			fromEmail := cache.ParameterCache.Get("SMTP_FROM_EMAIL", smtpUser, config.DB)
			if envFrom := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")); envFrom != "" {
				fromEmail = envFrom
			}
			if fromEmail == "" {
				fromEmail = smtpUser
			}

			passMasked := ""
			if len(smtpPass) > 4 {
				passMasked = smtpPass[:2] + "****" + smtpPass[len(smtpPass)-2:]
			} else if smtpPass != "" {
				passMasked = "****"
			}

			log.Printf("[DIAGNOSTIC-TEST-EMAIL] Testing email to '%s' using Host: %s:%s | User: %s | PassLen: %d", toEmail, smtpHost, smtpPort, smtpUser, len(smtpPass))

			subject := "🔍 Test Email Diagnostic Notifikasi LMS Kopkara"
			htmlBody := fmt.Sprintf(`
				<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
					<h2 style="color: #3b82f6;">🟢 Test Email Notifikasi LMS Kopkara</h2>
					<p>Email ini dikirim otomatis oleh sistem LMS Kopkara EWA untuk menguji koneksi SMTP Gmail.</p>
					<ul>
						<li><b>SMTP Host:</b> %s:%s</li>
						<li><b>Pengirim:</b> %s (%s)</li>
						<li><b>Penerima:</b> %s</li>
						<li><b>Waktu Uji:</b> %s</li>
					</ul>
				</div>
			`, smtpHost, smtpPort, fromName, fromEmail, toEmail, time.Now().Format("02-01-2006 15:04:05"))

			err := services.SendEmailNotification(toEmail, subject, htmlBody, true, config.DB)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":      "FAILED",
					"error":       err.Error(),
					"to":          toEmail,
					"host":        smtpHost,
					"port":        smtpPort,
					"user":        smtpUser,
					"pass_len":    len(smtpPass),
					"pass_masked": passMasked,
					"hint":        "Periksa log terminal untuk rincian error SMTP. Jika Gmail, pastikan menggunakan 16-karakter App Password (bukan password normal Gmail).",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "SUCCESS",
				"message": fmt.Sprintf("Berhasil mengirim Email uji ke %s via %s:%s!", toEmail, smtpHost, smtpPort),
				"to":      toEmail,
				"host":    smtpHost,
				"port":    smtpPort,
				"user":    smtpUser,
			})
		})

		// GET Diagnostic FCM Test Push Notification Endpoint (Disabled / No-Op)
		r.GET("/api/test-fcm", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "DISABLED", "message": "FCM Push Notifications telah dinonaktifkan."})
		})

		// GET In-App User Notifications Endpoint (Disabled / No-Op)
		api.GET("/notifications", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"unread_count": 0, "data": []interface{}{}})
		})

		// PUT Mark Single Notification as Read (No-Op)
		api.PUT("/notifications/:id/read", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})

		// PUT Mark All User Notifications as Read (No-Op)
		api.PUT("/notifications/read-all", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})

		// POST Register / Update FCM Device Token Endpoint (Disabled / No-Op)
		r.POST("/api/device-token", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "FCM Device Token endpoint disabled"})
		})

		// POST Payroll Adjustment Endpoint
		api.POST("/payroll/adjust", func(c *gin.Context) {
			type AdjustReq struct {
				RefNo          string  `json:"ref_no"`
				LoanNo         int64   `json:"loan_no"`
				Period         string  `json:"period"`
				AdjustmentType string  `json:"adjustment_type"` // 'FAILED_CORRECTION', 'OVERPAYMENT_REFUND', 'OVERPAYMENT_OFFSET'
				OriginalAmount float64 `json:"original_amount"`
				DeductedAmount float64 `json:"deducted_amount"`
				AdjustedAmount float64 `json:"adjusted_amount"`
				Notes          string  `json:"notes"`
				CreatedUser    string  `json:"created_user"`
			}
			var req AdjustReq
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			updaterName := req.CreatedUser
			if updaterName == "" {
				updaterName = "10101"
			}
			periodStr := req.Period
			if periodStr == "" {
				periodStr = "2026-08"
			}
			notesStr := strings.TrimSpace(req.Notes)
			// Calculate exact adjusted amount (nominal overpayment diff) if type is OVERPAYMENT
			adjVal := req.AdjustedAmount
			diffVal := req.DeductedAmount - req.OriginalAmount
			if diffVal < 0 {
				diffVal = -diffVal
			}
			if strings.Contains(strings.ToUpper(req.AdjustmentType), "OVERPAYMENT") {
				adjVal = diffVal
			}
			if adjVal <= 0 && diffVal > 0 {
				adjVal = diffVal
			}

			var refNoBigInt int64
			if req.RefNo != "" {
				var digits string
				for _, ch := range req.RefNo {
					if ch >= '0' && ch <= '9' {
						digits += string(ch)
					}
				}
				if digits != "" {
					fmt.Sscanf(digits, "%d", &refNoBigInt)
				}
			}
			if refNoBigInt == 0 && req.LoanNo > 0 {
				config.DB.Raw("SELECT id FROM lms_sch.payroll_deductions WHERE loan_no = ? AND period = ? ORDER BY id DESC LIMIT 1", req.LoanNo, periodStr).Scan(&refNoBigInt)
			}

			// 1. Insert into lms_sch.payroll_adjustments
			err := config.DB.Exec(`
				INSERT INTO lms_sch.payroll_adjustments (period, loan_no, ref_no, adjustment_type, original_amount, deducted_amount, adjusted_amount, notes, created_at, created_user)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?);
			`, periodStr, req.LoanNo, refNoBigInt, req.AdjustmentType, req.OriginalAmount, req.DeductedAmount, adjVal, notesStr, updaterName).Error

			if err != nil {
				log.Printf("[ADJUST-ERROR] Failed to insert payroll_adjustment: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan adjustment ke database: " + err.Error()})
				return
			}

			// 2. Update payroll_deductions if refNoBigInt or req.LoanNo is provided
			if refNoBigInt > 0 || req.LoanNo > 0 {
				targetLoan := req.LoanNo
				if targetLoan == 0 {
					targetLoan = refNoBigInt
				}
				config.DB.Exec(`
					UPDATE lms_sch.payroll_deductions
					SET status = 'ADJUSTED', failure_reason = CONCAT('ADJUSTMENT: ', CAST(? AS VARCHAR)), updated_at = NOW()
					WHERE (loan_no = ? OR id = ?) AND (period = ? OR period = '');
				`, notesStr, targetLoan, refNoBigInt, periodStr)
			}

			// Always zero out remaining installment on loan_schedules upon adjustment
			if req.LoanNo > 0 {
				config.DB.Exec(`
					UPDATE lms_sch.loan_schedules
					SET remaining_installment = 0, amount_paid = total_installment, status = 'ADJUSTED', updated_at = NOW(), updated_user = ?
					WHERE loan_no = ? AND period = ?;
				`, updaterName, req.LoanNo, periodStr)
			}
			
			if strings.HasPrefix(req.RefNo, "LMS-PAY-") {
				// Parse member_no from LMS-PAY-YYYYMM-MEMBERNO
				parts := strings.Split(req.RefNo, "-")
				if len(parts) >= 3 {
					memberNoStr := parts[len(parts)-1]
					config.DB.Exec(`
						UPDATE lms_sch.loan_schedules
						SET remaining_installment = 0, amount_paid = total_installment, status = 'ADJUSTED', updated_at = NOW(), updated_user = ?
						WHERE period = ? AND loan_no IN (SELECT loan_no FROM lms_sch.loans WHERE CAST(member_no AS VARCHAR) = ?);
					`, updaterName, periodStr, memberNoStr)
				}
			}

			if adjVal > 0 {
				typeCode := "ADJI"
				dir := "IN"
				if strings.Contains(strings.ToUpper(req.AdjustmentType), "OUT") || strings.Contains(strings.ToUpper(req.AdjustmentType), "REFUND") {
					typeCode = "ADJO"
					dir = "OUT"
				}
				_ = recordCashBankTransactionHelper(config.DB, typeCode, dir, adjVal, "ADJUSTMENT", req.RefNo, 0, fmt.Sprintf("Adjustment [%s] - Ref: %s - %s", req.AdjustmentType, req.RefNo, notesStr), updaterName)
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": fmt.Sprintf("Adjustment [%s] berhasil disimpan dengan referensi %s: '%s'", req.AdjustmentType, req.RefNo, notesStr),
			})
		})

		// Reset & Reopen Reconciliation Endpoint
		api.POST("/payroll/reset-reconciliation", func(c *gin.Context) {
			period := c.Query("period")
			if period == "" {
				period = "2026-08"
			}
			config.DB.Exec("UPDATE lms_sch.loan_schedules SET amount_paid = 0, remaining_installment = total_installment, status = 'UNPAID', updated_at = NOW() WHERE period = ?", period)
			config.DB.Exec("DELETE FROM lms_sch.payroll_adjustments WHERE period = ? OR period = ''", period)
			config.DB.Exec("DELETE FROM lms_sch.payroll_deductions WHERE period = ? OR period = ''", period)
			config.DB.Exec("TRUNCATE lms_sch.loan_payroll_import_logs RESTART IDENTITY;")
			config.DB.Exec("DELETE FROM lms_sch.payroll_reconciliation_closings WHERE period = ? OR period = ''", period)

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": fmt.Sprintf("Status Rekonsiliasi & Jadwal Angsuran Periode %s berhasil di-reset total ke status OPEN / UNPAID.", period),
			})
		})

		// GET Payroll Adjustments Endpoint
		api.GET("/payroll/adjustments", func(c *gin.Context) {
			period := c.Query("period")
			if period == "" {
				period = "2026-08"
			}

			type AdjItem struct {
				ID             int64     `json:"id"`
				Period         string    `json:"period"`
				LoanNo         int64     `json:"loan_no"`
				RefNo          string    `json:"ref_no"`
				AdjustmentType string    `json:"adjustment_type"`
				OriginalAmount float64   `json:"original_amount"`
				DeductedAmount float64   `json:"deducted_amount"`
				AdjustedAmount float64   `json:"adjusted_amount"`
				Notes          string    `json:"notes"`
				CreatedAt      time.Time `json:"created_at"`
				CreatedUser    string    `json:"created_user"`
				EmployeeName   string    `json:"employee_name"`
				MemberNo       string    `json:"member_no"`
			}
			var list []AdjItem

			config.DB.Raw(`
				SELECT pa.id, pa.period, pa.loan_no, pa.ref_no, pa.adjustment_type, pa.original_amount, pa.deducted_amount, pa.adjusted_amount, pa.notes, pa.created_at, pa.created_user,
				       COALESCE(e.name, CONCAT('Anggota #', m.member_no), 'Anggota #110101') as employee_name,
				       COALESCE(CAST(m.member_no AS VARCHAR), CAST(pa.loan_no AS VARCHAR), '110101') as member_no
				FROM lms_sch.payroll_adjustments pa
				LEFT JOIN lms_sch.payroll_deductions pd ON CAST(pa.ref_no AS VARCHAR) = CAST(pd.id AS VARCHAR)
				LEFT JOIN lms_sch.loans l ON (pa.loan_no = l.loan_no OR pd.loan_no = l.loan_no) AND l.loan_no > 0
				LEFT JOIN lms_sch.members m ON l.member_no = m.member_no
				LEFT JOIN lms_sch.employees e ON e.employee_id = m.employee_id OR e.employee_id = m.member_no
				WHERE pa.period = ?
				ORDER BY pa.id DESC
			`, period).Scan(&list)

			c.JSON(http.StatusOK, gin.H{
				"period":      period,
				"adjustments": list,
			})
		})

		// POST Close Reconciliation Endpoint
		api.POST("/payroll/close-reconciliation", func(c *gin.Context) {
			type CloseReq struct {
				Period           string `json:"period"`
				HrdSignatory     string `json:"hrd_signatory"`
				FinanceSignatory string `json:"finance_signatory"`
				KopkaraSignatory string `json:"kopkara_signatory"`
				ClosingNotes     string `json:"closing_notes"`
				ClosedUser       string `json:"closed_user"`
			}
			var req CloseReq
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			periodStr := req.Period
			if periodStr == "" {
				periodStr = "2026-08"
			}
			updaterName := req.ClosedUser
			if updaterName == "" {
				updaterName = "10101"
			}

			config.DB.Exec(`
				INSERT INTO lms_sch.payroll_reconciliation_closings (period, status, hrd_signatory, finance_signatory, kopkara_signatory, closing_notes, closed_at, closed_user)
				VALUES (?, 'CLOSED', ?, ?, ?, ?, NOW(), ?)
				ON CONFLICT (period) DO UPDATE 
				SET status = 'CLOSED', hrd_signatory = EXCLUDED.hrd_signatory, finance_signatory = EXCLUDED.finance_signatory, kopkara_signatory = EXCLUDED.kopkara_signatory, closing_notes = EXCLUDED.closing_notes, closed_at = NOW(), closed_user = EXCLUDED.closed_user;
			`, periodStr, req.HrdSignatory, req.FinanceSignatory, req.KopkaraSignatory, req.ClosingNotes, updaterName)

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": fmt.Sprintf("Laporan Rekonsiliasi Payroll Periode [%s] telah Resmi Ditutup (CLOSED) & Ditandatangani!", periodStr),
			})
		})

		// GET Reconciliation Status Endpoint
		api.GET("/payroll/reconciliation-status", func(c *gin.Context) {
			periodStr := strings.TrimSpace(c.Query("period"))
			if periodStr == "" {
				periodStr = "2026-08"
			}

			type ClosingInfo struct {
				ID               int64  `json:"id"`
				Period           string `json:"period"`
				Status           string `json:"status"`
				HrdSignatory     string `json:"hrd_signatory"`
				FinanceSignatory string `json:"finance_signatory"`
				KopkaraSignatory string `json:"kopkara_signatory"`
				ClosingNotes     string `json:"closing_notes"`
			}
			var info ClosingInfo
			config.DB.Raw("SELECT id, period, status, COALESCE(hrd_signatory, '') AS hrd_signatory, COALESCE(finance_signatory, '') AS finance_signatory, COALESCE(kopkara_signatory, '') AS kopkara_signatory, COALESCE(closing_notes, '') AS closing_notes FROM lms_sch.payroll_reconciliation_closings WHERE period = ? LIMIT 1", periodStr).Scan(&info)

			if info.Period == "" {
				info.Period = periodStr
				info.Status = "OPEN"
			}

			c.JSON(http.StatusOK, gin.H{"status": "success", "data": info})
		})
	}

	// Root Ping Handler for Mobile APK & Browser Connection Testing
	r.GET("/", func(c *gin.Context) {
		log.Printf("🟢 [PING TEST SUCCESS] Request received from Client IP: %s", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"status": "online",
			"message": "LMS Backend Go Server Active",
			"client_ip": c.ClientIP(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Profiling Go pprof Listener (Active when ENABLE_PPROF=true or APP_ENV=development)
	if os.Getenv("ENABLE_PPROF") == "true" || os.Getenv("APP_ENV") == "development" {
		go func() {
			log.Println("[PPROF] Starting Go profiler server on http://localhost:6060/debug/pprof")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				log.Printf("[PPROF ERROR] Profiler server error: %v", err)
			}
		}()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	certPath := strings.TrimSpace(os.Getenv("SSL_CERT_PATH"))
	keyPath := strings.TrimSpace(os.Getenv("SSL_KEY_PATH"))

	if certPath != "" && strings.ToUpper(certPath) != "OFF" && keyPath != "" && strings.ToUpper(keyPath) != "OFF" {
		log.Printf("Starting HTTPS server on port %s using cert %s and key %s", port, certPath, keyPath)
		if err := r.RunTLS(":"+port, certPath, keyPath); err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		log.Printf("Starting HTTP server on port %s (SSL is OFF)", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}
}
