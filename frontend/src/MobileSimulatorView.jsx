import React, { useState } from 'react';
import axios from 'axios';

export default function MobileSimulatorView({ apiBaseUrl = 'https://localhost:8086' }) {
  const [activeStep, setActiveStep] = useState(1);

  // Form States
  const [registerCheckForm, setRegisterCheckForm] = useState({
    no_ktp: '234567',
    employee_id: 100001,
    name: 'Nur Kholim',
    phone_number: '085882500073'
  });

  const [otpForm, setOtpForm] = useState({
    phone_number: '085882500073',
    otp_code: '123456',
    channel: 'whatsapp'
  });

  const [pinForm, setPinForm] = useState({
    employee_id: 100001,
    no_ktp: '234567',
    phone_number: '085882500073',
    pin: '859204'
  });

  const [loginForm, setLoginForm] = useState({
    phone_number: '085882500073',
    pin: '859204'
  });

  // Response States
  const [loading, setLoading] = useState(false);
  const [apiResponse, setApiResponse] = useState(null);
  const [apiError, setApiError] = useState(null);
  const [loggedInMobileUser, setLoggedInMobileUser] = useState(null);

  const handleRegisterCheck = async (e) => {
    e.preventDefault();
    setLoading(true);
    setApiResponse(null);
    setApiError(null);
    try {
      const payload = {
        ...registerCheckForm,
        employee_id: parseInt(registerCheckForm.employee_id) || 0
      };
      const res = await axios.post(`${apiBaseUrl}/api/v1/auth/mobile-register-check`, payload);
      setApiResponse(res.data);

      if (res.data && res.data.status === 'MATCH_SUCCESS') {
        if (res.data.is_registered && res.data.has_pin) {
          alert('ℹ️ Anda SUDAH TERDAFTAR di sistem EWA!\n\nSilakan langsung login menggunakan Nomor Handphone & PIN 6-Digit Anda.');
          setLoginForm(prev => ({ ...prev, phone_number: registerCheckForm.phone_number }));
          setActiveStep(4); // Switch directly to PIN Login!
        } else {
          alert('✅ Data karyawan terverifikasi valid dengan HRD!\n\nSilakan lanjutkan verifikasi WhatsApp OTP dan pendaftaran PIN 6-Digit Anda.');
          setOtpForm(prev => ({ ...prev, phone_number: registerCheckForm.phone_number }));
          setPinForm(prev => ({
            ...prev,
            employee_id: registerCheckForm.employee_id,
            no_ktp: registerCheckForm.no_ktp,
            phone_number: registerCheckForm.phone_number
          }));
          setActiveStep(2); // Switch to OTP & Setup PIN!
        }
      }
    } catch (err) {
      setApiError(err.response?.data || { error: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleRequestOTP = async () => {
    setLoading(true);
    setApiResponse(null);
    setApiError(null);
    try {
      const res = await axios.post(`${apiBaseUrl}/api/v1/auth/request-otp`, {
        phone_number: otpForm.phone_number,
        channel: otpForm.channel
      });
      setApiResponse(res.data);
      alert(`📲 ${res.data.message || 'Kode OTP telah dikirimkan ke WhatsApp Anda!'}`);
    } catch (err) {
      setApiError(err.response?.data || { error: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOTP = async (e) => {
    e.preventDefault();
    setLoading(true);
    setApiResponse(null);
    setApiError(null);
    try {
      const res = await axios.post(`${apiBaseUrl}/api/v1/auth/verify-otp`, {
        phone_number: otpForm.phone_number,
        otp_code: otpForm.otp_code
      });
      setApiResponse(res.data);
      if (res.data && res.data.status === 'SUCCESS') {
        alert('✅ Verifikasi OTP Berhasil!\n\nSilakan daftarkan PIN 6-Digit rahasia Anda pada langkah berikutnya.');
        setActiveStep(3); // Switch to Setup PIN!
      }
    } catch (err) {
      setApiError(err.response?.data || { error: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleSetupPIN = async (e) => {
    e.preventDefault();
    setLoading(true);
    setApiResponse(null);
    setApiError(null);
    try {
      const payload = {
        ...pinForm,
        employee_id: parseInt(pinForm.employee_id) || 0
      };
      const res = await axios.post(`${apiBaseUrl}/api/v1/auth/setup-pin`, payload);
      setApiResponse(res.data);
      if (res.data && res.data.status === 'SUCCESS') {
        alert('🎉 PIN 6-Digit Berhasil Didaftarkan!\n\nAnda kini dapat login menggunakan Nomor Handphone & PIN 6-Digit.');
        setLoginForm({
          phone_number: pinForm.phone_number,
          pin: pinForm.pin
        });
        setActiveStep(4); // Switch to Mobile Login!
      }
    } catch (err) {
      setApiError(err.response?.data || { error: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleMobileLogin = async (e) => {
    e.preventDefault();
    setLoading(true);
    setApiResponse(null);
    setApiError(null);
    try {
      const res = await axios.post(`${apiBaseUrl}/api/v1/auth/mobile-login`, loginForm);
      setApiResponse(res.data);
      if (res.data && res.data.status === 'SUCCESS') {
        setLoggedInMobileUser(res.data);
      }
    } catch (err) {
      setApiError(err.response?.data || { error: err.message });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      {/* Header Banner */}
      <div style={{
        background: 'linear-gradient(135deg, #0B2545 0%, #134074 100%)',
        color: 'white',
        padding: '24px 28px',
        borderRadius: '12px',
        marginBottom: '24px',
        boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.2)'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <h2 style={{ margin: 0, fontSize: '1.4rem', fontWeight: 800, display: 'flex', alignItems: 'center', gap: '10px' }}>
              📱 Portal Self-Service & Mobile Auth EWA
            </h2>
            <p style={{ margin: '6px 0 0 0', opacity: 0.9, fontSize: '0.9rem' }}>
              Pendaftaran Karyawan Baru, Pengiriman OTP WhatsApp, Setup PIN 6-Digit, dan Mobile Login.
            </p>
          </div>
          <div style={{ background: 'rgba(255,255,255,0.15)', padding: '6px 14px', borderRadius: '20px', fontSize: '0.8rem', fontWeight: 600 }}>
            Target API: <code style={{ color: '#38bdf8' }}>{apiBaseUrl}/api/v1/auth/*</code>
          </div>
        </div>
      </div>

      {/* Step Navigation Tabs */}
      <div style={{ display: 'flex', gap: '10px', marginBottom: '24px', flexWrap: 'wrap' }}>
        {[
          { step: 1, title: 'Step 1: Cek Data Karyawan', desc: '4-Factor Matching HRD' },
          { step: 2, title: 'Step 2: Verifikasi OTP WA', desc: 'Kode OTP WhatsApp' },
          { step: 3, title: 'Step 3: Setup PIN 6-Digit', desc: 'Pendaftaran PIN Bcrypt' },
          { step: 4, title: 'Step 4: Mobile Login', desc: 'Login No. HP + PIN' }
        ].map((s) => {
          const isActive = activeStep === s.step;
          return (
            <button
              key={s.step}
              onClick={() => {
                setActiveStep(s.step);
                setApiResponse(null);
                setApiError(null);
              }}
              style={{
                flex: '1 1 200px',
                padding: '12px 16px',
                borderRadius: '8px',
                border: isActive ? '2px solid #2563eb' : '1px solid #cbd5e1',
                background: isActive ? '#eff6ff' : 'white',
                color: isActive ? '#1e40af' : '#475569',
                cursor: 'pointer',
                textAlign: 'left',
                boxShadow: isActive ? '0 4px 6px -1px rgba(37, 99, 235, 0.15)' : 'none',
                transition: 'all 0.2s'
              }}
            >
              <div style={{ fontWeight: 700, fontSize: '0.9rem' }}>{s.title}</div>
              <div style={{ fontSize: '0.75rem', opacity: 0.8, marginTop: '2px' }}>{s.desc}</div>
            </button>
          );
        })}
      </div>

      {/* Main Grid: Form Left, Response Right */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(360px, 1fr))', gap: '24px' }}>
        
        {/* Left Column: Interactive Form */}
        <div style={{ background: 'white', borderRadius: '10px', padding: '24px', boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.05)', border: '1px solid #e2e8f0' }}>
          
          {/* STEP 1: REGISTER CHECK */}
          {activeStep === 1 && (
            <form onSubmit={handleRegisterCheck}>
              <h3 style={{ margin: '0 0 16px 0', fontSize: '1.1rem', color: '#1e293b' }}>
                🔍 Step 1: Cek Status Pendaftaran Karyawan (4-Factor)
              </h3>
              <p style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '16px' }}>
                Masukkan data karyawan Anda. Sistem akan memeriksa apakah Anda sudah terdaftar atau perlu registrasi awal.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>No. KTP (NIK 16 Digit)</label>
                  <input
                    type="text"
                    value={registerCheckForm.no_ktp}
                    onChange={(e) => setRegisterCheckForm({ ...registerCheckForm, no_ktp: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Employee ID (NIP Karyawan)</label>
                  <input
                    type="number"
                    value={registerCheckForm.employee_id}
                    onChange={(e) => setRegisterCheckForm({ ...registerCheckForm, employee_id: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Nama Lengkap Karyawan</label>
                  <input
                    type="text"
                    value={registerCheckForm.name}
                    onChange={(e) => setRegisterCheckForm({ ...registerCheckForm, name: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>No. Handphone WhatsApp</label>
                  <input
                    type="text"
                    value={registerCheckForm.phone_number}
                    onChange={(e) => setRegisterCheckForm({ ...registerCheckForm, phone_number: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  style={{
                    marginTop: '8px',
                    padding: '12px',
                    background: '#2563eb',
                    color: 'white',
                    border: 'none',
                    borderRadius: '6px',
                    fontWeight: 700,
                    cursor: 'pointer'
                  }}
                >
                  {loading ? '⏳ Memproses...' : '🚀 Cek Status Pendaftaran HRD'}
                </button>
              </div>
            </form>
          )}

          {/* STEP 2: VERIFY OTP */}
          {activeStep === 2 && (
            <div>
              <h3 style={{ margin: '0 0 16px 0', fontSize: '1.1rem', color: '#1e293b' }}>
                💬 Step 2: WhatsApp OTP Request & Verification
              </h3>
              <p style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '16px' }}>
                Pengiriman pesan OTP resmi via WhatsApp. Gunakan kode OTP <code>123456</code> untuk verifikasi.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>No. Handphone WhatsApp</label>
                  <input
                    type="text"
                    value={otpForm.phone_number}
                    onChange={(e) => setOtpForm({ ...otpForm, phone_number: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                  />
                </div>

                <button
                  onClick={handleRequestOTP}
                  disabled={loading}
                  style={{
                    padding: '10px',
                    background: '#059669',
                    color: 'white',
                    border: 'none',
                    borderRadius: '6px',
                    fontWeight: 700,
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '8px'
                  }}
                >
                  {loading ? '⏳ Mengirim...' : '📲 Minta Kode OTP WhatsApp'}
                </button>

                <hr style={{ border: 'none', borderTop: '1px solid #e2e8f0', margin: '8px 0' }} />

                <form onSubmit={handleVerifyOTP}>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Masukkan Kode OTP (6 Digit)</label>
                  <input
                    type="text"
                    value={otpForm.otp_code}
                    onChange={(e) => setOtpForm({ ...otpForm, otp_code: e.target.value })}
                    placeholder="Contoh: 123456"
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '1.1rem', letterSpacing: '4px', textAlign: 'center', fontWeight: 'bold', marginBottom: '12px' }}
                    required
                  />

                  <button
                    type="submit"
                    disabled={loading}
                    style={{
                      width: '100%',
                      padding: '12px',
                      background: '#2563eb',
                      color: 'white',
                      border: 'none',
                      borderRadius: '6px',
                      fontWeight: 700,
                      cursor: 'pointer'
                    }}
                  >
                    {loading ? '⏳ Verifikasi...' : '✅ Verifikasi Kode OTP'}
                  </button>
                </form>
              </div>
            </div>
          )}

          {/* STEP 3: SETUP PIN */}
          {activeStep === 3 && (
            <form onSubmit={handleSetupPIN}>
              <h3 style={{ margin: '0 0 16px 0', fontSize: '1.1rem', color: '#1e293b' }}>
                🔐 Step 3: Setup PIN 6-Digit
              </h3>
              <p style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '16px' }}>
                Mendaftarkan PIN 6-digit rahasia dengan enkripsi Bcrypt.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Employee ID (NIP)</label>
                  <input
                    type="number"
                    value={pinForm.employee_id}
                    onChange={(e) => setPinForm({ ...pinForm, employee_id: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>No. KTP</label>
                  <input
                    type="text"
                    value={pinForm.no_ktp}
                    onChange={(e) => setPinForm({ ...pinForm, no_ktp: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>No. Handphone</label>
                  <input
                    type="text"
                    value={pinForm.phone_number}
                    onChange={(e) => setPinForm({ ...pinForm, phone_number: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>PIN Baru (6 Digit Angka)</label>
                  <input
                    type="password"
                    maxLength={6}
                    value={pinForm.pin}
                    onChange={(e) => setPinForm({ ...pinForm, pin: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '1.1rem', letterSpacing: '6px', textAlign: 'center' }}
                    required
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  style={{
                    marginTop: '8px',
                    padding: '12px',
                    background: '#2563eb',
                    color: 'white',
                    border: 'none',
                    borderRadius: '6px',
                    fontWeight: 700,
                    cursor: 'pointer'
                  }}
                >
                  {loading ? '⏳ Daftarkan...' : '🔒 Daftarkan PIN 6-Digit'}
                </button>
              </div>
            </form>
          )}

          {/* STEP 4: MOBILE LOGIN */}
          {activeStep === 4 && (
            <form onSubmit={handleMobileLogin}>
              <h3 style={{ margin: '0 0 16px 0', fontSize: '1.1rem', color: '#1e293b' }}>
                🔑 Step 4: Mobile Login (No. HP + PIN)
              </h3>
              <p style={{ fontSize: '0.85rem', color: '#64748b', marginBottom: '16px' }}>
                Login harian aplikasi mobile EWA menggunakan Nomor HP & PIN 6-Digit tanpa perlu OTP.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>Nomor Handphone</label>
                  <input
                    type="text"
                    value={loginForm.phone_number}
                    onChange={(e) => setLoginForm({ ...loginForm, phone_number: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    required
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#334155', marginBottom: '4px' }}>PIN 6-Digit</label>
                  <input
                    type="password"
                    maxLength={6}
                    value={loginForm.pin}
                    onChange={(e) => setLoginForm({ ...loginForm, pin: e.target.value })}
                    style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '1.1rem', letterSpacing: '6px', textAlign: 'center' }}
                    required
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  style={{
                    marginTop: '8px',
                    padding: '12px',
                    background: '#059669',
                    color: 'white',
                    border: 'none',
                    borderRadius: '6px',
                    fontWeight: 700,
                    cursor: 'pointer'
                  }}
                >
                  {loading ? '⏳ Otentikasi...' : '🔓 Login Mobile App'}
                </button>
              </div>
            </form>
          )}

        </div>

        {/* Right Column: Live API Response & Mobile App Preview */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          
          {/* Realtime API Response Box */}
          <div style={{ background: '#1e293b', borderRadius: '10px', padding: '20px', color: '#f8fafc', boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
              <h4 style={{ margin: 0, color: '#38bdf8', fontSize: '0.95rem' }}>📡 Real-time API Response:</h4>
              {apiResponse && (
                <span style={{
                  background: apiResponse.status === 'SUCCESS' || apiResponse.status === 'MATCH_SUCCESS' ? '#10b981' : '#f59e0b',
                  color: 'white',
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '0.75rem',
                  fontWeight: 'bold'
                }}>
                  {apiResponse.status || '200 OK'}
                </span>
              )}
            </div>

            {loading && <div style={{ color: '#94a3b8', fontStyle: 'italic', fontSize: '0.9rem' }}>⏳ Memanggil API Backend LMS...</div>}

            {apiError && (
              <div style={{ background: '#7f1d1d', border: '1px solid #ef4444', color: '#fca5a5', padding: '12px', borderRadius: '6px', fontSize: '0.85rem' }}>
                ❌ <strong>Error Response:</strong>
                <pre style={{ margin: '6px 0 0 0', whiteSpace: 'pre-wrap', fontFamily: 'monospace' }}>
                  {JSON.stringify(apiError, null, 2)}
                </pre>
              </div>
            )}

            {apiResponse && !loading && (
              <pre style={{ margin: 0, whiteSpace: 'pre-wrap', background: '#0f172a', padding: '12px', borderRadius: '6px', fontSize: '0.8rem', color: '#4ade80', maxHeight: '280px', overflowY: 'auto' }}>
                {JSON.stringify(apiResponse, null, 2)}
              </pre>
            )}

            {!apiResponse && !apiError && !loading && (
              <div style={{ color: '#64748b', fontSize: '0.85rem', fontStyle: 'italic' }}>
                Klik tombol aksi di sebelah kiri untuk melihat respon JSON dari server.
              </div>
            )}
          </div>

          {/* Logged in Mobile App Interface Preview */}
          {loggedInMobileUser && (
            <div style={{ background: '#0f172a', borderRadius: '24px', padding: '16px', border: '4px solid #334155', color: 'white', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.3)' }}>
              <div style={{ textAlign: 'center', fontSize: '0.75rem', color: '#94a3b8', marginBottom: '12px', fontWeight: 600 }}>
                📱 PREVIEW LAYAR MOBILE EWA (LOGGED IN)
              </div>
              <div style={{ background: '#1e293b', borderRadius: '16px', padding: '16px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <div>
                    <div style={{ fontSize: '0.75rem', color: '#94a3b8' }}>Selamat Datang,</div>
                    <div style={{ fontSize: '1rem', fontWeight: 800, color: '#38bdf8' }}>{loggedInMobileUser.user?.name}</div>
                  </div>
                  <span style={{ background: '#059669', padding: '4px 8px', borderRadius: '12px', fontSize: '0.7rem', fontWeight: 'bold' }}>
                    🟢 ONLINE
                  </span>
                </div>

                <div style={{ background: 'linear-gradient(135deg, #2563eb, #1d4ed8)', padding: '14px', borderRadius: '12px', marginBottom: '12px' }}>
                  <div style={{ fontSize: '0.75rem', opacity: 0.9 }}>Gaji Terdaftar (HRD)</div>
                  <div style={{ fontSize: '1.2rem', fontWeight: 800, marginTop: '2px' }}>
                    Rp {(loggedInMobileUser.employee?.salary || 0).toLocaleString('id-ID')}
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', marginTop: '10px', opacity: 0.9 }}>
                    <span>Total Outstanding Pinjaman:</span>
                    <span style={{ fontWeight: 'bold', color: '#fef08a' }}>Rp {(loggedInMobileUser.employee?.total_loan || 0).toLocaleString('id-ID')}</span>
                  </div>
                </div>

                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: '#94a3b8', background: '#0f172a', padding: '8px 12px', borderRadius: '8px' }}>
                  <span>Auto Lock Idle Timeout:</span>
                  <span style={{ color: '#38bdf8', fontWeight: 'bold' }}>{loggedInMobileUser.idle_timeout_minutes} Menit</span>
                </div>
              </div>
            </div>
          )}

        </div>

      </div>
    </div>
  );
}
