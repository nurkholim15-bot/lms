import fs from 'fs';

const code = fs.readFileSync('src/App.jsx', 'utf8');

// Find content-body start
const startIdx = code.indexOf('<div className="content-body">');
const endIdx = code.indexOf('</main>');

const bodyCode = code.substring(startIdx, endIdx);

// Count <div> and </div>
const openDivs = (bodyCode.match(/<div[\s>]/g) || []).length;
const closeDivs = (bodyCode.match(/<\/div>/g) || []).length;

console.log(`Content-Body to Main End -> Open divs: ${openDivs}, Close divs: ${closeDivs}`);

// Let's find balance line by line
const lines = bodyCode.split('\n');
let depth = 0;
lines.forEach((l, idx) => {
  const opens = (l.match(/<div[\s>]/g) || []).length;
  const closes = (l.match(/<\/div>/g) || []).length;
  depth += (opens - closes);
  if (l.includes('Modal') || l.includes('activeTab') || opens !== closes) {
    console.log(`Line ${startIdx ? 2150 + idx : idx}: depth=${depth} | ${l.trim().substring(0, 80)}`);
  }
});
