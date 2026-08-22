# Python script to implement in-memory cache lookup for APP_TOKEN and global Axios token interceptor

# 1. Update backend/main.go
with open('backend/main.go', 'r', encoding='utf-8') as f:
    main_code = f.read()

# Replace getAppTokenName with cached version
old_get_token_name = """func getAppTokenName() string {
	var tokenName string
	if config.DB != nil {
		config.DB.Raw("SELECT key_value FROM lms_sch.global_parameters WHERE key_name = 'APP_TOKEN' AND deleted_at IS NULL LIMIT 1").Scan(&tokenName)
	}
	tokenName = strings.TrimSpace(tokenName)
	if tokenName == "" {
		tokenName = "ewa_token"
	}
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
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return strings.TrimSpace(authHeader)
}"""

new_get_token_name = """var (
	cachedAppTokenName string
	cachedAppTokenMutex sync.RWMutex
)

func getAppTokenNameFromCache(paramRepo repositories.ParameterRepository) string {
	cachedAppTokenMutex.RLock()
	if cachedAppTokenName != "" {
		defer cachedAppTokenMutex.RUnlock()
		return cachedAppTokenName
	}
	cachedAppTokenMutex.RUnlock()

	tokenName := "ewa_token"
	if paramRepo != nil {
		if p, err := paramRepo.FindByKey("APP_TOKEN"); err == nil && strings.TrimSpace(p.KeyValue) != "" {
			tokenName = strings.TrimSpace(p.KeyValue)
		}
	}

	cachedAppTokenMutex.Lock()
	cachedAppTokenName = tokenName
	cachedAppTokenMutex.Unlock()
	return tokenName
}

func getTokenFromRequest(c *gin.Context, paramRepo repositories.ParameterRepository) string {
	tokenName := getAppTokenNameFromCache(paramRepo)
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
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return strings.TrimSpace(authHeader)
}"""

main_code = main_code.replace(old_get_token_name, new_get_token_name)

# Update callers of getTokenFromRequest in main.go
main_code = main_code.replace("getTokenFromRequest(c)", "getTokenFromRequest(c, paramRepo)")
main_code = main_code.replace("tokenCookieName := getAppTokenName()", "tokenCookieName := getAppTokenNameFromCache(paramRepo)")

with open('backend/main.go', 'w', encoding='utf-8') as f:
    f.write(main_code)

print("Successfully updated main.go token cache!")

# 2. Update frontend/src/App.jsx to add global Axios Interceptor & token persistence in sessionStorage
with open('frontend/src/App.jsx', 'r', encoding='utf-8') as f:
    app_code = f.read()

old_axios_setup = """axios.defaults.withCredentials = true;"""

new_axios_setup = """axios.defaults.withCredentials = true;

// Axios Request Interceptor: Automatically attach Authorization Bearer Token to all API requests
axios.interceptors.request.use((config) => {
  config.withCredentials = true;
  const token = sessionStorage.getItem('lms_auth_token') || localStorage.getItem('lms_auth_token') || '';
  if (token && !config.headers['Authorization']) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
}, (error) => Promise.reject(error));"""

app_code = app_code.replace(old_axios_setup, new_axios_setup)

# Update verifySession & handleLogin to store token in sessionStorage
old_verify_store = """      const user = res.data.user;
      setCurrentUser(user);"""

new_verify_store = """      const user = res.data.user;
      if (overrideToken) {
        sessionStorage.setItem('lms_auth_token', overrideToken);
      }
      setCurrentUser(user);"""

app_code = app_code.replace(old_verify_store, new_verify_store)

old_login_store = """      const receivedToken = res.data?.token || '';
      await verifySession(receivedToken);"""

new_login_store = """      const receivedToken = res.data?.token || '';
      if (receivedToken) {
        sessionStorage.setItem('lms_auth_token', receivedToken);
      }
      await verifySession(receivedToken);"""

app_code = app_code.replace(old_login_store, new_login_store)

with open('frontend/src/App.jsx', 'w', encoding='utf-8') as f:
    f.write(app_code)

print("Successfully updated App.jsx Axios interceptor!")
