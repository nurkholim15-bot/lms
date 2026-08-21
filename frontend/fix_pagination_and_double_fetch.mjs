import fs from 'fs';

let code = fs.readFileSync('src/App.jsx', 'utf8');

// 1. Fix line 1624 pageSize prioritization in fetchMasterData
const oldPageSize = `const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || parseInt(getParamVal('PAGINATION_LIMIT', '10')) || 10;`;
const newPageSize = `const pageSize = parseInt(getParamVal('PAGINATION_LIMIT', getParamVal('DEFAULT_PAGE_SIZE', '5'))) || 5;`;

if (code.includes(oldPageSize)) {
  code = code.replace(oldPageSize, newPageSize);
  console.log("Fixed pageSize to prioritize PAGINATION_LIMIT!");
}

// 2. Remove duplicate fetchMasterData calls from activeTab useEffect
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
  console.log("Removed duplicate fetchMasterData from activeTab useEffect!");
}

fs.writeFileSync('src/App.jsx', code, 'utf8');
console.log("Successfully applied pagination limit & double fetch fixes to App.jsx!");
