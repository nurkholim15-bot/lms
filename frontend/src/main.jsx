import React, { StrictMode, Component } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error("App Crash ErrorBoundary:", error, errorInfo);
    this.setState({ errorInfo });
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          padding: '24px', background: '#fff1f2', color: '#991b1b',
          fontFamily: 'sans-serif', minHeight: '100vh', boxSizing: 'border-box'
        }}>
          <h2 style={{ fontSize: '1.4rem', marginTop: 0, color: '#e11d48' }}>⚠️ Aplikasi Mengalami Error (Crash):</h2>
          <div style={{
            background: '#ffffff', padding: '16px', borderRadius: '12px',
            border: '1px solid #fecdd3', fontSize: '0.85rem', fontFamily: 'monospace',
            whiteSpace: 'pre-wrap', wordBreak: 'break-all', marginBottom: '16px'
          }}>
            {this.state.error && this.state.error.toString()}
          </div>
          <p style={{ fontSize: '0.8rem', color: '#475569' }}>
            Detail Komponen:
          </p>
          <div style={{
            background: '#0f172a', color: '#38bdf8', padding: '12px',
            borderRadius: '8px', fontSize: '0.75rem', fontFamily: 'monospace',
            whiteSpace: 'pre-wrap', wordBreak: 'break-all', maxHeight: '200px', overflowY: 'auto'
          }}>
            {this.state.errorInfo && this.state.errorInfo.componentStack}
          </div>
          <button
            onClick={() => window.location.reload()}
            style={{
              marginTop: '20px', padding: '12px 20px', background: '#2563eb',
              color: 'white', border: 'none', borderRadius: '8px', fontWeight: 800, cursor: 'pointer'
            }}
          >
            🔄 Muat Ulang Aplikasi
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </StrictMode>,
)
