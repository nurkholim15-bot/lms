import fs from 'fs';

let code = fs.readFileSync('src/App.jsx', 'utf8');

// 1. Add formatRupiahInput and parseRupiahInput helper functions if not present
const rupiahHelpers = `
  const formatRupiahInput = (val) => {
    if (val === null || val === undefined || val === '') return '';
    const numbersOnly = String(val).replace(/[^0-9]/g, '');
    if (!numbersOnly) return '';
    const number = parseInt(numbersOnly, 10);
    return isNaN(number) ? '' : \`Rp \${number.toLocaleString('id-ID')}\`;
  };

  const parseRupiahInput = (str) => {
    const numbersOnly = String(str).replace(/[^0-9]/g, '');
    return numbersOnly ? parseInt(numbersOnly, 10) : 0;
  };
`;

if (!code.includes('formatRupiahInput')) {
  code = code.replace('const saveMasterData = async (e) => {', `${rupiahHelpers}\n  const saveMasterData = async (e) => {`);
  console.log("Added formatRupiahInput and parseRupiahInput helpers!");
}

// 2. Fix fetchMasterData pageSize logic to prioritize PAGINATION_LIMIT
const oldPageSizeLogic = `const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || parseInt(getParamVal('PAGINATION_LIMIT', '10')) || 10;`;
const newPageSizeLogic = `const pageSize = parseInt(getParamVal('PAGINATION_LIMIT', '')) || parseInt(getParamVal('DEFAULT_PAGE_SIZE', '')) || 5;`;

if (code.includes(oldPageSizeLogic)) {
  code = code.replace(oldPageSizeLogic, newPageSizeLogic);
  console.log("Updated pageSize logic to prioritize PAGINATION_LIMIT!");
}

// 3. Fix double fetch in activeTab useEffect
const oldEffect1 = `  // Sync masterTab with activeTab for dynamic master menus
  useEffect(() => {
    if (activeTab.startsWith('master-')) {
      const tab = activeTab.replace('master-', '');
      setMasterTab(tab);
      setCurrentPage(1);
      setMasterSearchQuery('');
      fetchMasterData(tab, '', 1);
      setMasterForm({});
    } else if (activeTab === 'master') {
      fetchMasterData(masterTab, masterSearchQuery, currentPage);
    }
  }, [activeTab]);`;

const newEffect1 = `  // Sync masterTab with activeTab for dynamic master menus
  useEffect(() => {
    if (activeTab.startsWith('master-')) {
      const tab = activeTab.replace('master-', '');
      setMasterTab(tab);
      setCurrentPage(1);
      setMasterSearchQuery('');
      setMasterForm({});
    }
  }, [activeTab]);`;

if (code.includes(oldEffect1)) {
  code = code.replace(oldEffect1, newEffect1);
  console.log("Removed duplicate fetchMasterData call from activeTab useEffect!");
}

// 4. Update fields definition for employees tab to show Salary (IDR)
code = code.replace(
  "{k:'salary', l:'Salary (Number)', type:'number'}",
  "{k:'salary', l:'Salary (IDR)', type:'idr'}"
);

// 5. Update input rendering in modal form to handle type 'idr'
const oldInputRender = `<input 
                                    type={f.type || 'text'}
                                    disabled={isDisabled}
                                    required={f.type !== 'checkbox'}
                                    checked={f.type === 'checkbox' ? Boolean(val) : undefined}
                                    value={f.type !== 'checkbox' ? (val ?? '') : undefined}
                                    onChange={e => setMasterForm({
                                      ...masterForm, 
                                      [f.k]: f.type === 'checkbox' ? e.target.checked : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value)
                                    })}`;

const newInputRender = `<input 
                                    type={f.type === 'idr' ? 'text' : (f.type || 'text')}
                                    disabled={isDisabled}
                                    required={f.type !== 'checkbox'}
                                    checked={f.type === 'checkbox' ? Boolean(val) : undefined}
                                    value={f.type === 'idr' ? formatRupiahInput(val) : (f.type !== 'checkbox' ? (val ?? '') : undefined)}
                                    onChange={e => setMasterForm({
                                      ...masterForm, 
                                      [f.k]: f.type === 'checkbox' ? e.target.checked : (f.type === 'idr' ? parseRupiahInput(e.target.value) : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value))
                                    })}`;

if (code.includes(oldInputRender)) {
  code = code.replace(oldInputRender, newInputRender);
  console.log("Updated input rendering in modal form to support type idr!");
}

fs.writeFileSync('src/App.jsx', code, 'utf8');
console.log("Successfully applied all UI Karyawan fixes to App.jsx!");
