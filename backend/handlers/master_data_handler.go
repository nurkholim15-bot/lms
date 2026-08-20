package handlers

import (
	"fmt"
	"lms-backend/models"
	"lms-backend/repositories"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MasterDataHandler struct {
	db        *gorm.DB
	paramRepo repositories.ParameterRepository
}

func NewMasterDataHandler(db *gorm.DB, pRepo repositories.ParameterRepository) *MasterDataHandler {
	return &MasterDataHandler{db: db, paramRepo: pRepo}
}

func (h *MasterDataHandler) GetAll(c *gin.Context) {
	table := c.Param("table")

	// Determine pagination parameters
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	limit := 10
	if h.paramRepo != nil {
		if p, err := h.paramRepo.FindByKey("PAGINATION_LIMIT"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(p.KeyValue)); err == nil && parsed > 0 {
				limit = parsed
			}
		} else if p, err := h.paramRepo.FindByKey("DEFAULT_PAGE_SIZE"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			if parsed, err := strconv.Atoi(strings.TrimSpace(p.KeyValue)); err == nil && parsed > 0 {
				limit = parsed
			}
		}
	}

	if userLimit, _ := strconv.Atoi(c.Query("limit")); userLimit > 0 {
		limit = userLimit
	}

	offset := (page - 1) * limit
	q := strings.TrimSpace(c.Query("q"))
	likeStr := "%" + strings.ToLower(q) + "%"

	var totalRecords int64 = 0

	switch table {
	case "departments":
		var data []models.Department
		dQuery := h.db.Model(&models.Department{}).Where("deleted_at IS NULL")
		if q != "" {
			dQuery = dQuery.Where("(LOWER(deptno) LIKE ? OR LOWER(dept_name) LIKE ?)", likeStr, likeStr)
		}
		dQuery.Count(&totalRecords)
		dQuery.Order("deptno ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "employee-statuses":
		var data []models.EmployeeStatus
		stQuery := h.db.Model(&models.EmployeeStatus{})
		if q != "" {
			stQuery = stQuery.Where("(LOWER(status_code) LIKE ? OR LOWER(description) LIKE ?)", likeStr, likeStr)
		}
		stQuery.Count(&totalRecords)
		stQuery.Order("status_code ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "kopkara-statuses":
		var data []models.KopkaraStatus
		kpQuery := h.db.Model(&models.KopkaraStatus{})
		if q != "" {
			kpQuery = kpQuery.Where("(LOWER(status_code) LIKE ? OR LOWER(description) LIKE ?)", likeStr, likeStr)
		}
		kpQuery.Count(&totalRecords)
		kpQuery.Order("status_code ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "employee-categories":
		var data []models.EmployeeCategory
		catQuery := h.db.Model(&models.EmployeeCategory{})
		if q != "" {
			catQuery = catQuery.Where("(LOWER(category_code) LIKE ? OR LOWER(description) LIKE ?)", likeStr, likeStr)
		}
		catQuery.Count(&totalRecords)
		catQuery.Order("category_code ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "employees":
		var data []models.Employee
		empQuery := h.db.Model(&models.Employee{}).Where("deleted_at IS NULL")
		if q != "" {
			empIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if empIDSearch > 0 {
				empQuery = empQuery.Where("(employee_id = ? OR LOWER(name) LIKE ? OR LOWER(employee_status) LIKE ? OR LOWER(category_code) LIKE ?)", empIDSearch, likeStr, likeStr, likeStr)
			} else {
				empQuery = empQuery.Where("(LOWER(name) LIKE ? OR LOWER(employee_status) LIKE ? OR LOWER(category_code) LIKE ?)", likeStr, likeStr, likeStr)
			}
		}
		empQuery.Count(&totalRecords)
		empQuery.Order("employee_id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "members":
		var data []models.Member
		memQuery := h.db.Model(&models.Member{})
		if q != "" {
			memIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if memIDSearch > 0 {
				memQuery = memQuery.Where("(member_no = ? OR employee_id = ? OR LOWER(bank_account_name) LIKE ? OR LOWER(bank_name) LIKE ?)", memIDSearch, memIDSearch, likeStr, likeStr)
			} else {
				memQuery = memQuery.Where("(LOWER(bank_account_name) LIKE ? OR LOWER(bank_name) LIKE ?)", likeStr, likeStr)
			}
		}
		memQuery.Count(&totalRecords)
		memQuery.Order("member_no ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "roles":
		var data []models.Role
		roleQuery := h.db.Model(&models.Role{})
		if q != "" {
			roleIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if roleIDSearch > 0 {
				roleQuery = roleQuery.Where("(role_id = ? OR LOWER(role_name) LIKE ? OR LOWER(description) LIKE ?)", roleIDSearch, likeStr, likeStr)
			} else {
				roleQuery = roleQuery.Where("(LOWER(role_name) LIKE ? OR LOWER(description) LIKE ?)", likeStr, likeStr)
			}
		}
		roleQuery.Count(&totalRecords)
		roleQuery.Order("role_id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "menus":
		var data []models.Menu
		menuQuery := h.db.Model(&models.Menu{})
		if q != "" {
			menuIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if menuIDSearch > 0 {
				menuQuery = menuQuery.Where("(menu_id = ? OR LOWER(title) LIKE ? OR LOWER(path) LIKE ?)", menuIDSearch, likeStr, likeStr)
			} else {
				menuQuery = menuQuery.Where("(LOWER(title) LIKE ? OR LOWER(path) LIKE ?)", likeStr, likeStr)
			}
		}
		menuQuery.Count(&totalRecords)
		menuQuery.Order("menu_id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "role-menus":
		var data []models.RoleMenu
		rmQuery := h.db.Model(&models.RoleMenu{})
		if q != "" {
			rmIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if rmIDSearch > 0 {
				rmQuery = rmQuery.Where("(role_id = ? OR menu_id = ?)", rmIDSearch, rmIDSearch)
			}
		}
		rmQuery.Count(&totalRecords)
		rmQuery.Order("role_id ASC, menu_id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "parameters":
		var data []models.GlobalParameter
		paramQuery := h.db.Model(&models.GlobalParameter{})
		if q != "" {
			paramQuery = paramQuery.Where("(LOWER(key_name) LIKE ? OR LOWER(key_value) LIKE ? OR LOWER(description) LIKE ?)", likeStr, likeStr, likeStr)
		}
		paramQuery.Count(&totalRecords)
		paramQuery.Order("id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "users":
		var data []models.User
		uQuery := h.db.Model(&models.User{}).Where("deleted_at IS NULL")
		if q != "" {
			uIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if uIDSearch > 0 {
				uQuery = uQuery.Where("(id = ? OR member_no = ? OR LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(role) LIKE ?)", uIDSearch, uIDSearch, likeStr, likeStr, likeStr)
			} else {
				uQuery = uQuery.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(role) LIKE ?)", likeStr, likeStr, likeStr)
			}
		}
		uQuery.Count(&totalRecords)
		uQuery.Order("id ASC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	case "sessions":
		var data []models.Session
		sQuery := h.db.Model(&models.Session{})
		if q != "" {
			sQuery = sQuery.Where("(LOWER(username) LIKE ? OR LOWER(ip_address) LIKE ? OR LOWER(user_agent) LIKE ?)", likeStr, likeStr, likeStr)
		}
		sQuery.Count(&totalRecords)
		sQuery.Order("id DESC").Limit(limit).Offset(offset).Find(&data)
		totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
		if totalPages < 1 { totalPages = 1 }
		c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total_records": totalRecords, "total_pages": totalPages})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown master table"})
	}
}

func (h *MasterDataHandler) Save(c *gin.Context) {
	table := c.Param("table")
	switch table {
	case "departments":
		var data models.Department
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employee-statuses":
		var data models.EmployeeStatus
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "kopkara-statuses":
		var data models.KopkaraStatus
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employee-categories":
		var data models.EmployeeCategory
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employees":
		var data models.Employee
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "members":
		var data models.Member
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if data.EmployeeID > 0 {
			var existing models.Member
			query := h.db.Where("employee_id = ?", data.EmployeeID)
			if data.MemberNo > 0 {
				query = query.Where("member_no != ?", data.MemberNo)
			}
			if err := query.First(&existing).Error; err == nil && existing.MemberNo > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("employee_id %d sudah menjadi anggota dengan nomor anggota %d", data.EmployeeID, existing.MemberNo),
				})
				return
			}
		}

		if err := h.db.Save(&data).Error; err != nil {
			if strings.Contains(err.Error(), "members_employee_id_key") || strings.Contains(err.Error(), "23505") {
				var existing models.Member
				_ = h.db.Where("employee_id = ?", data.EmployeeID).First(&existing)
				memberNoStr := fmt.Sprintf("%d", existing.MemberNo)
				if existing.MemberNo == 0 {
					memberNoStr = "terkait"
				}
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("employee_id %d sudah menjadi anggota dengan nomor anggota %s", data.EmployeeID, memberNoStr),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "roles":
		var data models.Role
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "menus":
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var menu models.Menu
		if idVal, ok := req["menu_id"]; ok && idVal != nil {
			idNum := int64(0)
			switch v := idVal.(type) {
			case float64:
				idNum = int64(v)
			case int64:
				idNum = v
			}
			if idNum > 0 {
				h.db.Where("menu_id = ?", idNum).First(&menu)
			}
		}
		if pVal, ok := req["parent_id"]; ok && pVal != nil {
			switch v := pVal.(type) {
			case float64:
				p := int64(v)
				if p > 0 { menu.ParentID = &p } else { menu.ParentID = nil }
			case int64:
				if v > 0 { menu.ParentID = &v } else { menu.ParentID = nil }
			case string:
				if parsed, errP := strconv.ParseInt(v, 10, 64); errP == nil && parsed > 0 {
					menu.ParentID = &parsed
				} else {
					menu.ParentID = nil
				}
			}
		} else {
			menu.ParentID = nil
		}

		if t, ok := req["title"].(string); ok { menu.Title = t }
		if ic, ok := req["icon"].(string); ok { menu.Icon = ic }
		if p, ok := req["path"].(string); ok { menu.Path = p }
		if oVal, ok := req["order"]; ok && oVal != nil {
			switch v := oVal.(type) {
			case float64:
				menu.Order = int(v)
			case int64:
				menu.Order = int(v)
			}
		} else if oSeqVal, ok := req["order_seq"]; ok && oSeqVal != nil {
			switch v := oSeqVal.(type) {
			case float64:
				menu.Order = int(v)
			case int64:
				menu.Order = int(v)
			}
		}

		if err := h.db.Save(&menu).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": menu})
	case "role-menus":
		var data models.RoleMenu
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "parameters":
		var data models.GlobalParameter
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "users":
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var user models.User
		if idVal, ok := req["id"]; ok && idVal != nil {
			idNum := int64(0)
			switch v := idVal.(type) {
			case float64:
				idNum = int64(v)
			case int64:
				idNum = v
			}
			if idNum > 0 {
				h.db.Where("id = ?", idNum).First(&user)
			}
		}
		if username, ok := req["username"].(string); ok && username != "" {
			user.Username = username
		}
		if name, ok := req["name"].(string); ok && name != "" {
			user.Name = name
		}
		if role, ok := req["role"].(string); ok && role != "" {
			user.Role = role
		}
		if memNo, ok := req["member_no"]; ok && memNo != nil {
			switch v := memNo.(type) {
			case float64:
				m := int64(v)
				user.MemberNo = &m
			case int64:
				user.MemberNo = &v
			}
		} else {
			user.MemberNo = nil
		}
		if pwd, ok := req["password"].(string); ok && strings.TrimSpace(pwd) != "" {
			hashed, errH := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(pwd)), bcrypt.DefaultCost)
			if errH == nil {
				user.Password = string(hashed)
				user.PasswordChangedAt = time.Now()
			}
		} else if user.ID == 0 {
			hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			user.Password = string(hashed)
			user.PasswordChangedAt = time.Now()
		}

		if err := h.db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": user})
	case "sessions":
		var data models.Session
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown master table"})
	}
}

func (h *MasterDataHandler) Delete(c *gin.Context) {
	table := c.Param("table")
	id := c.Param("id")

	var err error
	switch table {
	case "departments":
		err = h.db.Where("deptno = ?", id).Delete(&models.Department{}).Error
	case "employee-statuses":
		err = h.db.Where("status_code = ?", id).Delete(&models.EmployeeStatus{}).Error
	case "kopkara-statuses":
		err = h.db.Where("status_code = ?", id).Delete(&models.KopkaraStatus{}).Error
	case "employee-categories":
		err = h.db.Where("category_code = ?", id).Delete(&models.EmployeeCategory{}).Error
	case "employees":
		err = h.db.Where("employee_id = ?", id).Delete(&models.Employee{}).Error
	case "members":
		err = h.db.Where("member_no = ?", id).Delete(&models.Member{}).Error
	case "roles":
		err = h.db.Where("role_id = ?", id).Delete(&models.Role{}).Error
	case "menus":
		err = h.db.Where("menu_id = ?", id).Delete(&models.Menu{}).Error
	case "role-menus":
		err = h.db.Where("role_id = ? AND menu_id = ?", c.Query("role_id"), c.Query("menu_id")).Delete(&models.RoleMenu{}).Error
	case "parameters":
		err = h.db.Where("id = ?", id).Delete(&models.GlobalParameter{}).Error
	case "users":
		err = h.db.Where("id = ?", id).Delete(&models.User{}).Error
	case "sessions":
		err = h.db.Where("id = ?", id).Delete(&models.Session{}).Error
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown master table"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (h *MasterDataHandler) UnlockUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User account unlocked successfully"})
}

func (h *MasterDataHandler) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	hashed, errH := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if errH != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password":             string(hashed),
		"password_changed_at": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Password reset to default 'password123' successfully"})
}

func (h *MasterDataHandler) RevokeSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Model(&models.Session{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Session revoked successfully"})
}

func (h *MasterDataHandler) CleanupSessions(c *gin.Context) {
	if err := h.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Expired sessions cleaned up successfully"})
}

func (h *MasterDataHandler) GetUserInfo(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	if employeeIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id required"})
		return
	}

	var emp models.Employee
	// 1. Query employee table with filter
	errEmp := h.db.Where("employee_id = ?", employeeIDStr).First(&emp).Error
	
	// 2. Query member table with filter to check membership
	var mem models.Member
	isMember := false
	if errMem := h.db.Where("employee_id = ?", employeeIDStr).First(&mem).Error; errMem == nil {
		isMember = true
	}

	var roleID int64
	var roleName string = "Bukan Anggota"

	if errEmp == nil && emp.RoleID > 0 {
		roleID = emp.RoleID
	} else if isMember {
		roleID = 2 // Default Anggota for registered members
	} else if employeeIDStr == "admin" || employeeIDStr == "9999" {
		roleID = 1 // Admin
	}

	if roleID > 0 {
		var role models.Role
		if errRole := h.db.Where("role_id = ?", roleID).First(&role).Error; errRole == nil {
			roleName = role.RoleName
		} else if roleID == 1 {
			roleName = "Admin"
		} else if roleID == 2 {
			roleName = "Anggota"
		}
	}

	// 3. Query menus for this role using SQL JOIN filter
	var menus []models.Menu
	if roleID > 0 {
		h.db.Table("lms_sch.menus").
			Select("lms_sch.menus.menu_id, lms_sch.menus.parent_id, lms_sch.menus.title, lms_sch.menus.icon, lms_sch.menus.path, lms_sch.menus.order_seq").
			Joins("JOIN lms_sch.role_menus ON lms_sch.menus.menu_id = lms_sch.role_menus.menu_id").
			Where("lms_sch.role_menus.role_id = ?", roleID).
			Order("lms_sch.menus.order_seq asc").
			Scan(&menus)
	}

	name := emp.Name
	if name == "" {
		name = "Karyawan " + employeeIDStr
	}

	c.JSON(http.StatusOK, gin.H{
		"employee_id": employeeIDStr,
		"name":        name,
		"role_id":     roleID,
		"role_name":   roleName,
		"is_member":   isMember,
		"menus":       menus,
	})
}

func (h *MasterDataHandler) CheckEmployeeMember(c *gin.Context) {
	empIDStr := c.Param("employee_id")
	empID, err := strconv.ParseInt(empIDStr, 10, 64)
	if err != nil || empID <= 0 {
		c.JSON(http.StatusOK, gin.H{"exists": false})
		return
	}

	var member models.Member
	if err := h.db.Where("employee_id = ?", empID).First(&member).Error; err == nil && member.MemberNo > 0 {
		c.JSON(http.StatusOK, gin.H{
			"exists":    true,
			"member_no": member.MemberNo,
			"message":   fmt.Sprintf("employee_id %d sudah menjadi anggota dengan nomor anggota %d", empID, member.MemberNo),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": false})
}
