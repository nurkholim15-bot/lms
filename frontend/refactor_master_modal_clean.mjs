import fs from 'fs';

let code = fs.readFileSync('src/App.jsx', 'utf8');

// 1. Check helper functions in App
const helpersCode = `  const getMasterTitle = (tab) => {
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
    return \`Master \${tab.replace('-', ' ')}\`;
  };

  const getPrimaryKeyKey = (tab) => {
    if (tab === 'departments') return 'deptno';
    if (tab === 'employee-statuses' || tab === 'kopkara-statuses') return 'status_code';
    if (tab === 'employee-categories') return 'category_code';
    if (tab === 'employees') return 'employee_id';
    if (tab === 'members') return 'member_no';
    if (tab === 'roles') return 'role_id';
    if (tab === 'menus') return 'menu_id';
    if (tab === 'parameters') return 'key_name';
    if (tab === 'users') return 'user_id';
    return 'id';
  };

  const getFieldValue = (form, key) => {
    if (!form) return '';
    if (form[key] !== undefined && form[key] !== null) return form[key];
    const match = Object.keys(form).find(m => m.toLowerCase() === key.toLowerCase());
    if (match && form[match] !== undefined && form[match] !== null) return form[match];
    return '';
  };
`;

// Insert helpers before saveMasterData
if (!code.includes('const getMasterTitle =')) {
  code = code.replace('const saveMasterData = async (e) => {', `${helpersCode}\n  const saveMasterData = async (e) => {`);
}

// 2. Replace the master modal block at the bottom with pure JSX
const modalStartStr = "{/* Top-Level Stable Non-Flickering Master Modal UI */}";
const modalEndStr = "</main>";

const modalStartIdx = code.indexOf(modalStartStr);
const modalEndIdx = code.indexOf(modalEndStr, modalStartIdx);

if (modalStartIdx === -1 || modalEndIdx === -1) {
  console.error("Could not find modal positions!");
  process.exit(1);
}

const cleanModalJSX = `{/* Top-Level Stable Non-Flickering Master Modal UI */}
        {isMasterModalOpen && activeTab.startsWith('master-') && masterTab !== 'role-menus' && (
          <div 
            style={{ 
              position: 'fixed', 
              top: 0, 
              left: 0, 
              width: '100vw', 
              height: '100vh', 
              backgroundColor: 'rgba(15, 23, 42, 0.65)', 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center', 
              zIndex: 9999, 
              backdropFilter: 'blur(4px)',
              pointerEvents: 'auto'
            }}
          >
            <div 
              onClick={e => e.stopPropagation()}
              style={{ 
                backgroundColor: '#ffffff', 
                borderRadius: '12px', 
                width: '90%', 
                maxWidth: '560px', 
                boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)', 
                border: '1px solid #cbd5e1', 
                overflow: 'hidden' 
              }}
            >
              {/* Modal Header dengan background HEADER_BG & font putih */}
              <div style={{ background: 'var(--header-bg, #0B2545)', color: '#ffffff', padding: '16px 20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h3 style={{ margin: 0, color: '#ffffff', fontSize: '1.1rem', fontWeight: 700 }}>
                  {Boolean(getFieldValue(masterForm, getPrimaryKeyKey(masterTab)) || masterForm?.id || masterForm?.ID) ? 'Edit' : 'Tambah'} {getMasterTitle(masterTab)}
                </h3>
                <button 
                  type="button"
                  onClick={() => setIsMasterModalOpen(false)} 
                  style={{ background: 'none', border: 'none', color: '#ffffff', fontSize: '1.2rem', cursor: 'pointer', opacity: 0.85 }}
                >
                  ✕
                </button>
              </div>

              {/* Modal Body Form */}
              <div style={{ padding: '24px' }}>
                <form onSubmit={saveMasterData} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                  {(() => {
                    const pkFieldKey = getPrimaryKeyKey(masterTab);
                    const currentPkVal = getFieldValue(masterForm, pkFieldKey);
                    const isEditMode = Boolean(currentPkVal || (masterForm && (masterForm.id || masterForm.ID)));

                    let fields = [];
                    if (masterTab === 'departments') fields = [{k:'deptno', l:'Dept No', type: 'text'}, {k:'dept_name', l:'Dept Name'}];
                    else if (masterTab === 'employee-statuses' || masterTab === 'kopkara-statuses') fields = [{k:'status_code', l:'Status Code'}, {k:'description', l:'Description'}];
                    else if (masterTab === 'employee-categories') fields = [{k:'category_code', l:'Category Code'}, {k:'description', l:'Description'}, {k:'max_limit', l:'Max Limit (Number)', type:'number'}, {k:'is_eligible', l:'Is Eligible (Check for Yes)', type:'checkbox'}];
                    else if (masterTab === 'employees') fields = [
                      {k:'employee_id', l:'Employee ID (Number)', type:'number'}, 
                      {k:'name', l:'Name'}, 
                      {k:'employee_status', l:'Employee Status', type:'select', options: referenceData.employeeStatuses.map(d => ({val: d.status_code, label: d.description}))}, 
                      {k:'deptno', l:'Department', type:'select', options: referenceData.departments.map(d => ({val: d.deptno, label: d.dept_name}))}, 
                      {k:'category_code', l:'Category', type:'select', options: referenceData.employeeCategories.map(d => ({val: d.category_code, label: d.description}))},
                      {k:'role_id', l:'Role', type:'select', options: referenceData.roles.map(d => ({val: d.role_id, label: d.role_name}))},
                      {k:'salary', l:'Salary (Number)', type:'number'}
                    ];
                    else if (masterTab === 'members') fields = [
                      {k:'member_no', l:'Member No (Number)', type:'number'}, 
                      {k:'employee_id', l:'Employee', type:'select', options: referenceData.employees.map(d => ({val: d.employee_id, label: \`\${d.employee_id} - \${d.name}\`}))}, 
                      {k:'kopkara_status', l:'Kopkara Status', type:'select', options: referenceData.kopkaraStatuses.map(d => ({val: d.status_code, label: d.description}))}, 
                      {k:'join_date', l:'Join Date (YYYY-MM-DD)', type:'date'},
                      {k:'bank_name', l:'Nama Bank (Contoh: BCA, Mandiri)'},
                      {k:'bank_account_no', l:'No. Rekening Bank'},
                      {k:'bank_account_name', l:'Nama Pemilik Rekening'}
                    ];
                    else if (masterTab === 'roles') fields = [
                      {k:'role_name', l:'Role Name'},
                      {k:'description', l:'Description'}
                    ];
                    else if (masterTab === 'menus') fields = [
                      {k:'title', l:'Menu Title'},
                      {k:'icon', l:'Icon (Emoji)'},
                      {k:'path', l:'Path (Route)'},
                      {k:'order', l:'Order Sequence (Number)', type:'number'}
                    ];
                    else if (masterTab === 'role-menus') fields = [
                      {k:'role_id', l:'Role', type:'select', options: referenceData.roles.map(d => ({val: d.role_id, label: d.role_name}))},
                      {k:'menu_id', l:'Menu', type:'select', options: referenceData.menus.map(d => ({val: d.menu_id, label: \`\${d.title} (\${d.path})\`}))}
                    ];
                    else if (masterTab === 'parameters') fields = [
                      {k:'key_name', l:'Key Name'},
                      {k:'key_value', l:'Key Value'},
                      {k:'description', l:'Description'}
                    ];
                    
                    return fields.map(f => {
                      const val = getFieldValue(masterForm, f.k);
                      const isPk = (f.k === pkFieldKey);
                      const isDisabled = isEditMode && isPk;

                      return (
                        <div key={f.k} style={{ display: 'flex', flexDirection: f.type === 'checkbox' ? 'row' : 'column', alignItems: f.type === 'checkbox' ? 'center' : 'flex-start', gap: f.type === 'checkbox' ? '8px' : '4px' }}>
                          <label style={{ fontSize: '0.9rem', fontWeight: 600, color: '#334155' }}>
                            {f.l}{isDisabled ? ' (Read Only / Locked)' : ''}
                          </label>

                          {f.k === 'employee_id' && masterTab === 'members' ? (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%', backgroundColor: '#f8fafc', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1' }}>
                              <div style={{ display: 'flex', gap: '6px' }}>
                                <input 
                                  type="text"
                                  disabled={isDisabled}
                                  placeholder="🔍 Cari Employee (ID / Nama)..."
                                  value={empSelectSearchQuery}
                                  onChange={e => setEmpSelectSearchQuery(e.target.value)}
                                  onKeyDown={e => {
                                    if (e.key === 'Enter') {
                                      e.preventDefault();
                                      setEmpSelectPage(1);
                                      fetchPaginatedEmployeesForSelect(empSelectSearchQuery, 1);
                                    }
                                  }}
                                  style={{ flex: 1, padding: '8px', borderRadius: '4px', border: '1px solid #cbd5e1', fontSize: '0.85rem' }}
                                />
                                <button 
                                  type="button"
                                  disabled={isDisabled}
                                  onClick={() => {
                                    setEmpSelectPage(1);
                                    fetchPaginatedEmployeesForSelect(empSelectSearchQuery, 1);
                                  }}
                                  style={{ padding: '8px 12px', background: 'var(--button-bg, #10b981)', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 600 }}
                                >
                                  Cari
                                </button>
                              </div>

                              <select 
                                required
                                disabled={isDisabled}
                                value={val || ''}
                                onChange={e => setMasterForm({...masterForm, [f.k]: parseInt(e.target.value) || ''})}
                                style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', backgroundColor: isDisabled ? '#f1f5f9' : 'white', fontWeight: 500 }}
                              >
                                <option value="">-- Pilih Employee ({empSelectTotalRecords} data ditemukan) --</option>
                                {val && !empSelectList.some(e => String(e.employee_id) === String(val)) && (
                                  <option value={val}>
                                    Selected: Employee ID #{val}
                                  </option>
                                )}
                                {empSelectList.map(emp => (
                                  <option key={emp.employee_id} value={emp.employee_id}>
                                    {emp.employee_id} - {emp.name} ({emp.employee_id})
                                  </option>
                                ))}
                              </select>

                              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.8rem', color: '#64748b', marginTop: '2px' }}>
                                <button 
                                  type="button"
                                  disabled={empSelectPage <= 1 || empSelectLoading}
                                  onClick={() => {
                                    const newPage = Math.max(1, empSelectPage - 1);
                                    setEmpSelectPage(newPage);
                                    fetchPaginatedEmployeesForSelect(empSelectSearchQuery, newPage);
                                  }}
                                  style={{ padding: '4px 8px', borderRadius: '4px', border: '1px solid #cbd5e1', background: empSelectPage <= 1 ? '#f1f5f9' : 'white', cursor: empSelectPage <= 1 ? 'not-allowed' : 'pointer' }}
                                >
                                  ◄ Prev Page
                                </button>
                                <span>Halaman <strong>{empSelectPage}</strong> dari <strong>{empSelectTotalPages}</strong></span>
                                <button 
                                  type="button"
                                  disabled={empSelectPage >= empSelectTotalPages || empSelectLoading}
                                  onClick={() => {
                                    const newPage = Math.min(empSelectTotalPages, empSelectPage + 1);
                                    setEmpSelectPage(newPage);
                                    fetchPaginatedEmployeesForSelect(empSelectSearchQuery, newPage);
                                  }}
                                  style={{ padding: '4px 8px', borderRadius: '4px', border: '1px solid #cbd5e1', background: empSelectPage >= empSelectTotalPages ? '#f1f5f9' : 'white', cursor: empSelectPage >= empSelectTotalPages ? 'not-allowed' : 'pointer' }}
                                >
                                  Next Page ►
                                </button>
                              </div>
                            </div>
                          ) : f.type === 'select' ? (
                            <select 
                              required
                              disabled={isDisabled}
                              value={val || ''}
                              onChange={e => setMasterForm({...masterForm, [f.k]: f.k === 'employee_id' ? parseInt(e.target.value) : e.target.value})}
                              style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', backgroundColor: isDisabled ? '#f1f5f9' : 'white' }}
                            >
                              <option value="">-- Pilih {f.l} --</option>
                              {f.options && f.options.map(opt => (
                                <option key={opt.val} value={opt.val}>{opt.label} ({opt.val})</option>
                              ))}
                            </select>
                          ) : (
                            <input 
                              type={f.type || 'text'}
                              disabled={isDisabled}
                              required={f.type !== 'checkbox'}
                              checked={f.type === 'checkbox' ? Boolean(val) : undefined}
                              value={f.type !== 'checkbox' ? (val ?? '') : undefined}
                              onChange={e => setMasterForm({
                                ...masterForm, 
                                [f.k]: f.type === 'checkbox' ? e.target.checked : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value)
                              })}
                              style={f.type !== 'checkbox' ? { 
                                width: '100%', 
                                padding: '10px', 
                                borderRadius: '6px', 
                                border: '1px solid #cbd5e1',
                                backgroundColor: isDisabled ? '#f1f5f9' : '#ffffff',
                                color: isDisabled ? '#64748b' : '#0f172a',
                                fontWeight: isDisabled ? 600 : 400,
                                cursor: isDisabled ? 'not-allowed' : 'text'
                              } : { width: '20px', height: '20px' }}
                            />
                          )}
                        </div>
                      );
                    });
                  })()}
                  
                  {/* Modal Footer Buttons - Left Aligned */}
                  <div style={{ display: 'flex', justifyContent: 'flex-start', gap: '10px', marginTop: '20px' }}>
                    <button 
                      type="submit" 
                      style={{ padding: '8px 20px', background: 'var(--button-bg, #10b981)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                    >
                      Simpan Data
                    </button>
                    <button 
                      type="button" 
                      onClick={() => {
                        setMasterForm({});
                        setIsMasterModalOpen(false);
                      }} 
                      style={{ padding: '8px 20px', background: 'var(--button-cancel-bg, #64748b)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}\n        `;

code = code.substring(0, modalStartIdx) + cleanModalJSX + code.substring(modalEndIdx);
fs.writeFileSync('src/App.jsx', code, 'utf8');
console.log("Successfully refactored Master Modal to clean JSX!");
