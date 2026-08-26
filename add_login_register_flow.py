# Python script to add Register link & Self-Service Flow to App.jsx

with open('frontend/src/App.jsx', 'r', encoding='utf-8') as f:
    code = f.read()

# Replace Login Card rendering with Register Link & EWA Self-Service Toggle
old_login_card = """          <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {loginError && <div style={{ padding: '12px', background: '#fee2e2', color: '#991B1B', borderRadius: '8px', fontSize: '0.875rem' }}>{loginError}</div>}
            
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <label style={{ width: '100px', fontSize: '0.75rem', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em' }}>USERNAME</label>
              <input 
                type="text" required 
                value={loginForm.username} 
                onChange={e => setLoginForm({...loginForm, username: e.target.value})} 
                style={{ flex: 1, padding: '10px 14px', borderRadius: '10px', border: '1px solid #bfdbfe', background: '#f0f7ff', fontSize: '0.95rem', outline: 'none' }} 
                placeholder="Username" 
              />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <label style={{ width: '100px', fontSize: '0.75rem', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em' }}>PASSWORD</label>
              <div style={{ flex: 1, position: 'relative', display: 'flex', alignItems: 'center' }}>
                <input 
                  type={showPassword ? 'text' : 'password'} required 
                  value={loginForm.password} 
                  onChange={e => setLoginForm({...loginForm, password: e.target.value})} 
                  style={{ width: '100%', padding: '10px 38px 10px 14px', borderRadius: '10px', border: '1px solid #bfdbfe', background: '#f0f7ff', fontSize: '0.95rem', outline: 'none' }} 
                  placeholder="••••••••" 
                />
                <button 
                  type="button" 
                  onClick={() => setShowPassword(!showPassword)}
                  title={showPassword ? 'Sembunyikan Password' : 'Tampilkan Password'}
                  style={{ position: 'absolute', right: '10px', background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.15rem', color: '#64748b', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '2px', userSelect: 'none' }}
                >
                  {showPassword ? '👁️' : '👁️‍🗨️'}
                </button>
              </div>
            </div>
            
            <div style={{ display: 'flex', gap: '12px', marginTop: '12px' }}>
              <button 
                type="button" 
                onClick={() => alert('Ganti Password: Masukkan Username & Password lama Anda, lalu hubungi admin atau ubah di menu profil setelah login.')}
                style={{ flex: 1, padding: '12px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '12px', fontSize: '0.95rem', fontWeight: 700, cursor: 'pointer' }}
              >
                Ganti Password
              </button>
              
              <button 
                type="submit" 
                disabled={loginLoading}
                style={{ flex: 1, padding: '12px', background: 'var(--button-bg, #0284c7)', color: '#ffffff', border: 'none', borderRadius: '12px', fontSize: '0.95rem', fontWeight: 700, cursor: 'pointer' }}
              >
                {loginLoading ? 'Loading...' : 'Login'}
              </button>
            </div>
            <div style={{ textAlign: 'center', marginTop: '8px' }}>
              <span 
                onClick={() => alert('Lupa Password: Silakan hubungi tim Admin IT Kopkara untuk mereset akun Anda.')} 
                style={{ color: '#4338ca', fontSize: '0.85rem', cursor: 'pointer', textDecoration: 'underline' }}
              >
                Forgot?
              </span>
            </div>
          </form>"""

new_login_card = """          <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {loginError && <div style={{ padding: '12px', background: '#fee2e2', color: '#991B1B', borderRadius: '8px', fontSize: '0.875rem' }}>{loginError}</div>}
            
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <label style={{ width: '100px', fontSize: '0.75rem', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em' }}>USERNAME</label>
              <input 
                type="text" required 
                value={loginForm.username} 
                onChange={e => setLoginForm({...loginForm, username: e.target.value})} 
                style={{ flex: 1, padding: '10px 14px', borderRadius: '10px', border: '1px solid #bfdbfe', background: '#f0f7ff', fontSize: '0.95rem', outline: 'none' }} 
                placeholder="Username" 
              />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <label style={{ width: '100px', fontSize: '0.75rem', fontWeight: 700, color: '#475569', textTransform: 'uppercase', letterSpacing: '0.05em' }}>PASSWORD</label>
              <div style={{ flex: 1, position: 'relative', display: 'flex', alignItems: 'center' }}>
                <input 
                  type={showPassword ? 'text' : 'password'} required 
                  value={loginForm.password} 
                  onChange={e => setLoginForm({...loginForm, password: e.target.value})} 
                  style={{ width: '100%', padding: '10px 38px 10px 14px', borderRadius: '10px', border: '1px solid #bfdbfe', background: '#f0f7ff', fontSize: '0.95rem', outline: 'none' }} 
                  placeholder="••••••••" 
                />
                <button 
                  type="button" 
                  onClick={() => setShowPassword(!showPassword)}
                  title={showPassword ? 'Sembunyikan Password' : 'Tampilkan Password'}
                  style={{ position: 'absolute', right: '10px', background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.15rem', color: '#64748b', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '2px', userSelect: 'none' }}
                >
                  {showPassword ? '👁️' : '👁️‍🗨️'}
                </button>
              </div>
            </div>
            
            <div style={{ display: 'flex', gap: '12px', marginTop: '12px' }}>
              <button 
                type="button" 
                onClick={() => alert('Ganti Password: Masukkan Username & Password lama Anda, lalu hubungi admin atau ubah di menu profil setelah login.')}
                style={{ flex: 1, padding: '12px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '12px', fontSize: '0.95rem', fontWeight: 700, cursor: 'pointer' }}
              >
                Ganti Password
              </button>
              
              <button 
                type="submit" 
                disabled={loginLoading}
                style={{ flex: 1, padding: '12px', background: 'var(--button-bg, #0284c7)', color: '#ffffff', border: 'none', borderRadius: '12px', fontSize: '0.95rem', fontWeight: 700, cursor: 'pointer' }}
              >
                {loginLoading ? 'Loading...' : 'Login'}
              </button>
            </div>
            
            {/* Tombol / Link Pendaftaran & Login Mobile EWA */}
            <div style={{ marginTop: '14px', paddingTop: '14px', borderTop: '1px solid #e2e8f0', textAlign: 'center' }}>
              <button
                type="button"
                onClick={() => setActiveTab('mobile-simulator')}
                style={{
                  width: '100%',
                  padding: '10px',
                  background: '#eff6ff',
                  border: '1px solid #3b82f6',
                  color: '#1d4ed8',
                  borderRadius: '10px',
                  fontSize: '0.85rem',
                  fontWeight: 700,
                  cursor: 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '6px'
                }}
              >
                📱 Belum punya Akun EWA? Register / Login PIN di sini
              </button>
            </div>

            <div style={{ textAlign: 'center', marginTop: '4px' }}>
              <span 
                onClick={() => alert('Lupa Password: Silakan hubungi tim Admin IT Kopkara untuk mereset akun Anda.')} 
                style={{ color: '#4338ca', fontSize: '0.85rem', cursor: 'pointer', textDecoration: 'underline' }}
              >
                Forgot?
              </span>
            </div>
          </form>"""

code = code.replace(old_login_card, new_login_card)

with open('frontend/src/App.jsx', 'w', encoding='utf-8') as f:
    f.write(code)

print("Successfully updated App.jsx with Register link on Login Card!")
