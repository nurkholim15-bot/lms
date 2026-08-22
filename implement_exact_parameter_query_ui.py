# Python script to convert Global Parameter screen to exact standard Master Data Query UI

with open('frontend/src/App.jsx', 'r', encoding='utf-8') as f:
    app_code = f.read()

# 1. Update getMasterTitle to map 'parameters' to 'Pengaturan Parameter Global'
old_get_master_title = """  const getMasterTitle = (tab) => {
    if (tab === 'departments') return 'Master Department';
    if (tab === 'employee-statuses') return 'Master Status Karyawan';
    if (tab === 'kopkara-statuses') return 'Master Status Kopkara';
    if (tab === 'employee-categories') return 'Master Kategori Karyawan';
    if (tab === 'employees') return 'Master Data Karyawan';
    if (tab === 'members') return 'Master Anggota Kopkara';
    if (tab === 'roles') return 'Master Role Sistem';
    if (tab === 'menus') return 'Master Menu Navigation';
    if (tab === 'parameters') return 'Parameter Global';
    if (tab === 'users') return 'Master User Accounts';
    if (tab === 'sessions') return 'Master Active Sessions';
    return `Master ${tab.replace('-', ' ')}`;
  };"""

new_get_master_title = """  const getMasterTitle = (tab) => {
    if (tab === 'departments') return 'Master Department';
    if (tab === 'employee-statuses') return 'Master Status Karyawan';
    if (tab === 'kopkara-statuses') return 'Master Status Kopkara';
    if (tab === 'employee-categories') return 'Master Kategori Karyawan';
    if (tab === 'employees') return 'Master Data Karyawan';
    if (tab === 'members') return 'Master Anggota Kopkara';
    if (tab === 'roles') return 'Master Role Sistem';
    if (tab === 'menus') return 'Master Menu Navigation';
    if (tab === 'parameters') return 'Pengaturan Parameter Global';
    if (tab === 'users') return 'Master User Accounts';
    if (tab === 'sessions') return 'Master Active Sessions';
    return `Master ${tab.replace('-', ' ')}`;
  };"""

app_code = app_code.replace(old_get_master_title, new_get_master_title)

# 2. Update activeTab === 'parameters' condition to route to master-parameters container
old_param_tab_condition = """          {activeTab === 'parameters' && ("""
new_param_tab_condition = """          {(activeTab === 'parameters' || activeTab === 'master-parameters') && (() => {
            const q = (masterSearchQuery || '').toLowerCase().trim();
            const filteredParams = (parameters || []).filter(p => {
              if (!q) return true;
              const kn = String(p.key_name || p.KeyName || '').toLowerCase();
              const kv = String(p.key_value || p.KeyValue || '').toLowerCase();
              const kd = String(p.description || p.Description || '').toLowerCase();
              return kn.includes(q) || kv.includes(q) || kd.includes(q);
            });

            const limit = parseInt(getParamVal('PAGINATION_LIMIT', getParamVal('DEFAULT_PAGE_SIZE', '5'))) || 5;
            const totalRecords = filteredParams.length;
            const totalPages = Math.ceil(totalRecords / limit) || 1;
            const startIdx = (currentPage - 1) * limit;
            const paginatedParams = filteredParams.slice(startIdx, startIdx + limit);

            return (
              <div className="card" style={{ maxWidth: '1200px', padding: 0, overflow: 'hidden' }}>
                {/* Header Container khusus dengan background HEADER_BG & font putih */}
                <div style={{ background: 'var(--header-bg, #0B2545)', color: '#ffffff', padding: '16px 20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '12px' }}>
                  <div>
                    <div style={{ fontSize: '1.2rem', fontWeight: 'bold', color: '#ffffff' }}>
                      Pengaturan Parameter Global
                    </div>
                    <div style={{ fontSize: '0.85rem', color: 'rgba(255, 255, 255, 0.75)', marginTop: '2px' }}>
                      Kelola data parameter sistem
                    </div>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                    {/* Search Filter Box */}
                    <div style={{ background: '#ffffff', borderRadius: '8px', padding: '4px 6px', display: 'inline-flex', alignItems: 'center', gap: '6px', border: '1px solid #cbd5e1' }}>
                      <input 
                        type="text" 
                        placeholder="Cari Nama atau ID..."
                        value={masterSearchQuery}
                        onChange={e => setMasterSearchQuery(e.target.value)}
                        onKeyDown={e => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            setCurrentPage(1);
                          }
                        }}
                        style={{ border: 'none', outline: 'none', padding: '4px 8px', fontSize: '0.85rem', color: '#0f172a', width: '170px' }}
                      />
                      <button
                        type="button"
                        onClick={() => setCurrentPage(1)}
                        style={{ padding: '5px 14px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.85rem' }}
                      >
                        Filter
                      </button>
                    </div>

                    {/* Tombol + Tambah Data (mengikuti BUTTON_BG dan font putih) */}
                    <button
                      type="button"
                      onClick={() => {
                        setParamForm({ id: 0, key_name: '', key_value: '', description: '' });
                        setIsMasterModalOpen(true);
                      }}
                      style={{ padding: '8px 18px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.85rem', display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                    >
                      ➕ Tambah Data
                    </button>

                    {/* Tombol ✕ Tutup (mengikuti BUTTON_CLOSED_BG dan font putih) */}
                    <button
                      type="button"
                      onClick={() => {
                        setIsMasterModalOpen(false);
                        setActiveTab('dashboard');
                      }}
                      style={{ padding: '8px 18px', background: 'var(--button-closed-bg, #475569)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.85rem', display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                    >
                      ✕ Tutup
                    </button>
                  </div>
                </div>

                <div style={{ padding: '20px' }}>
                  <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', marginBottom: '12px', fontSize: '0.85rem', color: '#64748B' }}>
                    <span>Pencarian: <strong>{paginatedParams.length}</strong> ditampilkan dari <strong>{totalRecords}</strong> total data</span>
                  </div>

                  <div style={{ overflowX: 'auto', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '700px' }}>
                      <thead>
                        <tr style={{ background: '#0B2545', color: '#ffffff' }}>
                          <th style={{ padding: '12px 14px', fontSize: '0.85rem', textAlign: 'left', color: '#ffffff', width: '60px' }}>id</th>
                          <th style={{ padding: '12px 14px', fontSize: '0.85rem', textAlign: 'left', color: '#ffffff' }}>key_name</th>
                          <th style={{ padding: '12px 14px', fontSize: '0.85rem', textAlign: 'left', color: '#ffffff' }}>key_value</th>
                          <th style={{ padding: '12px 14px', fontSize: '0.85rem', textAlign: 'left', color: '#ffffff' }}>description</th>
                          <th style={{ padding: '12px 14px', fontSize: '0.85rem', textAlign: 'center', color: '#ffffff', width: '140px' }}>Aksi</th>
                        </tr>
                      </thead>
                      <tbody>
                        {paginatedParams.map((param, idx) => {
                          const pId = param.id || param.ID || (idx + 1);
                          const kName = param.key_name || param.KeyName || '';
                          const kVal = param.key_value || param.KeyValue || '';
                          const kDesc = param.description || param.Description || '';

                          return (
                            <tr key={pId} style={{ borderBottom: '1px solid #e2e8f0' }}>
                              <td style={{ padding: '10px 14px', fontSize: '0.85rem', color: '#0f172a', fontWeight: 600 }}>{pId}</td>
                              <td style={{ padding: '10px 14px', fontSize: '0.85rem', color: '#0f172a', fontWeight: 600 }}>{kName}</td>
                              <td style={{ padding: '10px 14px', fontSize: '0.85rem', color: '#334155' }}>{kVal}</td>
                              <td style={{ padding: '10px 14px', fontSize: '0.85rem', color: '#64748b' }}>{kDesc}</td>
                              <td style={{ padding: '10px 14px', textAlign: 'center' }}>
                                <button
                                  type="button"
                                  onClick={() => {
                                    setParamForm({ id: pId, key_name: kName, key_value: kVal, description: kDesc });
                                    setIsMasterModalOpen(true);
                                  }}
                                  style={{ padding: '4px 10px', marginRight: '6px', background: '#fef3c7', color: '#92400E', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}
                                >
                                  Edit
                                </button>
                                <button
                                  type="button"
                                  onClick={() => deleteParameter(pId)}
                                  style={{ padding: '4px 10px', background: '#fee2e2', color: '#991B1B', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}
                                >
                                  Hapus
                                </button>
                              </td>
                            </tr>
                          );
                        })}
                        {paginatedParams.length === 0 && (
                          <tr>
                            <td colSpan="5" style={{ textAlign: 'center', padding: '24px', color: '#64748B' }}>
                              Belum ada data parameter
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>

                  {/* Pagination Footer (mengikuti PAGINATION_LIMIT) */}
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '16px', flexWrap: 'wrap', gap: '12px' }}>
                    <div style={{ fontSize: '0.85rem', color: '#64748B' }}>
                      Halaman <strong>{currentPage}</strong> dari <strong>{totalPages}</strong> (Total <strong>{totalRecords}</strong> data)
                    </div>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button
                        type="button"
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                        disabled={currentPage <= 1}
                        style={{ padding: '6px 14px', background: currentPage <= 1 ? '#e2e8f0' : 'var(--button-bg, #10b981)', color: currentPage <= 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: currentPage <= 1 ? 'not-allowed' : 'pointer', fontWeight: 600 }}
                      >
                        ◄ Sebelumnya
                      </button>
                      <button
                        type="button"
                        onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                        disabled={currentPage >= totalPages}
                        style={{ padding: '6px 14px', background: currentPage >= totalPages ? '#e2e8f0' : 'var(--button-bg, #10b981)', color: currentPage >= totalPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: currentPage >= totalPages ? 'not-allowed' : 'pointer', fontWeight: 600 }}
                      >
                        Selanjutnya ►
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            );
          })()} //"""

# Find end of old parameter section and replace
old_param_full_block = """          {activeTab === 'parameters' && (
            <div style={{ display: 'flex', gap: '24px', flexWrap: 'wrap' }}>
              <div className="card" style={{ flex: '1', minWidth: '300px' }}>
                <h2>{paramForm.id ? 'Edit Parameter' : 'Tambah Parameter Baru'}</h2>
                <form onSubmit={submitParameter} style={{ display: 'flex', flexDirection: 'column', gap: '16px', marginTop: '24px' }}>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Key Name (Misal: max_loan_limit)</label>
                    <input 
                      type="text" required
                      value={paramForm.key_name}
                      onChange={e => setParamForm({...paramForm, key_name: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                      disabled={paramForm.id > 0} // Key tidak bisa diubah jika edit
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Key Value</label>
                    <input 
                      type="text" required
                      value={paramForm.key_value}
                      onChange={e => setParamForm({...paramForm, key_value: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Deskripsi</label>
                    <textarea 
                      value={paramForm.description}
                      onChange={e => setParamForm({...paramForm, description: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', minHeight: '80px' }}
                    />
                  </div>
                  <div style={{ display: 'flex', gap: '12px' }}>
                    <button type="submit" style={{ padding: '10px 20px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '4px', fontWeight: 600, cursor: 'pointer' }}>
                      Simpan Parameter
                    </button>
                    {paramForm.id > 0 && (
                      <button type="button" onClick={() => setParamForm({ id: 0, key_name: '', key_value: '', description: '' })} style={{ padding: '10px 20px', background: 'var(--button-cancel-bg, #64748b)', color: '#ffffff', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>
                        Cancel
                      </button>
                    )}
                  </div>
                </form>
              </div>

              <div className="table-container" style={{ flex: '2', minWidth: '500px', borderRadius: '8px', overflow: 'hidden', border: '1px solid #cbd5e1' }}>
                <div className="table-header" style={{ background: '#0B2545', color: '#ffffff', padding: '14px 20px', fontWeight: 'bold', fontSize: '1.1rem' }}>Daftar Parameter Global</div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ background: '#0B2545', color: '#ffffff' }}>
                      <th style={{ padding: '10px 14px', fontSize: '0.875rem', textAlign: 'left', color: '#ffffff' }}>Key Name</th>
                      <th style={{ padding: '10px 14px', fontSize: '0.875rem', textAlign: 'left', color: '#ffffff' }}>Value</th>
                      <th style={{ padding: '10px 14px', fontSize: '0.875rem', textAlign: 'left', color: '#ffffff' }}>Deskripsi</th>
                      <th style={{ padding: '10px 14px', fontSize: '0.875rem', textAlign: 'right', color: '#ffffff' }}>Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {parameters.length === 0 ? (
                      <tr><td colSpan="4" style={{ textAlign: 'center', padding: '16px', color: '#64748b' }}>Belum ada konfigurasi parameter</td></tr>
                    ) : (
                      parameters.map((param, idx) => {
                        const pId = param.id || param.ID || (idx + 1);
                        const kName = param.key_name || param.KeyName || '';
                        const kVal = param.key_value || param.KeyValue || '';
                        const kDesc = param.description || param.Description || '';
                        return (
                          <tr key={pId} style={{ borderBottom: '1px solid #e2e8f0' }}>
                            <td style={{ padding: '8px 12px', fontSize: '0.875rem' }}><strong>{kName}</strong></td>
                            <td style={{ padding: '8px 12px', fontSize: '0.875rem' }}>{kVal}</td>
                            <td style={{ padding: '8px 12px', fontSize: '0.875rem' }}>{kDesc}</td>
                            <td style={{ padding: '8px 12px' }}>
                              <button onClick={() => setParamForm({ id: pId, key_name: kName, key_value: kVal, description: kDesc })} style={{ padding: '4px 10px', marginRight: '8px', background: '#fef3c7', color: '#92400E', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem', fontWeight: 600 }}>Edit</button>
                              <button onClick={() => deleteParameter(pId)} style={{ padding: '4px 10px', background: '#fee2e2', color: '#991B1B', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem', fontWeight: 600 }}>Hapus</button>
                            </td>
                          </tr>
                        );
                      })
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}"""

app_code = app_code.replace(old_param_full_block, new_param_tab_condition)

# 3. Update Modal for parameter form when isMasterModalOpen is true
old_modal_content = """                <h3 style={{ margin: '0 0 16px 0', color: 'var(--primary-blue)' }}>
                  {isEditMasterMode ? `Edit ${getMasterTitle(masterTab)}` : `Tambah ${getMasterTitle(masterTab)}`}
                </h3>"""

new_modal_content = """                <h3 style={{ margin: '0 0 16px 0', color: 'var(--primary-blue)' }}>
                  {paramForm.id > 0 ? 'Edit Pengaturan Parameter Global' : (isEditMasterMode ? `Edit ${getMasterTitle(masterTab)}` : `Tambah ${getMasterTitle(masterTab)}`)}
                </h3>"""

app_code = app_code.replace(old_modal_content, new_modal_content)

# Update Modal Form submit and Cancel buttons to match requirement 3, 4, 5
old_modal_form_footer = """                      <div style={{ display: 'flex', gap: '12px', marginTop: '10px' }}>
                        <button
                          type="submit"
                          style={{ padding: '8px 20px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Simpan Data
                        </button>
                        <button
                          type="button"
                          onClick={() => {
                            setIsMasterModalOpen(false);
                            setIsEditMasterMode(false);
                            setMasterForm({});
                          }}
                          style={{ padding: '8px 20px', background: 'var(--button-cancel-bg, #64748b)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Cancel
                        </button>
                      </div>"""

new_modal_form_footer = """                      <div style={{ display: 'flex', gap: '12px', marginTop: '10px' }}>
                        <button
                          type="submit"
                          style={{ padding: '10px 22px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Simpan Data
                        </button>
                        <button
                          type="button"
                          onClick={() => {
                            setIsMasterModalOpen(false);
                            setIsEditMasterMode(false);
                            setMasterForm({});
                            setParamForm({ id: 0, key_name: '', key_value: '', description: '' });
                          }}
                          style={{ padding: '10px 22px', background: 'var(--button-closed-bg, #f59e0b)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Cancel
                        </button>
                      </div>"""

app_code = app_code.replace(old_modal_form_footer, new_modal_form_footer)

with open('frontend/src/App.jsx', 'w', encoding='utf-8') as f:
    f.write(app_code)

print("Successfully updated Parameter UI to exact standard Master Query layout!")
