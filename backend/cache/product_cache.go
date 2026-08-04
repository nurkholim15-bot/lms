package cache

import (
	"log"
	"sync"

	"lms-backend/models"
)

// ProductCache adalah in-memory cache untuk tabel loan_products.
// Cache diload sekali saat backend startup dan di-invalidate setiap kali
// ada operasi write (Create/Save/Delete), sehingga query berikutnya
// akan reload dari database.
var ProductCache = &productCache{}

type productCache struct {
	mu       sync.RWMutex
	data     []models.LoanProduct
	loaded   bool
}

// Set menyimpan seluruh data produk ke cache.
func (c *productCache) Set(products []models.LoanProduct) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = products
	c.loaded = true
	log.Printf("[PRODUCT-CACHE] Cache diperbarui: %d produk dimuat.", len(products))
}

// Get mengembalikan data cache dan apakah cache sudah terisi.
func (c *productCache) Get() ([]models.LoanProduct, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return nil, false
	}
	// Kembalikan salinan slice agar aman dari modifikasi luar
	result := make([]models.LoanProduct, len(c.data))
	copy(result, c.data)
	return result, true
}

// GetByID mencari produk berdasarkan ID dari cache.
// Mengembalikan produk dan true jika ditemukan.
func (c *productCache) GetByID(id int64) (models.LoanProduct, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return models.LoanProduct{}, false
	}
	for _, p := range c.data {
		if p.ID == id {
			return p, true
		}
	}
	return models.LoanProduct{}, false
}

// Invalidate mengosongkan cache sehingga request berikutnya
// akan reload dari database.
func (c *productCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = nil
	c.loaded = false
	log.Println("[PRODUCT-CACHE] Cache di-invalidate.")
}

// IsLoaded mengembalikan status apakah cache sudah terisi.
func (c *productCache) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}
