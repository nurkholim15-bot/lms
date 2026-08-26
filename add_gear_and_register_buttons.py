# Python script to add Gear icon ⚙️ and Register Mobile button to Login Card in App.jsx

with open('frontend/src/App.jsx', 'r', encoding='utf-8') as f:
    code = f.read()

# 1. Add gear button to login card header
old_card_start = """        <div style={{ width: '420px', padding: '36px', background: '#f8fafc', borderRadius: '24px', boxShadow: '0 8px 32px rgba(0, 0, 0, 0.25)' }}>"""

new_card_start = """        <div style={{ width: '420px', padding: '36px', background: '#f8fafc', borderRadius: '24px', boxShadow: '0 8px 32px rgba(0, 0, 0, 0.25)', position: 'relative' }}>
          <button
            type="button"
            onClick={() => setServerModalOpen(true)}
            title="Pengaturan Server API (IP & Port HP Mobile)"
            style={{
              position: 'absolute', top: '16px', right: '16px',
              background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '50%',
              cursor: 'pointer', fontSize: '1.2rem', width: '36px', height: '36px',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
            }}
          >
            ⚙️
          </button>"""

code = code.replace(old_card_start, new_card_start)

# 2. Add Register Mobile EWA button under Login form
old_forgot = """            <div style={{ textAlign: 'center', marginTop: '4px' }}>
              <button 
                type="button" 
                onClick={() => alert('Akun Tersedia (Simulator):\\n- User Anggota: ID Angka Berapa Saja (Contoh: 10101) / password123\\n- User Admin: admin / admin123\\n- User HRD: hrd / hrd123')}
                style={{ background: 'none', border: 'none', color: '#6366f1', fontSize: '0.85rem', cursor: 'pointer', fontWeight: 600, textDecoration: 'underline' }}
              >
                Forgot?
              </button>
            </div>"""

new_forgot = """            <button 
              type="button" 
              onClick={() => {
                setCurrentUser({ role: 'anggota', username: 'mobile_user', employee_id: 100001 });
                setActiveTab('mobile-app-enterprise');
              }}
              style={{ width: '100%', padding: '12px', background: '#2563eb', color: '#ffffff', border: 'none', borderRadius: '12px', fontSize: '0.9rem', fontWeight: 800, cursor: 'pointer', marginTop: '4px' }}
            >
              📱 Register / Login Mobile EWA (PIN 6-Digit)
            </button>

            <div style={{ textAlign: 'center', marginTop: '4px' }}>
              <button 
                type="button" 
                onClick={() => alert('Akun Tersedia (Simulator):\\n- User Anggota: ID Angka Berapa Saja (Contoh: 10101) / password123\\n- User Admin: admin / admin123\\n- User HRD: hrd / hrd123')}
                style={{ background: 'none', border: 'none', color: '#6366f1', fontSize: '0.85rem', cursor: 'pointer', fontWeight: 600, textDecoration: 'underline' }}
              >
                Forgot?
              </button>
            </div>"""

code = code.replace(old_forgot, new_forgot)

with open('frontend/src/App.jsx', 'w', encoding='utf-8') as f:
    f.write(code)

print("Successfully added Gear ⚙️ button and Mobile Register button to Login Card!")
