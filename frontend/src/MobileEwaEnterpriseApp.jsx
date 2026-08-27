import React, { useState, useEffect } from 'react';
import axios from 'axios';

export default function MobileEwaEnterpriseApp({ currentUser, onLogout, defaultApiBaseUrl = 'https://localhost:8086' }) {
  // Rule 1: Dynamic Server Config State
  const [apiConfig, setApiConfig] = useState(() => {
    const saved = localStorage.getItem('ewa_mobile_api_config');
    if (saved) {
      try { return JSON.parse(saved); } catch (e) {}
    }
    let currentHost = typeof window !== 'undefined' && window.location.hostname ? window.location.hostname : '192.168.0.104';
    if (!currentHost || currentHost === 'localhost' || currentHost === '127.0.0.1') {
      currentHost = '192.168.0.104';
    }
    return {
      mode: 'wifi', // 'wifi' or 'internet'
      ip: currentHost,
      port: '8086',
      customUrl: `http://${currentHost}:8086`
    };
  });

  const [serverModalOpen, setServerModalOpen] = useState(false);
  const [testStatus, setTestStatus] = useState('Belum Diuji'); // 'Belum Diuji', 'Testing', 'Sukses', 'Gagal'
  const [testMessage, setTestMessage] = useState('');

  // Active Tab & Data States
  const [activeBottomNav, setActiveBottomNav] = useState('beranda'); // 'beranda', 'pengajuan', 'pinjaman', 'profil'
  const [searchQuery, setSearchQuery] = useState('');
  
  // Data States
  const [loading, setLoading] = useState(false);
  const [employeeProfile, setEmployeeProfile] = useState(null);
  const [myLoans, setMyLoans] = useState([]);
  const [products, setProducts] = useState([]);

  // Form State for Pengajuan EWA
  const [loanForm, setLoanForm] = useState({
    product_id: '',
    requested_amount: '',
    tenor: 1,
    purpose: 'Pencairan EWA / Gaji Di Muka'
  });
  const [submittingLoan, setSubmittingLoan] = useState(false);
  const [submissionSuccess, setSubmissionSuccess] = useState(null);

  // Active Base URL for API Calls
  const activeApiUrl = apiConfig.customUrl || defaultApiBaseUrl;

  // Fetch Member Data on Mount or API Config Change
  useEffect(() => {
    fetchMemberData();
  }, [activeApiUrl]);

  const fetchMemberData = async () => {
    setLoading(true);
    try {
      // Fetch products
      const prodRes = await axios.get(`${activeApiUrl}/api/products`);
      if (Array.isArray(prodRes.data)) {
        setProducts(prodRes.data);
        if (prodRes.data.length > 0 && !loanForm.product_id) {
          setLoanForm(prev => ({ ...prev, product_id: prodRes.data[0].id }));
        }
      }

      // Fetch my loans
      const loanRes = await axios.get(`${activeApiUrl}/api/applications`);
      if (Array.isArray(loanRes.data)) {
        setMyLoans(loanRes.data);
      }
    } catch (err) {
      console.log('Error fetching mobile data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Test Connection Helper (Rule 1)
  const [traceLogDetails, setTraceLogDetails] = useState([]);

  const handleTestConnection = async () => {
    setTestStatus('Testing');
    setTestMessage('Menghubungkan ke server...');
    setTraceLogDetails([]);
    
    const ip = apiConfig.ip || '192.168.0.104';
    const port = apiConfig.port || '8086';

    const candidateHttp = `http://${ip}:${port}`;
    const candidateHttps = `https://${ip}:${port}`;

    const addLog = (msg) => setTraceLogDetails(prev => [...prev, `[${new Date().toLocaleTimeString('id-ID')}] ${msg}`]);

    addLog(`Target IP: ${ip}:${port}`);
    addLog(`Mencoba GET ${candidateHttp}/ ...`);

    try {
      const res = await axios.get(`${candidateHttp}/`, { timeout: 4000 });
      setTestStatus('Sukses');
      setTestMessage(`🟢 Koneksi HTTP Berhasil! Target: ${candidateHttp}`);
      addLog(`SUCCESS: Response Data => ${JSON.stringify(res.data)}`);
      setApiConfig(prev => ({ ...prev, customUrl: candidateHttp }));
    } catch (errHttp) {
      addLog(`HTTP FAIL: ${errHttp.message} (${errHttp.code || 'NO_CODE'})`);
      addLog(`Mencoba Fallback GET ${candidateHttps}/ ...`);
      try {
        const res2 = await axios.get(`${candidateHttps}/`, { timeout: 4000 });
        setTestStatus('Sukses');
        setTestMessage(`🟢 Koneksi HTTPS Berhasil! Target: ${candidateHttps}`);
        addLog(`HTTPS SUCCESS: ${JSON.stringify(res2.data)}`);
        setApiConfig(prev => ({ ...prev, customUrl: candidateHttps }));
      } catch (errHttps) {
        setTestStatus('Gagal');
        setTestMessage(`🔴 Gagal terhubung ke ${candidateHttp}: ${errHttp.message}`);
        addLog(`HTTPS FAIL: ${errHttps.message}`);
        addLog(`Saran Tracing: Pastikan Backend Go (lms-backend.exe) aktif di laptop, dan HP terhubung ke Wi-Fi yang sama.`);
      }
    }
  };

  const handleSaveServerConfig = () => {
    const defaultWifi = `http://${apiConfig.ip}:${apiConfig.port}`;
    const finalUrl = apiConfig.customUrl || (apiConfig.mode === 'wifi' ? defaultWifi : apiConfig.customUrl);
    
    const newConfig = { ...apiConfig, customUrl: finalUrl };
    setApiConfig(newConfig);
    localStorage.setItem('ewa_mobile_api_config', JSON.stringify(newConfig));
    localStorage.setItem('ewa_custom_api_url', finalUrl);
    setServerModalOpen(false);
    alert(`✅ Pengaturan Server Disimpan! Target API: ${finalUrl}`);
  };

  const handleSubmitLoan = async (e) => {
    e.preventDefault();
    setSubmittingLoan(true);
    setSubmissionSuccess(null);
    try {
      const payload = {
        member_no: currentUser?.member_no || currentUser?.employee_id || 100001,
        product_id: parseInt(loanForm.product_id),
        requested_amount: parseFloat(loanForm.requested_amount),
        tenor: parseInt(loanForm.tenor),
        notes: loanForm.purpose
      };
      const res = await axios.post(`${activeApiUrl}/api/applications`, payload);
      setSubmissionSuccess('🎉 Pengajuan EWA / Pinjaman Berhasil Dikirim!');
      fetchMemberData();
      setTimeout(() => {
        setActiveBottomNav('pinjaman');
      }, 1500);
    } catch (err) {
      alert('❌ Gagal Mengajukan: ' + (err.response?.data?.error || err.message));
    } finally {
      setSubmittingLoan(false);
    }
  };

  // Compute Summary Statistics
  const draftLoans = myLoans.filter(l => l.status === 'SUBMITTED' || l.status === 'DRAFT');
  const activeLoans = myLoans.filter(l => l.status === 'APPROVED' || l.status === 'DISBURSED');

  return (
    <div style={{
      minHeight: '100vh',
      backgroundColor: '#f1f5f9',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      paddingBottom: '80px',
      maxWidth: '480px',
      margin: '0 auto',
      boxShadow: '0 0 20px rgba(0,0,0,0.1)',
      position: 'relative'
    }}>
      {/* 🟢 TOP BAR HEADER WITH BRAND & SERVER CONFIG ICON */}
      <div style={{
        background: '#044B36',
        color: 'white',
        padding: '16px 20px 24px 20px',
        borderBottomLeftRadius: '24px',
        borderBottomRightRadius: '24px',
        boxShadow: '0 4px 12px rgba(4, 75, 54, 0.2)'
      }}>
        {/* Search Bar & Gear Settings */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
          <div style={{
            flex: 1,
            background: 'white',
            borderRadius: '20px',
            padding: '8px 16px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
          }}>
            <span style={{ color: '#94a3b8' }}>🔍</span>
            <input
              type="text"
              placeholder="Cari pinjaman, pengajuan, atau profil..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ border: 'none', outline: 'none', width: '100%', fontSize: '0.85rem', color: '#1e293b' }}
            />
          </div>
          <button
            onClick={() => setServerModalOpen(true)}
            title="Pengaturan Server API"
            style={{
              background: 'rgba(255,255,255,0.2)',
              border: 'none',
              borderRadius: '50%',
              width: '38px',
              height: '38px',
              color: 'white',
              cursor: 'pointer',
              fontSize: '1.1rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0
            }}
          >
            ⚙️
          </button>
          <button
            onClick={onLogout}
            title="Keluar / Logout dari Aplikasi"
            style={{
              background: '#ef4444',
              border: 'none',
              borderRadius: '50%',
              width: '38px',
              height: '38px',
              color: 'white',
              cursor: 'pointer',
              fontSize: '1rem',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
              boxShadow: '0 2px 6px rgba(239, 68, 68, 0.4)'
            }}
          >
            🚪
          </button>
        </div>

        {/* Title */}
        <div style={{ fontSize: '1.1rem', fontWeight: 800, letterSpacing: '-0.01em' }}>
          Laboratory & Loan Management System
        </div>
        <div style={{ fontSize: '0.75rem', opacity: 0.85, marginTop: '2px' }}>
          Kopkara Mobile Enterprise EWA
        </div>
      </div>

      {/* 👤 MEMBER PROFILE GREETING CARD (Rule 3: Strictly Mode Anggota) */}
      <div style={{ padding: '0 16px', marginTop: '-18px' }}>
        <div style={{
          background: 'white',
          borderRadius: '16px',
          padding: '16px 20px',
          boxShadow: '0 4px 12px rgba(0,0,0,0.05)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div style={{
              width: '46px',
              height: '46px',
              borderRadius: '50%',
              background: '#ecfdf5',
              border: '2px solid #10b981',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '1.3rem'
            }}>
              👤
            </div>
            <div>
              <div style={{ fontSize: '0.75rem', color: '#64748b' }}>Selamat Datang,</div>
              <div style={{ fontSize: '1.05rem', fontWeight: 800, color: '#0f172a' }}>
                {currentUser?.name || currentUser?.username || 'Nur Kholim'}
              </div>
            </div>
          </div>
          <span style={{
            background: '#dcfce7',
            color: '#15803d',
            padding: '4px 10px',
            borderRadius: '12px',
            fontSize: '0.7rem',
            fontWeight: 800,
            letterSpacing: '0.05em'
          }}>
            ANGGOTA
          </span>
        </div>

        {/* Summary Badges */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr 1fr',
          gap: '10px',
          marginTop: '12px'
        }}>
          <div style={{ background: 'white', padding: '12px', borderRadius: '12px', textAlign: 'center', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
            <div style={{ fontSize: '1.2rem', fontWeight: 800, color: '#2563eb' }}>{draftLoans.length}</div>
            <div style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600 }}>Draft Pengajuan</div>
          </div>
          <div style={{ background: 'white', padding: '12px', borderRadius: '12px', textAlign: 'center', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
            <div style={{ fontSize: '1.2rem', fontWeight: 800, color: '#10b981' }}>{activeLoans.length}</div>
            <div style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600 }}>Pinjaman Aktif</div>
          </div>
          <div style={{ background: 'white', padding: '12px', borderRadius: '12px', textAlign: 'center', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
            <div style={{ fontSize: '1.2rem', fontWeight: 800, color: '#d97706' }}>{myLoans.length}</div>
            <div style={{ fontSize: '0.7rem', color: '#64748b', fontWeight: 600 }}>Total Riwayat</div>
          </div>
        </div>
      </div>

      {/* 🚀 RULE 2: MENU SHORTCUT GRID */}
      <div style={{ padding: '20px 16px 10px 16px' }}>
        <div style={{ fontSize: '0.85rem', fontWeight: 800, color: '#334155', marginBottom: '12px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          MENU SHORTCUT
        </div>

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: '12px'
        }}>
          {[
            { id: 'beranda', label: 'Beranda', icon: '🏠', color: '#ecfdf5', border: '#10b981' },
            { id: 'pengajuan', label: 'Pengajuan EWA', icon: '📝', color: '#eff6ff', border: '#3b82f6' },
            { id: 'pinjaman', label: 'Daftar Pinjaman', icon: '💰', color: '#fef3c7', border: '#f59e0b' },
            { id: 'katalog', label: 'Katalog Produk', icon: '📦', color: '#f3e8ff', border: '#a855f7' },
            { id: 'keamanan', label: 'Keamanan PIN', icon: '🔒', color: '#fee2e2', border: '#ef4444' },
            { id: 'server', label: 'Set Server API', icon: '⚙️', color: '#f1f5f9', border: '#64748b' }
          ].map(m => (
            <div
              key={m.id}
              onClick={() => {
                if (m.id === 'server') setServerModalOpen(true);
                else if (m.id === 'katalog') alert('📦 Katalog Produk EWA: Tersedia produk Pinjaman Reguler & EWA Potong Gaji.');
                else if (m.id === 'keamanan') alert('🔒 Keamanan PIN: PIN 6-Digit Anda dilindungi dengan enkripsi Bcrypt.');
                else setActiveBottomNav(m.id);
              }}
              style={{
                background: 'white',
                borderRadius: '14px',
                padding: '12px 6px',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                boxShadow: '0 2px 6px rgba(0,0,0,0.04)',
                border: activeBottomNav === m.id ? `2px solid ${m.border}` : '1px solid #e2e8f0',
                transition: 'transform 0.1s'
              }}
            >
              <div style={{
                width: '40px',
                height: '40px',
                borderRadius: '12px',
                background: m.color,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '1.2rem',
                marginBottom: '6px'
              }}>
                {m.icon}
              </div>
              <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#334155', textAlign: 'center' }}>
                {m.label}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* 📄 MAIN CONTENT AREA BASED ON ACTIVE TAB */}
      <div style={{ padding: '0 16px' }}>

        {/* BERANDA / HOME TAB */}
        {activeBottomNav === 'beranda' && (
          <div>
            <div style={{ background: 'white', borderRadius: '16px', padding: '16px', marginBottom: '16px', boxShadow: '0 2px 6px rgba(0,0,0,0.04)' }}>
              <div style={{ fontSize: '0.85rem', fontWeight: 800, color: '#1e293b', marginBottom: '8px' }}>
                📌 Pengumuman Kopkara EWA
              </div>
              <p style={{ margin: 0, fontSize: '0.8rem', color: '#475569', lineHeight: 1.4 }}>
                Pencairan EWA (Earned Wage Access) dapat dilakukan setiap saat dan dipotong otomatis pada payroll akhir bulan.
              </p>
            </div>

            {/* Recent Loans */}
            <div style={{ fontSize: '0.85rem', fontWeight: 800, color: '#334155', marginBottom: '10px' }}>
              Riwayat Pinjaman Terbaru
            </div>
            {myLoans.length === 0 ? (
              <div style={{ background: 'white', padding: '20px', borderRadius: '12px', textAlign: 'center', color: '#94a3b8', fontSize: '0.8rem' }}>
                Belum ada riwayat pengajuan pinjaman.
              </div>
            ) : (
              myLoans.slice(0, 3).map(loan => (
                <div key={loan.id} style={{ background: 'white', padding: '12px 16px', borderRadius: '12px', marginBottom: '8px', boxShadow: '0 2px 4px rgba(0,0,0,0.03)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontSize: '0.85rem', fontWeight: 700, color: '#1e293b' }}>
                      Rp {(loan.requested_amount || 0).toLocaleString('id-ID')}
                    </div>
                    <div style={{ fontSize: '0.7rem', color: '#64748b' }}>Tenor: {loan.tenor} Bulan</div>
                  </div>
                  <span style={{
                    padding: '3px 8px',
                    borderRadius: '8px',
                    fontSize: '0.7rem',
                    fontWeight: 700,
                    background: loan.status === 'APPROVED' ? '#dcfce7' : (loan.status === 'REJECTED' ? '#fee2e2' : '#fef3c7'),
                    color: loan.status === 'APPROVED' ? '#15803d' : (loan.status === 'REJECTED' ? '#b91c1c' : '#b45309')
                  }}>
                    {loan.status}
                  </span>
                </div>
              ))
            )}
          </div>
        )}

        {/* PENGAJUAN TAB */}
        {activeBottomNav === 'pengajuan' && (
          <div style={{ background: 'white', borderRadius: '16px', padding: '20px', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}>
            <h3 style={{ margin: '0 0 14px 0', fontSize: '1.05rem', color: '#0f172a', fontWeight: 800 }}>
              📝 Pengajuan EWA / Pinjaman Baru
            </h3>

            {submissionSuccess && (
              <div style={{ background: '#dcfce7', color: '#15803d', padding: '12px', borderRadius: '8px', fontSize: '0.85rem', fontWeight: 700, marginBottom: '14px' }}>
                {submissionSuccess}
              </div>
            )}

            <form onSubmit={handleSubmitLoan} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div>
                <label style={{ fontSize: '0.75rem', fontWeight: 700, color: '#475569', display: 'block', marginBottom: '4px' }}>Produk Pinjaman</label>
                <select
                  value={loanForm.product_id}
                  onChange={(e) => setLoanForm({ ...loanForm, product_id: e.target.value })}
                  style={{ width: '100%', padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.85rem' }}
                  required
                >
                  {products.map(p => (
                    <option key={p.id} value={p.id}>{p.name} (Bunga: {p.interest_rate}%)</option>
                  ))}
                </select>
              </div>

              <div>
                <label style={{ fontSize: '0.75rem', fontWeight: 700, color: '#475569', display: 'block', marginBottom: '4px' }}>Nominal Pinjaman (Rp)</label>
                <input
                  type="number"
                  placeholder="Contoh: 1000000"
                  value={loanForm.requested_amount}
                  onChange={(e) => setLoanForm({ ...loanForm, requested_amount: e.target.value })}
                  style={{ width: '100%', padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.9rem', fontWeight: 'bold' }}
                  required
                />
              </div>

              <div>
                <label style={{ fontSize: '0.75rem', fontWeight: 700, color: '#475569', display: 'block', marginBottom: '4px' }}>Tenor (Bulan)</label>
                <select
                  value={loanForm.tenor}
                  onChange={(e) => setLoanForm({ ...loanForm, tenor: e.target.value })}
                  style={{ width: '100%', padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.85rem' }}
                >
                  <option value={1}>1 Bulan</option>
                  <option value={3}>3 Bulan</option>
                  <option value={6}>6 Bulan</option>
                  <option value={12}>12 Bulan</option>
                </select>
              </div>

              <button
                type="submit"
                disabled={submittingLoan}
                style={{
                  marginTop: '10px',
                  padding: '12px',
                  background: '#044B36',
                  color: 'white',
                  border: 'none',
                  borderRadius: '10px',
                  fontWeight: 800,
                  fontSize: '0.9rem',
                  cursor: 'pointer'
                }}
              >
                {submittingLoan ? '⏳ Memproses...' : '🚀 Kirim Pengajuan'}
              </button>
            </form>
          </div>
        )}

        {/* PINJAMAN TAB */}
        {activeBottomNav === 'pinjaman' && (
          <div>
            <div style={{ fontSize: '0.9rem', fontWeight: 800, color: '#0f172a', marginBottom: '12px' }}>
              💰 Daftar Pinjaman Saya
            </div>
            {myLoans.map(loan => (
              <div key={loan.id} style={{ background: 'white', padding: '14px', borderRadius: '12px', marginBottom: '10px', boxShadow: '0 2px 6px rgba(0,0,0,0.04)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                  <span style={{ fontSize: '0.75rem', color: '#64748b' }}>App ID: #{loan.id}</span>
                  <span style={{
                    padding: '2px 8px',
                    borderRadius: '6px',
                    fontSize: '0.65rem',
                    fontWeight: 800,
                    background: loan.status === 'APPROVED' ? '#dcfce7' : '#fef3c7',
                    color: loan.status === 'APPROVED' ? '#15803d' : '#b45309'
                  }}>
                    {loan.status}
                  </span>
                </div>
                <div style={{ fontSize: '1.1rem', fontWeight: 800, color: '#0f172a' }}>
                  Rp {(loan.requested_amount || 0).toLocaleString('id-ID')}
                </div>
                <div style={{ fontSize: '0.75rem', color: '#475569', marginTop: '4px' }}>
                  Tenor: {loan.tenor} Bulan · Tanggal: {new Date(loan.created_at || Date.now()).toLocaleDateString('id-ID')}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* ACCOUNT TAB */}
        {(activeBottomNav === 'account' || activeBottomNav === 'profil') && (
          <div style={{ background: 'white', borderRadius: '16px', padding: '20px', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}>
            <h3 style={{ margin: '0 0 16px 0', fontSize: '1.05rem', color: '#0f172a', fontWeight: 800 }}>
              👤 Account & Profil Anggota
            </h3>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '0.85rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>1. Nama Lengkap:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.name || 'Nur Kholim'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>2. NIK Karyawan:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.nik || currentUser?.employee_id || currentUser?.username || '100001'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>3. No. HP / WhatsApp:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.phone_number || currentUser?.username || '085882500073'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>4. No. KTP:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.no_ktp || '3201012345670001'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>5. Nama Bank (bank_name):</span>
                <span style={{ fontWeight: 700, color: '#2563eb' }}>{currentUser?.bank_name || 'BANK MANDIRI'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>6. No. Rekening Bank:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.bank_account_no || '123000456789'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', paddingBottom: '8px', borderBottom: '1px solid #f1f5f9' }}>
                <span style={{ color: '#64748b' }}>7. Nama Pemilik Rekening:</span>
                <span style={{ fontWeight: 700, color: '#0f172a' }}>{currentUser?.bank_account_name || currentUser?.name || 'NUR KHOLIM'}</span>
              </div>

              <button
                onClick={onLogout}
                style={{
                  marginTop: '16px',
                  padding: '12px',
                  background: '#ef4444',
                  color: 'white',
                  border: 'none',
                  borderRadius: '10px',
                  fontWeight: 800,
                  cursor: 'pointer'
                }}
              >
                🚪 Keluar / Logout
              </button>
            </div>
          </div>
        )}

      </div>

      {/* 📱 BOTTOM NAVIGATION BAR (Shopee 3 Tabs Footer) */}
      <div style={{
        position: 'fixed',
        bottom: 0,
        left: '50%',
        transform: 'translateX(-50%)',
        width: '100%',
        maxWidth: '480px',
        background: 'white',
        borderTop: '1px solid #e2e8f0',
        display: 'flex',
        justifyContent: 'space-around',
        padding: '8px 0',
        zIndex: 9000,
        boxShadow: '0 -4px 12px rgba(0,0,0,0.05)'
      }}>
        {[
          { id: 'beranda', label: 'Home', icon: '🏠' },
          { id: 'pinjaman', label: 'Daftar Pinjaman', icon: '📋' },
          { id: 'account', label: 'Account', icon: '👤' }
        ].map(item => {
          const isActive = activeBottomNav === item.id || (item.id === 'account' && activeBottomNav === 'profil');
          return (
            <div
              key={item.id}
              onClick={() => setActiveBottomNav(item.id)}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                cursor: 'pointer',
                color: isActive ? '#044B36' : '#94a3b8',
                fontWeight: isActive ? 800 : 500
              }}
            >
              <span style={{ fontSize: '1.2rem' }}>{item.icon}</span>
              <span style={{ fontSize: '0.65rem', marginTop: '2px' }}>{item.label}</span>
            </div>
          );
        })}
      </div>

      {/* ⚙️ RULE 1 MODAL: PENGATURAN SERVER API MODAL (MATCHING SCREENSHOT 1) */}
      {serverModalOpen && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          backgroundColor: 'rgba(15, 23, 42, 0.75)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          zIndex: 100000, backdropFilter: 'blur(4px)', padding: '16px'
        }}>
          <div style={{
            backgroundColor: 'white', borderRadius: '24px', padding: '24px',
            width: '100%', maxWidth: '380px', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)'
          }}>
            {/* Modal Icon Header */}
            <div style={{ textAlign: 'center', marginBottom: '16px' }}>
              <div style={{ width: '56px', height: '56px', borderRadius: '50%', background: '#eff6ff', color: '#2563eb', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '1.6rem', margin: '0 auto 10px' }}>
                🌐
              </div>
              <h3 style={{ margin: 0, fontSize: '1.2rem', color: '#0f172a', fontWeight: 800 }}>
                Pengaturan Server API
              </h3>
              <p style={{ margin: '4px 0 0 0', fontSize: '0.8rem', color: '#64748b' }}>
                Atur alamat koneksi untuk HP Anda
              </p>
            </div>

            {/* Toggle Wi-Fi Lokal vs Internet */}
            <div style={{ display: 'flex', background: '#f1f5f9', padding: '4px', borderRadius: '12px', marginBottom: '16px' }}>
              <button
                onClick={() => setApiConfig({ ...apiConfig, mode: 'wifi' })}
                style={{
                  flex: 1, padding: '8px', borderRadius: '8px', border: 'none',
                  background: apiConfig.mode === 'wifi' ? 'white' : 'transparent',
                  fontWeight: apiConfig.mode === 'wifi' ? 800 : 600,
                  color: apiConfig.mode === 'wifi' ? '#2563eb' : '#64748b',
                  fontSize: '0.8rem', cursor: 'pointer', boxShadow: apiConfig.mode === 'wifi' ? '0 2px 4px rgba(0,0,0,0.05)' : 'none'
                }}
              >
                📶 Wi-Fi Lokal
              </button>
              <button
                onClick={() => setApiConfig({ ...apiConfig, mode: 'internet' })}
                style={{
                  flex: 1, padding: '8px', borderRadius: '8px', border: 'none',
                  background: apiConfig.mode === 'internet' ? 'white' : 'transparent',
                  fontWeight: apiConfig.mode === 'internet' ? 800 : 600,
                  color: apiConfig.mode === 'internet' ? '#2563eb' : '#64748b',
                  fontSize: '0.8rem', cursor: 'pointer', boxShadow: apiConfig.mode === 'internet' ? '0 2px 4px rgba(0,0,0,0.05)' : 'none'
                }}
              >
                🌐 Internet (VPS / Ngrok)
              </button>
            </div>

            {/* Form Input Address */}
            {apiConfig.mode === 'wifi' ? (
              <div style={{ marginBottom: '16px' }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 800, color: '#334155', display: 'block', marginBottom: '4px' }}>ALAMAT & PORT SERVER LOKAL</label>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <input
                    type="text"
                    value={apiConfig.ip}
                    onChange={(e) => setApiConfig({ ...apiConfig, ip: e.target.value })}
                    placeholder="192.168.0.103"
                    style={{ flex: 1, padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.85rem', fontWeight: 700 }}
                  />
                  <span style={{ alignSelf: 'center', fontWeight: 'bold' }}>:</span>
                  <input
                    type="text"
                    value={apiConfig.port}
                    onChange={(e) => setApiConfig({ ...apiConfig, port: e.target.value })}
                    placeholder="8086"
                    style={{ width: '70px', padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.85rem', fontWeight: 700 }}
                  />
                </div>
                <div style={{ fontSize: '0.7rem', color: '#64748b', marginTop: '6px', fontStyle: 'italic' }}>
                  * Gunakan port 8086 jika menembak Go Backend HTTPS langsung.
                </div>
              </div>
            ) : (
              <div style={{ marginBottom: '16px' }}>
                <label style={{ fontSize: '0.75rem', fontWeight: 800, color: '#334155', display: 'block', marginBottom: '4px' }}>URL PUBLIC (VPS / NGROK)</label>
                <input
                  type="text"
                  value={apiConfig.customUrl}
                  onChange={(e) => setApiConfig({ ...apiConfig, customUrl: e.target.value })}
                  placeholder="https://lims.domain.com"
                  style={{ width: '100%', padding: '10px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '0.85rem', fontWeight: 700 }}
                />
              </div>
            )}

            {/* Connection Test Status */}
            <div style={{ background: '#f8fafc', padding: '10px 14px', borderRadius: '10px', marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '0.8rem', color: '#475569', fontWeight: 600 }}>Status Koneksi:</span>
              <span style={{
                fontSize: '0.8rem', fontWeight: 800,
                color: testStatus === 'Sukses' ? '#16a34a' : (testStatus === 'Gagal' ? '#dc2626' : '#2563eb')
              }}>
                {testStatus}
              </span>
            </div>

            {testMessage && (
              <div style={{ fontSize: '0.75rem', marginBottom: '10px', color: testStatus === 'Sukses' ? '#15803d' : '#b91c1c', fontWeight: 700 }}>
                {testMessage}
              </div>
            )}

            {/* Real-time Tracing Log Box */}
            {traceLogDetails.length > 0 && (
              <div style={{
                background: '#0f172a', color: '#38bdf8', padding: '10px',
                borderRadius: '8px', fontSize: '0.7rem', fontFamily: 'monospace',
                maxHeight: '120px', overflowY: 'auto', marginBottom: '14px',
                border: '1px solid #334155'
              }}>
                <div style={{ fontWeight: 'bold', color: '#fef08a', marginBottom: '4px' }}>📋 Log Tracing Real-time (HP):</div>
                {traceLogDetails.map((logLine, idx) => (
                  <div key={idx} style={{ marginBottom: '2px', wordBreak: 'break-all' }}>{logLine}</div>
                ))}
              </div>
            )}

            {/* Buttons Grid Matching Screenshot 1 */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
              <button
                onClick={handleTestConnection}
                style={{ padding: '10px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '10px', fontWeight: 800, fontSize: '0.8rem', cursor: 'pointer' }}
              >
                Tes Koneksi
              </button>
              <button
                onClick={handleSaveServerConfig}
                style={{ padding: '10px', background: '#10b981', color: 'white', border: 'none', borderRadius: '10px', fontWeight: 800, fontSize: '0.8rem', cursor: 'pointer' }}
              >
                Simpan & Terapkan
              </button>
              <button
                onClick={() => {
                  setApiConfig({ mode: 'wifi', ip: '192.168.0.103', port: '8086', customUrl: defaultApiBaseUrl });
                  setTestStatus('Belum Diuji');
                  setTestMessage('');
                }}
                style={{ padding: '10px', background: '#94a3b8', color: 'white', border: 'none', borderRadius: '10px', fontWeight: 800, fontSize: '0.8rem', cursor: 'pointer' }}
              >
                Reset Default
              </button>
              <button
                onClick={() => setServerModalOpen(false)}
                style={{ padding: '10px', background: '#ef4444', color: 'white', border: 'none', borderRadius: '10px', fontWeight: 800, fontSize: '0.8rem', cursor: 'pointer' }}
              >
                Tutup
              </button>
            </div>

          </div>
        </div>
      )}

    </div>
  );
}
