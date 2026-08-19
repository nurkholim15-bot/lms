// Config File for Grafana k6 LMS Load Test

export const BASE_URL = __ENV.BASE_URL || "https://localhost:8086";

// Konfigurasi Range Data Master (Dapat disesuaikan sesuai isi Database untuk analisis)
export const MEMBER_NO_START = parseInt(__ENV.MEMBER_NO_START || "200001");
export const MEMBER_NO_COUNT = parseInt(__ENV.MEMBER_NO_COUNT || "50000");
export const PRODUCT_ID = parseInt(__ENV.PRODUCT_ID || "1");

// Konfigurasi Target SLA Response Time (dalam milidetik / ms)
export const SIMULATION_LATENCY_SLA = parseInt(__ENV.SIMULATION_LATENCY_SLA || "300"); // Target SLA Simulasi (ms)
export const SUBMIT_LATENCY_SLA = parseInt(__ENV.SUBMIT_LATENCY_SLA || "500");       // Target SLA Submit (ms)
export const APPROVE_LATENCY_SLA = parseInt(__ENV.APPROVE_LATENCY_SLA || "500");     // Target SLA Approve (ms)
export const DISBURSE_LATENCY_SLA = parseInt(__ENV.DISBURSE_LATENCY_SLA || "1000");   // Target SLA Disburse (ms)

export const TEST_THRESHOLDS = {
  checks: ["rate>0.95"], // 95% check pass rate SLA requirement
  http_req_duration: ["p(95)<500", "p(99)<5000"], // p95 < 500ms, p99 < 5000ms for single-machine 500 VUs
  http_req_under_300ms: ["rate>0.90"], // Displays % response time < 300ms directly in K6 CLI summary
  http_req_under_500ms: ["rate>0.95"], // Displays % response time < 500ms directly in K6 CLI summary
  http_req_above_500ms: ["rate<0.05"], // Displays % response time >= 500ms directly in K6 CLI summary
};

// Ramp-up Scenarios for Load & Stress Testing
export const LOAD_SCENARIOS = {
  normal_load: {
    executor: "ramping-vus",
    startVUs: 10,
    stages: [
      { duration: "30s", target: 200 },
      { duration: "2m", target: 500 },
      { duration: "30s", target: 0 },
    ],
    gracefulStop: "10s",
  },
  stress_test: {
    executor: "ramping-vus",
    startVUs: 50,
    stages: [
      { duration: "1m", target: 1000 },
      { duration: "2m", target: 3000 },
      { duration: "1m", target: 5000 },
      { duration: "1m", target: 0 },
    ],
    gracefulStop: "30s",
  },
};
