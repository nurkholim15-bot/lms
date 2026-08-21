with open('main.go', 'r', encoding='utf-8') as f:
    main_code = f.read()

# Update RBACRoleMenuMiddleware in main.go
old_rbac_block = """		// Look up real session and user role from DB if token is a session token
		if config.DB != nil && !strings.HasPrefix(token, "mock-token-") {
			var sessionRecord models.Session
			if errSess := config.DB.Where("token = ? AND is_active = ?", token, true).First(&sessionRecord).Error; errSess == nil {
				username = sessionRecord.Username
				var u models.User
				if errU := config.DB.Where("id = ?", sessionRecord.UserID).First(&u).Error; errU == nil {
					roleLower := strings.ToLower(strings.TrimSpace(u.Role))
					if roleLower == "admin" {
						roleID = 1
					} else if roleLower == "hrd" {
						roleID = 3
					} else if roleLower == "anggota" {
						roleID = 2
					}
				}
			}
		}

		usernameLower := strings.ToLower(strings.TrimSpace(username))

		if usernameLower == "admin" || usernameLower == "9999" {
			roleID = 1 // Role 1: Admin
		} else if usernameLower == "hrd" {
			roleID = 3 // Role 3: HRD
		} else if empID, err := strconv.ParseInt(usernameLower, 10, 64); err == nil {
			var dbRoleID int64 = 0
			config.DB.Raw("SELECT role_id FROM lms_sch.employees WHERE employee_id = ? LIMIT 1", empID).Scan(&dbRoleID)
			if dbRoleID > 0 {
				roleID = dbRoleID
			}
			if empID == 10101 && roleID != 1 {
				roleID = 2 // Role 2: Anggota
			}
		}"""

new_rbac_block = """		// Look up real session and user role from DB if token is a session token
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
			config.DB.Raw("SELECT role_id FROM lms_sch.users WHERE LOWER(username) = ? AND deleted_at IS NULL LIMIT 1", usernameLower).Scan(&userRoleID)
			if userRoleID > 0 {
				roleID = userRoleID
			}
		}"""

main_code = main_code.replace(old_rbac_block, new_rbac_block)

# Update verify token in main.go
old_verify_user = """						var user models.User
						if errU := config.DB.Where("id = ?", session.UserID).First(&user).Error; errU == nil {
							empID := int64(0)
							name := user.Name
							role := user.Role"""

new_verify_user = """						var user models.User
						if errU := config.DB.Where("id = ?", session.UserID).First(&user).Error; errU == nil {
							empID := int64(0)
							name := user.Name
							role := "anggota"
							if user.RoleID > 0 {
								var r models.Role
								if errR := config.DB.Where("role_id = ?", user.RoleID).First(&r).Error; errR == nil {
									role = r.RoleName
								}
							}"""

main_code = main_code.replace(old_verify_user, new_verify_user)

with open('main.go', 'w', encoding='utf-8') as f:
    f.write(main_code)

print("Successfully updated main.go!")

# Update handlers/master_data_handler.go
with open('handlers/master_data_handler.go', 'r', encoding='utf-8') as f:
    handler_code = f.read()

# Update case "users" in GetAll
old_users_get = """	case "users":
		var data []models.User
		uQuery := h.db.Model(&models.User{}).Where("deleted_at IS NULL")
		if q != "" {
			uIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if uIDSearch > 0 {
				uQuery = uQuery.Where("(id = ? OR member_no = ? OR LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(role) LIKE ?)", uIDSearch, uIDSearch, likeStr, likeStr, likeStr)
			} else {
				uQuery = uQuery.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ? OR LOWER(role) LIKE ?)", likeStr, likeStr, likeStr)
			}
		}"""

new_users_get = """	case "users":
		var data []models.User
		uQuery := h.db.Model(&models.User{}).Where("deleted_at IS NULL")
		if q != "" {
			uIDSearch, _ := strconv.ParseInt(q, 10, 64)
			if uIDSearch > 0 {
				uQuery = uQuery.Where("(id = ? OR member_no = ? OR role_id = ? OR LOWER(username) LIKE ? OR LOWER(name) LIKE ?)", uIDSearch, uIDSearch, uIDSearch, likeStr, likeStr)
			} else {
				uQuery = uQuery.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ?)", likeStr, likeStr)
			}
		}"""

handler_code = handler_code.replace(old_users_get, new_users_get)

# Update case "users" in Save
old_users_save = """		if role, ok := req["role"].(string); ok && role != "" {
			user.Role = role
		}"""

new_users_save = """		if roleIDVal, ok := req["role_id"]; ok && roleIDVal != nil {
			switch v := roleIDVal.(type) {
			case float64:
				user.RoleID = int64(v)
			case int64:
				user.RoleID = v
			}
		} else if roleStr, ok := req["role"].(string); ok && roleStr != "" {
			var r models.Role
			if errR := h.db.Where("LOWER(role_name) = ?", strings.ToLower(roleStr)).First(&r).Error; errR == nil {
				user.RoleID = r.RoleID
			} else if strings.ToLower(roleStr) == "admin" {
				user.RoleID = 1
			} else if strings.ToLower(roleStr) == "hrd" {
				user.RoleID = 3
			} else {
				user.RoleID = 2
			}
		}
		if user.RoleID <= 0 {
			user.RoleID = 2
		}"""

handler_code = handler_code.replace(old_users_save, new_users_save)

# Update GetUserInfo in master_data_handler.go
old_user_info = """	if errEmp == nil && emp.RoleID > 0 {
		roleID = emp.RoleID
	} else if isMember {
		roleID = 2 // Default Anggota for registered members
	} else if employeeIDStr == "admin" || employeeIDStr == "9999" {
		roleID = 1 // Admin
	}"""

new_user_info = """	var userRecord models.User
	if errU := h.db.Where("username = ? OR id = ?", employeeIDStr, employeeIDStr).First(&userRecord).Error; errU == nil && userRecord.RoleID > 0 {
		roleID = userRecord.RoleID
	} else if isMember {
		roleID = 2 // Default Anggota for registered members
	} else if employeeIDStr == "admin" || employeeIDStr == "9999" {
		roleID = 1 // Admin
	}"""

handler_code = handler_code.replace(old_user_info, new_user_info)

with open('handlers/master_data_handler.go', 'w', encoding='utf-8') as f:
    f.write(handler_code)

print("Successfully updated master_data_handler.go!")
