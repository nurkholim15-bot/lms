import re

# 1. Update backend/handlers/master_data_handler.go
with open('handlers/master_data_handler.go', 'r', encoding='utf-8') as f:
    go_code = f.read()

# Update Delete case for sessions
old_del_sess = 'case "sessions":\n\terr = h.db.Where("id = ?", id).Delete(&models.Session{}).Error'
new_del_sess = """case "sessions":
		if id == "cleanup" || id == "clean-up" || id == "expired" {
			res := h.db.Where("expires_at < ? OR is_active = ?", time.Now(), false).Delete(&models.Session{})
			err = res.Error
		} else {
			err = h.db.Where("id = ?", id).Delete(&models.Session{}).Error
		}"""

go_code = go_code.replace(old_del_sess, new_del_sess)

# Update CleanupSessions func
old_clean_func = """func (h *MasterDataHandler) CleanupSessions(c *gin.Context) {
	if err := h.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Expired sessions cleaned up successfully"})
}"""

new_clean_func = """func (h *MasterDataHandler) CleanupSessions(c *gin.Context) {
	res := h.db.Where("expires_at < ? OR is_active = ?", time.Now(), false).Delete(&models.Session{})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("Berhasil membersihkan %d session expired / non-aktif", res.RowsAffected), "rows_affected": res.RowsAffected})
}"""

go_code = go_code.replace(old_clean_func, new_clean_func)

with open('handlers/master_data_handler.go', 'w', encoding='utf-8') as f:
    f.write(go_code)

print("Successfully updated master_data_handler.go!")
