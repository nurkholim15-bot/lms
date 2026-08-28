package cache

import (
	"strings"
	"sync"

	"lms-backend/models"

	"gorm.io/gorm"
)

var IdentityCache = &identityCache{
	usersByUsername:  make(map[string]*models.User),
	usersByID:        make(map[int64]*models.User),
	membersByNo:      make(map[int64]*models.Member),
	membersByEmp:     make(map[int64]*models.Member),
	employeesByID:    make(map[int64]*models.Employee),
	categoriesByCode: make(map[string]*models.EmployeeCategory),
}

type identityCache struct {
	mu               sync.RWMutex
	usersByUsername  map[string]*models.User
	usersByID        map[int64]*models.User
	membersByNo      map[int64]*models.Member
	membersByEmp     map[int64]*models.Member
	employeesByID    map[int64]*models.Employee
	categoriesByCode map[string]*models.EmployeeCategory
}

// SetUser caches user in RAM
func (c *identityCache) SetUser(u *models.User) {
	if u == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.usersByID[u.ID] = u
	if u.Username != "" {
		c.usersByUsername[strings.ToLower(strings.TrimSpace(u.Username))] = u
	}
	if u.PhoneNumber != nil && *u.PhoneNumber != "" {
		c.usersByUsername[strings.TrimSpace(*u.PhoneNumber)] = u
	}
}

// GetUserByUsername gets user from RAM cache or DB
func (c *identityCache) GetUserByUsername(db *gorm.DB, username string) (*models.User, error) {
	key := strings.ToLower(strings.TrimSpace(username))
	c.mu.RLock()
	if u, exists := c.usersByUsername[key]; exists && u != nil {
		c.mu.RUnlock()
		return u, nil
	}
	c.mu.RUnlock()

	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var u models.User
	phone08 := username
	phone62 := username
	if strings.HasPrefix(username, "0") {
		phone62 = "62" + username[1:]
	} else if strings.HasPrefix(username, "62") {
		phone08 = "0" + username[2:]
	}

	if err := db.Where("(LOWER(username) = ? OR phone_number = ? OR phone_number = ?) AND deleted_at IS NULL", key, phone08, phone62).First(&u).Error; err != nil {
		return nil, err
	}

	c.SetUser(&u)
	return &u, nil
}

// GetUserByID gets user from RAM cache or DB
func (c *identityCache) GetUserByID(db *gorm.DB, id int64) (*models.User, error) {
	c.mu.RLock()
	if u, exists := c.usersByID[id]; exists && u != nil {
		c.mu.RUnlock()
		return u, nil
	}
	c.mu.RUnlock()

	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var u models.User
	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}

	c.SetUser(&u)
	return &u, nil
}

// SetMember caches member in RAM
func (c *identityCache) SetMember(m *models.Member) {
	if m == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if m.MemberNo > 0 {
		c.membersByNo[m.MemberNo] = m
	}
	if m.EmployeeID > 0 {
		c.membersByEmp[m.EmployeeID] = m
	}
}

// GetMemberByNo gets member from RAM cache or DB
func (c *identityCache) GetMemberByNo(db *gorm.DB, memberNo int64) (*models.Member, error) {
	c.mu.RLock()
	if m, exists := c.membersByNo[memberNo]; exists && m != nil {
		c.mu.RUnlock()
		return m, nil
	}
	if m, exists := c.membersByEmp[memberNo]; exists && m != nil {
		c.mu.RUnlock()
		return m, nil
	}
	c.mu.RUnlock()

	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var m models.Member
	if err := db.Where("member_no = ? AND deleted_at IS NULL", memberNo).First(&m).Error; err != nil {
		if err2 := db.Where("employee_id = ? AND deleted_at IS NULL", memberNo).First(&m).Error; err2 != nil {
			return nil, err
		}
	}

	c.SetMember(&m)
	return &m, nil
}

// SetEmployee caches employee in RAM
func (c *identityCache) SetEmployee(e *models.Employee) {
	if e == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if e.EmployeeID > 0 {
		c.employeesByID[e.EmployeeID] = e
	}
}

// GetEmployeeByID gets employee from RAM cache or DB
func (c *identityCache) GetEmployeeByID(db *gorm.DB, empID int64) (*models.Employee, error) {
	c.mu.RLock()
	if e, exists := c.employeesByID[empID]; exists && e != nil {
		c.mu.RUnlock()
		return e, nil
	}
	c.mu.RUnlock()

	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var e models.Employee
	if err := db.Where("employee_id = ? AND deleted_at IS NULL", empID).First(&e).Error; err != nil {
		return nil, err
	}

	c.SetEmployee(&e)
	return &e, nil
}

// SetCategory caches category in RAM
func (c *identityCache) SetCategory(cat *models.EmployeeCategory) {
	if cat == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if cat.CategoryCode != "" {
		c.categoriesByCode[strings.ToUpper(strings.TrimSpace(cat.CategoryCode))] = cat
	}
}

// GetCategoryByCode gets category from RAM cache or DB
func (c *identityCache) GetCategoryByCode(db *gorm.DB, code string) (*models.EmployeeCategory, error) {
	key := strings.ToUpper(strings.TrimSpace(code))
	c.mu.RLock()
	if cat, exists := c.categoriesByCode[key]; exists && cat != nil {
		c.mu.RUnlock()
		return cat, nil
	}
	c.mu.RUnlock()

	if db == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var cat models.EmployeeCategory
	if err := db.Where("category_code = ? AND deleted_at IS NULL", key).First(&cat).Error; err != nil {
		return nil, err
	}

	c.SetCategory(&cat)
	return &cat, nil
}
