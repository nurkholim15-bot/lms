package handlers

import (
	"lms-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MasterDataHandler struct {
	db *gorm.DB
}

func NewMasterDataHandler(db *gorm.DB) *MasterDataHandler {
	return &MasterDataHandler{db: db}
}

func (h *MasterDataHandler) GetAll(c *gin.Context) {
	table := c.Param("table")

	// Dynamic Filtering Map
	filters := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && values[0] != "" {
			filters[key] = values[0]
		}
	}
	query := h.db.Where(filters)

	switch table {
	case "departments":
		var data []models.Department
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employee-statuses":
		var data []models.EmployeeStatus
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "kopkara-statuses":
		var data []models.KopkaraStatus
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employee-categories":
		var data []models.EmployeeCategory
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "employees":
		var data []models.Employee
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "members":
		var data []models.Member
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "roles":
		var data []models.Role
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "menus":
		var data []models.Menu
		query.Order("order_seq asc").Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "role-menus":
		var data []models.RoleMenu
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
	case "parameters":
		var data []models.GlobalParameter
		query.Find(&data)
		c.JSON(http.StatusOK, gin.H{"data": data})
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
		if err := h.db.Save(&data).Error; err != nil {
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
		var data models.Menu
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.db.Save(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": data})
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
