package cache

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"lms-backend/models"

	"gorm.io/gorm"
)

var ParameterCache = &parameterCache{}

type parameterCache struct {
	mu     sync.RWMutex
	cache  map[string]string
	loaded bool
}

func (c *parameterCache) Init(db *gorm.DB) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]string)
	if db != nil {
		var params []models.GlobalParameter
		if err := db.Where("deleted_at IS NULL").Find(&params).Error; err == nil {
			for _, p := range params {
				c.cache[strings.TrimSpace(p.KeyName)] = strings.TrimSpace(p.KeyValue)
			}
			c.loaded = true
			log.Printf("[PARAM-CACHE] RAM Cache initialized with %d global parameters.", len(params))
			return
		}
	}
	c.loaded = true
}

func (c *parameterCache) Get(key string, defaultVal string, db ...*gorm.DB) string {
	key = strings.TrimSpace(key)
	c.mu.RLock()
	if c.loaded && c.cache != nil {
		if val, exists := c.cache[key]; exists {
			c.mu.RUnlock()
			return val
		}
		c.mu.RUnlock()
		// Cache is fully loaded from DB. Missing key means it is not in DB table.
		// Cache defaultVal in RAM memory so we NEVER query DB for missing keys.
		c.Set(key, defaultVal)
		return defaultVal
	}
	c.mu.RUnlock()

	// Only query DB if cache is NOT YET loaded
	if len(db) > 0 && db[0] != nil {
		var param models.GlobalParameter
		if err := db[0].Where("key_name = ? AND deleted_at IS NULL", key).First(&param).Error; err == nil && strings.TrimSpace(param.KeyValue) != "" {
			val := strings.TrimSpace(param.KeyValue)
			c.Set(key, val)
			return val
		}
	}

	c.Set(key, defaultVal)
	return defaultVal
}

func (c *parameterCache) GetInt(key string, defaultVal int, db ...*gorm.DB) int {
	strVal := c.Get(key, "", db...)
	if strVal == "" {
		return defaultVal
	}
	if parsed, err := strconv.Atoi(strVal); err == nil {
		return parsed
	}
	return defaultVal
}

func (c *parameterCache) Set(key string, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = make(map[string]string)
	}
	c.cache[key] = value
}

func (c *parameterCache) Delete(key string) {
	key = strings.TrimSpace(key)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache != nil {
		delete(c.cache, key)
	}
}

func (c *parameterCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]string)
	c.loaded = false
}
