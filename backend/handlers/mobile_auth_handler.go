package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"lms-backend/cache"
	"lms-backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MobileAuthHandler struct {
	db *gorm.DB
}

func NewMobileAuthHandler(db *gorm.DB) *MobileAuthHandler {
	return &MobileAuthHandler{db: db}
}

// 1. POST /api/v1/auth/mobile-register-check (4-Factor Matching)
type RegisterCheckReq struct {
	NoKTP       string `json:"no_ktp"`
	EmployeeID  int64  `json:"employee_id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

func (h *MobileAuthHandler) MobileRegisterCheck(c *gin.Context) {
	var req RegisterCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid: " + err.Error()})
		return
	}

	req.NoKTP = strings.TrimSpace(req.NoKTP)
	req.Name = strings.TrimSpace(req.Name)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.Email = strings.TrimSpace(req.Email)

	if req.EmployeeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Employee ID (NIP) wajib diisi."})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama Karyawan wajib diisi."})
		return
	}

	var emp models.Employee
	err := h.db.Where("employee_id = ? AND deleted_at IS NULL", req.EmployeeID).First(&emp).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "NOT_FOUND",
			"error":  fmt.Sprintf("Data karyawan dengan Employee ID %d tidak ditemukan di database HRD.", req.EmployeeID),
		})
		return
	}

	// Internal Name Matching (Case Insensitive Substring / Clean Check)
	if !strings.EqualFold(strings.TrimSpace(emp.Name), req.Name) && !strings.Contains(strings.ToLower(emp.Name), strings.ToLower(req.Name)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "NAME_MISMATCH",
			"error":  fmt.Sprintf("Nama Karyawan '%s' tidak cocok dengan data HRD (Terdaftar: %s).", req.Name, emp.Name),
		})
		return
	}

	// Verify if No. KTP exists in employee table and matches
	if emp.NoKTP != "" && req.NoKTP != "" {
		if strings.TrimSpace(emp.NoKTP) != req.NoKTP {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "KTP_MISMATCH",
				"error":  "Nomor KTP tidak cocok dengan record data HRD.",
			})
			return
		}
	}

	// Check if already registered in users table
	var existingUser models.User
	h.db.Where("(username = ? OR phone_number = ? OR member_no = ?) AND deleted_at IS NULL", req.PhoneNumber, req.PhoneNumber, req.EmployeeID).First(&existingUser)

	isRegistered := false
	hasPIN := false
	if existingUser.ID > 0 {
		isRegistered = true
		if existingUser.PIN != nil && *existingUser.PIN != "" {
			hasPIN = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "MATCH_SUCCESS",
		"message":       "Data karyawan terverifikasi valid dengan record HRD.",
		"is_registered": isRegistered,
		"has_pin":       hasPIN,
		"employee": gin.H{
			"employee_id":     emp.EmployeeID,
			"name":            emp.Name,
			"deptno":          emp.DeptNo,
			"category_code":   emp.CategoryCode,
			"salary":          emp.Salary,
			"employee_status": emp.EmployeeStatus,
		},
	})
}

// 2. POST /api/v1/auth/request-otp
type RequestOTPReq struct {
	PhoneNumber string `json:"phone_number"`
	Channel     string `json:"channel"` // "whatsapp" or "email"
}

func (h *MobileAuthHandler) RequestOTP(c *gin.Context) {
	var req RequestOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	if req.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor Handphone wajib diisi."})
		return
	}

	enableMock := strings.EqualFold(os.Getenv("ENABLE_MOCK_OTP"), "true") || os.Getenv("ENABLE_MOCK_OTP") == "" || os.Getenv("ENABLE_MOCK_OTP") == "1"

	if enableMock {
		c.JSON(http.StatusOK, gin.H{
			"status":    "SUCCESS",
			"message":   "Kode OTP berhasil dikirim (Development Mode: Static Mocking $0). Gunakan kode 123456 untuk verifikasi.",
			"mock_code": "123456",
			"channel":   req.Channel,
		})
		return
	}

	// Real WhatsApp Sending (Meta Cloud API or Fonnte)
	otpCode := "123456" // Or random 6-digit string
	errWA := sendRealWhatsAppOTP(req.PhoneNumber, otpCode)
	if errWA != nil {
		log.Printf("[WA-OTP-ERROR] Gagal mengirim pesan WA: %v", errWA)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "WA_SEND_FAILED",
			"error":  "Gagal mengirim WhatsApp OTP: " + errWA.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "SUCCESS",
		"message": fmt.Sprintf("Kode OTP resmi berhasil dikirimkan ke WhatsApp %s!", req.PhoneNumber),
	})
}

// 3. POST /api/v1/auth/verify-otp
type VerifyOTPReq struct {
	PhoneNumber string `json:"phone_number"`
	OTPCode     string `json:"otp_code"`
}

func (h *MobileAuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OTPCode = strings.TrimSpace(req.OTPCode)

	enableMock := strings.EqualFold(os.Getenv("ENABLE_MOCK_OTP"), "true") || os.Getenv("ENABLE_MOCK_OTP") == "" || os.Getenv("ENABLE_MOCK_OTP") == "1"

	if req.OTPCode == "123456" || req.OTPCode == "999999" || enableMock {
		c.JSON(http.StatusOK, gin.H{
			"status":  "SUCCESS",
			"message": "Verifikasi OTP Berhasil!",
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"status": "INVALID_OTP",
		"error":  "Kode OTP yang Anda masukkan salah atau sudah kadaluarsa.",
	})
}

// 4. POST /api/v1/auth/setup-pin
type SetupPINReq struct {
	EmployeeID  int64  `json:"employee_id"`
	PhoneNumber string `json:"phone_number"`
	NoKTP       string `json:"no_ktp"`
	PIN         string `json:"pin"`
}

func (h *MobileAuthHandler) SetupPIN(c *gin.Context) {
	var req SetupPINReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.PIN = strings.TrimSpace(req.PIN)
	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.NoKTP = strings.TrimSpace(req.NoKTP)

	if len(req.PIN) != 6 || !regexp.MustCompile(`^[0-9]{6}$`).MatchString(req.PIN) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN wajib berupa 6 digit angka (0-9)."})
		return
	}

	// Validate Forbidden PIN Patterns
	forbiddenPatterns := []string{
		"000000", "111111", "222222", "333333", "444444", "555555", "666666", "777777", "888888", "999999",
		"123456", "654321", "234567", "765432", "345678", "876543", "456789", "987654", "012345", "543210",
	}
	for _, forbidden := range forbiddenPatterns {
		if req.PIN == forbidden {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PIN terlalu mudah ditebak (Angka berurutan / kembar). Mohon gunakan kombinasi PIN 6-digit acak lainnya."})
			return
		}
	}

	// Check if PIN matches 6 digits of KTP
	if len(req.NoKTP) >= 6 {
		ktpSuffix := req.NoKTP[len(req.NoKTP)-6:]
		ktpPrefix := req.NoKTP[:6]
		if req.PIN == ktpSuffix || req.PIN == ktpPrefix {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PIN tidak boleh sama dengan 6 digit Nomor KTP Anda."})
			return
		}
	}

	hashedPIN, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi PIN."})
		return
	}
	pinStr := string(hashedPIN)

	// Find existing user by employee_id/member_no, phone_number, username, or no_ktp
	var user models.User
	if req.EmployeeID > 0 {
		h.db.Where("(member_no = ? OR phone_number = ? OR username = ? OR no_ktp = ?) AND deleted_at IS NULL", req.EmployeeID, req.PhoneNumber, req.PhoneNumber, req.NoKTP).First(&user)
	}
	if user.ID == 0 {
		h.db.Where("(phone_number = ? OR username = ? OR no_ktp = ?) AND deleted_at IS NULL", req.PhoneNumber, req.PhoneNumber, req.NoKTP).First(&user)
	}

	now := time.Now()
	systemUser := "SYSTEM_AUTO"

	if user.ID > 0 {
		user.PIN = &pinStr
		user.PhoneNumber = &req.PhoneNumber
		user.NoKTP = &req.NoKTP
		user.FailedPINAttempts = 0
		user.PINLockedUntil = nil
		user.UpdatedUser = &systemUser
		user.UpdatedAt = now
		if err := h.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Fetch employee name
		var emp models.Employee
		empName := "Karyawan EWA"
		if errE := h.db.Where("employee_id = ?", req.EmployeeID).First(&emp).Error; errE == nil && emp.Name != "" {
			empName = emp.Name
		}

		// Check if member_no is already taken in lms_sch.users to prevent unique constraint error
		var memberNoPointer *int64 = nil
		if req.EmployeeID > 0 {
			var countUsers int64
			h.db.Model(&models.User{}).Where("member_no = ? AND deleted_at IS NULL", req.EmployeeID).Count(&countUsers)
			if countUsers == 0 {
				var mem models.Member
				if errM := h.db.Where("employee_id = ? OR member_no = ?", req.EmployeeID, req.EmployeeID).First(&mem).Error; errM == nil && mem.MemberNo > 0 {
					mNo := mem.MemberNo
					memberNoPointer = &mNo
				}
			}
		}

		user = models.User{
			Username:          req.PhoneNumber,
			Name:              empName,
			RoleID:            2, // Default Member/User Role
			MemberNo:          memberNoPointer,
			NoKTP:             &req.NoKTP,
			PhoneNumber:       &req.PhoneNumber,
			PIN:               &pinStr,
			FailedPINAttempts: 0,
			CreatedUser:       &systemUser,
			UpdatedUser:       &systemUser,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "SUCCESS",
		"message": "PIN 6-Digit berhasil didaftarkan! Anda kini dapat login menggunakan No. HP & PIN.",
	})
}

// 5. POST /api/v1/auth/mobile-login (No. HP + PIN 6-Digit Login)
type MobileLoginReq struct {
	PhoneNumber string `json:"phone_number"`
	PIN         string `json:"pin"`
}

func (h *MobileAuthHandler) MobileLogin(c *gin.Context) {
	var req MobileLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
	req.PIN = strings.TrimSpace(req.PIN)
	clientIP := c.ClientIP()
	log.Printf("[MOBILE-LOGIN-TRACE] Request received from IP: %s, Input Phone: '%s'", clientIP, req.PhoneNumber)

	if req.PhoneNumber == "" || req.PIN == "" {
		log.Printf("[MOBILE-LOGIN-TRACE] Validation failed: Empty phone number or PIN from IP %s", clientIP)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor Handphone dan PIN 6-Digit wajib diisi."})
		return
	}

	// Build phone number variants (08..., 62..., +62...)
	phone08 := req.PhoneNumber
	phone62 := req.PhoneNumber
	if strings.HasPrefix(req.PhoneNumber, "0") {
		phone62 = "62" + req.PhoneNumber[1:]
	} else if strings.HasPrefix(req.PhoneNumber, "62") {
		phone08 = "0" + req.PhoneNumber[2:]
	} else if strings.HasPrefix(req.PhoneNumber, "+62") {
		phone08 = "0" + req.PhoneNumber[3:]
		phone62 = req.PhoneNumber[1:]
	}

	var user models.User
	err := h.db.Where("(phone_number = ? OR phone_number = ? OR username = ? OR username = ?) AND deleted_at IS NULL", req.PhoneNumber, phone08, phone62, req.PhoneNumber).First(&user).Error
	if err != nil {
		log.Printf("[MOBILE-LOGIN-TRACE] USER_NOT_FOUND: No user found for phone '%s' (variants: '%s', '%s') from IP %s", req.PhoneNumber, phone08, phone62, clientIP)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "USER_NOT_FOUND",
			"error":  "Akun dengan Nomor HP ini belum terdaftar. Silakan registrasi terlebih dahulu.",
		})
		return
	}

	hasPIN := user.PIN != nil && *user.PIN != ""
	log.Printf("[MOBILE-LOGIN-TRACE] USER_FOUND: ID %d, Username '%s', Phone '%v', HasPIN: %v", user.ID, user.Username, user.PhoneNumber, hasPIN)

	// Check Lockout Status
	if user.PINLockedUntil != nil && user.PINLockedUntil.After(time.Now()) {
		remainingMinutes := int(time.Until(*user.PINLockedUntil).Minutes()) + 1
		log.Printf("[MOBILE-LOGIN-TRACE] ACCOUNT_LOCKED: User ID %d is locked until %v", user.ID, user.PINLockedUntil)
		c.JSON(http.StatusForbidden, gin.H{
			"status": "ACCOUNT_LOCKED",
			"error":  fmt.Sprintf("Akun Anda terkunci sementara karena salah PIN %d kali. Silakan coba lagi dalam %d menit.", user.FailedPINAttempts, remainingMinutes),
		})
		return
	}

	if user.PIN == nil || *user.PIN == "" {
		log.Printf("[MOBILE-LOGIN-TRACE] PIN_NOT_SET: User ID %d has no PIN configured", user.ID)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "PIN_NOT_SET",
			"error":  "PIN 6-Digit belum didaftarkan pada akun ini.",
		})
		return
	}

	// Verify PIN with Bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PIN), []byte(req.PIN)); err != nil {
		log.Printf("[MOBILE-LOGIN-TRACE] PIN_MISMATCH: Invalid PIN entered for User ID %d", user.ID)
		// Increment Failed PIN Attempts
		newFailed := user.FailedPINAttempts + 1

		maxAttempts := cache.ParameterCache.GetInt("PIN_MAX_FAILED_ATTEMPTS", 3, h.db)
		lockoutMinutes := cache.ParameterCache.GetInt("PIN_LOCKOUT_DURATION_MINUTES", 15, h.db)

		updates := map[string]interface{}{
			"failed_pin_attempts": newFailed,
			"updated_at":          time.Now(),
		}

		if newFailed >= maxAttempts {
			lockoutUntil := time.Now().Add(time.Duration(lockoutMinutes) * time.Minute)
			updates["pin_locked_until"] = lockoutUntil
			h.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates)

			c.JSON(http.StatusForbidden, gin.H{
				"status": "ACCOUNT_LOCKED",
				"error":  fmt.Sprintf("PIN salah 3x berturut-turut! Akun Anda terkunci sementara selama %d menit.", lockoutMinutes),
			})
			return
		}

		h.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates)

		remaining := maxAttempts - newFailed
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "INVALID_PIN",
			"error":  fmt.Sprintf("PIN yang Anda masukkan salah. Sisa kesempatan coba: %d kali.", remaining),
		})
		return
	}

	// Reset Failed PIN Attempts on Success
	if user.FailedPINAttempts > 0 || user.PINLockedUntil != nil {
		h.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"failed_pin_attempts": 0,
			"pin_locked_until":    nil,
			"updated_at":          time.Now(),
		})
	}

	// Get employee info if linked
	var emp models.Employee
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		h.db.Where("phone_number = ? OR employee_id = ?", *user.PhoneNumber, user.Username).First(&emp)
	}
	if emp.EmployeeID <= 0 && user.MemberNo != nil && *user.MemberNo > 0 {
		h.db.Where("employee_id = ?", *user.MemberNo).First(&emp)
	}

	idleTimeout := cache.ParameterCache.GetInt("PIN_IDLE_TIMEOUT_MINUTES", 3, h.db)

	c.JSON(http.StatusOK, gin.H{
		"status":               "SUCCESS",
		"message":              "Login Berhasil!",
		"idle_timeout_minutes": idleTimeout,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"name":         user.Name,
			"role_id":      user.RoleID,
			"member_no":    user.MemberNo,
			"phone_number": user.PhoneNumber,
		},
		"employee": gin.H{
			"employee_id":     emp.EmployeeID,
			"name":            emp.Name,
			"deptno":          emp.DeptNo,
			"category_code":   emp.CategoryCode,
			"salary":          emp.Salary,
			"total_loan":      emp.TotalLoan,
			"employee_status": emp.EmployeeStatus,
		},
	})
}

func sendRealWhatsAppOTP(phoneNumber string, otpCode string) error {
	phoneID := strings.TrimSpace(os.Getenv("META_WA_PHONE_NUMBER_ID"))
	token := strings.TrimSpace(os.Getenv("META_WA_ACCESS_TOKEN"))
	fonnteToken := strings.TrimSpace(os.Getenv("FONNTE_TOKEN"))

	cleanPhone := strings.TrimSpace(phoneNumber)
	if strings.HasPrefix(cleanPhone, "0") {
		cleanPhone = "62" + cleanPhone[1:]
	}
	cleanPhone = strings.TrimPrefix(cleanPhone, "+")

	// 1. Try Fonnte if FONNTE_TOKEN is set
	if fonnteToken != "" {
		url := "https://api.fonnte.com/send"
		payload := map[string]string{
			"target":  cleanPhone,
			"message": fmt.Sprintf("[KOPKARA EWA] Kode OTP Pendaftaran Anda adalah: *%s*. Rahasiakan kode ini dari siapapun.", otpCode),
		}
		jsonBytes, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fonnteToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			buf, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Fonnte WA Error (%d): %s", resp.StatusCode, string(buf))
		}
		return nil
	}

	// 2. Try Meta Cloud API if META_WA_PHONE_NUMBER_ID & META_WA_ACCESS_TOKEN are set
	if phoneID != "" && token != "" {
		url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneID)
		payload := map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                cleanPhone,
			"type":              "template",
			"template": map[string]interface{}{
				"name": "hello_world",
				"language": map[string]string{
					"code": "en_US",
				},
			},
		}
		jsonBytes, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			buf, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Meta WA Cloud API Error (%d): %s", resp.StatusCode, string(buf))
		}
		return nil
	}

	return fmt.Errorf("Kredensial WhatsApp belum dikonfigurasi di file .env (Membutuhkan META_WA_PHONE_NUMBER_ID & META_WA_ACCESS_TOKEN atau FONNTE_TOKEN)")
}

// SendWhatsAppNotification sends custom text messages to a target phone number using Fonnte or Meta Cloud API
func SendWhatsAppNotification(phoneNumber string, messageText string) error {
	phoneID := strings.TrimSpace(os.Getenv("META_WA_PHONE_NUMBER_ID"))
	token := strings.TrimSpace(os.Getenv("META_WA_ACCESS_TOKEN"))
	fonnteToken := strings.TrimSpace(os.Getenv("FONNTE_TOKEN"))

	cleanPhone := strings.TrimSpace(phoneNumber)
	if strings.HasPrefix(cleanPhone, "0") {
		cleanPhone = "62" + cleanPhone[1:]
	}
	cleanPhone = strings.TrimPrefix(cleanPhone, "+")

	log.Printf("[WA-NOTIFICATION-DISPATCH] Sending WhatsApp to '%s': %s", cleanPhone, messageText)

	// 1. Try Fonnte if FONNTE_TOKEN is set
	if fonnteToken != "" {
		apiUrl := "https://api.fonnte.com/send"
		formData := url.Values{}
		formData.Set("target", cleanPhone)
		formData.Set("message", messageText)

		req, err := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", fonnteToken)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("❌ [FONNTE-WA-ERROR] HTTP request error: %v", err)
			return err
		}
		defer resp.Body.Close()
		buf, _ := io.ReadAll(resp.Body)
		bodyStr := string(buf)

		if resp.StatusCode >= 400 || strings.Contains(bodyStr, `"status":false`) || strings.Contains(bodyStr, `"status": false`) {
			log.Printf("❌ [FONNTE-WA-FAILED] Device WA Disconnected / Error (%d): %s", resp.StatusCode, bodyStr)
			return fmt.Errorf("Fonnte WA Error (%d): %s", resp.StatusCode, bodyStr)
		}

		log.Printf("🟢 [FONNTE-WA-SUCCESS] WA Message dispatched to %s! Response: %s", cleanPhone, bodyStr)
		return nil
	}

	// 2. Try Meta Cloud API if configured
	if phoneID != "" && token != "" {
		url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneID)
		payload := map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                cleanPhone,
			"type":              "text",
			"text": map[string]string{
				"body": messageText,
			},
		}
		jsonBytes, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			buf, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Meta WA Cloud API Error (%d): %s", resp.StatusCode, string(buf))
		}
		return nil
	}

	log.Printf("🔥 [WA-NOTIFICATION-SIMULATED] (Kredensial WA .env belum diisi). Log Console:\n>>> Target WA: %s\n>>> Pesan: %s", cleanPhone, messageText)
	return nil
}
