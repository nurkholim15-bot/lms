import fs from 'fs';

let code = fs.readFileSync('src/App.jsx', 'utf8');

const modalMarker = "        {/* Top-Level Stable Non-Flickering Master Modal UI */}";
const trackingModalEndMarker = "            <div style={{ display: 'flex', justifyContent: 'flex-start', marginTop: '20px', paddingTop: '12px', borderTop: '1px solid #e2e8f0' }}>\n              <button onClick={() => setTrackingModalOpen(false)} style={{ padding: '8px 20px', background: 'var(--button-closed-bg, #475569)', color: '#ffffff', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Tutup</button>\n            </div>\n          </div>";

const idxModal = code.indexOf(modalMarker);
const idxTrackingEnd = code.indexOf(trackingModalEndMarker);

if (idxModal === -1 || idxTrackingEnd === -1) {
  console.error("Could not find markers!");
  process.exit(1);
}

// Find where master modal block ends: "        })()}"
const modalBlockEndMarker = "        })()}";
const idxModalEnd = code.indexOf(modalBlockEndMarker, idxModal);

if (idxModalEnd === -1) {
  console.error("Could not find modalBlockEndMarker!");
  process.exit(1);
}

const masterModalCode = code.substring(idxModal, idxModalEnd + modalBlockEndMarker.length);

// Remove master modal code from inside tracking modal
code = code.substring(0, idxModal) + code.substring(idxModalEnd + modalBlockEndMarker.length);

// Now find the end of tracking modal: "      )}" after idxTrackingEnd
const trackingCloseIdx = code.indexOf("      )}", idxTrackingEnd);

if (trackingCloseIdx === -1) {
  console.error("Could not find trackingCloseIdx!");
  process.exit(1);
}

const insertPos = trackingCloseIdx + "      )}".length;

code = code.substring(0, insertPos) + "\n\n" + masterModalCode + code.substring(insertPos);
fs.writeFileSync('src/App.jsx', code, 'utf8');
console.log("Successfully fixed master modal placement in App.jsx!");
