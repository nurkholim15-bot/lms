package services

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"lms-backend/cache"
	"lms-backend/models"

	"gorm.io/gorm"
)

// SendEmailNotification sends an HTML/Plain Email using Gmail SMTP or any Custom SMTP configured in lms_sch.global_parameters or .env
func SendEmailNotification(toEmail string, subject string, bodyContent string, isHTML bool, db *gorm.DB) error {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("Email tujuan kosong")
	}

	smtpHost := cache.ParameterCache.Get("SMTP_HOST", "smtp.gmail.com", db)
	if envHost := strings.TrimSpace(os.Getenv("SMTP_HOST")); envHost != "" {
		smtpHost = envHost
	}

	smtpPort := cache.ParameterCache.Get("SMTP_PORT", "587", db)
	if envPort := strings.TrimSpace(os.Getenv("SMTP_PORT")); envPort != "" {
		smtpPort = envPort
	}

	smtpUser := strings.TrimSpace(os.Getenv("SMTP_USER"))
	if smtpUser == "" {
		smtpUser = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	}
	if smtpUser == "" {
		dbUser := cache.ParameterCache.Get("SMTP_USERNAME", "", db)
		if dbUser != "" && dbUser != "your-email@gmail.com" {
			smtpUser = dbUser
		}
	}

	smtpPass := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	if smtpPass == "" {
		smtpPass = strings.TrimSpace(os.Getenv("SMTP_PASS"))
	}
	if smtpPass == "" {
		dbPass := cache.ParameterCache.Get("SMTP_PASSWORD", "", db)
		if dbPass != "" && !strings.HasPrefix(dbPass, "xxxx") {
			smtpPass = dbPass
		}
	}

	fromName := cache.ParameterCache.Get("SMTP_FROM_NAME", "Kopkara LMS System", db)
	if envName := strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")); envName != "" {
		fromName = envName
	}

	fromEmail := cache.ParameterCache.Get("SMTP_FROM_EMAIL", smtpUser, db)
	if envFrom := strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")); envFrom != "" {
		fromEmail = envFrom
	}
	if fromEmail == "" {
		fromEmail = smtpUser
	}

	if smtpUser == "" || smtpPass == "" {
		log.Printf("⚠️ [EMAIL-WARNING] Kredensial SMTP belum diisi (SMTP_USERNAME / SMTP_PASSWORD). Email ke '%s' dibatalkan.", toEmail)
		return fmt.Errorf("Kredensial SMTP_USERNAME / SMTP_PASSWORD belum dikonfigurasi di Parameter Global atau .env")
	}

	if fromEmail == "" {
		fromEmail = smtpUser
	}

	mimeType := "text/plain; charset=UTF-8"
	if isHTML {
		mimeType = "text/html; charset=UTF-8"
	}

	// Build RFC 822 Email Header & Content
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	header["To"] = toEmail
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = mimeType

	var message string
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + bodyContent

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	log.Printf("[EMAIL-DISPATCH] Sending email to '%s' via %s (Host: %s)...", toEmail, fromName, addr)

	// Auth setup
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// 1. SSL/TLS Direct Connection (Port 465)
	if smtpPort == "465" {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         smtpHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			log.Printf("❌ [EMAIL-ERROR] SSL Dial Failed to %s: %v", addr, err)
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			log.Printf("❌ [EMAIL-ERROR] SMTP SSL Client Error: %v", err)
			return err
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			log.Printf("❌ [EMAIL-ERROR] SMTP SSL Auth Error: %v", err)
			return err
		}
		if err = client.Mail(fromEmail); err != nil {
			return err
		}
		if err = client.Rcpt(toEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(message))
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		log.Printf("🟢 [EMAIL-SUCCESS] Email successfully sent to %s via SSL 465!", toEmail)
		return nil
	}

	// 2. STARTTLS / Standard SMTP (Port 587 or 25)
	c, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("❌ [EMAIL-ERROR] SMTP Dial Failed to %s: %v", addr, err)
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: smtpHost, InsecureSkipVerify: true}
		if err = c.StartTLS(config); err != nil {
			log.Printf("❌ [EMAIL-ERROR] STARTTLS Error: %v", err)
			return err
		}
	}

	if err = c.Auth(auth); err != nil {
		log.Printf("❌ [EMAIL-ERROR] SMTP Auth Error: %v", err)
		return err
	}

	if err = c.Mail(fromEmail); err != nil {
		return err
	}
	if err = c.Rcpt(toEmail); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	_ = w.Close()

	log.Printf("🟢 [EMAIL-SUCCESS] Email successfully sent to %s via STARTTLS %s!", toEmail, smtpPort)
	return nil
}

// SendWhatsAppNotification sends WhatsApp text message via Fonnte or Meta Cloud API
func SendWhatsAppNotification(phoneNumber string, messageText string, db *gorm.DB) error {
	fonnteToken := cache.ParameterCache.Get("FONNTE_TOKEN", os.Getenv("FONNTE_TOKEN"), db)
	phoneID := cache.ParameterCache.Get("META_WA_PHONE_NUMBER_ID", os.Getenv("META_WA_PHONE_NUMBER_ID"), db)
	token := cache.ParameterCache.Get("META_WA_ACCESS_TOKEN", os.Getenv("META_WA_ACCESS_TOKEN"), db)

	cleanPhone := strings.TrimSpace(phoneNumber)
	if strings.HasPrefix(cleanPhone, "0") {
		cleanPhone = "62" + cleanPhone[1:]
	}
	cleanPhone = strings.TrimPrefix(cleanPhone, "+")

	if cleanPhone == "" {
		return fmt.Errorf("Nomor WhatsApp tujuan kosong")
	}

	log.Printf("[WA-DISPATCH] Sending WhatsApp to '%s': %s", cleanPhone, messageText)

	// 1. Try Fonnte Gateway if token configured
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

		log.Printf("🟢 [FONNTE-WA-SUCCESS] WA Message sent to %s!", cleanPhone)
		return nil
	}

	// 2. Try Meta Cloud API if configured
	if phoneID != "" && token != "" {
		urlStr := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", phoneID)
		bodyPayload := fmt.Sprintf(`{"messaging_product":"whatsapp","to":"%s","type":"text","text":{"body":%q}}`, cleanPhone, messageText)
		req, err := http.NewRequest("POST", urlStr, strings.NewReader(bodyPayload))
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
			return fmt.Errorf("Meta WA Error (%d): %s", resp.StatusCode, string(buf))
		}
		log.Printf("🟢 [META-WA-SUCCESS] WA Message sent to %s via Meta Cloud API!", cleanPhone)
		return nil
	}

	log.Printf("⚠️ [WA-WARNING] Kredensial WhatsApp belum dikonfigurasi (FONNTE_TOKEN / META_WA_TOKEN). WA ke '%s' dibatalkan.", cleanPhone)
	return fmt.Errorf("Kredensial WhatsApp belum dikonfigurasi")
}

// DispatchMenuNotification checks lms_sch.menus.notification_type for the given menu path/id
// notification_type: 0 = OFF/None, 1 = Email Only, 2 = WhatsApp Only, 3 = Email & WhatsApp
func DispatchMenuNotification(menuPath string, targetPhone string, targetEmail string, subject string, messageText string, htmlBody string, db *gorm.DB) {
	menuPath = strings.TrimSpace(strings.ToLower(menuPath))

	var notifType int = -1
	if db != nil && menuPath != "" {
		var menu models.Menu
		if err := db.Where("LOWER(path) = ? OR CAST(menu_id AS VARCHAR) = ?", menuPath, menuPath).First(&menu).Error; err == nil {
			notifType = menu.NotificationType
		}
	}

	// Master notification toggle from global parameters
	masterChannels := strings.ToUpper(cache.ParameterCache.Get("NOTIFICATION_ENABLED_CHANNELS", "EMAIL,WA", db))
	if masterChannels == "OFF" || masterChannels == "NONE" || masterChannels == "" {
		log.Printf("[NOTIF-SKIP] All notifications disabled globally (NOTIFICATION_ENABLED_CHANNELS = 'OFF')")
		return
	}

	var sendEmail, sendWA bool

	if notifType >= 0 {
		switch notifType {
		case 0:
			log.Printf("[NOTIF-SKIP] Menu '%s' has notification_type = 0 (OFF). Notification skipped.", menuPath)
			return
		case 1:
			sendEmail = true
			sendWA = false
			log.Printf("[NOTIF-CONFIG] Menu '%s' notification_type = 1 (EMAIL ONLY)", menuPath)
		case 2:
			sendEmail = false
			sendWA = true
			log.Printf("[NOTIF-CONFIG] Menu '%s' notification_type = 2 (WHATSAPP ONLY)", menuPath)
		case 3:
			sendEmail = true
			sendWA = true
			log.Printf("[NOTIF-CONFIG] Menu '%s' notification_type = 3 (EMAIL & WHATSAPP)", menuPath)
		default:
			sendEmail = strings.Contains(masterChannels, "EMAIL")
			sendWA = strings.Contains(masterChannels, "WA")
		}
	} else {
		// Fallback to global parameter
		sendEmail = strings.Contains(masterChannels, "EMAIL")
		sendWA = strings.Contains(masterChannels, "WA")
	}

	// Execute asynchronously in background goroutine so backend APIs remain lightning fast (<50ms)
	go func() {
		if sendEmail && strings.TrimSpace(targetEmail) != "" {
			bodyToSend := htmlBody
			if bodyToSend == "" {
				bodyToSend = fmt.Sprintf("<div><p>%s</p></div>", strings.ReplaceAll(messageText, "\n", "<br/>"))
			}
			if err := SendEmailNotification(targetEmail, subject, bodyToSend, true, db); err != nil {
				log.Printf("❌ [ASYNC-EMAIL-FAILED] Menu '%s' to %s: %v", menuPath, targetEmail, err)
			}
		}

		if sendWA && strings.TrimSpace(targetPhone) != "" {
			if err := SendWhatsAppNotification(targetPhone, messageText, db); err != nil {
				log.Printf("❌ [ASYNC-WA-FAILED] Menu '%s' to %s: %v", menuPath, targetPhone, err)
			}
		}
	}()
}

// DispatchEventNotification is alias wrapper for DispatchMenuNotification
func DispatchEventNotification(eventType string, targetPhone string, targetEmail string, subject string, messageText string, htmlBody string, db *gorm.DB) {
	DispatchMenuNotification(eventType, targetPhone, targetEmail, subject, messageText, htmlBody, db)
}

// FormatIDR formats float64 amount to Indonesian Rupiah currency string (e.g. 69000 -> "Rp 69.000")
func FormatIDR(amount float64) string {
	intAmt := int64(amount)
	isNegative := false
	if intAmt < 0 {
		isNegative = true
		intAmt = -intAmt
	}
	str := strconv.FormatInt(intAmt, 10)
	var result []string
	length := len(str)
	for i, c := range str {
		if i > 0 && (length-i)%3 == 0 {
			result = append(result, ".")
		}
		result = append(result, string(c))
	}
	prefix := "Rp "
	if isNegative {
		prefix = "-Rp "
	}
	return prefix + strings.Join(result, "")
}
