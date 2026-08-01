package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// Dummy Database User
var mockUsers = map[string]map[string]interface{}{
	"1001": {
		"employee_id": 1001,
		"name":        "Budi Santoso",
		"eligible":    true,
		"role":        "anggota",
		"password":    "password123",
	},
	"admin": {
		"employee_id": 9999,
		"name":        "Administrator Koperasi",
		"eligible":    false,
		"role":        "admin",
		"password":    "admin123",
	},
	"hrd": {
		"employee_id": 8888,
		"name":        "Tim HRD Adira",
		"eligible":    false,
		"role":        "hrd",
		"password":    "hrd123",
	},
}

// Dummy Token mapping
var tokenStore = map[string]string{} // token -> username

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())

	karisma := r.Group("/api/karisma")
	{
		karisma.POST("/login", func(c *gin.Context) {
			var creds struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&creds); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
				return
			}

			user, exists := mockUsers[creds.Username]
			if !exists {
				// Cek jika username adalah angka (employee_id dinamis)
				if _, err := strconv.Atoi(creds.Username); err == nil {
					// Bebas login asal passwordnya password123
					if creds.Password != "password123" {
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Password untuk Anggota harus 'password123'"})
						return
					}
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
					return
				}
			} else {
				if user["password"] != creds.Password {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
					return
				}
			}

			token := "mock-token-" + creds.Username
			tokenStore[token] = creds.Username

			c.JSON(http.StatusOK, gin.H{"token": token})
		})

		karisma.POST("/verify", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
				return
			}
			token := authHeader[7:]
			username, exists := tokenStore[token]
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}

			user, userExists := mockUsers[username]
			if !userExists {
				if empID, err := strconv.Atoi(username); err == nil {
					user = map[string]interface{}{
						"employee_id": empID,
						"name":        "Karyawan " + username,
						"eligible":    true,
						"role":        "anggota",
					}
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User data not found"})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"user": gin.H{
					"employee_id": user["employee_id"],
					"name":        user["name"],
					"eligible":    user["eligible"],
					"role":        user["role"],
				},
			})
		})
	}

	log.Printf("Karisma Simulator is starting on port 8087")
	if err := r.Run(":8087"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
