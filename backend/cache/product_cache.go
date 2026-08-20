package cache

import (
	"log"
	"sync"

	"lms-backend/models"
)

// ProductCache adalah in-memory cache untuk tabel loan_products.
var ProductCache = &productCache{}

type productCache struct {
	mu     sync.RWMutex
	data   []models.LoanProduct
	loaded bool
}

// Set menyimpan data produk ke cache.
func (c *productCache) Set(products []models.LoanProduct) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(products) == 0 {
		c.data = nil
		c.loaded = false
		log.Println("[PRODUCT-CACHE] Database produk kosong, cache ditandai un-loaded.")
		return
	}
	c.data = products
	c.loaded = true
	log.Printf("[PRODUCT-CACHE] Cache diperbarui: %d produk dimuat ke RAM.", len(products))
}

// Get mengembalikan data cache dan statusnya.
func (c *productCache) Get() ([]models.LoanProduct, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded || len(c.data) == 0 {
		return nil, false
	}
	result := make([]models.LoanProduct, len(c.data))
	copy(result, c.data)
	return result, true
}

// GetByID mencari produk dari cache.
func (c *productCache) GetByID(id int64) (models.LoanProduct, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded || len(c.data) == 0 {
		return models.LoanProduct{}, false
	}
	for _, p := range c.data {
		if p.ID == id {
			return p, true
		}
	}
	return models.LoanProduct{}, false
}

// Invalidate mengosongkan cache.
func (c *productCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = nil
	c.loaded = false
	log.Println("[PRODUCT-CACHE] Cache produk di-invalidate.")
}

func (c *productCache) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded && len(c.data) > 0
}
