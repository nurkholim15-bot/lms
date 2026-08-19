import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate } from 'k6/metrics';
import { BASE_URL, TEST_THRESHOLDS, LOAD_SCENARIOS, MEMBER_NO_START, MEMBER_NO_COUNT, PRODUCT_ID, SIMULATION_LATENCY_SLA, SUBMIT_LATENCY_SLA, APPROVE_LATENCY_SLA, DISBURSE_LATENCY_SLA } from './config.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

export const reqUnder300ms = new Rate('http_req_under_300ms');
export const reqUnder500ms = new Rate('http_req_under_500ms');
export const reqAbove500ms = new Rate('http_req_above_500ms');

// Skenario Stress Test & Breaking Point Test (Hingga 5.000 VUs)
export const options = {
  thresholds: TEST_THRESHOLDS,
  insecureSkipTLSVerify: true,
  scenarios: {
    stress_test_5000: LOAD_SCENARIOS.stress_test,
  },
};

export default function () {
  const headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Authorization': 'Bearer mock-token-admin',
  };

  function track(res) {
    if (res && res.timings) {
      reqUnder300ms.add(res.timings.waiting < 300);
      reqUnder500ms.add(res.timings.waiting < 500);
      reqAbove500ms.add(res.timings.waiting >= 500);
    }
    return res;
  }

  // Group 1: Health Check & Loan Products
  group('01_HealthCheck_and_Products', function () {
    const resHealth = track(http.get(`${BASE_URL}/api/health`, { headers }));
    check(resHealth, {
      'Health check status 200': (r) => r.status === 200,
    });

    const resProducts = track(http.get(`${BASE_URL}/api/products`, { headers }));
    check(resProducts, {
      'Loan products fetched successfully (200)': (r) => r.status === 200,
    });
  });

  sleep(1);

  // Group 2: Loan Simulation Calculator
  group('02_Loan_Simulation_Calculator', function () {
    const randomMemberNo = MEMBER_NO_START + Math.floor(Math.random() * MEMBER_NO_COUNT);
    const simPayload = JSON.stringify({
      member_no: randomMemberNo,
      product_id: PRODUCT_ID,
      requested_amount: 1000000,
      tenor: 1,
      salary: 8000000,
    });

    const resSim = track(http.post(`${BASE_URL}/api/applications/simulate`, simPayload, { headers }));
    check(resSim, {
      'Loan simulation status 200': (r) => r.status === 200,
      [`Simulation latency < ${SIMULATION_LATENCY_SLA}ms`]: (r) => r.timings.duration < SIMULATION_LATENCY_SLA,
    });
  });

  sleep(1);

  // Group 3: Full End-to-End Loan Flow (Submit -> Approve -> Disburse)
  group('03_Full_E2E_Loan_Lifecycle', function () {
    const randomMemberNo = MEMBER_NO_START + Math.floor(Math.random() * MEMBER_NO_COUNT);
    const submitPayload = JSON.stringify({
      member_no: randomMemberNo,
      product_id: PRODUCT_ID,
      requested_amount: 1000000,
      tenor: 1,
      notes: 'Stress Test Automated E2E Submission',
    });

    const resSubmit = track(http.post(`${BASE_URL}/api/applications`, submitPayload, { headers }));
    const isStatusOk = resSubmit.status === 201 || resSubmit.status === 200;
    check(resSubmit, {
      'Submit application status 201': () => isStatusOk,
      [`Submit response time < ${SUBMIT_LATENCY_SLA}ms`]: (r) => r.timings.duration < SUBMIT_LATENCY_SLA,
    });

    if (isStatusOk && resSubmit.json() && resSubmit.json().data) {
      const appData = resSubmit.json().data;
      const appNo = appData.application_no || appData.ApplicationNo;

      if (appNo) {
        const approvePayload = JSON.stringify({
          action: 'APPROVED',
          approved_amount: 1000000,
          notes: 'Stress Test Automated Approval',
          updated_user: '10101',
        });
        const resApprove = track(http.post(`${BASE_URL}/api/applications/${appNo}/approve`, approvePayload, { headers }));
        check(resApprove, {
          'Approve application status 200': (r) => r.status === 200,
          [`Approve response time < ${APPROVE_LATENCY_SLA}ms`]: (r) => r.timings.duration < APPROVE_LATENCY_SLA,
        });

        if (resApprove.status === 200) {
          sleep(0.01);
          const disbursePayload = JSON.stringify({
            bank_account_no: '1234567890',
            bank_name: 'BCA',
            notes: 'Stress Test Automated Disbursement',
            updated_user: '10101',
          });
          const resDisburse = track(http.post(`${BASE_URL}/api/applications/${appNo}/disburse`, disbursePayload, { headers }));
          check(resDisburse, {
            'Disburse application status 200': (r) => r.status === 200,
            [`Disburse response time < ${DISBURSE_LATENCY_SLA}ms`]: (r) => r.timings.duration < DISBURSE_LATENCY_SLA,
          });
        }
      }
    }
  });

  sleep(1);

  // Group 4: Members & Parameters Query
  group('04_Query_Members_and_Parameters', function () {
    const resMembers = track(http.get(`${BASE_URL}/api/members`, { headers }));
    check(resMembers, {
      'Members list status 200': (r) => r.status === 200,
    });

    const resParams = track(http.get(`${BASE_URL}/api/parameters`, { headers }));
    check(resParams, {
      'Parameters status 200': (r) => r.status === 200,
    });
  });

  sleep(2);
}

export function handleSummary(data) {
  return {
    'stress_test_report_5000.html': htmlReport(data, { title: 'LMS Kopkara Stress Test (5,000 VUs) Report' }),
    'stress_test_summary_5000.json': JSON.stringify(data, null, 2),
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}
