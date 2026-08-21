import re

# 1. Update backend/handlers/master_data_handler.go for id ASC
with open('handlers/master_data_handler.go', 'r', encoding='utf-8') as f:
    go_code = f.read()

go_code = go_code.replace('sQuery.Order("id DESC")', 'sQuery.Order("id ASC")')

with open('handlers/master_data_handler.go', 'w', encoding='utf-8') as f:
    f.write(go_code)

print("Successfully updated master_data_handler.go sessions ORDER BY id ASC!")
