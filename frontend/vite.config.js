import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'fs'
import path from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const port = parseInt(env.VITE_PORT || 3005)
  
  const rawCert = env.SSL_CERT_PATH || process.env.SSL_CERT_PATH || ''
  const rawKey = env.SSL_KEY_PATH || process.env.SSL_KEY_PATH || ''

  let certPath = ''
  let keyPath = ''

  if (rawCert && rawKey) {
    const candidateCert1 = path.resolve(process.cwd(), rawCert)
    const candidateCert2 = path.resolve(process.cwd(), '../backend', rawCert)
    
    if (fs.existsSync(candidateCert1)) {
      certPath = candidateCert1
    } else if (fs.existsSync(candidateCert2)) {
      certPath = candidateCert2
    }

    const candidateKey1 = path.resolve(process.cwd(), rawKey)
    const candidateKey2 = path.resolve(process.cwd(), '../backend', rawKey)

    if (fs.existsSync(candidateKey1)) {
      keyPath = candidateKey1
    } else if (fs.existsSync(candidateKey2)) {
      keyPath = candidateKey2
    }
  }

  const httpsOptions = (certPath && keyPath) ? {
    cert: fs.readFileSync(certPath),
    key: fs.readFileSync(keyPath),
  } : false

  return {
    plugins: [react()],
    server: {
      host: true,
      allowedHosts: true,
      port,
      strictPort: true,
      https: httpsOptions,
    },
  }
})


