import fs from 'fs';

let code = fs.readFileSync('src/App.jsx', 'utf8');

// 1. Ensure useEffect updates CSS variables on :root
const cssVarEffect = `
  useEffect(() => {
    const headerBg = getParamVal('HEADER_BG', '#0B2545');
    const buttonBg = getParamVal('BUTTON_BG', '#10b981');
    const buttonCancelBg = getParamVal('BUTTON_CANCEL_BG', '#64748b');
    const buttonClosedBg = getParamVal('BUTTON_CLOSED_BG', '#475569');

    document.documentElement.style.setProperty('--header-bg', headerBg);
    document.documentElement.style.setProperty('--button-bg', buttonBg);
    document.documentElement.style.setProperty('--button-cancel-bg', buttonCancelBg);
    document.documentElement.style.setProperty('--button-closed-bg', buttonClosedBg);
  }, [parameters]);
`;

// Insert cssVarEffect after parameters definition if not already present
if (!code.includes("document.documentElement.style.setProperty('--header-bg'")) {
  code = code.replace("}, [activeTab]);", "}, [activeTab]);\n" + cssVarEffect);
}

fs.writeFileSync('src/App.jsx', code, 'utf8');
console.log('App.jsx updated with CSS Var Effect!');
