import React, { useState, useEffect, useRef } from 'react'
import axios from 'axios'

const API_PROTOCOL = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'https:' : 'http:';
const API_BASE_URL = import.meta.env.VITE_API_URL || `${API_PROTOCOL}//localhost:8086`;
const KARISMA_SIMULATOR_URL = import.meta.env.VITE_KARISMA_URL || `${API_PROTOCOL}//localhost:8087`;

// Enable credentials (HttpOnly Cookie Transmission) across all Axios requests
axios.defaults.withCredentials = true;

// Konfigurasi Role dan Menu
const MENU_CONFIG = [
  { id: 'dashboard', label: 'Dashboard', icon: '📊', roles: ['anggota', 'admin', 'hrd'] },
  { id: 'pengajuan', label: 'Pengajuan Pinjaman', icon: '📝', roles: ['anggota', 'admin'] },
  { id: 'pinjaman', label: 'Daftar Pinjaman', icon: '💰', roles: ['anggota', 'admin', 'hrd'] },
  { id: 'approval', label: 'Approval Pinjaman', icon: '✅', roles: ['admin'] },
  { id: 'payroll', label: 'Potong Gaji (HRD)', icon: '✂️', roles: ['hrd', 'admin'] },
  { id: 'manual-repayment', label: 'Pelunasan Manual', icon: '💳', roles: ['admin'] },
  { id: 'products', label: 'Produk Pinjaman', icon: '📦', roles: ['admin', 'anggota', 'hrd'] },
  { id: 'parameters', label: 'Pengaturan Parameter', icon: '⚙️', roles: ['admin'] },
  { id: 'master', label: 'Data Master', icon: '🗃️', roles: ['admin'] },
];

function App() {
  const fileInputRef = useRef(null);
  const [activeTab, setActiveTab] = useState('dashboard');
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [masterTab, setMasterTab] = useState('departments');
  const [expandedParents, setExpandedParents] = useState({ 10: true });
  const [masterSearchQuery, setMasterSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  // Autentikasi & Sesi
  const [authToken, setAuthToken] = useState(localStorage.getItem('karisma_token') || '');
  const [currentUser, setCurrentUser] = useState(() => {
    try {
      const savedUser = localStorage.getItem('karisma_user');
      if (!savedUser) return null;
      const parsed = JSON.parse(savedUser);
      return (parsed && (parsed.employee_id || parsed.username || parsed.role)) ? parsed : null;
    } catch {
      return null;
    }
  }); // { employee_id, name, eligible, role }
  const [loginForm, setLoginForm] = useState({ username: '', password: '' });
  const [loginError, setLoginError] = useState('');
  const [products, setProducts] = useState([]);
  const [applications, setApplications] = useState([]);
  const [parameters, setParameters] = useState([]);
  const getParamVal = (key, defaultVal) => {
    const p = (parameters || []).find(item => item.key_name === key);
    return (p && p.key_value !== undefined && p.key_value !== null) ? p.key_value : defaultVal;
  };
  const [loading, setLoading] = useState(false);
  const [simulation, setSimulation] = useState(null);
  const [dateError, setDateError] = useState('');
  const [payrollReconciled, setPayrollReconciled] = useState(false);
  const [importedRows, setImportedRows] = useState([]);
  const [trackingModalOpen, setTrackingModalOpen] = useState(false);
  const [trackingList, setTrackingList] = useState([]);
  const [trackingAppNo, setTrackingAppNo] = useState(null);
  const [manualForm, setManualForm] = useState({
    loan_no: '',
    member_no: '',
    period: '2026-08',
    payment_type: 'TRANSFER_BANK',
    nominal: '',
    reference_no: '',
    notes: '',
    is_full_settlement: false
  });
  const [manualLogs, setManualLogs] = useState([]);
  const [auditLogPage, setAuditLogPage] = useState(1);
  const [receiptData, setReceiptData] = useState(null);
  const [receiptModalOpen, setReceiptModalOpen] = useState(false);
  const [allMembers, setAllMembers] = useState([]);
  const [selectedMemberFilter, setSelectedMemberFilter] = useState('');
  const [manualMonth, setManualMonth] = useState('08');
  const [manualYear, setManualYear] = useState('2026');
  const [memberSearchQuery, setMemberSearchQuery] = useState('');
  const [reportMonthFilter, setReportMonthFilter] = useState({ year: '2026', month: '08' });
  const [reportStatusFilter, setReportStatusFilter] = useState('ALL');
  const [reportMemberSearchQuery, setReportMemberSearchQuery] = useState('');
  const [reportPage, setReportPage] = useState(1);
  const [memberPage, setMemberPage] = useState(1);
  const [paginatedMemberList, setPaginatedMemberList] = useState([]);
  const [memberTotalRecords, setMemberTotalRecords] = useState(0);
  const [memberTotalPages, setMemberTotalPages] = useState(1);
  const [memberDropdownLoaded, setMemberDropdownLoaded] = useState(false);
  const memberLoadingRef = useRef(false);

  const fetchPaginatedMembers = async (q, page) => {
    if (memberLoadingRef.current) return;
    memberLoadingRef.current = true;
    try {
      const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || 10;
      const res = await axios.get(`${API_BASE_URL}/api/members?q=${encodeURIComponent(q || '')}&page=${page || 1}&limit=${pageSize}`);
      setPaginatedMemberList(res.data.data || []);
      setMemberTotalRecords(res.data.total_records || 0);
      setMemberTotalPages(res.data.total_pages || 1);
      setMemberDropdownLoaded(true);
    } catch (err) {
      console.error("Error fetching members:", err);
    } finally {
      memberLoadingRef.current = false;
    }
  };

  const handleMemberSearchChange = (newQ) => {
    setMemberSearchQuery(newQ);
    setMemberPage(1);
  };

  // State & Fetcher untuk Paginated Searchable Employee Dropdown di Master Members
  const [empSelectSearchQuery, setEmpSelectSearchQuery] = useState('');
  const [empSelectPage, setEmpSelectPage] = useState(1);
  const [empSelectTotalPages, setEmpSelectTotalPages] = useState(1);
  const [empSelectTotalRecords, setEmpSelectTotalRecords] = useState(0);
  const [empSelectList, setEmpSelectList] = useState([]);
  const [empSelectLoading, setEmpSelectLoading] = useState(false);

  const fetchPaginatedEmployeesForSelect = async (q, page) => {
    setEmpSelectLoading(true);
    try {
      const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || parseInt(getParamVal('PAGINATION_LIMIT', '10')) || 10;
      const res = await axios.get(`${API_BASE_URL}/api/master/employees?q=${encodeURIComponent(q || '')}&page=${page || 1}&limit=${pageSize}`);
      setEmpSelectList(res.data.data || []);
      setEmpSelectTotalRecords(res.data.total_records || (res.data.data ? res.data.data.length : 0));
      setEmpSelectTotalPages(res.data.total_pages || 1);
    } catch (err) {
      console.error("Error fetching employees for select:", err);
    } finally {
      setEmpSelectLoading(false);
    }
  };
  const [loanSearchQuery, setLoanSearchQuery] = useState('');
  const [loanPage, setLoanPage] = useState(1);
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [exportCutoffDate, setExportCutoffDate] = useState('');
  const [exportPeriodMonth, setExportPeriodMonth] = useState('08');
  const [exportPeriodYear, setExportPeriodYear] = useState('2026');
  const [exportCustomFolder, setExportCustomFolder] = useState('');
  const [productModalOpen, setProductModalOpen] = useState(false);
  const [productSearchQuery, setProductSearchQuery] = useState('');
  const [editingProduct, setEditingProduct] = useState(null);
  const [productForm, setProductForm] = useState({
    name: '',
    loan_type: 'FLAT',
    max_tenor_months: 24,
    submission_period_start: 1,
    submission_period_end: 25,
    max_percentage_salary: 40.0,
    interest_rate: 1.5,
    status: 'ACTIVE'
  });

  const [reconcileClosingInfo, setReconcileClosingInfo] = useState({ status: 'OPEN', hrd_signatory: '', finance_signatory: '', kopkara_signatory: '', closing_notes: '' });
  const [historyModalOpen, setHistoryModalOpen] = useState(false);
  const [adjustmentsList, setAdjustmentsList] = useState([]);
  const [adjustModalOpen, setAdjustModalOpen] = useState(false);
  const [adjustTargetItem, setAdjustTargetItem] = useState(null);
  const [adjustNotes, setAdjustNotes] = useState('Data dikembalikan ke HRD-Adira');
  const [adjustType, setAdjustType] = useState('FAILED_CORRECTION');
  const [closeModalOpen, setCloseModalOpen] = useState(false);
  const [closeForm, setCloseForm] = useState({
    hrd_signatory: 'Adira HRD Officer',
    finance_signatory: 'LMS Finance Officer',
    kopkara_signatory: 'Pengurus Kopkara',
    closing_notes: 'Laporan Rekonsiliasi Payroll telah diperiksa, disetujui, dan ditandatangani secara sah.'
  });

  const fetchReconciliationStatus = async () => {
    try {
      const res = await axios.get(`${API_BASE_URL}/api/payroll/reconciliation-status?period=2026-08`);
      if (res.data && res.data.data) {
        setReconcileClosingInfo(res.data.data);
      }
    } catch (err) {
      console.error("Error fetching reconciliation status:", err);
    }
  };

  useEffect(() => {
    fetchReconciliationStatus();
  }, []);

  const fetchAdjustments = async () => {
    try {
      const res = await axios.get(`${API_BASE_URL}/api/payroll/adjustments?period=2026-08`);
      if (res.data && res.data.adjustments) {
        setAdjustmentsList(res.data.adjustments || []);
      }
    } catch (err) {
      console.error("Error fetching adjustments:", err);
    }
  };

  const handleSaveAdjustment = async (e) => {
    e.preventDefault();
    if (!adjustTargetItem) return;
    try {
      const activeUserName = currentUser ? String(currentUser.employee_id || '10101') : '10101';
      const tagihan = Math.round(adjustTargetItem?.amount || 0);
      const terpotong = Math.round(adjustTargetItem?.deducted || 0);
      const isOverpay = adjustType.includes('OVERPAYMENT') || (terpotong - tagihan > 0);
      const refundNominal = Math.abs(terpotong - tagihan);
      
      const targetLoanNo = adjustTargetItem?.loanNo || adjustTargetItem?.loan_no || 0;
      const payload = {
        ref_no: String(targetLoanNo || adjustTargetItem?.refNo || ''),
        loan_no: targetLoanNo,
        period: adjustTargetItem.period || '2026-08',
        adjustment_type: adjustType,
        original_amount: tagihan,
        deducted_amount: terpotong,
        adjusted_amount: isOverpay ? refundNominal : (adjustType === 'FAILED_CORRECTION' ? 0 : tagihan),
        notes: adjustNotes,
        created_user: activeUserName
      };

      const res = await axios.post(`${API_BASE_URL}/api/payroll/adjust`, payload);
      alert(`✅ ${res.data.message}`);
      setAdjustModalOpen(false);
      setAdjustTargetItem(null);
      fetchAdjustments();
    } catch (err) {
      alert(`❌ Gagal menyimpan adjustment: ${err.response?.data?.error || err.message}`);
    }
  };

  const handleSaveCloseReconciliation = async (e) => {
    e.preventDefault();
    try {
      const activeUserName = currentUser ? String(currentUser.employee_id || '10101') : '10101';
      const payload = {
        period: '2026-08',
        hrd_signatory: closeForm.hrd_signatory,
        finance_signatory: closeForm.finance_signatory,
        kopkara_signatory: closeForm.kopkara_signatory,
        closing_notes: closeForm.closing_notes,
        closed_user: activeUserName
      };

      const res = await axios.post(`${API_BASE_URL}/api/payroll/close-reconciliation`, payload);
      alert(`🔒 ${res.data.message}`);
      setCloseModalOpen(false);
      fetchReconciliationStatus();
      fetchPayrollSchedules();
    } catch (err) {
      alert(`❌ Gagal menutup rekonsiliasi: ${err.response?.data?.error || err.message}`);
    }
  };

  const handleOpenTracking = async (applicationNo) => {
    setTrackingAppNo(applicationNo);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/applications/${applicationNo}/trackings`);
      setTrackingList(response.data.data || []);
      setTrackingModalOpen(true);
    } catch (err) {
      alert("Gagal mengambil riwayat tracking: " + (err.response?.data?.error || err.message));
    }
  };

  // Disbursement UI Modal State
  const [disbursementModalOpen, setDisbursementModalOpen] = useState(false);
  const [selectedDisburseApp, setSelectedDisburseApp] = useState(null);
  const [disburseForm, setDisburseForm] = useState({
    bank_name: 'BCA',
    custom_bank: '',
    bank_account_no: '',
    account_holder_name: '',
    notes: ''
  });
  const [disburseLoading, setDisburseLoading] = useState(false);

  // Master Data Generic State
  const [masterDataList, setMasterDataList] = useState([]);
  const [masterForm, setMasterForm] = useState({});
  const [referenceData, setReferenceData] = useState({
    departments: [],
    employeeStatuses: [],
    kopkaraStatuses: [],
    employeeCategories: [],
    employees: [],
    members: [],
    roles: [],
    menus: [],
    role_menus: []
  });

  // Form State
  const [form, setForm] = useState({ member_no: 1001, product_id: '', requested_amount: '', tenor: '' });
  const [paramForm, setParamForm] = useState({ id: 0, key_name: '', key_value: '', description: '' });

  const [userInfo, setUserInfo] = useState(null);
  const [dashboardSummary, setDashboardSummary] = useState({
    credit_limit: 15000000,
    available_limit: 15000000,
    total_debt: 0,
    active_loans: 0,
    recent_loans: []
  });

  const fetchDashboardSummary = async () => {
    try {
      const empId = currentUser?.employee_id || '';
      const currentRole = roleId || userInfo?.role_id || userInfo?.role_name || realRoleName || '';
      const res = await axios.get(`${API_BASE_URL}/api/dashboard/summary?employee_id=${encodeURIComponent(empId)}&role_id=${encodeURIComponent(currentRole)}`);
      if (res.data) {
        setDashboardSummary(res.data);
      }
    } catch (err) {
      console.error("Error fetching dashboard summary:", err);
    }
  };

  // Fetch User Info & APM Menus directly from Backend API with SQL filters
  useEffect(() => {
    if (currentUser?.employee_id) {
      axios.get(`${API_BASE_URL}/api/user-info/${currentUser.employee_id}`)
        .then(res => {
          const info = res.data;
          setUserInfo(info);

          // Pengecekan Keanggotaan: Jika bukan anggota (is_member === false) dan bukan akun superadmin
          const isSuperAdmin = String(currentUser.username || '').toLowerCase() === 'admin';
          if (!info.is_member && !isSuperAdmin) {
            alert("Anda bukan anggota sehingga tidak boleh login");
            handleLogout("Anda bukan anggota sehingga tidak boleh login");
          }
        })
        .catch(err => console.error("Error fetching user info:", err));
    } else {
      setUserInfo(null);
    }
  }, [currentUser]);

  // Cek role asli Karyawan dari Database LMS
  const realEmployee = (currentUser && referenceData.employees.length > 0)
    ? referenceData.employees.find(e => String(e.employee_id) === String(currentUser.employee_id))
    : null;
    
  let roleId = userInfo ? userInfo.role_id : (realEmployee && realEmployee.role_id ? realEmployee.role_id : null);
  let realRoleName = userInfo ? userInfo.role_name : (realEmployee && realEmployee.role ? realEmployee.role : 'Anggota');

  // Menyusun Dynamic Menus dari APM (High performance backend API fallback to frontend state)
  const allowedMenuIds = referenceData.role_menus
    .filter(rm => String(rm.role_id) === String(roleId))
    .map(rm => String(rm.menu_id));

  const visibleMenus = (userInfo && userInfo.menus && userInfo.menus.length > 0) 
    ? userInfo.menus 
    : referenceData.menus
        .filter(m => allowedMenuIds.includes(String(m.menu_id)))
        .sort((a, b) => (a.order || 0) - (b.order || 0));

  // Update HTML Document Title
  useEffect(() => {
    if (!currentUser) {
      document.title = 'Kopkara LMS - Login';
      return;
    }
    const roleStr = String(realRoleName || 'Anggota');
    document.title = `Kopkara LMS - ${roleStr.charAt(0).toUpperCase() + roleStr.slice(1)}`;
  }, [currentUser, realRoleName]);

  // Fungsi untuk memanggil Backend (API Products)
  const fetchProducts = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/products`);
      setProducts(response.data.data || []);
    } catch (error) {
      console.error("Error fetching products:", error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status) => {
    if (status === 'APPROVED' || status === 'DISBURSED') return 'badge badge-success';
    if (status === 'SUBMITTED' || status === 'REVISION_REQUIRED') return 'badge badge-warning';
    return 'badge badge-danger';
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const d = new Date(dateStr);
    if (isNaN(d.getTime()) || d.getFullYear() <= 1970) return '-';
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    const day = String(d.getDate()).padStart(2, '0');
    const month = months[d.getMonth()];
    const year = d.getFullYear();
    const hours = String(d.getHours()).padStart(2, '0');
    const mins = String(d.getMinutes()).padStart(2, '0');
    const secs = String(d.getSeconds()).padStart(2, '0');
    return `${day}-${month}-${year} ${hours}:${mins}:${secs}`;
  };

  const [approvalMonthFilter, setApprovalMonthFilter] = useState({ month: '08', year: '2026' });
  const [disbursementMonthFilter, setDisbursementMonthFilter] = useState({ month: '08', year: '2026', status: 'ALL' });
  const [loanMonthFilter, setLoanMonthFilter] = useState({ month: '08', year: '2026' });

  const [loanSearchMemberNo, setLoanSearchMemberNo] = useState('');
  const [hasSearchedLoans, setHasSearchedLoans] = useState(false);

  // Sync loan search member no when user changes or tab changes
  useEffect(() => {
    if (activeTab === 'pinjaman') {
      setHasSearchedLoans(false);
      setApplications([]);
      const isHighPriv = Boolean(
        roleId === 1 || roleId === 3 || 
        String(realRoleName).toLowerCase().includes('admin') || 
        String(realRoleName).toLowerCase().includes('hrd')
      );
      if (!isHighPriv && currentUser?.employee_id) {
        setLoanSearchMemberNo(String(currentUser.employee_id));
      } else if (isHighPriv && !loanSearchMemberNo) {
        setLoanSearchMemberNo('');
      }
    }
  }, [activeTab, currentUser, roleId, realRoleName]);

  const handleSearchLoans = (overrideNo, overrideYear, overrideMonth) => {
    const isHighPriv = Boolean(
      roleId === 1 || roleId === 3 || 
      String(realRoleName).toLowerCase().includes('admin') || 
      String(realRoleName).toLowerCase().includes('hrd')
    );
    
    let targetNo = overrideNo !== undefined ? overrideNo : loanSearchMemberNo;
    if (!isHighPriv && currentUser?.employee_id) {
      targetNo = String(currentUser.employee_id);
      setLoanSearchMemberNo(targetNo);
    }

    const yr = overrideYear || loanMonthFilter.year;
    const mo = overrideMonth || loanMonthFilter.month;

    setHasSearchedLoans(true);
    fetchApplications(yr, mo, targetNo);
  };

  const isReportTab = (tabPath) => {
    const p = String(tabPath || activeTab || '').toLowerCase();
    const title = String((visibleMenus || []).find(m => m.path === (tabPath || activeTab))?.title || '').toLowerCase();
    return p === 'report-loan-applications' || 
           p === 'report-applications' || 
           p === 'laporan-pengajuan-pinjaman' || 
           p === 'laporan-pengajuan' || 
           p.includes('report-loan') || 
           p.includes('laporan-pengajuan') || 
           title.includes('laporan pengajuan');
  };

  const fetchApplications = async (targetYear, targetMonth, filterMemberNo, limit, offset) => {
    try {
      const isReport = isReportTab();
      const yr = targetYear || (activeTab === 'disbursement' ? disbursementMonthFilter.year : (activeTab === 'pinjaman' ? loanMonthFilter.year : (isReport ? reportMonthFilter.year : approvalMonthFilter.year))) || '2026';
      const mo = targetMonth || (activeTab === 'disbursement' ? disbursementMonthFilter.month : (activeTab === 'pinjaman' ? loanMonthFilter.month : (isReport ? reportMonthFilter.month : approvalMonthFilter.month))) || '08';
      const period = `${yr}${mo}`;
      let url = `${API_BASE_URL}/api/applications?period=${period}`;
      if (activeTab === 'approval') {
        url += `&status=SUBMITTED`;
      } else if (activeTab === 'disbursement') {
        url += `&status=APPROVED`;
      } else if (isReport && reportStatusFilter && reportStatusFilter !== 'ALL') {
        url += `&status=${encodeURIComponent(reportStatusFilter)}`;
      }

      // Pass user role and employee_id for confidentiality & privilege checks
      const currentEmpId = currentUser?.employee_id || '';
      const currentRole = roleId || realRoleName || '';
      url += `&role_id=${encodeURIComponent(currentRole)}&user_employee_id=${encodeURIComponent(currentEmpId)}`;

      const searchNo = filterMemberNo !== undefined ? filterMemberNo : (isReport ? reportMemberSearchQuery : loanSearchMemberNo);
      if (searchNo) {
        url += `&member_no=${encodeURIComponent(searchNo)}`;
      }

      // Pass pagination parameters LIMIT & OFFSET to SQL
      if (limit) {
        url += `&limit=${limit}&offset=${offset || 0}`;
      }

      const response = await axios.get(url);
      setApplications(response.data.data || []);
    } catch (error) {
      console.error("Error fetching applications:", error);
      setApplications([]);
    }
  };

  const [payrollSchedules, setPayrollSchedules] = useState([]);

  const fetchPayrollSchedules = async (targetMemberNo) => {
    try {
      let url = `${API_BASE_URL}/api/payroll/schedules?period=2026-08`;
      if (targetMemberNo) {
        url += `&member_no=${encodeURIComponent(targetMemberNo)}`;
      }
      const response = await axios.get(url);
      if (response.data) {
        if (Array.isArray(response.data)) {
          setPayrollSchedules(response.data);
        } else if (response.data.data) {
          setPayrollSchedules(response.data.data || []);
        }
        if (response.data.adjustments) {
          setAdjustmentsList(response.data.adjustments || []);
        }
      }
    } catch (error) {
      console.error("Error fetching payroll schedules:", error);
    }
  };

  // Fetch parameters and products once on initial mount
  useEffect(() => {
    fetchParameters();
    fetchProducts();
  }, []);

  useEffect(() => {
    if (activeTab === 'dashboard') {
      fetchDashboardSummary();
    }

    if (activeTab === 'approval' || activeTab === 'disbursement' || activeTab === 'pengajuan' || isReportTab()) {
      fetchApplications();
    }

    if (activeTab === 'pengajuan' || activeTab === 'master') {
      axios.get(`${API_BASE_URL}/api/members/all`)
        .then(res => setAllMembers(res.data.data || []))
        .catch(err => console.error("Error fetching all members:", err));
    }

  }, [activeTab, currentUser, userInfo, roleId, realRoleName]);

  const handleProcessApproval = async (applicationNo, action) => {
    const labels = {
      APPROVED: 'SETUJU',
      REJECTED: 'TOLAK',
      REVISION_REQUIRED: 'REVISI'
    };

    const notes = window.prompt(`Masukkan Catatan Approval untuk tindakan [${labels[action]}] (Wajib):`);
    if (notes === null) return; // User canceled
    if (!notes.trim()) {
      alert("❌ Catatan approval wajib diisi!");
      return;
    }

    try {
      const activeUserId = currentUser ? String(currentUser.employee_id || '10101') : '10101';
      await axios.post(`${API_BASE_URL}/api/applications/${applicationNo}/approve`, {
        action: action,
        notes: notes.trim(),
        updated_user: activeUserId
      });
      alert(`Status pengajuan berhasil diubah menjadi [${labels[action]}]!`);
      fetchApplications();
    } catch (err) {
      alert("Gagal memproses approval: " + (err.response?.data?.error || err.message));
    }
  };

  const openDisburseModal = (app) => {
    try {
      console.log("openDisburseModal triggered for application:", app);
      const appNo = app.application_no || app.ApplicationNo;
      const memberNo = app.member_no || app.MemberNo;
      const approvedAmount = app.approved_amount ?? app.ApprovedAmount ?? app.requested_amount ?? app.RequestedAmount ?? 0;
      const tenor = app.tenor ?? app.Tenor ?? 0;

      const memberObj = (referenceData.members || []).find(m => String(m.member_no || m.MemberNo) === String(memberNo)) || {};
      const defaultBank = memberObj.bank_name || memberObj.BankName || "BCA";
      const defaultAcc = memberObj.bank_account_no || memberObj.BankAccountNo || "1234567890";
      const defaultHolder = memberObj.name || memberObj.Name || `Member #${memberNo}`;

      setSelectedDisburseApp({
        ...app,
        application_no: appNo,
        member_no: memberNo,
        approved_amount: approvedAmount,
        tenor: tenor
      });

      setDisburseForm({
        bank_name: defaultBank,
        custom_bank: '',
        bank_account_no: defaultAcc,
        account_holder_name: defaultHolder,
        notes: `Pencairan dana pinjaman #${appNo} via Transfer ${defaultBank} No. Rek: ${defaultAcc}`
      });

      setDisbursementModalOpen(true);
    } catch (err) {
      console.error("Error opening disbursement modal:", err);
      alert("Terjadi kesalahan saat membuka modal pencairan: " + err.message);
    }
  };

  const handleDisburse = (app) => {
    openDisburseModal(app);
  };

  const submitDisbursement = async (e) => {
    e.preventDefault();
    if (!selectedDisburseApp) return;

    try {
      setDisburseLoading(true);
      const appNo = selectedDisburseApp.application_no;
      const finalBank = disburseForm.bank_name === 'Lainnya' ? disburseForm.custom_bank : disburseForm.bank_name;
      const activeUserId = currentUser ? String(currentUser.employee_id || '100001') : '100001';
      const authToken = localStorage.getItem('token') || `mock-token-${activeUserId}`;

      console.log(`Submitting disbursement POST to /api/applications/${appNo}/disburse...`, disburseForm);

      const response = await axios.post(
        `${API_BASE_URL}/api/applications/${appNo}/disburse`,
        {
          bank_name: finalBank,
          bank_account_no: disburseForm.bank_account_no,
          notes: disburseForm.notes || `Pencairan dana via Transfer ${finalBank} a/n ${disburseForm.account_holder_name} (${disburseForm.bank_account_no})`,
          updated_user: activeUserId
        },
        {
          headers: {
            'Authorization': `Bearer ${authToken}`,
            'Content-Type': 'application/json'
          }
        }
      );

      console.log("Disbursement POST Response:", response.data);
      setDisbursementModalOpen(false);
      alert(`✅ Pencairan dana pinjaman #${appNo} sebesar Rp ${Number(selectedDisburseApp.approved_amount).toLocaleString('id-ID')} berhasil diproses!`);
      fetchApplications();
    } catch (err) {
      console.error("Error submitting disbursement:", err);
      alert("Gagal memproses pencairan: " + (err.response?.data?.error || err.message));
    } finally {
      setDisburseLoading(false);
    }
  };

  const handleCSVUpload = (e) => {
    const targetInput = e.target;
    const file = targetInput.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (evt) => {
      const text = evt.target.result;
      const lines = text.split(/\r?\n/).map(l => l.trim()).filter(l => l.length > 0);
      if (lines.length <= 1) {
        alert("File CSV kosong atau format header tidak sesuai.");
        return;
      }

      const delimiter = lines[0].includes(';') ? ';' : ',';
      const cleanHeaderStr = (str) => (str || '').replace(/^\uFEFF/, '').replace(/"/g, '').trim().toUpperCase();
      const headers = lines[0].split(delimiter).map(h => cleanHeaderStr(h));
      const parsedData = [];
      const importPayload = [];

      for (let i = 1; i < lines.length; i++) {
        const cols = lines[i].split(delimiter).map(c => c.replace(/^"/, '').replace(/"$/, '').trim());
        if (cols.length >= 2) {
          const rowObj = {};
          headers.forEach((h, idx) => {
            rowObj[h] = cols[idx] || '';
          });

          const cleanNum = (str) => {
            if (!str) return 0;
            const cleaned = String(str).replace(/"/g, '').replace(/[^0-9.,-]/g, '').trim();
            if (!cleaned) return 0;
            return parseFloat(cleaned.replace(',', '.')) || 0;
          };

          const refNo = rowObj['NO_REFERENSI'] || rowObj['NO_REF'] || `LMS-PAY-202608-${i}`;
          const rawEmpId = rowObj['EMPLOYEE_ID'] || rowObj['EMPLOYEE'] || (refNo.includes('-') ? refNo.split('-').pop() : '110102');
          const empId = parseInt(rawEmpId) || 110102;
          const period = rowObj['PERIODE'] || '2026-08';
          const status = rowObj['STATUS_POTONGAN'] || rowObj['STATUS'] || 'SUCCESS';
          
          const rawDeducted = rowObj['NOMINAL_TERPOTONG'] || rowObj['NOMINAL_POTONGAN'] || rowObj['DEDUCTED'] || '0';
          const rawAmount = rowObj['NOMINAL_TAGIHAN'] || rowObj['NOMINAL_POTONGAN'] || rowObj['AMOUNT'] || '0';

          let deducted = cleanNum(rawDeducted);
          let amount = cleanNum(rawAmount);

          // If NOMINAL_TERPOTONG is blank/0 but NOMINAL_TAGIHAN is provided, fallback deducted = amount
          if (deducted === 0 && amount > 0) {
            deducted = amount;
          }
          if (amount === 0 && deducted > 0) {
            amount = deducted;
          }

          const keterangan = rowObj['KETERANGAN'] || 'Potongan gaji diproses';

          parsedData.push({
            refNo: refNo,
            nik: rowObj['NIK_ADIRA'] || rowObj['NIK'] || '3171012345670001',
            name: rowObj['NAMA_KARYAWAN'] || rowObj['NAMA'] || `Employee #${cols[1] || i}`,
            period: period,
            amount: amount,
            deducted: deducted,
            status: status,
            keterangan: keterangan
          });

          const rawLoanNo = rowObj['LOAN_NO'] || rowObj['LOAN_ID'] || rowObj['NO_LOAN'] || '0';
          const loanNo = parseInt(rawLoanNo) || 0;

          importPayload.push({
            ref_no: refNo,
            employee_id: empId,
            loan_no: loanNo,
            period: period,
            nominal_original: amount,
            deducted: deducted,
            status: status,
            keterangan: keterangan
          });
        }
      }

      setImportedRows(parsedData);
      setPayrollReconciled(true);

      // Call Backend to update lms_sch.loan_schedules & insert lms_sch.payroll_deductions!
      try {
        const activeUserName = currentUser ? String(currentUser.employee_id || '10101') : '10101';
        const res = await axios.post(`${API_BASE_URL}/api/payroll/import`, {
          file_name: file.name,
          updated_user: activeUserName,
          rows: importPayload
        });
        if (res.data && res.data.logs && res.data.logs.length > 0) {
          setImportedRows(res.data.logs);
        }
        alert(`✅ Berhasil mengimpor file [${file.name}] dari folder komputer Anda!\n\n${res.data.message}`);
      } catch (err) {
        console.error("Error updating DB schedule status:", err);
        const errMsg = err.response?.data?.error || err.message;
        if (errMsg.includes("sudah pernah diupdate")) {
          alert(`⚠️ Notifikasi Import: File import [${file.name}] sudah pernah diupdate.`);
        } else {
          alert(`❌ Gagal mengimpor file [${file.name}]: ${errMsg}`);
        }
      } finally {
        if (targetInput) targetInput.value = '';
      }
    };
    reader.readAsText(file);
  };

  const handlePrintReconciliationReport = async (displayList, totalAmount, totalDeducted, totalShortage) => {
    let adjList = [];
    try {
      const res = await axios.get(`${API_BASE_URL}/api/payroll/adjustments?period=2026-08`);
      adjList = res.data.adjustments || [];
    } catch (e) {
      console.error("Error fetching adjustments for print report:", e);
    }

    let adjRowsHtml = '';
    if (adjList.length === 0) {
      adjRowsHtml = `<tr><td colspan="6" style="text-align: center; color: #64748b; padding: 10px;">Belum ada record adjustment selisih pada periode ini.</td></tr>`;
    } else {
      adjList.forEach((adj, idx) => {
        adjRowsHtml += `
          <tr>
            <td style="padding: 6px 8px; text-align: center;">${idx + 1}</td>
            <td style="padding: 6px 8px;"><strong>${adj.ref_no || '-'}</strong></td>
            <td style="padding: 6px 8px;">${adj.employee_name || '-'} (#${adj.member_no || ''})</td>
            <td style="padding: 6px 8px;"><span style="background: #e0e7ff; color: #3730a3; padding: 2px 6px; border-radius: 4px; font-weight: bold; font-size: 0.75rem;">${adj.adjustment_type || '-'}</span></td>
            <td style="padding: 6px 8px; text-align: right;"><strong>Rp ${Math.round(adj.adjusted_amount || 0).toLocaleString('id-ID')}</strong></td>
            <td style="padding: 6px 8px; font-style: italic; color: #334155;">"${adj.notes || '-'}"</td>
          </tr>
        `;
      });
    }

    const reportWin = window.open('', '_blank');
    const todayStr = new Date().toLocaleDateString('id-ID', { year: 'numeric', month: 'long', day: 'numeric' });

    let rowsHtml = '';
    (displayList || []).forEach((item, idx) => {
      const tagihanHrd = Math.round(item?.amount || 0);
      const isFailed = item?.status === 'FAILED';
      const isAdjusted = item?.status === 'ADJUSTED';
      const terpotongHrd = isFailed ? 0 : Math.round(item?.deducted || 0);
      const selisih = isAdjusted ? 0 : (tagihanHrd - terpotongHrd);
      
      let statusRecon = '✅ MATCH (SELESAI)';
      let statusColor = '#047857';
      if (isAdjusted) {
        statusRecon = '⚙️ ADJUSTED (SELESAI)';
        statusColor = '#6366f1';
      } else if (isFailed) {
        statusRecon = '❌ FAILED (PINJAMAN TIDAK DITEMUKAN)';
        statusColor = '#b91c1c';
      } else if (terpotongHrd === 0 && tagihanHrd > 0) {
        statusRecon = '⏳ PENDING / BELUM TERPOTONG';
        statusColor = '#b45309';
      } else if (selisih > 0) {
        statusRecon = `⚠️ PARTIAL (KURANG Rp ${selisih.toLocaleString('id-ID')})`;
        statusColor = '#d97706';
      } else if (selisih < 0) {
        statusRecon = `🔵 OVERPAYMENT (LEBIH Rp ${Math.abs(selisih).toLocaleString('id-ID')})`;
        statusColor = '#0284c7';
      }

      const refNoCol = item?.loanNo || item?.loan_no || item?.refNo || '-';
      rowsHtml += `
        <tr style="border-bottom: 1px solid #cbd5e1;">
          <td style="padding: 6px 10px; text-align: center; white-space: nowrap;">${idx + 1}</td>
          <td style="padding: 6px 10px; white-space: nowrap; font-weight: bold;">${refNoCol}</td>
          <td style="padding: 6px 10px; white-space: nowrap;">${item?.nik || '-'}</td>
          <td style="padding: 6px 10px; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${item?.name || '-'}</td>
          <td style="padding: 6px 10px; text-align: center; white-space: nowrap;">${item?.period || '2026-08'}</td>
          <td style="padding: 6px 10px; text-align: right; white-space: nowrap;">Rp ${tagihanHrd.toLocaleString('id-ID')}</td>
          <td style="padding: 6px 10px; text-align: right; white-space: nowrap; font-weight: bold; color: ${statusColor};">Rp ${terpotongHrd.toLocaleString('id-ID')}</td>
          <td style="padding: 6px 10px; text-align: right; white-space: nowrap; color: ${selisih > 0 ? '#b91c1c' : (selisih < 0 ? '#0284c7' : '#047857')}; font-weight: bold;">
            ${selisih === 0 ? 'Rp 0' : (selisih > 0 ? `+Rp ${selisih.toLocaleString('id-ID')}` : `-Rp ${Math.abs(selisih).toLocaleString('id-ID')}`)}
          </td>
          <td style="padding: 6px 10px; font-weight: bold; color: ${statusColor}; white-space: nowrap;">${statusRecon}</td>
          <td style="padding: 6px 10px; font-style: italic; font-size: 0.8rem; color: ${isFailed ? '#b91c1c' : '#334155'};">${item?.keterangan || (isFailed ? 'Nomor Pinjaman tidak ditemukan di DB LMS' : '-')}</td>
        </tr>
      `;
    });

    reportWin.document.write(`
      <!DOCTYPE html>
      <html>
      <head>
        <title>Laporan Rekonsiliasi Payroll - LMS Kopkara</title>
        <style>
          body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding: 24px; color: #1e293b; }
          .header { text-align: center; border-bottom: 3px double #0284c7; padding-bottom: 12px; margin-bottom: 20px; }
          .header h2 { margin: 0; color: #0f172a; font-size: 1.3rem; }
          .header p { margin: 4px 0 0 0; color: #64748b; font-size: 0.85rem; }
          .summary-table { width: 100%; border-collapse: collapse; margin-bottom: 20px; font-size: 0.85rem; }
          .summary-table td { padding: 8px 12px; border: 1px solid #cbd5e1; }
          .main-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
          .main-table th { background: #0f172a; color: white; padding: 8px 10px; border: 1px solid #0f172a; text-align: left; }
          .main-table td { border: 1px solid #cbd5e1; }
          .signatures { display: flex; justify-content: space-between; margin-top: 45px; text-align: center; font-size: 0.85rem; page-break-inside: avoid; }
          .sig-box { width: 30%; }
          .sig-space { height: 60px; }
          @media print {
            @page { size: landscape; margin: 10mm; }
            button { display: none; }
          }
        </style>
      </head>
      <body>
        <div style="text-align: right; margin-bottom: 12px;">
          <button onclick="window.print()" style="padding: 8px 18px; background: #0284c7; color: white; border: none; border-radius: 4px; cursor: pointer; font-weight: bold;">🖨️ Cetak / Print PDF</button>
        </div>

        <div class="header">
          <h2>KOPERASI KARYAWAN ADIRA (KOPKARA)</h2>
          <h2>LAPORAN REKONSILIASI PEMOTONGAN GAJI PAYROLL</h2>
          <p>Perbandingan Hasil Import HRD-Adira vs Database LMS Kopkara | Tanggal Cetak: ${todayStr}</p>
        </div>

        <table class="summary-table">
          <tr style="background: #f8fafc; font-weight: bold;">
            <td>Total Record Transaksi</td>
            <td>Total Tagihan Payroll (LMS)</td>
            <td>Total Terpotong (HRD Adira)</td>
            <td>Total Selisih / Tunggakan</td>
          </tr>
          <tr>
            <td><strong>${(displayList || []).length} Transaksi</strong></td>
            <td><strong>Rp ${Math.round(totalAmount || 0).toLocaleString('id-ID')}</strong></td>
            <td style="color: #047857; font-weight: bold;">Rp ${Math.round(totalDeducted || 0).toLocaleString('id-ID')}</td>
            <td style="color: ${(totalShortage || 0) > 0 ? '#b91c1c' : '#047857'}; font-weight: bold;">Rp ${Math.round(totalShortage || 0).toLocaleString('id-ID')}</td>
          </tr>
        </table>

        <table class="main-table">
          <thead>
            <tr>
              <th style="text-align: center;">No</th>
              <th>No. Ref / Loan No</th>
              <th>NIK Adira</th>
              <th>Anggota / Karyawan</th>
              <th style="text-align: center;">Periode</th>
              <th style="text-align: right;">Tagihan LMS</th>
              <th style="text-align: right;">Potongan HRD</th>
              <th style="text-align: right;">Selisih (Varian)</th>
              <th>Status Rekonsiliasi</th>
              <th>Keterangan HRD / LMS</th>
            </tr>
          </thead>
          <tbody>
            ${rowsHtml}
          </tbody>
        </table>

        <div style="margin-top: 24px; margin-bottom: 20px; page-break-inside: avoid;">
          <h3 style="margin-top: 0; font-size: 0.95rem; color: #1e293b; border-bottom: 2px solid #6366f1; padding-bottom: 4px;">
            📋 RIWAYAT ADJUSTMENT SELISIH PAYROLL (AUDIT TRAIL & HASIL KOREKSI UNTUK DITANDATANGANI)
          </h3>
          <table class="main-table">
            <thead>
              <tr style="background: #475569;">
                <th style="text-align: center; width: 40px;">No</th>
                <th>No. Referensi</th>
                <th>Anggota / Karyawan</th>
                <th>Tipe Adjustment</th>
                <th style="text-align: right;">Nominal Adjusted</th>
                <th>Keterangan User / Catatan Audit</th>
              </tr>
            </thead>
            <tbody>
              ${adjRowsHtml}
            </tbody>
          </table>
        </div>

        <div class="signatures">
          <div class="sig-box">
            <p>Dibuat Oleh (HRD Adira)</p>
            <div class="sig-space"></div>
            <p>_______________________</p>
            <p>Staff Payroll HRD</p>
          </div>
          <div class="sig-box">
            <p>Diperiksa Oleh (LMS Kopkara)</p>
            <div class="sig-space"></div>
            <p>_______________________</p>
            <p>Analis Keuangan / Admin</p>
          </div>
          <div class="sig-box">
            <p>Disetujui Oleh</p>
            <div class="sig-space"></div>
            <p>_______________________</p>
            <p>Pengurus Kopkara</p>
          </div>
        </div>
      </body>
      </html>
    `);
    reportWin.document.close();
  };

  const renderPayrollContent = () => {
    const getParamVal = (key, fallback) => {
      const p = (parameters || []).find(item => item.key_name === key || item.KeyName === key);
      return (p?.key_value || p?.KeyValue) || fallback;
    };

    const exportFolder = getParamVal('FOLDER_BILL_EXPORT', 'D:/Data_NK/Project5/LMS/export_payroll');
    const importFolder = getParamVal('FOLDER_BILL_IMPORT', 'D:/Data_NK/Project5/LMS/import_payroll');

    const realSchedulesList = (payrollSchedules || []).map(ps => {
      const empName = ps.employee_name || referenceData.employees.find(e => String(e.employee_id) === String(ps.member_no))?.name || `Member #${ps.member_no || ''}`;
      const safePeriod = ps.period || '2026-08';
      const safeRef = ps.ref_no || `LMS-PAY-${safePeriod.replace(/-/g, '')}-${ps.member_no || ''}`;
      const paidAmt = ps.amount_paid !== undefined ? ps.amount_paid : (ps.status === 'PAID' ? (ps.total_installment || 0) : 0);
      const remAmt = ps.status === 'ADJUSTED' ? 0 : (ps.remaining_installment !== undefined ? ps.remaining_installment : (paidAmt - (ps.total_installment || 0)));
      return {
        id: ps.id,
        loanNo: ps.loan_no,
        memberNo: ps.member_no,
        refNo: safeRef,
        nik: ps.nik || `317101234567${String(ps.member_no || '1001').padStart(4, '0')}`,
        name: `${empName} (#${ps.member_no || '1001'})`,
        period: safePeriod,
        amount: ps.total_installment || 0,
        deducted: ps.status === 'ADJUSTED' ? (ps.total_installment || 0) : paidAmt,
        remaining: remAmt,
        status: ps.status === 'ADJUSTED' ? 'ADJUSTED' : (ps.status === 'PAID' ? 'SUCCESS' : (ps.status === 'PARTIAL' ? 'PARTIAL' : (ps.status === 'UNPAID' ? 'PENDING' : (ps.status || 'PENDING')))),
        keterangan: ps.status === 'ADJUSTED' ? '⚙️ Adjustment Koreksi (SELESAI / Selisih Rp 0)' : (ps.status === 'PAID' ? 'Potongan gaji diproses penuh (LUNAS)' : (ps.status === 'PARTIAL' ? `Potongan parsial (Sisa/Credit: Rp ${Math.round(remAmt).toLocaleString('id-ID')})` : 'Menunggu eksekusi payroll HRD Adira'))
      };
    });

    const rawList = importedRows.length > 0 ? importedRows : realSchedulesList;
    const displayList = rawList.map(item => {
      const hasAdj = (adjustmentsList || []).some(a => a.ref_no === item.refNo || (a.loan_no > 0 && String(a.loan_no) === String(item.loanNo)));
      if (hasAdj || item.status === 'ADJUSTED') {
        return {
          ...item,
          deducted: item.amount,
          remaining: 0,
          status: 'SUCCESS',
          keterangan: '⚙️ Adjustment Koreksi (SELESAI / Selisih Rp 0)'
        };
      }
      return item;
    });

    // Create Final Table Rows including explicit Adjustment Reference rows
    const finalTableRows = [...displayList];
    (adjustmentsList || []).forEach(adj => {
      const targetRef = adj.ref_no || `LMS-PAY-202608-${adj.member_no}`;
      finalTableRows.push({
        isAdjustmentRef: true,
        id: `adj-ref-${adj.id}`,
        refNo: `${targetRef}-ADJ`,
        parentRefNo: targetRef,
        nik: adj.member_no,
        name: `${adj.employee_name || 'Karyawan'} (#${adj.member_no})`,
        period: adj.period || '2026-08',
        amount: adj.original_amount || 0,
        deducted: adj.adjusted_amount || adj.original_amount || 0,
        remaining: 0,
        status: 'ADJUSTED',
        adjustmentType: adj.adjustment_type,
        notes: adj.notes,
        keterangan: `⚙️ Record Adjustment: ${adj.adjustment_type} ("${adj.notes}")`
      });
    });

    const totalAmount = displayList.reduce((sum, item) => sum + (item?.amount || 0), 0);
    const totalDeducted = displayList.reduce((sum, item) => sum + (item?.deducted || 0), 0);
    const totalShortage = totalAmount - totalDeducted;
    const successCount = displayList.filter(i => i?.status === 'SUCCESS').length;
    const partialCount = displayList.filter(i => i?.status === 'PARTIAL').length;
    const failedCount = displayList.filter(i => i?.status === 'FAILED').length;
    const pendingCount = displayList.filter(i => i?.status === 'PENDING' || i?.status === 'UNPAID').length;

    return (
      <div>
        {/* Folder Config Banner */}
        <div style={{ padding: '10px 16px', background: '#e0f2fe', color: '#0369a1', borderRadius: '6px', marginBottom: '16px', fontSize: '0.85rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontWeight: 500 }}>
          <div>📁 <strong>Folder Export Default (FOLDER_BILL_EXPORT):</strong> <span style={{ fontFamily: 'monospace' }}>{exportFolder}</span></div>
          <div>📂 <strong>Folder Import Default (FOLDER_BILL_IMPORT):</strong> <span style={{ fontFamily: 'monospace' }}>{importFolder}</span></div>
        </div>

        {/* KPI Cards Row */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '20px' }}>
          <div className="card" style={{ padding: '16px', background: '#f8fafc', borderLeft: '4px solid #3b82f6', marginBottom: 0 }}>
            <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Total Tagihan Payroll</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 'bold', color: '#0f172a', marginTop: '4px' }}>Rp {Math.round(totalAmount).toLocaleString('id-ID')}</div>
            <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '4px' }}>{displayList.length} Transaksi Database (lms_sch.loan_schedules)</div>
          </div>

          <div className="card" style={{ padding: '16px', background: '#f0fdf4', borderLeft: '4px solid #10b981', marginBottom: 0 }}>
            <div style={{ fontSize: '0.8rem', color: '#047857', fontWeight: 600, textTransform: 'uppercase' }}>Terpotong Sukses</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 'bold', color: '#065f46', marginTop: '4px' }}>Rp {Math.round(totalDeducted).toLocaleString('id-ID')}</div>
            <div style={{ fontSize: '0.75rem', color: '#047857', marginTop: '4px' }}>{successCount} Sukses | {partialCount} Partial</div>
          </div>

          <div className="card" style={{ padding: '16px', background: totalShortage > 0 ? '#fef2f2' : '#f8fafc', borderLeft: `4px solid ${totalShortage > 0 ? '#ef4444' : '#64748b'}`, marginBottom: 0 }}>
            <div style={{ fontSize: '0.8rem', color: totalShortage > 0 ? '#991b1b' : '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Selisih / Tunggakan</div>
            <div style={{ fontSize: '1.4rem', fontWeight: 'bold', color: totalShortage > 0 ? '#991b1b' : '#0f172a', marginTop: '4px' }}>Rp {Math.round(totalShortage).toLocaleString('id-ID')}</div>
            <div style={{ fontSize: '0.75rem', color: totalShortage > 0 ? '#991b1b' : '#64748b', marginTop: '4px' }}>{failedCount} Gagal Potong</div>
          </div>

          <div className="card" style={{ padding: '16px', background: reconcileClosingInfo.status === 'CLOSED' ? '#ecfdf5' : '#fefce8', borderLeft: `4px solid ${reconcileClosingInfo.status === 'CLOSED' ? '#10b981' : '#eab308'}`, marginBottom: 0 }}>
            <div style={{ fontSize: '0.8rem', color: reconcileClosingInfo.status === 'CLOSED' ? '#047857' : '#854d0e', fontWeight: 600, textTransform: 'uppercase' }}>Status Rekonsiliasi</div>
            <div style={{ fontSize: '1.2rem', fontWeight: 'bold', color: reconcileClosingInfo.status === 'CLOSED' ? '#047857' : ((failedCount > 0 || totalShortage > 0) ? '#b91c1c' : '#047857'), marginTop: '4px' }}>
              {reconcileClosingInfo.status === 'CLOSED' ? '🔒 REKONSILIASI CLOSED' : ((payrollReconciled || successCount > 0 || partialCount > 0 || failedCount > 0) ? ((totalShortage > 0 || failedCount > 0) ? '⚠️ UNRECONCILED (ADA SELISIH)' : '✅ TEREKONSILIASI') : '⏳ PENDING')}
            </div>
            <div style={{ fontSize: '0.75rem', color: reconcileClosingInfo.status === 'CLOSED' ? '#047857' : '#854d0e', marginTop: '4px' }}>
              {reconcileClosingInfo.status === 'CLOSED' ? '🔒 Telah Ditandatangani & Disetujui Resmi' : ((payrollReconciled || successCount > 0 || partialCount > 0 || failedCount > 0) ? `${successCount + partialCount} Selesai / ${failedCount} Gagal Potong` : `${pendingCount} Menunggu Import HRD`)}
            </div>
          </div>
        </div>

        {/* Table Container */}
        <div className="table-container">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 20px', background: 'var(--primary-blue)', color: 'white', borderTopLeftRadius: '8px', borderTopRightRadius: '8px' }}>
            <div style={{ fontWeight: 'bold', fontSize: '1.1rem' }}>✂️ Detail Rekonsiliasi Pemotongan Gaji (lms_sch.loan_schedules)</div>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button 
                onClick={() => setCloseModalOpen(true)}
                disabled={reconcileClosingInfo.status === 'CLOSED'}
                style={{ padding: '6px 12px', background: reconcileClosingInfo.status === 'CLOSED' ? '#64748b' : '#059669', color: 'white', border: 'none', borderRadius: '4px', cursor: reconcileClosingInfo.status === 'CLOSED' ? 'not-allowed' : 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                title="Tanda Tangan & Tutup Status Rekonsiliasi Periode Ini secara Resmi"
              >
                {reconcileClosingInfo.status === 'CLOSED' ? '🔒 Rekon Closed' : '🔒 Tanda Tangan & Tutup Rekon (CLOSE)'}
              </button>
              <button 
                type="button"
                onClick={async () => {
                  if (window.confirm('Apakah Anda yakin ingin RESET & BUKA KEMBALI status rekonsiliasi agar bisa meng-import ulang file CSV HRD-Adira?')) {
                    try {
                      await axios.post(`${API_BASE_URL}/api/payroll/reset-reconciliation?period=2026-08`);
                      setImportedRows([]);
                      setReconcileClosingInfo({ status: 'OPEN' });
                      fetchPayrollSchedules();
                      alert('Status Rekonsiliasi & Tagihan 2026-08 BERHASIL DI-RESET! Tombol Import Result Rekonsiliasi kini aktif kembali.');
                    } catch (err) {
                      alert('Gagal me-reset status: ' + err.message);
                    }
                  }
                }}
                style={{ padding: '6px 12px', background: '#dc2626', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                title="Reset Total Data Rekonsiliasi Periode Ini & Buka Kembali Akses Import File CSV HRD"
              >
                🔄 Reset & Buka Rekon (RE-IMPORT)
              </button>
              <button 
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  fetchAdjustments();
                  setHistoryModalOpen(true);
                }}
                style={{ padding: '6px 12px', background: '#6366f1', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.85rem', zIndex: 10, position: 'relative' }}
                title="Lihat Daftar & Riwayat Audit Adjustment Periode Ini"
              >
                📋 Lihat Riwayat Adjustment ({adjustmentsList.length})
              </button>
              <button 
                onClick={() => handlePrintReconciliationReport(displayList, totalAmount, totalDeducted, totalShortage)}
                style={{ padding: '6px 12px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                title="Cetak / Download Laporan Resmi Rekonsiliasi Perbandingan HRD vs LMS"
              >
                🖨️ Cetak Laporan Rekonsiliasi
              </button>
              <button 
                onClick={() => {
                  const now = new Date();
                  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0);
                  const yyyy = lastDay.getFullYear();
                  const mm = String(lastDay.getMonth() + 1).padStart(2, '0');
                  const dd = String(lastDay.getDate()).padStart(2, '0');
                  setExportCutoffDate(`${yyyy}-${mm}-${dd}`);
                  setExportCustomFolder(exportFolder);
                  setExportModalOpen(true);
                }}
                style={{ padding: '6px 12px', background: '#10b981', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                title={`Export File CSV Tagihan s/d Tanggal Cut-off ke Folder ${exportFolder}`}
              >
                📥 Export File Payroll (.csv)
              </button>
              <input 
                type="file" 
                ref={fileInputRef} 
                accept=".csv,.txt" 
                onChange={handleCSVUpload} 
                style={{ display: 'none' }} 
              />
              <button 
                onClick={() => fileInputRef.current && fileInputRef.current.click()}
                disabled={reconcileClosingInfo.status === 'CLOSED'}
                style={{ padding: '6px 12px', background: reconcileClosingInfo.status === 'CLOSED' ? '#64748b' : '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: reconcileClosingInfo.status === 'CLOSED' ? 'not-allowed' : 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                title={`Buka Folder Import ${importFolder} & Pilih File CSV Rekonsiliasi`}
              >
                📤 Import Result Rekonsiliasi
              </button>
            </div>
          </div>
          <table style={{ tableLayout: 'auto', width: '100%' }}>
            <thead>
              <tr>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>No. Ref</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>NIK Adira</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', maxWidth: '180px' }}>Anggota / Karyawan</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Periode</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Nominal Tagihan</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Terpotong HRD</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Selisih (Varian)</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Status</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Hasil Analisis / Keterangan HRD</th>
                <th style={{ padding: '8px 10px', fontSize: '0.85rem', textAlign: 'center' }}>Aksi Adjustment</th>
              </tr>
            </thead>
            <tbody>
              {finalTableRows.length === 0 ? (
                <tr><td colSpan="10" style={{ textAlign: 'center', padding: '24px', color: '#94a3b8' }}>Belum ada tagihan angsuran pinjaman aktif di lms_sch.loan_schedules.</td></tr>
              ) : (
                finalTableRows.map((item, idx) => {
                  if (item.isAdjustmentRef) {
                    return (
                      <tr key={`adj-ref-row-${item.id}-${idx}`} style={{ background: '#f5f3ff', borderBottom: '1px solid #c7d2fe' }}>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>
                          <strong style={{ color: '#4f46e5' }}>↳ {item.refNo}</strong>
                        </td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>{item.nik}</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{item.name}</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>{item.period}</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>Rp {Math.round(item.amount).toLocaleString('id-ID')}</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap', color: '#059669', fontWeight: 600 }}>Rp {Math.round(item.deducted).toLocaleString('id-ID')}</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap', fontWeight: 600, color: '#059669' }}>Rp 0</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>
                          <span className="badge badge-success" style={{ background: '#6366f1' }}>⚙️ ADJUSTED</span>
                        </td>
                        <td style={{ padding: '8px 10px', fontSize: '0.85rem', color: '#4338ca', fontStyle: 'italic' }}>
                          🔗 Ref ke <strong>{item.parentRefNo}</strong>: "{item.notes}"
                        </td>
                        <td style={{ padding: '8px 10px', textAlign: 'center', fontSize: '0.75rem', color: '#6366f1', fontWeight: 600 }}>
                          ⚙️ Record Reference
                        </td>
                      </tr>
                    );
                  }
                  const tagihan = Math.round(item?.amount || 0);
                  const isFailed = item?.status === 'FAILED';
                  const terpotong = isFailed ? 0 : Math.round(item?.deducted || 0);
                  const diff = tagihan - terpotong;
                  const displayRefNo = String(item?.loanNo || item?.loan_no || item?.refNo || '-');
                  return (
                    <tr key={item?.id ? `sched-${item.id}` : `${item?.refNo || 'pay'}-${idx}`} style={{ borderBottom: '1px solid #e2e8f0' }}>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}><strong>{displayRefNo}</strong></td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>{item?.nik || '-'}</td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={item?.name || ''}>
                        {item?.name || '-'}
                      </td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>{item?.period || '-'}</td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}><strong>Rp {tagihan.toLocaleString('id-ID')}</strong></td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap', color: isFailed ? '#dc2626' : (item?.status === 'SUCCESS' ? '#047857' : '#b45309'), fontWeight: 600 }}>
                        Rp {terpotong.toLocaleString('id-ID')}
                      </td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap', fontWeight: 600, color: diff > 0 ? '#b91c1c' : (diff < 0 ? '#0284c7' : '#047857') }}>
                        {diff === 0 ? 'Rp 0' : (diff > 0 ? `+Rp ${diff.toLocaleString('id-ID')}` : `-Rp ${Math.abs(diff).toLocaleString('id-ID')}`)}
                      </td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', whiteSpace: 'nowrap' }}>
                        <span className={item?.status === 'SUCCESS' ? 'badge badge-success' : (isFailed ? 'badge badge-danger' : 'badge badge-warning')}>
                          {item?.status === 'UNPAID' ? 'PENDING' : (item?.status || 'PENDING')}
                        </span>
                      </td>
                      <td style={{ padding: '8px 10px', fontSize: '0.85rem', color: isFailed ? '#991b1b' : '#475569' }}>
                        <i>{item?.keterangan || (isFailed ? 'Gagal Potong: Pinjaman tidak ditemukan' : 'Potongan diproses')}</i>
                      </td>
                      <td style={{ padding: '8px 10px', textAlign: 'center', whiteSpace: 'nowrap' }}>
                        {(() => {
                          const itemLoanNo = String(item?.loanNo || item?.loan_no || '');
                          const isItemAdjusted = adjustmentsList.some(a => 
                            (itemLoanNo !== '' && itemLoanNo !== '0' && (String(a.loan_no) === itemLoanNo || String(a.ref_no) === itemLoanNo)) ||
                            (a.ref_no && String(a.ref_no) === displayRefNo)
                          );
                          if (isItemAdjusted) {
                            return (
                              <button disabled style={{ padding: '4px 10px', background: '#94a3b8', color: 'white', border: 'none', borderRadius: '4px', cursor: 'not-allowed', fontSize: '0.8rem', fontWeight: 600 }} title="Record ini sudah diadjust">
                                ⚙️ Adjusted
                              </button>
                            );
                          }
                          if ((isFailed || diff < 0) && reconcileClosingInfo.status !== 'CLOSED') {
                            return (
                              <button
                                onClick={() => {
                                  setAdjustTargetItem(item);
                                  if (isFailed) {
                                    setAdjustType('FAILED_CORRECTION');
                                    setAdjustNotes('data dikembalikan ke HRD-Adira');
                                  } else if (diff < 0) {
                                    setAdjustType('OVERPAYMENT_REFUND');
                                    setAdjustNotes('data dikembalikan ke HRD-Adira');
                                  } else {
                                    setAdjustType('FAILED_CORRECTION');
                                    setAdjustNotes('data dikembalikan ke HRD-Adira');
                                  }
                                  setAdjustModalOpen(true);
                                }}
                                style={{ padding: '4px 10px', background: '#6366f1', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem', fontWeight: 600 }}
                                title="Proses Adjustment Selisih Rekonsiliasi"
                              >
                                ⚙️ Adjust
                              </button>
                            );
                          }
                          if (reconcileClosingInfo.status === 'CLOSED') {
                            return <span style={{ fontSize: '0.75rem', color: '#64748b' }}>🔒 Locked</span>;
                          }
                          return null;
                        })()}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Modal Adjustment Selisih */}
        {adjustModalOpen && adjustTargetItem && (
          <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
            <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '90%', maxWidth: '520px', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                <h3 style={{ margin: 0, fontSize: '1.15rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  ⚙️ Form Adjustment Selisih Rekonsiliasi
                </h3>
                <button onClick={() => setAdjustModalOpen(false)} style={{ background: 'none', border: 'none', fontSize: '1.25rem', cursor: 'pointer', color: '#64748b' }}>&times;</button>
              </div>

              <form onSubmit={handleSaveAdjustment}>
                <div style={{ padding: '12px', background: '#f8fafc', borderRadius: '6px', marginBottom: '16px', fontSize: '0.85rem', borderLeft: '4px solid #6366f1' }}>
                  <div><strong>No. Referensi:</strong> {adjustTargetItem.refNo}</div>
                  <div><strong>Karyawan / Anggota:</strong> {adjustTargetItem.name}</div>
                  <div><strong>Nominal Tagihan LMS:</strong> Rp {Math.round(adjustTargetItem.amount || 0).toLocaleString('id-ID')}</div>
                  <div><strong>Terpotong HRD:</strong> Rp {Math.round(adjustTargetItem.deducted || 0).toLocaleString('id-ID')}</div>
                  <div><strong>Status Rekon:</strong> <span className="badge badge-danger">{adjustTargetItem.status}</span></div>
                </div>

                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.85rem' }}>Tipe Adjustment:</label>
                  <select 
                    value={adjustType} 
                    onChange={e => setAdjustType(e.target.value)}
                    style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', fontSize: '0.85rem' }}
                  >
                    <option value="FAILED_CORRECTION">Koreksi Gagal (Membuat Record Adjustment & Kirim Balik Ke HRD)</option>
                    <option value="OVERPAYMENT_REFUND">Koreksi Overpayment (Set Remaining = 0 & Buat Record Adjustment)</option>
                    <option value="OVERPAYMENT_OFFSET">Koreksi Overpayment (Offset Angsuran Pokok Berikutnya)</option>
                  </select>
                </div>

                <div style={{ marginBottom: '20px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.85rem' }}>Keterangan Adjustment (Keterangan Dibuat User):</label>
                  <textarea 
                    required
                    value={adjustNotes}
                    onChange={e => setAdjustNotes(e.target.value)}
                    placeholder="misal: data dikembalikan ke HRD-Adira"
                    style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', minHeight: '80px', fontSize: '0.85rem' }}
                  />
                  <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '4px' }}>
                    * Record adjustment baru akan mereferensikan transaksi ini ({adjustTargetItem.refNo}) untuk jejak audit (tracking).
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
                  <button type="button" onClick={() => setAdjustModalOpen(false)} style={{ padding: '8px 16px', background: '#e2e8f0', color: '#475569', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>
                    Batal
                  </button>
                  <button type="submit" style={{ padding: '8px 16px', background: '#6366f1', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>
                    💾 Simpan Adjustment
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Modal Closing & Tanda Tangan Rekonsiliasi */}
        {closeModalOpen && (
          <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
            <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '90%', maxWidth: '540px', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                <h3 style={{ margin: 0, fontSize: '1.15rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  🔒 Tanda Tangan & Tutup Rekonsiliasi (CLOSE)
                </h3>
                <button onClick={() => setCloseModalOpen(false)} style={{ background: 'none', border: 'none', fontSize: '1.25rem', cursor: 'pointer', color: '#64748b' }}>&times;</button>
              </div>

              <form onSubmit={handleSaveCloseReconciliation}>
                <div style={{ padding: '12px', background: '#f0fdf4', borderRadius: '6px', marginBottom: '16px', fontSize: '0.85rem', borderLeft: '4px solid #10b981', color: '#047857' }}>
                  <strong>Catatan Persetujuan:</strong> Menutup periode rekonsiliasi ini akan mengunci (lock) status rekonsiliasi secara resmi dan merekam nama para penandatangan sah.
                </div>

                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', marginBottom: '4px', fontWeight: 600, fontSize: '0.85rem' }}>Penandatangan HRD Adira:</label>
                  <input 
                    type="text" required
                    value={closeForm.hrd_signatory}
                    onChange={e => setCloseForm({...closeForm, hrd_signatory: e.target.value})}
                    style={{ width: '100%', padding: '8px 10px', borderRadius: '4px', border: '1px solid var(--border-color)', fontSize: '0.85rem' }}
                  />
                </div>

                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', marginBottom: '4px', fontWeight: 600, fontSize: '0.85rem' }}>Penandatangan Finance LMS:</label>
                  <input 
                    type="text" required
                    value={closeForm.finance_signatory}
                    onChange={e => setCloseForm({...closeForm, finance_signatory: e.target.value})}
                    style={{ width: '100%', padding: '8px 10px', borderRadius: '4px', border: '1px solid var(--border-color)', fontSize: '0.85rem' }}
                  />
                </div>

                <div style={{ marginBottom: '12px' }}>
                  <label style={{ display: 'block', marginBottom: '4px', fontWeight: 600, fontSize: '0.85rem' }}>Penandatangan Pengurus Kopkara:</label>
                  <input 
                    type="text" required
                    value={closeForm.kopkara_signatory}
                    onChange={e => setCloseForm({...closeForm, kopkara_signatory: e.target.value})}
                    style={{ width: '100%', padding: '8px 10px', borderRadius: '4px', border: '1px solid var(--border-color)', fontSize: '0.85rem' }}
                  />
                </div>

                <div style={{ marginBottom: '20px' }}>
                  <label style={{ display: 'block', marginBottom: '4px', fontWeight: 600, fontSize: '0.85rem' }}>Catatan Persetujuan Rekonsiliasi:</label>
                  <textarea 
                    value={closeForm.closing_notes}
                    onChange={e => setCloseForm({...closeForm, closing_notes: e.target.value})}
                    style={{ width: '100%', padding: '8px 10px', borderRadius: '4px', border: '1px solid var(--border-color)', minHeight: '60px', fontSize: '0.85rem' }}
                  />
                </div>

                <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
                  <button type="button" onClick={() => setCloseModalOpen(false)} style={{ padding: '8px 16px', background: '#e2e8f0', color: '#475569', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>
                    Batal
                  </button>
                  <button type="submit" style={{ padding: '8px 16px', background: '#059669', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>
                    🔒 Tanda Tangan & CLOSE Rekon
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Modal Riwayat Adjustment Payroll */}
        {historyModalOpen && (
          <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
            <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '90%', maxWidth: '850px', maxHeight: '85vh', overflowY: 'auto', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                <h3 style={{ margin: 0, fontSize: '1.15rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  📋 Riwayat Audit Adjustment Selisih Payroll (lms_sch.payroll_adjustments)
                </h3>
                <button onClick={() => setHistoryModalOpen(false)} style={{ background: 'none', border: 'none', fontSize: '1.25rem', cursor: 'pointer', color: '#64748b' }}>&times;</button>
              </div>

              <div style={{ padding: '10px 14px', background: '#eff6ff', borderRadius: '6px', marginBottom: '16px', fontSize: '0.85rem', color: '#1e40af' }}>
                ℹ️ Seluruh catatan koreksi & adjustment yang dibuat user akan tersimpan secara otomatis sebagai audit log resmi untuk proses rekonsiliasi.
              </div>

              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
                <thead>
                  <tr style={{ background: '#f1f5f9', borderBottom: '2px solid #cbd5e1' }}>
                    <th style={{ padding: '8px 10px', textAlign: 'left' }}>No. Ref</th>
                    <th style={{ padding: '8px 10px', textAlign: 'left' }}>Karyawan / Anggota</th>
                    <th style={{ padding: '8px 10px', textAlign: 'left' }}>Tipe Adjustment</th>
                    <th style={{ padding: '8px 10px', textAlign: 'right' }}>Nominal Tagihan</th>
                    <th style={{ padding: '8px 10px', textAlign: 'right' }}>Nominal Adjusted</th>
                    <th style={{ padding: '8px 10px', textAlign: 'left' }}>Keterangan User</th>
                    <th style={{ padding: '8px 10px', textAlign: 'left' }}>User & Waktu</th>
                  </tr>
                </thead>
                <tbody>
                  {adjustmentsList.length === 0 ? (
                    <tr>
                      <td colSpan="7" style={{ textAlign: 'center', padding: '24px', color: '#94a3b8' }}>Belum ada data adjustment pada periode ini.</td>
                    </tr>
                  ) : (
                    adjustmentsList.map(adj => (
                      <tr key={adj.id || Math.random()} style={{ borderBottom: '1px solid #e2e8f0' }}>
                        <td style={{ padding: '8px 10px', whiteSpace: 'nowrap' }}><strong>{adj.ref_no || '-'}</strong></td>
                        <td style={{ padding: '8px 10px' }}>{adj.employee_name || '-'} <span style={{ color: '#64748b', fontSize: '0.75rem' }}>(#{adj.member_no || ''})</span></td>
                        <td style={{ padding: '8px 10px', whiteSpace: 'nowrap' }}>
                          <span style={{ padding: '2px 6px', background: '#e0e7ff', color: '#3730a3', borderRadius: '4px', fontWeight: 600, fontSize: '0.75rem' }}>
                            {adj.adjustment_type}
                          </span>
                        </td>
                        <td style={{ padding: '8px 10px', textAlign: 'right', whiteSpace: 'nowrap' }}>Rp {Math.round(adj.original_amount || 0).toLocaleString('id-ID')}</td>
                        <td style={{ padding: '8px 10px', textAlign: 'right', whiteSpace: 'nowrap', fontWeight: 'bold', color: '#059669' }}>Rp {Math.round(adj.adjusted_amount || 0).toLocaleString('id-ID')}</td>
                        <td style={{ padding: '8px 10px', color: '#334155', fontStyle: 'italic' }}>"{adj.notes || '-'}"</td>
                        <td style={{ padding: '8px 10px', fontSize: '0.75rem', color: '#64748b', whiteSpace: 'nowrap' }}>
                          <div>👤 {adj.created_user || 'Admin'}</div>
                          <div>📅 {adj.created_at ? new Date(adj.created_at).toLocaleString('id-ID') : '-'}</div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>

              <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '20px', paddingTop: '12px', borderTop: '1px solid #e2e8f0' }}>
                <button onClick={() => setHistoryModalOpen(false)} style={{ padding: '8px 20px', background: '#64748b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Tutup</button>
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  const handlePrintContract = (app) => {
    const appNo = app.application_no || app.ApplicationNo;
    const memberNo = app.member_no || app.MemberNo;
    const approvedAmount = app.approved_amount ?? app.ApprovedAmount ?? app.requested_amount ?? app.RequestedAmount ?? 0;
    const tenor = app.tenor ?? app.Tenor ?? 0;
    const notes = app.approval_notes || app.ApprovalNotes || 'Disetujui';
    const dateStr = formatDate(app.approved_at || app.ApprovedAt || app.submission_date || app.SubmissionDate);
    const amtStr = Number(approvedAmount).toLocaleString('id-ID');

    const contractWin = window.open('', '_blank');
    if (!contractWin) {
      alert("Pop-up diblokir oleh browser. Silakan izinkan pop-up untuk mencetak kontrak.");
      return;
    }

    contractWin.document.write(`
      <!DOCTYPE html>
      <html>
        <head>
          <title>Surat Perjanjian Kredit & Kontrak Pinjaman #${appNo}</title>
          <style>
            body { font-family: 'Segoe UI', Arial, sans-serif; padding: 40px; line-height: 1.6; color: #1e293b; max-width: 800px; margin: 0 auto; }
            .header { text-align: center; border-bottom: 3px double #0f172a; padding-bottom: 16px; margin-bottom: 24px; }
            .header h2 { margin: 0; font-size: 1.4rem; color: #0f172a; }
            .header h3 { margin: 6px 0; font-size: 1.1rem; color: #2563eb; }
            .header p { margin: 4px 0 0 0; font-size: 0.9rem; color: #64748b; }
            .section { margin-bottom: 20px; }
            .info-table { width: 100%; border-collapse: collapse; margin: 16px 0; }
            .info-table td { padding: 8px 12px; border: 1px solid #cbd5e1; font-size: 0.95rem; }
            .info-table td.label { background: #f8fafc; font-weight: 600; width: 35%; color: #334155; }
            .signature-table { width: 100%; margin-top: 60px; text-align: center; font-size: 0.9rem; }
            .signature-space { height: 80px; }
          </style>
        </head>
        <body>
          <div class="header">
            <h2>KOPERASI KARYAWAN KOPKARA (LMS)</h2>
            <h3>SURAT PERJANJIAN KREDIT & KONTRAK PINJAMAN</h3>
            <p>No. Kontrak: <strong>CTR-${appNo}</strong></p>
          </div>
          <div class="section">
            <p>Pada hari ini, disepakati Perjanjian Pinjaman antara <strong>KOPERASI KARYAWAN KOPKARA</strong> (Pemberi Pinjaman) dan <strong>Anggota #${memberNo}</strong> (Penerima Pinjaman) dengan rincian sebagai berikut:</p>
            <table class="info-table">
              <tr><td class="label">No. Pengajuan Pinjaman</td><td><strong>#${appNo}</strong></td></tr>
              <tr><td class="label">Plafon Pinjaman Disetujui</td><td><strong style="color:#15803d; font-size:1.1rem;">Rp ${amtStr}</strong></td></tr>
              <tr><td class="label">Tenor Jangka Waktu</td><td><strong>${tenor} Bulan</strong></td></tr>
              <tr><td class="label">Status Persetujuan HRD</td><td><strong>APPROVED</strong> (${notes})</td></tr>
              <tr><td class="label">Tanggal Cetak Kontrak</td><td>${dateStr}</td></tr>
            </table>
          </div>
          <div class="section">
            <p><strong>Ketentuan Pelunasan:</strong> Penerima Pinjaman menyatakan setuju & memberikan kuasa penuh untuk melunasi angsuran bulanan melalui mekanisme pemotongan gaji (*payroll deduction*) setiap tanggal 25 hingga lunas.</p>
          </div>
          <table class="signature-table">
            <tr>
              <td width="50%">
                Penerima Pinjaman (Karyawan),
                <div class="signature-space"></div>
                ( ________________________ )<br/>
                Member #${memberNo}
              </td>
              <td width="50%">
                Pemberi Pinjaman (Pengurus Koperasi),
                <div class="signature-space"></div>
                ( ________________________ )<br/>
                HRD / Treasury Kopkara
              </td>
            </tr>
          </table>
          <script>
            window.onload = function() {
              window.print();
            };
          </script>
        </body>
      </html>
    `);
    contractWin.document.close();
  };

  const fetchParameters = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/parameters`);
      setParameters(response.data.data || []);
    } catch (error) {
      console.error("Error fetching parameters:", error);
    }
  };

  const toggleRoleMenu = async (roleId, menuId, isGranted) => {
    try {
      if (isGranted) {
        // Revoke access
        await axios.delete(`${API_BASE_URL}/api/master/role-menus/0?role_id=${roleId}&menu_id=${menuId}`);
        setReferenceData(prev => ({
          ...prev,
          role_menus: prev.role_menus.filter(rm => !(String(rm.role_id) === String(roleId) && String(rm.menu_id) === String(menuId)))
        }));
      } else {
        // Grant access
        await axios.post(`${API_BASE_URL}/api/master/role-menus`, {
          role_id: parseInt(roleId),
          menu_id: parseInt(menuId)
        });
        setReferenceData(prev => ({
          ...prev,
          role_menus: [...prev.role_menus, { role_id: parseInt(roleId), menu_id: parseInt(menuId) }]
        }));
      }
    } catch (error) {
      alert("Gagal memperbarui APM Matrix: " + (error.response?.data?.error || error.message));
    }
  };

  const [masterTotalRecords, setMasterTotalRecords] = useState(0);
  const [masterTotalPages, setMasterTotalPages] = useState(1);

  const fetchMasterData = async (table, q = '', page = 1) => {
    try {
      const targetTable = table || masterTab;
      const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || parseInt(getParamVal('PAGINATION_LIMIT', '10')) || 10;
      const response = await axios.get(`${API_BASE_URL}/api/master/${targetTable}?q=${encodeURIComponent(q || '')}&page=${page || 1}&limit=${pageSize}`);
      setMasterDataList(response.data.data || []);
      setMasterTotalRecords(response.data.total_records || (response.data.data ? response.data.data.length : 0));
      setMasterTotalPages(response.data.total_pages || 1);
    } catch (error) {
      console.error('Error fetching master data:', error);
    }
  };

  const fetchReferenceData = async () => {
    try {
      const [deptRes, empStatusRes, kopkaraStatusRes, empCatRes, empRes, memRes, roleRes, menuRes, roleMenuRes] = await Promise.all([
        axios.get(`${API_BASE_URL}/api/master/departments`),
        axios.get(`${API_BASE_URL}/api/master/employee-statuses`),
        axios.get(`${API_BASE_URL}/api/master/kopkara-statuses`),
        axios.get(`${API_BASE_URL}/api/master/employee-categories`),
        axios.get(`${API_BASE_URL}/api/master/employees`),
        axios.get(`${API_BASE_URL}/api/master/members`),
        axios.get(`${API_BASE_URL}/api/master/roles`),
        axios.get(`${API_BASE_URL}/api/master/menus`),
        axios.get(`${API_BASE_URL}/api/master/role-menus`)
      ]);
      setReferenceData({
        departments: deptRes.data.data || [],
        employeeStatuses: empStatusRes.data.data || [],
        kopkaraStatuses: kopkaraStatusRes.data.data || [],
        employeeCategories: empCatRes.data.data || [],
        employees: empRes.data.data || [],
        members: memRes.data.data || [],
        roles: roleRes.data.data || [],
        menus: menuRes.data.data || [],
        role_menus: roleMenuRes.data.data || []
      });
    } catch (error) {
      console.error('Error fetching reference data:', error);
    }
  };

  // Sync masterTab with activeTab for dynamic master menus
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
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'master' || activeTab.startsWith('master-')) {
      fetchMasterData(masterTab, masterSearchQuery, currentPage);
    }
    if (masterTab === 'members') {
      fetchPaginatedEmployeesForSelect('', 1);
    }
  }, [masterTab, currentPage]);

  const submitApplication = async (e) => {
    e.preventDefault();
    try {
      await axios.post(`${API_BASE_URL}/api/applications`, {
        member_no: parseInt(form.member_no),
        product_id: parseInt(form.product_id),
        requested_amount: parseFloat(form.requested_amount),
        tenor: parseInt(form.tenor)
      });
      alert('Pengajuan berhasil dikirim!');
      setForm({ ...form, product_id: '', requested_amount: '', tenor: '' });
      setSimulation(null);
      fetchApplications();
    } catch (error) {
      alert('Gagal mengirim pengajuan: ' + (error.response?.data?.error || error.message));
    }
  };

  // Validasi Tanggal
  useEffect(() => {
    const getParamVal = (key, fallback) => {
      const p = (parameters || []).find(item => item.key_name === key || item.KeyName === key);
      return (p?.key_value || p?.KeyValue) || fallback;
    };
    const startPeriod = parseInt(getParamVal('LOAN_START_PERIOD', '1'));
    const endPeriod = parseInt(getParamVal('LOAN_END_PERIOD', '31'));
    const currentDay = new Date().getDate();
    if (currentDay < startPeriod || currentDay > endPeriod) {
      setDateError(`Pengajuan pinjaman ditutup. Hanya dibuka tanggal ${startPeriod} - ${endPeriod}`);
    } else {
      setDateError('');
    }
  }, [parameters]);

  // Simulasi Dinamis
  useEffect(() => {
    if (form.product_id && form.requested_amount > 0 && form.tenor > 0) {
      const fetchSimulation = async () => {
        try {
          const response = await axios.post(`${API_BASE_URL}/api/applications/simulate`, {
            member_no: parseInt(form.member_no),
            product_id: parseInt(form.product_id),
            requested_amount: parseFloat(form.requested_amount),
            tenor: parseInt(form.tenor)
          });
          setSimulation(response.data.data);
        } catch (error) {
          setSimulation({ error: error.response?.data?.error || error.message });
        }
      };
      // Simple debounce
      const timeoutId = setTimeout(() => fetchSimulation(), 500);
      return () => clearTimeout(timeoutId);
    } else {
      setSimulation(null);
    }
  }, [form.product_id, form.requested_amount, form.tenor]);

  useEffect(() => {
    fetchProducts();
    fetchApplications();
    fetchParameters();
    fetchReferenceData();
  }, []);

  // Login Authentication Verification
  useEffect(() => {
    if (authToken) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
      const verifyAuthToken = async () => {
        try {
          // Try Karisma Simulator port 8087 first
          const res = await axios.post(`${KARISMA_SIMULATOR_URL}/api/karisma/verify`, {}, {
            headers: { Authorization: `Bearer ${authToken}` }
          });
          const user = res.data.user;
          setCurrentUser(user);
          localStorage.setItem('karisma_user', JSON.stringify(user));
          setForm(prev => ({...prev, member_no: user.employee_id}));
        } catch (err8087) {
          try {
            // Fallback to LMS Core Backend port 8086
            const res = await axios.post(`${API_BASE_URL}/api/karisma/verify`, {}, {
              headers: { Authorization: `Bearer ${authToken}` }
            });
            const user = res.data.user;
            setCurrentUser(user);
            localStorage.setItem('karisma_user', JSON.stringify(user));
            setForm(prev => ({...prev, member_no: user.employee_id}));
          } catch (err8086) {
            setAuthToken('');
            localStorage.removeItem('karisma_token');
            localStorage.removeItem('karisma_user');
            setCurrentUser(null);
            delete axios.defaults.headers.common['Authorization'];
          }
        }
      };
      verifyAuthToken();
    } else {
      delete axios.defaults.headers.common['Authorization'];
    }
  }, [authToken]);

  const handleLogin = async (e) => {
    e.preventDefault();
    setLoginError('');
    try {
      let res;
      try {
        // Try Karisma Simulator port 8087 first
        res = await axios.post(`${KARISMA_SIMULATOR_URL}/api/karisma/login`, loginForm, { withCredentials: true });
      } catch (err8087) {
        console.warn("Port 8087 unreachable, falling back to LMS Core Backend port 8086...", err8087);
        // Fallback to LMS Core Backend port 8086
        res = await axios.post(`${API_BASE_URL}/api/karisma/login`, loginForm, { withCredentials: true });
      }
      const token = res.data.token;
      localStorage.setItem('karisma_token', token);
      setAuthToken(token);
    } catch (err) {
      setLoginError(err.response?.data?.error || 'Username atau Password salah!');
    }
  };

  const handleLogout = async (reason = '') => {
    try {
      await axios.post(`${API_BASE_URL}/api/karisma/logout`, {}, { withCredentials: true });
    } catch (err) {
      console.warn("Error calling logout endpoint:", err);
    }
    localStorage.removeItem('karisma_token');
    localStorage.removeItem('karisma_user');
    setAuthToken('');
    setCurrentUser(null);
    setUserInfo(null);
    setActiveTab('dashboard');
    if (reason) {
      setLoginError(reason);
    }
    window.location.reload();
  };

  const saveMasterData = async (e) => {
    e.preventDefault();
    try {
      const payload = { ...masterForm };
      ['role_id', 'employee_id', 'member_no', 'menu_id', 'parent_id', 'salary', 'max_limit', 'order'].forEach(key => {
        if (payload[key] !== undefined && payload[key] !== '' && payload[key] !== null) {
          payload[key] = (key === 'salary' || key === 'max_limit') ? parseFloat(payload[key]) : parseInt(payload[key]);
        }
      });

      await axios.post(`${API_BASE_URL}/api/master/${masterTab}`, payload);
      alert('Data berhasil disimpan!');
      setMasterForm({});
      fetchMasterData(masterTab);
      fetchReferenceData();
    } catch (error) {
      alert('Gagal menyimpan: ' + (error.response?.data?.error || error.message));
    }
  };

  const deleteMasterData = async (pkField, pkValue) => {
    if (window.confirm(`Yakin ingin menghapus data dengan ${pkField} = ${pkValue}?`)) {
      try {
        await axios.delete(`${API_BASE_URL}/api/master/${masterTab}/${pkValue}`);
        alert('Data berhasil dihapus!');
        fetchMasterData(masterTab);
        fetchReferenceData();
      } catch (error) {
        alert('Gagal menghapus data: ' + (error.response?.data?.error || error.message));
      }
    }
  };

  useEffect(() => {
    if (activeTab === 'master') {
      fetchMasterData(masterTab);
    }
  }, [activeTab, masterTab]);

  const submitParameter = async (e) => {
    e.preventDefault();
    try {
      await axios.post(`${API_BASE_URL}/api/parameters`, {
        id: parseInt(paramForm.id),
        key_name: paramForm.key_name,
        key_value: paramForm.key_value,
        description: paramForm.description
      });
      alert('Parameter berhasil disimpan!');
      setParamForm({ id: 0, key_name: '', key_value: '', description: '' });
      fetchParameters();
    } catch (error) {
      alert('Gagal menyimpan parameter: ' + (error.response?.data?.error || error.message));
    }
  };

  const deleteParameter = async (id) => {
    if(window.confirm('Yakin ingin menghapus parameter ini?')) {
      try {
        await axios.delete(`${API_BASE_URL}/api/parameters/${id}`);
        fetchParameters();
      } catch (error) {
        alert('Gagal menghapus parameter');
      }
    }
  };

  if (!currentUser || (!currentUser.employee_id && !currentUser.username && !currentUser.role)) {
    return (
      <div style={{ display: 'flex', height: '100vh', width: '100vw', alignItems: 'center', justifyContent: 'center', backgroundColor: '#f1f5f9', position: 'fixed', top: 0, left: 0, zIndex: 99999 }}>
        <div style={{ width: '400px', padding: '40px', background: '#ffffff', borderRadius: '12px', boxShadow: '0 4px 20px rgba(0, 0, 0, 0.15)' }}>
          <div style={{ textAlign: 'center', marginBottom: '32px' }}>
            <div style={{ width: '48px', height: '48px', background: 'var(--primary-blue)', borderRadius: '12px', margin: '0 auto 16px' }}></div>
            <h2 style={{ color: '#0f172a', margin: 0 }}>LMS Karisma Login</h2>
            <p style={{ color: '#64748B', margin: '8px 0 0 0' }}>Masuk dengan akun HRIS Adira Anda</p>
          </div>
          
          <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {loginError && <div style={{ padding: '12px', background: '#fee2e2', color: '#991B1B', borderRadius: '8px', fontSize: '0.875rem' }}>{loginError}</div>}
            
            <div>
              <label style={{ display: 'block', marginBottom: '8px', color: '#334155', fontWeight: 500 }}>Username</label>
              <input type="text" required value={loginForm.username} onChange={e => setLoginForm({...loginForm, username: e.target.value})} style={{ width: '100%', padding: '12px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '1rem' }} placeholder="Contoh: 1001, admin, hrd" />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '8px', color: '#334155', fontWeight: 500 }}>Password</label>
              <input type="password" required value={loginForm.password} onChange={e => setLoginForm({...loginForm, password: e.target.value})} style={{ width: '100%', padding: '12px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '1rem' }} placeholder="Password" />
            </div>
            
            <button type="submit" style={{ width: '100%', padding: '12px', background: 'var(--primary-blue)', color: 'white', border: 'none', borderRadius: '8px', fontSize: '1rem', fontWeight: 600, cursor: 'pointer', marginTop: '8px' }}>
              Masuk ke LMS
            </button>
          </form>
          
          <div style={{ marginTop: '24px', fontSize: '0.875rem', color: '#64748B', textAlign: 'center' }}>
            <p><strong>Akun Tersedia (Simulator):</strong></p>
            <ul style={{ listStyle: 'none', padding: 0, margin: '8px 0 0 0' }}>
              <li>User Anggota: <code>[ID Angka Berapa Saja]</code> / <code>password123</code><br/><small>(Contoh: 10101, 1001, dsb)</small></li>
              <li>User Admin: <code>admin</code> / <code>admin123</code></li>
              <li>User HRD: <code>hrd</code> / <code>hrd123</code></li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="app-container">
      {/* Sidebar */}
      <aside className={`sidebar ${isSidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="sidebar-header" style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '16px 20px' }}>
          <img 
            src="/kopkara.jfif" 
            alt="Kopkara Logo" 
            style={{ width: '36px', height: '36px', borderRadius: '6px', objectFit: 'contain', flexShrink: 0, background: '#ffffff', padding: '2px', boxShadow: '0 1px 3px rgba(0,0,0,0.2)' }} 
          />
          {!isSidebarCollapsed && (
            <span style={{ display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontWeight: 'bold', fontSize: '1rem', color: '#ffffff' }}>
                {getParamVal('LMS_Title', 'Kopkara LMS')}
              </span>
              <span style={{ fontSize: '0.75rem', fontWeight: 'normal', color: '#94a3b8', textTransform: 'capitalize' }}>
                Mode: {realRoleName}
              </span>
            </span>
          )}
        </div>
        <nav className="sidebar-nav">
          {(() => {
            const topMenus = visibleMenus.filter(m => !m.parent_id);
            const getSubMenus = (parentId) => visibleMenus.filter(m => String(m.parent_id) === String(parentId));

            return topMenus.map(menu => {
              const subMenus = getSubMenus(menu.menu_id);
              const isExpanded = expandedParents[menu.menu_id];
              const isChildActive = subMenus.some(sub => sub.path === activeTab);
              const isActive = activeTab === menu.path || isChildActive;

              if (subMenus.length > 0) {
                return (
                  <div key={menu.menu_id} style={{ marginBottom: '4px' }}>
                    <div 
                      className={`nav-item ${isActive ? 'active' : ''}`}
                      onClick={() => setExpandedParents(prev => ({ ...prev, [menu.menu_id]: !prev[menu.menu_id] }))}
                      style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', cursor: 'pointer' }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                        <span className="nav-icon">{menu.icon || '🗃️'}</span>
                        {!isSidebarCollapsed && <span className="nav-label">{menu.title}</span>}
                      </div>
                      {!isSidebarCollapsed && (
                        <span style={{ fontSize: '0.65rem', opacity: 0.8 }}>{isExpanded ? '▼' : '▶'}</span>
                      )}
                    </div>
                    {isExpanded && !isSidebarCollapsed && (
                      <div style={{ paddingLeft: '16px', display: 'flex', flexDirection: 'column', gap: '2px', marginTop: '2px' }}>
                        {subMenus.map(sub => (
                          <div 
                            key={sub.menu_id}
                            className={`nav-item ${activeTab === sub.path ? 'active' : ''}`}
                            onClick={() => setActiveTab(sub.path)}
                            style={{ fontSize: '0.875rem', padding: '8px 12px' }}
                          >
                            <span className="nav-icon">{sub.icon || '📌'}</span>
                            <span className="nav-label">{sub.title}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                );
              }

              return (
                <div 
                  key={menu.menu_id}
                  className={`nav-item ${activeTab === menu.path ? 'active' : ''}`}
                  onClick={() => setActiveTab(menu.path)}
                  title={isSidebarCollapsed ? menu.title : ''}
                >
                  <span className="nav-icon">{menu.icon || '📌'}</span>
                  {!isSidebarCollapsed && <span className="nav-label">{menu.title}</span>}
                </div>
              );
            });
          })()}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="main-content">
        <header className="header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <button 
              className="toggle-sidebar-btn"
              onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            >
              ☰
            </button>
            <div className="header-title">
              {visibleMenus.find(m => m.path === activeTab)?.title || (activeTab.startsWith('master-') ? 'Data Master: ' + masterTab.replace('-', ' ') : 'Dashboard')}
            </div>
          </div>
          
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            {(() => {
              const displayName = realEmployee ? realEmployee.name : (currentUser ? currentUser.name : '');
              return (
                <>
                  <span style={{ fontWeight: 500, color: '#334155' }}>
                    {displayName} ({currentUser?.employee_id}) - <span style={{ textTransform: 'capitalize' }}>{realRoleName}</span>
                  </span>
                  <div style={{ width: '40px', height: '40px', borderRadius: '50%', background: 'var(--accent-blue)', color: 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 'bold' }}>
                    {displayName ? displayName.substring(0, 2).toUpperCase() : ''}
                  </div>
                </>
              );
            })()}
            <button 
              onClick={handleLogout}
              style={{ padding: '8px 16px', background: '#fee2e2', color: '#b91c1c', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}
            >
              Logout
            </button>
          </div>
        </header>

        <div className="content-body">
          {activeTab === 'dashboard' && (
            <>
              <div className="card-grid">
                <div className="card">
                  <div className="card-title">Available Credit Limit</div>
                  <div className="card-value" style={{ color: '#059669' }}>
                    Rp {dashboardSummary.available_limit ? dashboardSummary.available_limit.toLocaleString('id-ID') : '0'}
                  </div>
                </div>
                <div className="card">
                  <div className="card-title">Total Hutang</div>
                  <div className="card-value" style={{ color: '#dc2626' }}>
                    Rp {dashboardSummary.total_debt ? dashboardSummary.total_debt.toLocaleString('id-ID') : '0'}
                  </div>
                </div>
                <div className="card">
                  <div className="card-title">Pinjaman Aktif</div>
                  <div className="card-value" style={{ color: '#2563eb' }}>
                    {dashboardSummary.active_loans || 0}
                  </div>
                </div>
              </div>

              <div className="table-container">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px', paddingBottom: '8px', borderBottom: '1px solid #e2e8f0' }}>
                  <div className="table-header" style={{ marginBottom: 0 }}>Pinjaman Terbaru</div>
                  <button 
                    onClick={fetchDashboardSummary} 
                    style={{ padding: '4px 12px', background: '#f1f5f9', color: '#334155', border: '1px solid #cbd5e1', borderRadius: '4px', cursor: 'pointer', fontSize: '0.78rem', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '4px' }}
                  >
                    🔄 Refresh Real-Time Data
                  </button>
                </div>

                <table>
                  <thead>
                    <tr>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>No Pengajuan</th>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Employee ID</th>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Tanggal</th>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Nominal</th>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Tenor</th>
                      <th style={{ padding: '8px 10px', fontSize: '0.85rem' }}>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {!dashboardSummary.recent_loans || dashboardSummary.recent_loans.length === 0 ? (
                      <tr>
                        <td colSpan="6" style={{ textAlign: 'center', padding: '24px 16px', color: '#64748b' }}>
                          Belum ada pengajuan / pinjaman aktif.
                        </td>
                      </tr>
                    ) : (
                      dashboardSummary.recent_loans.map(app => (
                        <tr key={app.ApplicationNo || app.application_no} style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem' }}><strong>{app.ApplicationNo || app.application_no}</strong></td>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem', fontWeight: 600, color: '#0f172a' }}>{app.MemberNo || app.member_no}</td>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                            {formatDate(app.SubmissionDate || app.submission_date)}
                          </td>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem' }}>
                            Rp {(app.RequestedAmount || app.requested_amount) ? (app.RequestedAmount || app.requested_amount).toLocaleString('id-ID') : '0'}
                          </td>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem' }}>{app.Tenor || app.tenor} Bulan</td>
                          <td style={{ padding: '8px 10px', fontSize: '0.85rem' }}>
                            <span className={getStatusBadge(app.Status || app.status)}>
                              {app.Status || app.status}
                            </span>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              {/* Tampilan Integrasi Backend (Products) */}
              <div className="table-container" style={{ marginTop: '32px' }}>
                <div className="table-header">Katalog Produk Pinjaman (Dari Backend)</div>
                <div style={{ padding: '24px' }}>
                  {loading ? (
                    <p>Loading data dari API...</p>
                  ) : products.length === 0 ? (
                    <p>Belum ada produk pinjaman. Coba tambahkan lewat POST /api/products.</p>
                  ) : (
                    <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
                      {products.map(p => {
                        const pId = p.id || p.ID;
                        const pName = p.name || p.Name || `Produk #${pId}`;
                        const pTenor = p.max_tenor_months || p.MaxTenorMonths || 12;
                        const pRate = p.interest_rate !== undefined ? p.interest_rate : (p.InterestRate !== undefined ? p.InterestRate : 0);
                        const pStatus = p.status || p.Status || 'ACTIVE';
                        return (
                          <div key={pId} style={{ border: '1px solid var(--border-color)', padding: '16px', borderRadius: '8px', minWidth: '250px' }}>
                            <h4 style={{ color: 'var(--primary-blue)', marginBottom: '8px' }}>{pName}</h4>
                            <p style={{ fontSize: '0.875rem', color: '#64748B' }}>Max Tenor: {pTenor} bln</p>
                            <p style={{ fontSize: '0.875rem', color: '#64748B' }}>Bunga: {pRate}%</p>
                            <span className={getStatusBadge(pStatus)} style={{ marginTop: '12px', display: 'inline-block' }}>{pStatus}</span>
                          </div>
                        );
                      })}
                    </div>
                  )}
                  <button 
                    onClick={fetchProducts} 
                    style={{ marginTop: '16px', padding: '8px 16px', background: 'var(--primary-blue)', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                  >
                    Refresh Data
                  </button>
                </div>
              </div>
            </>
          )}

          {activeTab === 'pengajuan' && (
            <div className="card" style={{ maxWidth: '700px' }}>
              <h2>Formulir Pengajuan Pinjaman</h2>
              {dateError && (
                <div style={{ padding: '12px', background: '#fee2e2', color: '#991B1B', borderRadius: '4px', marginTop: '16px', fontWeight: 500 }}>
                  ⚠️ {dateError}
                </div>
              )}
              <form onSubmit={submitApplication} style={{ display: 'flex', flexDirection: 'column', gap: '16px', marginTop: '24px' }}>
                <div>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Produk Pinjaman</label>
                  <select 
                    required 
                    value={form.product_id} 
                    onChange={e => setForm({...form, product_id: e.target.value})}
                    style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                  >
                    <option value="">-- Pilih Produk --</option>
                    {products.map(p => {
                      const pId = p.id || p.ID;
                      const pName = p.name || p.Name || `Produk #${pId}`;
                      const pRate = p.interest_rate !== undefined ? p.interest_rate : (p.InterestRate !== undefined ? p.InterestRate : 0);
                      const pType = p.loan_type || p.LoanType || 'FLAT';
                      return (
                        <option key={pId} value={pId}>
                          {pName} - (Bunga {pRate}% / bln - {pType})
                        </option>
                      );
                    })}
                  </select>
                </div>
                <div>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Nominal Pinjaman (Rp)</label>
                  <input 
                    type="text" 
                    required 
                    placeholder="Contoh: 300.000"
                    value={form.requested_amount ? Number(form.requested_amount).toLocaleString('id-ID') : ''}
                    onChange={e => {
                      const rawVal = e.target.value.replace(/\D/g, '');
                      setForm({...form, requested_amount: rawVal});
                    }}
                    style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', fontWeight: 600, fontSize: '0.95rem' }}
                  />
                  {form.requested_amount && (
                    <span style={{ fontSize: '0.75rem', color: '#0284c7', display: 'block', marginTop: '4px' }}>
                      Nominal Terbaca: <strong>Rp {Number(form.requested_amount).toLocaleString('id-ID')}</strong>
                    </span>
                  )}
                </div>
                <div>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Tenor (Bulan)</label>
                  <select 
                    required
                    value={form.tenor}
                    onChange={e => setForm({...form, tenor: e.target.value})}
                    style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                  >
                    <option value="">-- Pilih Tenor --</option>
                    {(() => {
                      const maxParam = parameters.find(p => String(p.KeyName || p.key_name || '').toUpperCase() === 'LOAN_MAX_TENOR');
                      const maxGlobal = maxParam ? (parseInt(maxParam.KeyValue || maxParam.key_value) || 60) : 60;
                      const selectedProduct = products.find(p => String(p.ID || p.id) === String(form.product_id));
                      const maxProduct = selectedProduct ? (parseInt(selectedProduct.MaxTenorMonths || selectedProduct.max_tenor_months) || 60) : 60;
                      const maxTenor = Math.min(maxGlobal, maxProduct);
                      return Array.from({ length: Math.max(1, maxTenor) }, (_, i) => i + 1).map(m => (
                        <option key={m} value={m}>{m} Bulan</option>
                      ));
                    })()}
                  </select>
                </div>

                {/* Simulasi Breakdown */}
                {simulation && !simulation.error && (
                  <div style={{ marginTop: '16px', padding: '16px', background: '#f8fafc', border: '1px solid #cbd5e1', borderRadius: '8px' }}>
                    <h3 style={{ fontSize: '1.1rem', marginBottom: '12px', color: 'var(--primary-blue)' }}>Rincian Simulasi Pinjaman</h3>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                      <div><span style={{ color: '#64748B' }}>Pokok per Bulan:</span> <br/><strong>Rp {Math.round(simulation.principal_per_month).toLocaleString('id-ID')}</strong></div>
                      <div><span style={{ color: '#64748B' }}>Bunga per Bulan ({simulation.interest_rate}%):</span> <br/><strong>Rp {Math.round(simulation.interest_per_month).toLocaleString('id-ID')}</strong></div>
                      <div><span style={{ color: '#64748B' }}>Biaya Administrasi:</span> <br/><strong>Rp {Math.round(simulation.admin_fee).toLocaleString('id-ID')}</strong></div>
                      <div><span style={{ color: '#64748B' }}>Total Angsuran per Bulan:</span> <br/><strong>Rp {Math.round(simulation.total_installment).toLocaleString('id-ID')}</strong></div>
                      <div style={{ gridColumn: 'span 2', marginTop: '8px', paddingTop: '8px', borderTop: '1px dashed #cbd5e1' }}>
                        <span style={{ color: '#64748B' }}>Batas Plafon Maksimal (Berdasarkan Rumus):</span> <br/>
                        <strong style={{ fontSize: '1.1rem', color: '#047857' }}>Rp {simulation.credit_limit ? Math.round(simulation.credit_limit).toLocaleString('id-ID') : 'Tidak Dibatasi'}</strong>
                      </div>
                      <div style={{ gridColumn: 'span 2', marginTop: '4px' }}>
                        <span style={{ color: '#64748B' }}>Total Kewajiban (Pokok + Bunga + Admin):</span> <br/>
                        <strong style={{ fontSize: '1.2rem', color: '#0f172a' }}>Rp {Math.round(simulation.total_loan_cost).toLocaleString('id-ID')}</strong>
                      </div>
                    </div>
                  </div>
                )}
                {simulation && simulation.error && (
                  <div style={{ padding: '12px', background: '#fee2e2', color: '#991B1B', borderRadius: '4px', marginTop: '16px' }}>
                    ❌ {simulation.error}
                  </div>
                )}

                <button 
                  type="submit" 
                  disabled={dateError !== '' || (simulation && simulation.error)}
                  style={{ padding: '12px', background: dateError || (simulation && simulation.error) ? '#94a3b8' : 'var(--success-green)', color: 'white', border: 'none', borderRadius: '4px', fontWeight: 600, cursor: dateError || (simulation && simulation.error) ? 'not-allowed' : 'pointer', marginTop: '16px' }}>
                  Kirim Pengajuan
                </button>
              </form>
            </div>
          )}

          {activeTab === 'pinjaman' && (
            <div className="table-container">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px', paddingBottom: '8px', borderBottom: '1px solid #e2e8f0', flexWrap: 'wrap', gap: '8px' }}>
                <div style={{ fontSize: '1.1rem', fontWeight: 'bold', color: '#1e293b' }}>
                  Daftar Pengajuan & Pinjaman
                </div>

                {/* Badge Status Privilege Role */}
                {(roleId === 1 || roleId === 3 || String(realRoleName).toLowerCase().includes('admin') || String(realRoleName).toLowerCase().includes('hrd')) ? (
                  <span style={{ backgroundColor: '#e0e7ff', color: '#3730a3', border: '1px solid #c7d2fe', padding: '4px 12px', borderRadius: '16px', fontSize: '0.78rem', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                    ⭐ Role High-Privilege ({realRoleName}): Akses Seluruh Data Pinjaman
                  </span>
                ) : (
                  <span style={{ backgroundColor: '#f1f5f9', color: '#475569', border: '1px solid #cbd5e1', padding: '4px 12px', borderRadius: '16px', fontSize: '0.78rem', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                    🔒 Mode Anggota: Data Pinjaman Pribadi ({currentUser?.employee_id || 'Anggota'})
                  </span>
                )}
              </div>

              {/* Form Filter Pencarian No. Anggota / Employee ID & Periode */}
              {(() => {
                const isHighPriv = Boolean(
                  roleId === 1 || roleId === 3 || 
                  String(realRoleName).toLowerCase().includes('admin') || 
                  String(realRoleName).toLowerCase().includes('hrd')
                );
                return (
                  <div style={{ background: '#f8fafc', padding: '12px 14px', borderRadius: '8px', marginBottom: '16px', border: '1px solid #cbd5e1', display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span style={{ fontSize: '0.85rem', fontWeight: 600, color: '#334155' }}>
                        🔍 Cari berdasarkan No. Anggota / Employee ID:
                      </span>
                      <input 
                        type="text"
                        placeholder={isHighPriv ? "Kosongkan untuk semua, atau ketik No. Anggota..." : "No. Anggota Anda"}
                        value={isHighPriv ? loanSearchMemberNo : (currentUser?.employee_id || loanSearchMemberNo)}
                        disabled={!isHighPriv}
                        onChange={(e) => {
                          if (isHighPriv) setLoanSearchMemberNo(e.target.value);
                        }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            handleSearchLoans();
                          }
                        }}
                        style={{ 
                          padding: '6px 12px', 
                          fontSize: '0.85rem', 
                          borderRadius: '6px', 
                          border: '1px solid #94a3b8', 
                          minWidth: '220px',
                          backgroundColor: isHighPriv ? '#ffffff' : '#e2e8f0',
                          color: isHighPriv ? '#0f172a' : '#475569',
                          cursor: isHighPriv ? 'text' : 'not-allowed',
                          fontWeight: isHighPriv ? 400 : 600
                        }}
                      />
                    </div>

                    {/* Filter Periode (Bulan & Tahun) */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span style={{ fontSize: '0.85rem', fontWeight: 600, color: '#334155' }}>
                        📅 Periode:
                      </span>
                      <select
                        value={loanMonthFilter.month}
                        onChange={(e) => setLoanMonthFilter(prev => ({ ...prev, month: e.target.value }))}
                        style={{ padding: '6px 10px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid #94a3b8', background: '#fff' }}
                      >
                        <option value="01">01 - Januari</option>
                        <option value="02">02 - Februari</option>
                        <option value="03">03 - Maret</option>
                        <option value="04">04 - April</option>
                        <option value="05">05 - Mei</option>
                        <option value="06">06 - Juni</option>
                        <option value="07">07 - Juli</option>
                        <option value="08">08 - Agustus</option>
                        <option value="09">09 - September</option>
                        <option value="10">10 - Oktober</option>
                        <option value="11">11 - November</option>
                        <option value="12">12 - Desember</option>
                      </select>
                      <select
                        value={loanMonthFilter.year}
                        onChange={(e) => setLoanMonthFilter(prev => ({ ...prev, year: e.target.value }))}
                        style={{ padding: '6px 10px', fontSize: '0.85rem', borderRadius: '6px', border: '1px solid #94a3b8', background: '#fff' }}
                      >
                        <option value="2024">2024</option>
                        <option value="2025">2025</option>
                        <option value="2026">2026</option>
                        <option value="2027">2027</option>
                        <option value="2028">2028</option>
                      </select>
                    </div>

                    <button 
                      onClick={() => handleSearchLoans()}
                      style={{ padding: '6px 16px', background: '#2563eb', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                    >
                      Cari Pinjaman
                    </button>

                    {isHighPriv && (loanSearchMemberNo || loanMonthFilter.month !== '08' || loanMonthFilter.year !== '2026') && (
                      <button 
                        onClick={() => {
                          setLoanSearchMemberNo('');
                          setLoanMonthFilter({ month: '08', year: '2026' });
                          handleSearchLoans('', '2026', '08');
                        }}
                        style={{ padding: '6px 12px', background: '#64748b', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontSize: '0.85rem' }}
                      >
                        Reset
                      </button>
                    )}
                  </div>
                );
              })()}

              <table>
                <thead>
                  <tr>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>No. Pengajuan</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Employee ID</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tanggal Submit</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Nominal</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tenor</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Status</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Catatan HRD</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tanggal Approval</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem', textAlign: 'center' }}>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {!hasSearchedLoans ? (
                    <tr>
                      <td colSpan="9" style={{ textAlign: 'center', padding: '36px 16px', color: '#475569', fontSize: '0.9rem' }}>
                        🔍 Silakan pilih periode & klik tombol <strong>"Cari Pinjaman"</strong> untuk menampilkan data.
                      </td>
                    </tr>
                  ) : applications.length === 0 ? (
                    <tr>
                      <td colSpan="9" style={{ textAlign: 'center', padding: '24px 16px', color: '#dc2626', fontWeight: 600, backgroundColor: '#fef2f2', borderBottom: '1px solid #fecaca' }}>
                        ⚠️ Data pinjaman tidak ditemukan (Periode {loanMonthFilter.year}-{loanMonthFilter.month})
                      </td>
                    </tr>
                  ) : (
                    (() => {
                      const pageSize = parseInt(getParamVal('PAGINATION_LIMIT', getParamVal('DEFAULT_PAGE_SIZE', '5'))) || 5;
                      const totalPages = Math.ceil(applications.length / pageSize) || 1;
                      const safePage = Math.min(Math.max(loanPage, 1), totalPages);
                      const startIndex = (safePage - 1) * pageSize;
                      const paginatedApps = applications.slice(startIndex, startIndex + pageSize);
                      return paginatedApps.map(app => {
                        const appNo = app.application_no || app.ApplicationNo;
                        const memberNo = app.member_no || app.MemberNo;
                        const submissionDate = app.submission_date || app.SubmissionDate;
                        const requestedAmount = app.requested_amount ?? app.RequestedAmount ?? 0;
                        const tenor = app.tenor ?? app.Tenor ?? 0;
                        const status = app.status || app.Status || '';
                        const approvalNotes = app.approval_notes || app.ApprovalNotes || '-';
                        const approvedAt = app.approved_at || app.ApprovedAt;

                        return (
                          <tr key={appNo} style={{ borderBottom: '1px solid #e2e8f0' }}>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}><strong>{appNo}</strong></td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', fontWeight: 600, color: '#0f172a' }}>{memberNo}</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                              {formatDate(submissionDate)}
                            </td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Rp {requestedAmount ? Number(requestedAmount).toLocaleString('id-ID') : '0'}</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>{tenor} Bulan</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>
                              <span className={getStatusBadge(status)} style={{ backgroundColor: status === 'REVISION_REQUIRED' ? '#f59e0b' : undefined }}>
                                {status === 'REVISION_REQUIRED' ? 'PERLU REVISI' : status}
                              </span>
                            </td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#b45309', fontWeight: 500 }}>
                              {approvalNotes}
                            </td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                              {formatDate(approvedAt)}
                            </td>
                            <td style={{ padding: '6px 10px', textAlign: 'center', whiteSpace: 'nowrap' }}>
                              <div style={{ display: 'flex', gap: '4px', justifyContent: 'center' }}>
                                {status === 'REVISION_REQUIRED' && (
                                  <button 
                                    onClick={() => {
                                      setForm({
                                        member_no: memberNo,
                                        product_id: app.product_id || app.ProductID || '',
                                        requested_amount: requestedAmount,
                                        tenor: tenor
                                      });
                                      setActiveTab('pengajuan');
                                      alert(`Silakan revisi nominal/tenor pengajuan #${appNo} lalu klik Kirim Pengajuan.`);
                                    }}
                                    style={{ padding: '3px 8px', background: '#f59e0b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  >
                                    ✏️ Revisi
                                  </button>
                                )}
                                {(status === 'APPROVED' || status === 'DISBURSED') && (
                                  <button 
                                    onClick={() => handlePrintContract(app)}
                                    style={{ padding: '3px 8px', background: '#0f172a', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                    title="Cetak Surat Perjanjian Kredit & Kontrak Pinjaman"
                                  >
                                    📄 Kontrak
                                  </button>
                                )}
                                {status === 'APPROVED' && (
                                  <button 
                                    onClick={() => handleDisburse(app)}
                                    style={{ padding: '3px 8px', background: '#10b981', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                    title="Cairkan Dana Pinjaman"
                                  >
                                    💵 Cairkan
                                  </button>
                                )}
                                <button 
                                  onClick={() => handleOpenTracking(appNo)}
                                  style={{ padding: '3px 8px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Lihat Riwayat Status / Tracking"
                                >
                                  📜 Track
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      });
                    })()
                  )}
                </tbody>
              </table>

              {/* Pagination Controls */}
              {applications.length > 0 && (() => {
                const pageSize = parseInt(getParamVal('PAGINATION_LIMIT', getParamVal('DEFAULT_PAGE_SIZE', '5'))) || 5;
                const totalPages = Math.ceil(applications.length / pageSize) || 1;
                const safePage = Math.min(Math.max(loanPage, 1), totalPages);

                return (
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 16px', background: '#f8fafc', borderTop: '1px solid #cbd5e1', fontSize: '0.85rem', color: '#475569' }}>
                    <span>Halaman <strong>{safePage}</strong> dari <strong>{totalPages}</strong> ({applications.length} Data) | Limit per Halaman: <strong>{pageSize}</strong></span>
                    <div style={{ display: 'flex', gap: '6px' }}>
                      <button
                        disabled={safePage <= 1}
                        onClick={() => setLoanPage(prev => Math.max(prev - 1, 1))}
                        style={{ padding: '5px 12px', background: safePage <= 1 ? '#e2e8f0' : '#0284c7', color: safePage <= 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: safePage <= 1 ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                      >
                        ◀ Prev
                      </button>
                      <button
                        disabled={safePage >= totalPages}
                        onClick={() => setLoanPage(prev => Math.min(prev + 1, totalPages))}
                        style={{ padding: '5px 12px', background: safePage >= totalPages ? '#e2e8f0' : '#0284c7', color: safePage >= totalPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: safePage >= totalPages ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                      >
                        Next ▶
                      </button>
                    </div>
                  </div>
                );
              })()}
            </div>
          )}

          {activeTab === 'parameters' && (
            <div style={{ display: 'flex', gap: '24px', flexWrap: 'wrap' }}>
              <div className="card" style={{ flex: '1', minWidth: '300px' }}>
                <h2>{paramForm.id ? 'Edit Parameter' : 'Tambah Parameter Baru'}</h2>
                <form onSubmit={submitParameter} style={{ display: 'flex', flexDirection: 'column', gap: '16px', marginTop: '24px' }}>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Key Name (Misal: max_loan_limit)</label>
                    <input 
                      type="text" required
                      value={paramForm.key_name}
                      onChange={e => setParamForm({...paramForm, key_name: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                      disabled={paramForm.id > 0} // Key tidak bisa diubah jika edit
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Key Value</label>
                    <input 
                      type="text" required
                      value={paramForm.key_value}
                      onChange={e => setParamForm({...paramForm, key_value: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' }}
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500 }}>Deskripsi</label>
                    <textarea 
                      value={paramForm.description}
                      onChange={e => setParamForm({...paramForm, description: e.target.value})}
                      style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', minHeight: '80px' }}
                    />
                  </div>
                  <div style={{ display: 'flex', gap: '12px' }}>
                    <button type="submit" style={{ flex: 1, padding: '12px', background: 'var(--primary-blue)', color: 'white', border: 'none', borderRadius: '4px', fontWeight: 600, cursor: 'pointer' }}>
                      Simpan Parameter
                    </button>
                    {paramForm.id > 0 && (
                      <button type="button" onClick={() => setParamForm({ id: 0, key_name: '', key_value: '', description: '' })} style={{ padding: '12px', background: '#e2e8f0', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
                        Batal Edit
                      </button>
                    )}
                  </div>
                </form>
              </div>

              <div className="table-container" style={{ flex: '2', minWidth: '500px' }}>
                <div className="table-header">Daftar Parameter Global</div>
                <table>
                  <thead>
                    <tr>
                      <th style={{ padding: '8px 12px', fontSize: '0.875rem' }}>Key Name</th>
                      <th style={{ padding: '8px 12px', fontSize: '0.875rem' }}>Value</th>
                      <th style={{ padding: '8px 12px', fontSize: '0.875rem' }}>Deskripsi</th>
                      <th style={{ padding: '8px 12px', fontSize: '0.875rem' }}>Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    {parameters.length === 0 ? (
                      <tr><td colSpan="4" style={{ textAlign: 'center' }}>Belum ada konfigurasi parameter</td></tr>
                    ) : (
                      parameters.map(param => (
                        <tr key={param.ID} style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '6px 12px', fontSize: '0.875rem' }}><strong>{param.KeyName}</strong></td>
                          <td style={{ padding: '6px 12px', fontSize: '0.875rem' }}>{param.KeyValue}</td>
                          <td style={{ padding: '6px 12px', fontSize: '0.875rem' }}>{param.Description}</td>
                          <td style={{ padding: '6px 12px' }}>
                            <button onClick={() => setParamForm({id: param.ID, key_name: param.KeyName, key_value: param.KeyValue, description: param.Description})} style={{ padding: '3px 8px', marginRight: '8px', background: '#fef3c7', color: '#92400E', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem' }}>Edit</button>
                            <button onClick={() => deleteParameter(param.ID)} style={{ padding: '3px 8px', background: '#fee2e2', color: '#991B1B', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem' }}>Hapus</button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {activeTab === 'disbursement' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {/* Header Box Verifikasi Pencairan Dana & Filter Partisi (Matching Screenshot #2) */}
              <div style={{ background: '#ffffff', borderRadius: '12px', border: '1px solid #e2e8f0', padding: '20px', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
                <h3 style={{ fontSize: '1.1rem', color: '#1e293b', margin: '0 0 4px 0', fontWeight: 'bold' }}>Verifikasi Pencairan Dana</h3>
                <div style={{ fontSize: '0.8rem', color: '#2563eb', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span>🗄️ LMS System -</span>
                  <span style={{ fontWeight: 600 }}>Querying on: loan_applications_{disbursementMonthFilter.year}{disbursementMonthFilter.month}</span>
                </div>

                <div style={{ background: '#f8fafc', border: '1px solid #cbd5e1', borderRadius: '10px', padding: '16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '16px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>BULAN</label>
                      <select 
                        value={disbursementMonthFilter.month} 
                        onChange={e => setDisbursementMonthFilter({...disbursementMonthFilter, month: e.target.value})}
                        style={{ padding: '6px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', background: '#ffffff', fontSize: '0.9rem', minWidth: '120px' }}
                      >
                        <option value="01">Januari</option>
                        <option value="02">Februari</option>
                        <option value="03">Maret</option>
                        <option value="04">April</option>
                        <option value="05">Mei</option>
                        <option value="06">Juni</option>
                        <option value="07">Juli</option>
                        <option value="08">Agustus</option>
                        <option value="09">September</option>
                        <option value="10">Oktober</option>
                        <option value="11">November</option>
                        <option value="12">Desember</option>
                      </select>
                    </div>

                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>TAHUN</label>
                      <select 
                        value={disbursementMonthFilter.year} 
                        onChange={e => setDisbursementMonthFilter({...disbursementMonthFilter, year: e.target.value})}
                        style={{ padding: '6px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', background: '#ffffff', fontSize: '0.9rem', minWidth: '100px' }}
                      >
                        <option value="2024">2024</option>
                        <option value="2025">2025</option>
                        <option value="2026">2026</option>
                        <option value="2027">2027</option>
                        <option value="2028">2028</option>
                      </select>
                    </div>

                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>STATUS AKTIF</label>
                      <div style={{ padding: '6px 16px', background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '6px', color: '#2563eb', fontWeight: 'bold', fontSize: '0.85rem' }}>
                        APPROVED
                      </div>
                    </div>
                  </div>

                  <button 
                    type="button"
                    onClick={() => fetchApplications(disbursementMonthFilter.year, disbursementMonthFilter.month)}
                    style={{ background: '#10b981', color: '#ffffff', border: 'none', borderRadius: '8px', padding: '8px 24px', fontWeight: 'bold', fontSize: '0.9rem', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px', boxShadow: '0 2px 4px rgba(16,185,129,0.2)' }}
                  >
                    <span>🔍</span> Filter
                  </button>
                </div>
              </div>

              <div className="table-container">
                <div className="table-header">Pencairan Dana Pinjaman (Disbursement Treasury)</div>
                <table>
                  <thead>
                    <tr>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>No. Pengajuan</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tanggal Approved</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Pemohon</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Nominal Plafon</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tenor</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Status</th>
                      <th style={{ padding: '6px 10px', fontSize: '0.85rem', textAlign: 'center' }}>Aksi Pencairan</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(() => {
                      const filteredApps = applications.filter(a => {
                        const st = (a.status || a.Status || '').toUpperCase();
                        return st === 'APPROVED' || st === 'DISBURSED';
                      });

                      if (filteredApps.length === 0) {
                        return (
                          <tr>
                            <td colSpan="7" style={{ textAlign: 'center', padding: '24px', background: '#ffffff', color: '#64748b', fontWeight: 500, fontSize: '0.95rem' }}>
                              Data tidak ditemukan
                            </td>
                          </tr>
                        );
                      }

                      return filteredApps.map(app => {
                        const appNo = app.application_no || app.ApplicationNo;
                        const memberNo = app.member_no || app.MemberNo;
                        const submissionDate = app.submission_date || app.SubmissionDate;
                        const approvedAt = app.approved_at || app.ApprovedAt;
                        const approvedAmount = app.approved_amount ?? app.ApprovedAmount ?? app.requested_amount ?? app.RequestedAmount ?? 0;
                        const tenor = app.tenor ?? app.Tenor ?? 0;
                        const status = (app.status || app.Status || '').toUpperCase().trim();

                        return (
                          <tr key={appNo} style={{ borderBottom: '1px solid #e2e8f0' }}>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}><strong>{appNo}</strong></td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>{formatDate(approvedAt || submissionDate)}</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem', fontWeight: 600 }}>Member #{memberNo}</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}><strong>Rp {Number(approvedAmount).toLocaleString('id-ID')}</strong></td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>{tenor} Bulan</td>
                            <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>
                              <span className={getStatusBadge(status)}>
                                {status === 'DISBURSED' ? 'DICAIRKAN' : 'SETUJU (APPROVED)'}
                              </span>
                            </td>
                            <td style={{ padding: '6px 10px', textAlign: 'center', whiteSpace: 'nowrap' }}>
                              <div style={{ display: 'flex', gap: '6px', justifyContent: 'center' }}>
                                <button 
                                  onClick={() => handlePrintContract(app)}
                                  style={{ padding: '4px 10px', background: '#0f172a', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.8rem' }}
                                >
                                  📄 Cetak Kontrak
                                </button>
                                {status !== 'DISBURSED' && status !== 'REJECTED' && (
                                  <button 
                                    onClick={() => handleDisburse(app)}
                                    style={{ padding: '4px 10px', background: '#10b981', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.8rem' }}
                                  >
                                    💵 Cairkan Dana
                                  </button>
                                )}
                                <button 
                                  onClick={() => handleOpenTracking(appNo)}
                                  style={{ padding: '4px 10px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.8rem' }}
                                >
                                  📜 Track
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      });
                    })()}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {(activeTab === 'payroll' || activeTab === 'payroll-reconciliation') && renderPayrollContent()}

              {/* Modal UI Filter Cutoff Export Payroll HRD Adira */}
              {exportModalOpen && (
                <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
                  <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '90%', maxWidth: '540px', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                      <h3 style={{ margin: 0, fontSize: '1.15rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                        📥 Export File Tagihan Payroll HRD Adira
                      </h3>
                      <button onClick={() => setExportModalOpen(false)} style={{ background: '#f1f5f9', border: 'none', borderRadius: '50%', width: '32px', height: '32px', cursor: 'pointer', fontWeight: 'bold', color: '#64748b' }}>✕</button>
                    </div>

                    <form onSubmit={async (e) => {
                      e.preventDefault();
                      try {
                        const scanMode = getParamVal('SCAN_DUEDATE_BILLING', 'PERIOD').toUpperCase();
                        const effectiveCutoffDate = scanMode === 'DUEDATE' ? exportCutoffDate : `${exportPeriodYear}-${exportPeriodMonth}-31`;
                        const res = await axios.post(`${API_BASE_URL}/api/payroll/export`, {
                          custom_folder: exportCustomFolder,
                          cutoff_date: effectiveCutoffDate
                        });
                        const csvContent = "data:text/csv;charset=utf-8," + (res.data.csv_content || "");
                        const encodedUri = encodeURI(csvContent);
                        const link = document.createElement("a");
                        link.setAttribute("href", encodedUri);
                        link.setAttribute("download", res.data.file_name || `ADIRA_PAYROLL_KOPKARA_OUTGOING_${effectiveCutoffDate.replace(/-/g, '')}.csv`);
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);

                        setExportModalOpen(false);
                        const labelInfo = scanMode === 'DUEDATE' ? `Cut-Off Tanggal: s/d ${effectiveCutoffDate}` : `Cut-Off Periode: s/d ${exportPeriodYear}-${exportPeriodMonth}`;
                        alert(`✅ File CSV Tagihan Payroll HRD Adira BERHASIL digenerate!\n\n📁 Folder Simpan: ${res.data.file_path || exportCustomFolder}\n📅 ${labelInfo}\n📊 Total ${res.data.total_rows || 0} tagihan karyawan diproses.`);
                      } catch (err) {
                        alert("❌ Gagal mengeksport file payroll: " + (err.response?.data?.error || err.message));
                      }
                    }} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                      
                      {getParamVal('SCAN_DUEDATE_BILLING', 'PERIOD').toUpperCase() === 'DUEDATE' ? (
                        <div style={{ background: '#eff6ff', padding: '12px', borderRadius: '6px', border: '1px solid #bfdbfe' }}>
                          <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '6px', color: '#1e40af' }}>
                            📅 Batas Tanggal Jatuh Tempo (Cut-Off Due Date) *
                          </label>
                          <input 
                            type="date"
                            value={exportCutoffDate}
                            onChange={(e) => setExportCutoffDate(e.target.value)}
                            required
                            style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #93c5fd', fontSize: '0.9rem' }}
                          />
                          <div style={{ fontSize: '0.75rem', color: '#1e3a8a', marginTop: '6px', lineHeight: 1.4 }}>
                            💡 <strong>Ketentuan Cut-off:</strong> Default otomatis diset ke <strong>akhir bulan berjalan</strong> ({exportCutoffDate}). Sistem hanya mengeksport tagihan jatuh tempo <strong>s/d tanggal ini</strong> yang berstatus <strong>UNPAID / PARTIAL</strong>. Tagihan bulan-bulan berikutnya <strong>TIDAK AKAN DIEKSPORT</strong>.
                          </div>
                        </div>
                      ) : (
                        <div style={{ background: '#eff6ff', padding: '12px', borderRadius: '6px', border: '1px solid #bfdbfe' }}>
                          <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '6px', color: '#1e40af' }}>
                            📅 Periode Cut-Off Tagihan (Bulan & Tahun) *
                          </label>
                          <div style={{ display: 'flex', gap: '12px', marginBottom: '8px' }}>
                            <div style={{ flex: 1 }}>
                              <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 700, color: '#1e40af', marginBottom: '2px' }}>BULAN</label>
                              <select
                                value={exportPeriodMonth}
                                onChange={(e) => setExportPeriodMonth(e.target.value)}
                                style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #93c5fd', background: '#ffffff', fontSize: '0.9rem', fontWeight: 600 }}
                              >
                                <option value="01">Januari</option>
                                <option value="02">Februari</option>
                                <option value="03">Maret</option>
                                <option value="04">April</option>
                                <option value="05">Mei</option>
                                <option value="06">Juni</option>
                                <option value="07">Juli</option>
                                <option value="08">Agustus</option>
                                <option value="09">September</option>
                                <option value="10">Oktober</option>
                                <option value="11">November</option>
                                <option value="12">Desember</option>
                              </select>
                            </div>

                            <div style={{ flex: 1 }}>
                              <label style={{ display: 'block', fontSize: '0.7rem', fontWeight: 700, color: '#1e40af', marginBottom: '2px' }}>TAHUN</label>
                              <select
                                value={exportPeriodYear}
                                onChange={(e) => setExportPeriodYear(e.target.value)}
                                style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #93c5fd', background: '#ffffff', fontSize: '0.9rem', fontWeight: 600 }}
                              >
                                <option value="2024">2024</option>
                                <option value="2025">2025</option>
                                <option value="2026">2026</option>
                                <option value="2027">2027</option>
                                <option value="2028">2028</option>
                              </select>
                            </div>
                          </div>
                          <div style={{ fontSize: '0.75rem', color: '#1e3a8a', marginTop: '6px', lineHeight: 1.4 }}>
                            💡 <strong>Ketentuan Cut-off Periode:</strong> Sistem akan mengeksport tagihan <strong>s/d periode {exportPeriodYear}-{exportPeriodMonth}</strong> yang berstatus <strong>UNPAID / PARTIAL</strong>. Tagihan periode berikutnya <strong>TIDAK AKAN DIEKSPORT</strong>.
                          </div>
                        </div>
                      )}

                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px', color: '#0f172a' }}>
                          📁 Target Folder Penyimpanan (.csv File) *
                        </label>
                        <input 
                          type="text"
                          value={exportCustomFolder}
                          onChange={(e) => setExportCustomFolder(e.target.value)}
                          required
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1', fontSize: '0.85rem', fontFamily: 'monospace' }}
                        />
                      </div>

                      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '12px', borderTop: '1px solid #e2e8f0' }}>
                        <button type="button" onClick={() => setExportModalOpen(false)} style={{ padding: '8px 16px', background: '#64748b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Batal</button>
                        <button type="submit" style={{ padding: '8px 18px', background: '#10b981', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>📥 Proses Export File CSV</button>
                      </div>
                    </form>
                  </div>
                </div>
              )}
            

          {activeTab === 'approval' && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {/* Filter Bar Verifikasi Administratif Matching Screenshot #2 */}
              <div style={{ background: '#ffffff', borderRadius: '12px', border: '1px solid #e2e8f0', padding: '20px', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
                <h3 style={{ fontSize: '1.1rem', color: '#1e293b', margin: '0 0 4px 0', fontWeight: 'bold' }}>Verifikasi Administratif</h3>
                <div style={{ fontSize: '0.8rem', color: '#2563eb', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span>🗄️ LMS System -</span>
                  <span style={{ fontWeight: 600 }}>Querying on: loan_applications_{approvalMonthFilter.year}{approvalMonthFilter.month}</span>
                </div>

                <div style={{ background: '#f8fafc', border: '1px solid #cbd5e1', borderRadius: '10px', padding: '16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '16px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>BULAN</label>
                      <select 
                        value={approvalMonthFilter.month} 
                        onChange={e => setApprovalMonthFilter({...approvalMonthFilter, month: e.target.value})}
                        style={{ padding: '6px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', background: '#ffffff', fontSize: '0.9rem', minWidth: '120px' }}
                      >
                        <option value="01">Januari</option>
                        <option value="02">Februari</option>
                        <option value="03">Maret</option>
                        <option value="04">April</option>
                        <option value="05">Mei</option>
                        <option value="06">Juni</option>
                        <option value="07">Juli</option>
                        <option value="08">Agustus</option>
                        <option value="09">September</option>
                        <option value="10">Oktober</option>
                        <option value="11">November</option>
                        <option value="12">Desember</option>
                      </select>
                    </div>

                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>TAHUN</label>
                      <select 
                        value={approvalMonthFilter.year} 
                        onChange={e => setApprovalMonthFilter({...approvalMonthFilter, year: e.target.value})}
                        style={{ padding: '6px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', background: '#ffffff', fontSize: '0.9rem', minWidth: '100px' }}
                      >
                        <option value="2024">2024</option>
                        <option value="2025">2025</option>
                        <option value="2026">2026</option>
                        <option value="2027">2027</option>
                        <option value="2028">2028</option>
                      </select>
                    </div>

                    <div>
                      <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: '#475569', marginBottom: '4px', textTransform: 'uppercase' }}>STATUS AKTIF</label>
                      <div style={{ padding: '6px 16px', background: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '6px', color: '#2563eb', fontWeight: 'bold', fontSize: '0.85rem' }}>
                        SUBMITTED
                      </div>
                    </div>
                  </div>

                  <button 
                    type="button"
                    onClick={() => fetchApplications(approvalMonthFilter.year, approvalMonthFilter.month)}
                    style={{ background: '#10b981', color: '#ffffff', border: 'none', borderRadius: '8px', padding: '8px 24px', fontWeight: 'bold', fontSize: '0.9rem', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px', boxShadow: '0 2px 4px rgba(16,185,129,0.2)' }}
                  >
                    <span>🔍</span> Filter
                  </button>
                </div>
              </div>

              <div className="table-container">
                <div className="table-header">Approval Pengajuan Pinjaman (HRD / Approval)</div>
                <table>
                <thead>
                  <tr>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>No. Pengajuan</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tanggal Submit</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Pemohon</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Nominal</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tenor</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Status</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Catatan HRD</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Tanggal Diproses</th>
                    <th style={{ padding: '6px 10px', fontSize: '0.85rem', textAlign: 'center' }}>Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  {applications.length === 0 ? (
                    <tr><td colSpan="9" style={{ textAlign: 'center', padding: '20px' }}>Belum ada pengajuan pinjaman untuk diproses</td></tr>
                  ) : (
                    applications.map(app => {
                      const appNo = app.application_no || app.ApplicationNo;
                      const memberNo = app.member_no || app.MemberNo;
                      const submissionDate = app.submission_date || app.SubmissionDate;
                      const requestedAmount = app.requested_amount ?? app.RequestedAmount ?? 0;
                      const tenor = app.tenor ?? app.Tenor ?? 0;
                      const status = app.status || app.Status || '';
                      const approvalNotes = app.approval_notes || app.ApprovalNotes || '-';
                      const approvedAt = app.approved_at || app.ApprovedAt;

                      return (
                        <tr key={appNo} style={{ borderBottom: '1px solid #e2e8f0' }}>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}><strong>{appNo}</strong></td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                            {formatDate(submissionDate)}
                          </td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem', fontWeight: 600 }}>Member #{memberNo}</td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>Rp {requestedAmount ? Number(requestedAmount).toLocaleString('id-ID') : '0'}</td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>{tenor} Bulan</td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem' }}>
                            <span className={getStatusBadge(status)} style={{ backgroundColor: status === 'REVISION_REQUIRED' ? '#f59e0b' : undefined }}>
                              {status === 'REVISION_REQUIRED' ? 'REVISI' : status}
                            </span>
                          </td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                            {approvalNotes}
                          </td>
                          <td style={{ padding: '6px 10px', fontSize: '0.85rem', color: '#64748B' }}>
                            {formatDate(approvedAt)}
                          </td>
                          <td style={{ padding: '6px 10px', textAlign: 'center', whiteSpace: 'nowrap' }}>
                            {status === 'SUBMITTED' ? (
                              <div style={{ display: 'flex', gap: '4px', justifyContent: 'center' }}>
                                <button 
                                  onClick={() => handleProcessApproval(appNo, 'APPROVED')}
                                  style={{ padding: '3px 8px', background: 'var(--success-green)', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Setujui Pengajuan"
                                >
                                  ✅ Setuju
                                </button>
                                <button 
                                  onClick={() => handleProcessApproval(appNo, 'REVISION_REQUIRED')}
                                  style={{ padding: '3px 8px', background: '#f59e0b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Minta Revisi Pengajuan"
                                >
                                  ✏️ Revisi
                                </button>
                                <button 
                                  onClick={() => handleProcessApproval(appNo, 'REJECTED')}
                                  style={{ padding: '3px 8px', background: '#ef4444', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Tolak Pengajuan"
                                >
                                  ❌ Tolak
                                </button>
                                <button 
                                  onClick={() => handleOpenTracking(appNo)}
                                  style={{ padding: '3px 8px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Lihat Riwayat Status / Tracking"
                                >
                                  📜 Track
                                </button>
                              </div>
                            ) : (
                              <div style={{ display: 'flex', gap: '4px', justifyContent: 'center', alignItems: 'center' }}>
                                {(status === 'APPROVED' || status === 'DISBURSED') && (
                                  <button 
                                    onClick={() => handlePrintContract(app)}
                                    style={{ padding: '3px 8px', background: '#0f172a', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                    title="Cetak Surat Perjanjian Kredit & Kontrak Pinjaman"
                                  >
                                    📄 Kontrak
                                  </button>
                                )}
                                {status === 'APPROVED' && (
                                  <button 
                                    onClick={() => handleDisburse(app)}
                                    style={{ padding: '3px 8px', background: '#10b981', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                    title="Cairkan Dana Pinjaman"
                                  >
                                    💵 Cairkan
                                  </button>
                                )}
                                <button 
                                  onClick={() => handleOpenTracking(appNo)}
                                  style={{ padding: '3px 8px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                  title="Lihat Riwayat Status / Tracking"
                                >
                                  📜 Track
                                </button>
                              </div>
                            )}
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

          {activeTab.startsWith('master-') && (
            <div className="card" style={{ maxWidth: '1200px' }}>
              <h2 style={{ textTransform: 'capitalize', marginBottom: '24px', borderBottom: '1px solid #e2e8f0', paddingBottom: '12px' }}>
                Kelola Data Master: {masterTab.replace('-', ' ')}
              </h2>

              {masterTab === 'role-menus' ? (
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
                    <div>
                      <h3 style={{ margin: 0, color: 'var(--primary-blue)' }}>🔐 Access Privilege Matrix (APM)</h3>
                      <p style={{ margin: '4px 0 0 0', color: '#64748B', fontSize: '0.9rem' }}>
                        Centang kotak di bawah ini untuk mengatur hak akses menu setiap Role (Admin, Anggota, HRD/Approval). Perubahan langsung tersimpan ke Database!
                      </p>
                    </div>
                    <button onClick={fetchReferenceData} style={{ padding: '8px 16px', background: '#e0f2fe', color: '#0369a1', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 600 }}>
                      🔄 Refresh Matrix
                    </button>
                  </div>

                  <div style={{ overflowX: 'auto', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '700px' }}>
                      <thead>
                        <tr style={{ background: 'var(--primary-blue)', color: 'white' }}>
                          <th style={{ padding: '14px', textAlign: 'left', minWidth: '240px' }}>Modul / Menu LMS</th>
                          <th style={{ padding: '14px', textAlign: 'left', width: '160px' }}>Path Route</th>
                          {referenceData.roles.map(r => (
                            <th key={r.role_id} style={{ padding: '14px', textAlign: 'center', width: '150px' }}>
                              <div>{r.role_name}</div>
                              <div style={{ fontSize: '0.75rem', fontWeight: 'normal', opacity: 0.8 }}>ID: {r.role_id}</div>
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {referenceData.menus.map(menu => {
                          const isParent = !menu.parent_id;
                          return (
                            <tr key={menu.menu_id} style={{ borderBottom: '1px solid #e2e8f0', background: isParent ? '#f8fafc' : 'white' }}>
                              <td style={{ padding: '12px 16px', fontWeight: isParent ? '600' : 'normal', paddingLeft: isParent ? '16px' : '36px', color: isParent ? '#0f172a' : '#334155' }}>
                                <span style={{ marginRight: '8px' }}>{menu.icon || '📌'}</span>
                                {menu.title}
                              </td>
                              <td style={{ padding: '12px', fontSize: '0.85rem', color: '#64748B' }}>
                                <code>{menu.path}</code>
                              </td>
                              {referenceData.roles.map(role => {
                                const isGranted = referenceData.role_menus.some(
                                  rm => String(rm.role_id) === String(role.role_id) && String(rm.menu_id) === String(menu.menu_id)
                                );
                                return (
                                  <td key={role.role_id} style={{ padding: '12px', textAlign: 'center' }}>
                                    <input 
                                      type="checkbox"
                                      checked={isGranted}
                                      onChange={() => toggleRoleMenu(role.role_id, menu.menu_id, isGranted)}
                                      style={{ width: '20px', height: '20px', cursor: 'pointer', accentColor: 'var(--primary-blue)' }}
                                    />
                                  </td>
                                );
                              })}
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '24px' }}>
                {/* Form Input Dynamic */}
                <div style={{ background: '#f8fafc', padding: '16px', borderRadius: '8px', border: '1px solid #cbd5e1' }}>
                  <h3>Tambah / Edit {masterTab.replace('-', ' ')}</h3>
                  <form onSubmit={saveMasterData} style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginTop: '16px' }}>
                    
                    {(() => {
                      let fields = [];
                      if (masterTab === 'departments') fields = [{k:'deptno', l:'Dept No (Number)', type: 'number'}, {k:'dept_name', l:'Dept Name'}];
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
                        {k:'employee_id', l:'Employee', type:'select', options: referenceData.employees.map(d => ({val: d.employee_id, label: `${d.employee_id} - ${d.name}`}))}, 
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
                        {k:'menu_id', l:'Menu', type:'select', options: referenceData.menus.map(d => ({val: d.menu_id, label: `${d.title} (${d.path})`}))}
                      ];
                      else if (masterTab === 'parameters') fields = [
                        {k:'key_name', l:'Key Name'},
                        {k:'key_value', l:'Key Value'},
                        {k:'description', l:'Description'}
                      ];
                      
                      return fields.map(f => (
                        <div key={f.k} style={{ display: 'flex', flexDirection: f.type === 'checkbox' ? 'row' : 'column', alignItems: f.type === 'checkbox' ? 'center' : 'flex-start', gap: f.type === 'checkbox' ? '8px' : '4px' }}>
                          <label style={{ fontSize: '0.9rem', fontWeight: 500, color: '#334155' }}>{f.l}</label>
                          {f.k === 'employee_id' && masterTab === 'members' ? (
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', width: '100%', backgroundColor: '#f8fafc', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1' }}>
                              {/* Field Search Employee dengan Klik tombol Cari / Press Enter */}
                              <div style={{ display: 'flex', gap: '6px' }}>
                                <input 
                                  type="text"
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
                                  onClick={() => {
                                    setEmpSelectPage(1);
                                    fetchPaginatedEmployeesForSelect(empSelectSearchQuery, 1);
                                  }}
                                  style={{ padding: '8px 12px', background: '#2563eb', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.85rem', fontWeight: 600 }}
                                >
                                  Cari
                                </button>
                              </div>

                              {/* Dropdown mengikuti Pagination (mengambil limit dari global_parameters: DEFAULT_PAGE_SIZE / PAGINATION_LIMIT) */}
                              <select 
                                required
                                value={masterForm.employee_id || ''}
                                onChange={e => setMasterForm({...masterForm, employee_id: parseInt(e.target.value) || ''})}
                                style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', backgroundColor: 'white', fontWeight: 500 }}
                              >
                                <option value="">-- Pilih Employee ({empSelectTotalRecords} data ditemukan) --</option>
                                {masterForm.employee_id && !empSelectList.some(e => String(e.employee_id) === String(masterForm.employee_id)) && (
                                  <option value={masterForm.employee_id}>
                                    Selected: Employee ID #{masterForm.employee_id}
                                  </option>
                                )}
                                {empSelectList.map(emp => (
                                  <option key={emp.employee_id} value={emp.employee_id}>
                                    {emp.employee_id} - {emp.name} ({emp.employee_id})
                                  </option>
                                ))}
                              </select>

                              {/* User bisa klik Next Page / Prev Page */}
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
                              value={masterForm[f.k] || ''}
                              onChange={e => setMasterForm({...masterForm, [f.k]: f.k === 'employee_id' ? parseInt(e.target.value) : e.target.value})}
                              style={{ width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)', backgroundColor: 'white' }}
                            >
                              <option value="">-- Pilih {f.l} --</option>
                              {f.options && f.options.map(opt => (
                                <option key={opt.val} value={opt.val}>{opt.label} ({opt.val})</option>
                              ))}
                            </select>
                          ) : (
                            <input 
                              type={f.type || 'text'}
                              required={f.type !== 'checkbox'}
                              checked={f.type === 'checkbox' ? (masterForm[f.k] || false) : undefined}
                              value={f.type !== 'checkbox' ? (masterForm[f.k] || '') : undefined}
                              onChange={e => setMasterForm({
                                ...masterForm, 
                                [f.k]: f.type === 'checkbox' ? e.target.checked : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value)
                              })}
                              style={f.type !== 'checkbox' ? { width: '100%', padding: '10px', borderRadius: '4px', border: '1px solid var(--border-color)' } : { width: '20px', height: '20px' }}
                            />
                          )}
                        </div>
                      ));
                    })()}
                    
                    <button type="submit" style={{ padding: '10px', background: 'var(--success-green)', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, marginTop: '8px' }}>Simpan Data</button>
                    {Object.keys(masterForm).length > 0 && (
                      <button type="button" onClick={() => setMasterForm({})} style={{ padding: '10px', background: '#e2e8f0', color: '#475569', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Batal / Form Baru</button>
                    )}
                  </form>
                </div>

                {/* Table View Dynamic with Search Filter */}
                <div style={{ overflowX: 'auto', marginTop: '16px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                    <div style={{ display: 'flex', gap: '8px', flex: 1, maxWidth: '450px' }}>
                      <input 
                        type="text" 
                        placeholder="🔍 Cari data di seluruh database (ID / Nama / Parameter)..."
                        value={masterSearchQuery}
                        onChange={e => {
                          const newQ = e.target.value;
                          setMasterSearchQuery(newQ);
                          setCurrentPage(1);
                          fetchMasterData(masterTab, newQ, 1);
                        }}
                        onKeyDown={e => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            setCurrentPage(1);
                            fetchMasterData(masterTab, masterSearchQuery, 1);
                          }
                        }}
                        style={{ padding: '8px 14px', borderRadius: '6px', border: '1px solid #cbd5e1', flex: 1, fontSize: '0.9rem' }}
                      />
                      <button
                        type="button"
                        onClick={() => {
                          setCurrentPage(1);
                          fetchMasterData(masterTab, masterSearchQuery, 1);
                        }}
                        style={{ padding: '8px 14px', background: '#2563eb', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 600, fontSize: '0.85rem' }}
                      >
                        Cari
                      </button>
                    </div>
                    <span style={{ fontSize: '0.85rem', color: '#64748B' }}>
                      Pencarian: <strong>{masterDataList.length}</strong> ditampilkan dari <strong>{masterTotalRecords}</strong> total data
                    </span>
                  </div>

                  <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '600px' }}>
                    <thead>
                      <tr style={{ background: 'var(--primary-blue)', color: 'white' }}>
                        {(() => {
                          let keys = [];
                          if (masterDataList.length > 0) {
                            keys = Object.keys(masterDataList[0]).filter(k => k !== 'CreatedAt' && k !== 'UpdatedAt' && k !== 'DeletedAt' && k !== 'CreatedUser');
                          }
                          return keys.map(k => <th key={k} style={{ padding: '8px 12px', textAlign: 'left', textTransform: 'capitalize', fontSize: '0.875rem' }}>{k.replace('_', ' ')}</th>);
                        })()}
                        <th style={{ padding: '8px 12px', textAlign: 'center', width: '120px', fontSize: '0.875rem' }}>Aksi</th>
                      </tr>
                    </thead>
                    <tbody>
                      {masterDataList.map((row, idx) => {
                        const pkField = Object.keys(row).find(k => k.includes('id') || k.includes('no') || k.includes('code'));
                        const pkValue = row[pkField];
                        const keys = Object.keys(row).filter(k => k !== 'CreatedAt' && k !== 'UpdatedAt' && k !== 'DeletedAt' && k !== 'CreatedUser');
                        
                        return (
                          <tr key={idx} style={{ borderBottom: '1px solid #e2e8f0' }}>
                            {keys.map(k => (
                              <td key={k} style={{ padding: '6px 12px', color: '#334155', fontSize: '0.875rem' }}>
                                {typeof row[k] === 'boolean' 
                                  ? (row[k] ? 'Ya' : 'Tidak') 
                                  : (typeof row[k] === 'object' && row[k] !== null 
                                      ? JSON.stringify(row[k]) 
                                      : String(row[k] ?? ''))}
                              </td>
                            ))}
                            <td style={{ padding: '6px 12px', textAlign: 'center' }}>
                              <button onClick={() => setMasterForm(row)} style={{ padding: '3px 8px', background: '#e0f2fe', color: '#0369a1', border: 'none', borderRadius: '4px', cursor: 'pointer', marginRight: '4px', fontSize: '0.8rem' }}>Edit</button>
                              <button onClick={() => deleteMasterData(pkField, pkValue)} style={{ padding: '3px 8px', background: '#fee2e2', color: '#b91c1c', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem' }}>Hapus</button>
                            </td>
                          </tr>
                        );
                      })}
                      {masterDataList.length === 0 && (
                        <tr><td colSpan="100%" style={{ padding: '24px', textAlign: 'center', color: '#94a3b8' }}>Belum ada data ditemukan di database.</td></tr>
                      )}
                    </tbody>
                  </table>

                  {/* Navigasi Tombol Next & Previous selalu tampil */}
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '16px', paddingTop: '12px', borderTop: '1px solid #cbd5e1' }}>
                    <span style={{ fontSize: '0.85rem', color: '#64748B' }}>
                      Halaman <strong>{currentPage}</strong> dari <strong>{masterTotalPages}</strong> (Total <strong>{masterTotalRecords}</strong> data)
                    </span>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button 
                        disabled={currentPage <= 1}
                        onClick={() => {
                          const prevPage = Math.max(currentPage - 1, 1);
                          setCurrentPage(prevPage);
                          fetchMasterData(masterTab, masterSearchQuery, prevPage);
                        }}
                        style={{ padding: '6px 14px', background: currentPage <= 1 ? '#e2e8f0' : '#0369a1', color: currentPage <= 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: currentPage <= 1 ? 'not-allowed' : 'pointer', fontWeight: 600 }}
                      >
                        ◀ Sebelumnya
                      </button>
                      <button 
                        disabled={currentPage >= masterTotalPages}
                        onClick={() => {
                          const nextPage = Math.min(currentPage + 1, masterTotalPages);
                          setCurrentPage(nextPage);
                          fetchMasterData(masterTab, masterSearchQuery, nextPage);
                        }}
                        style={{ padding: '6px 14px', background: currentPage >= masterTotalPages ? '#e2e8f0' : '#0369a1', color: currentPage >= masterTotalPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: currentPage >= masterTotalPages ? 'not-allowed' : 'pointer', fontWeight: 600 }}
                      >
                        Selanjutnya ▶
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'manual-repayment' && (() => {
          const paymentSourcesParam = getParamVal('PAYMENT_SOURCES', 'TRANSFER_BANK:Transfer Bank BCA/Mandiri Kopkara,POTONG_PESANGON:Potong Langsung Uang Pesangon (HRD Adira),KOMPENSASI_SIMPANAN:Kompensasi Offset Simpanan Koperasi');
          const paymentSourceOptions = paymentSourcesParam.split(',').map(item => {
            const parts = item.split(':');
            return { code: (parts[0] || '').trim(), label: (parts[1] || parts[0] || '').trim() };
          }).filter(o => o.code.length > 0);

          const memberOptionsList = allMembers.length > 0 ? allMembers : (() => {
            const map = {};
            (payrollSchedules || []).forEach(s => {
              if (s.member_no && !map[s.member_no]) {
                map[s.member_no] = { member_no: s.member_no, employee_id: s.member_no, name: s.employee_name || `Anggota #${s.member_no}` };
              }
            });
            return Object.values(map);
          })();

          const pageSize = parseInt(getParamVal('DEFAULT_PAGE_SIZE', '10')) || 10;

          // Filter Anggota (LIKE Search on member_no, employee_id, or name) & Pagination
          const filteredMembers = memberOptionsList.filter(m => 
            String(m.member_no).toLowerCase().includes(memberSearchQuery.toLowerCase()) ||
            String(m.employee_id || '').toLowerCase().includes(memberSearchQuery.toLowerCase()) ||
            String(m.name || '').toLowerCase().includes(memberSearchQuery.toLowerCase())
          );
          const totalMemberPages = Math.ceil(filteredMembers.length / pageSize) || 1;
          const paginatedMembers = filteredMembers.slice((memberPage - 1) * pageSize, memberPage * pageSize);

          // Filter Pinjaman (Active/Outstanding loans matching selected member, excluding PAID/CLOSED) & Pagination
          const activeSchedules = (payrollSchedules || []).filter(s => s.status !== 'PAID' && s.status !== 'CLOSED');
          const baseLoans = selectedMemberFilter 
            ? activeSchedules.filter(s => String(s.member_no) === String(selectedMemberFilter))
            : activeSchedules;

          const filteredLoans = baseLoans.filter(s =>
            String(s.loan_no).toLowerCase().includes(loanSearchQuery.toLowerCase()) ||
            String(s.employee_name).toLowerCase().includes(loanSearchQuery.toLowerCase()) ||
            String(s.member_no).toLowerCase().includes(loanSearchQuery.toLowerCase()) ||
            String(s.period).toLowerCase().includes(loanSearchQuery.toLowerCase())
          );
          const totalLoanPages = Math.ceil(filteredLoans.length / pageSize) || 1;
          const paginatedLoans = filteredLoans.slice((loanPage - 1) * pageSize, loanPage * pageSize);

          const currentYearNum = new Date().getFullYear();
          const yearOptions = [currentYearNum - 3, currentYearNum - 2, currentYearNum - 1, currentYearNum, currentYearNum + 1, currentYearNum + 2, currentYearNum + 3];
          const monthOptions = [
            { code: '01', name: '01 - Januari' },
            { code: '02', name: '02 - Februari' },
            { code: '03', name: '03 - Maret' },
            { code: '04', name: '04 - April' },
            { code: '05', name: '05 - Mei' },
            { code: '06', name: '06 - Juni' },
            { code: '07', name: '07 - Juli' },
            { code: '08', name: '08 - Agustus' },
            { code: '09', name: '09 - September' },
            { code: '10', name: '10 - Oktober' },
            { code: '11', name: '11 - November' },
            { code: '12', name: '12 - Desember' },
          ];

          return (
          <div>
            <h2 style={{ marginBottom: '20px', color: '#0f172a', display: 'flex', alignItems: 'center', gap: '8px' }}>
              💳 Pelunasan Manual & Pelunasan Dipercepat Karyawan Resign
            </h2>

            {/* Centered Large Form Pelunasan Manual */}
            <div style={{ maxWidth: '900px', margin: '0 auto' }}>
              <div className="card" style={{ background: '#ffffff', padding: '28px', borderRadius: '12px', border: '1px solid #cbd5e1', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}>
                <h3 style={{ margin: 0, marginBottom: '20px', fontSize: '1.2rem', color: '#1e293b', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  📝 Form Pelunasan Manual
                </h3>

                <form onSubmit={async (e) => {
                  e.preventDefault();
                  if (!manualForm.loan_no || !manualForm.nominal) {
                    alert("Silakan masukan Nomor Pinjaman dan Nominal Pembayaran.");
                    return;
                  }
                  try {
                    const activeUserName = currentUser ? String(currentUser.employee_id || '10101') : '10101';
                    const res = await axios.post(`${API_BASE_URL}/api/payroll/manual-repayment`, {
                      ...manualForm,
                      period: `${manualYear}-${manualMonth}`,
                      loan_no: parseInt(manualForm.loan_no) || 0,
                      member_no: parseInt(manualForm.member_no) || 0,
                      nominal: parseFloat(manualForm.nominal) || 0,
                      updated_user: activeUserName
                    });
                    
                    // Build Kuitansi Data
                    const selectedMemberObj = paginatedMemberList.find(m => String(m.member_no) === String(manualForm.member_no)) ||
                                            allMembers.find(m => String(m.member_no) === String(manualForm.member_no));
                    const selectedLoanObj = (payrollSchedules || []).find(s => String(s.loan_no) === String(manualForm.loan_no));
                    const paymentLabel = paymentSourceOptions.find(p => p.code === manualForm.payment_type)?.label || manualForm.payment_type;
                    const todayStr = new Date().toISOString().slice(0, 10).replace(/-/g, '');
                    const kwtNo = `KWT/LMS/${todayStr}/${manualForm.loan_no || Math.floor(1000 + Math.random()*9000)}`;

                    setReceiptData({
                      kwtNo: kwtNo,
                      date: new Date().toLocaleString('id-ID'),
                      loanNo: manualForm.loan_no,
                      memberNo: manualForm.member_no,
                      memberName: selectedMemberObj?.name || selectedLoanObj?.employee_name || `Anggota #${manualForm.member_no}`,
                      paymentType: manualForm.payment_type,
                      paymentTypeLabel: paymentLabel,
                      nominal: parseFloat(manualForm.nominal) || 0,
                      referenceNo: manualForm.reference_no || '-',
                      notes: manualForm.notes || '-',
                      isFullSettlement: manualForm.is_full_settlement,
                      createdUser: activeUserName
                    });
                    setReceiptModalOpen(true);

                    setManualForm({ loan_no: '', member_no: '', period: `${manualYear}-${manualMonth}`, payment_type: paymentSourceOptions[0]?.code || 'TRANSFER_BANK', nominal: '', reference_no: '', notes: '', is_full_settlement: false });
                    if (selectedMemberFilter) {
                      fetchPayrollSchedules(selectedMemberFilter);
                    }
                  } catch (err) {
                    alert("❌ Gagal memproses pelunasan manual: " + (err.response?.data?.error || err.message));
                  }
                }} style={{ display: 'flex', flexDirection: 'column', gap: '18px' }}>
                  
                  {/* Dropdown 1: Input Search & Filter Button + Select Anggota */}
                  <div style={{ background: '#f8fafc', padding: '16px', borderRadius: '8px', border: '1px solid #cbd5e1' }}>
                    <label style={{ fontSize: '0.9rem', fontWeight: 700, display: 'block', marginBottom: '8px', color: '#0f172a' }}>
                      1. Pilih Anggota / Karyawan *
                    </label>
                    <div style={{ display: 'grid', gridTemplateColumns: '2fr auto 3fr', gap: '8px', alignItems: 'center' }}>
                      <input 
                        type="text"
                        value={memberSearchQuery}
                        onChange={(e) => handleMemberSearchChange(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            fetchPaginatedMembers(memberSearchQuery, 1);
                          }
                        }}
                        placeholder="Ketik NIK / Member No / Nama..."
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #94a3b8', fontSize: '0.9rem' }}
                      />
                      <button
                        type="button"
                        onClick={() => fetchPaginatedMembers(memberSearchQuery, 1)}
                        style={{ padding: '10px 16px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.9rem' }}
                      >
                        🔍 Cari
                      </button>
                      <select 
                        value={selectedMemberFilter}
                        onChange={(e) => {
                          const mNo = e.target.value;
                          setSelectedMemberFilter(mNo);
                          if (mNo) {
                            fetchPayrollSchedules(mNo);
                          }
                          const firstLoan = (payrollSchedules || []).find(s => String(s.member_no) === String(mNo) && s.status === 'PARTIAL') ||
                                            (payrollSchedules || []).find(s => String(s.member_no) === String(mNo) && s.status !== 'PAID' && s.status !== 'CLOSED');
                          if (firstLoan) {
                            const tot = parseFloat(firstLoan.total_installment) || 0;
                            const paid = parseFloat(firstLoan.amount_paid) || 0;
                            const rem = (tot - paid) > 0 ? (tot - paid) : tot;
                            setManualForm({
                              ...manualForm,
                              loan_no: firstLoan.loan_no,
                              member_no: firstLoan.member_no,
                              nominal: rem,
                              period: `${manualYear}-${manualMonth}`
                            });
                          } else {
                            setManualForm({ ...manualForm, loan_no: '', member_no: mNo, period: `${manualYear}-${manualMonth}` });
                          }
                        }}
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                      >
                        <option value="">-- Semua Anggota ({memberTotalRecords} Ditemukan) --</option>
                        {paginatedMemberList.map(m => (
                          <option key={m.member_no} value={m.member_no}>
                            {m.name} (ID: {m.employee_id || m.member_no})
                          </option>
                        ))}
                      </select>
                    </div>
                    
                    {/* Pagination Controls Anggota */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '8px', fontSize: '0.8rem', color: '#475569' }}>
                      <span>Hal {memberPage} dari {memberTotalPages} ({memberTotalRecords} Members Ditemukan)</span>
                      <div style={{ display: 'flex', gap: '6px' }}>
                        <button 
                          type="button"
                          disabled={memberPage <= 1}
                          onClick={() => {
                            const newP = Math.max(memberPage - 1, 1);
                            setMemberPage(newP);
                            fetchPaginatedMembers(memberSearchQuery, newP);
                          }}
                          style={{ padding: '4px 10px', fontSize: '0.8rem', background: memberPage <= 1 ? '#e2e8f0' : '#0369a1', color: memberPage <= 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: memberPage <= 1 ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                        >
                          ◀ Prev
                        </button>
                        <button 
                          type="button"
                          disabled={memberPage >= memberTotalPages}
                          onClick={() => {
                            const newP = memberPage + 1;
                            setMemberPage(newP);
                            fetchPaginatedMembers(memberSearchQuery, newP);
                          }}
                          style={{ padding: '4px 10px', fontSize: '0.8rem', background: memberPage >= memberTotalPages ? '#e2e8f0' : '#0369a1', color: memberPage >= memberTotalPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: memberPage >= memberTotalPages ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                        >
                          Next ▶
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Dropdown 2: Input Search & Select Pinjaman Aktif */}
                  <div style={{ background: '#f8fafc', padding: '16px', borderRadius: '8px', border: '1px solid #cbd5e1' }}>
                    <label style={{ fontSize: '0.9rem', fontWeight: 700, display: 'block', marginBottom: '8px', color: '#0f172a' }}>
                      2. Pilih Pinjaman Aktif *
                    </label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '8px', alignItems: 'center' }}>
                      <input 
                        type="text"
                        value={loanSearchQuery}
                        onChange={(e) => {
                          setLoanSearchQuery(e.target.value);
                          setLoanPage(1);
                        }}
                        placeholder="🔍 Filter No. Pinjaman / Periode..."
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #94a3b8', fontSize: '0.9rem' }}
                      />
                      <select 
                        value={manualForm.loan_no} 
                        onChange={(e) => {
                          const lNo = e.target.value;
                          const sel = payrollSchedules.find(s => String(s.loan_no) === String(lNo));
                          let remNominal = '';
                          if (sel) {
                            const tot = parseFloat(sel.total_installment) || 0;
                            const paid = parseFloat(sel.amount_paid) || 0;
                            remNominal = (tot - paid) > 0 ? (tot - paid) : tot;
                          }
                          setManualForm({
                            ...manualForm,
                            loan_no: lNo,
                            member_no: sel ? sel.member_no : selectedMemberFilter,
                            nominal: remNominal,
                            period: `${manualYear}-${manualMonth}`
                          });
                          if (sel && sel.member_no) {
                            setSelectedMemberFilter(String(sel.member_no));
                          }
                        }}
                        required
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                      >
                        <option value="">-- Pilih Pinjaman Aktif ({filteredLoans.length} Ditemukan) --</option>
                        {paginatedLoans.map(s => {
                          const tot = parseFloat(s.total_installment) || 0;
                          const paid = parseFloat(s.amount_paid) || 0;
                          const rem = (tot - paid) > 0 ? (tot - paid) : tot;
                          const hasPartial = paid > 0 && paid < tot;
                          return (
                            <option key={s.id} value={s.loan_no}>
                              Pinjaman #{s.loan_no} - {s.employee_name} ({s.member_no}) - Rp {Math.round(rem).toLocaleString('id-ID')}{hasPartial ? ' (Sisa Hutang)' : ''} ({s.period})
                            </option>
                          );
                        })}
                      </select>
                    </div>

                    {/* Pagination Controls Pinjaman */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '8px', fontSize: '0.8rem', color: '#475569' }}>
                      <span>Hal {loanPage} dari {totalLoanPages} ({filteredLoans.length} Pinjaman)</span>
                      <div style={{ display: 'flex', gap: '6px' }}>
                        <button 
                          type="button"
                          disabled={loanPage === 1}
                          onClick={() => setLoanPage(prev => Math.max(prev - 1, 1))}
                          style={{ padding: '4px 10px', fontSize: '0.8rem', background: loanPage === 1 ? '#e2e8f0' : '#0369a1', color: loanPage === 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: loanPage === 1 ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                        >
                          ◀ Prev
                        </button>
                        <button 
                          type="button"
                          disabled={loanPage >= totalLoanPages}
                          onClick={() => setLoanPage(prev => Math.min(prev + 1, totalLoanPages))}
                          style={{ padding: '4px 10px', fontSize: '0.8rem', background: loanPage >= totalLoanPages ? '#e2e8f0' : '#0369a1', color: loanPage >= totalLoanPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: loanPage >= totalLoanPages ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                        >
                          Next ▶
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Dropdown 3: Metode / Sumber Pelunasan */}
                  <div>
                    <label style={{ fontSize: '0.9rem', fontWeight: 600, display: 'block', marginBottom: '6px' }}>3. Metode / Sumber Pelunasan (Dinamis Master Parameter) *</label>
                    <select 
                      value={manualForm.payment_type}
                      onChange={(e) => setManualForm({ ...manualForm, payment_type: e.target.value })}
                      style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    >
                      {paymentSourceOptions.map(opt => (
                        <option key={opt.code} value={opt.code}>
                          {opt.label} ({opt.code})
                        </option>
                      ))}
                    </select>
                    <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: '4px' }}>
                      💡 <em>Dapat ditambah/diubah di menu Pengaturan Parameter (key: <code>PAYMENT_SOURCES</code>)</em>
                    </div>
                  </div>

                  {/* Dropdown 4: Periode Tagihan */}
                  <div>
                    <label style={{ fontSize: '0.9rem', fontWeight: 600, display: 'block', marginBottom: '6px' }}>4. Periode Tagihan (Bulan & Tahun) *</label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                      <select 
                        value={manualMonth}
                        onChange={(e) => {
                          const m = e.target.value;
                          setManualMonth(m);
                          setManualForm(prev => ({ ...prev, period: `${manualYear}-${m}` }));
                        }}
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                      >
                        {monthOptions.map(mo => (
                          <option key={mo.code} value={mo.code}>{mo.name}</option>
                        ))}
                      </select>

                      <select 
                        value={manualYear}
                        onChange={(e) => {
                          const y = e.target.value;
                          setManualYear(y);
                          setManualForm(prev => ({ ...prev, period: `${y}-${manualMonth}` }));
                        }}
                        style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                      >
                        {yearOptions.map(yr => (
                          <option key={yr} value={yr}>Tahun {yr}</option>
                        ))}
                      </select>
                    </div>
                    <div style={{ fontSize: '0.8rem', color: '#0369a1', marginTop: '6px', fontWeight: 500 }}>
                      📅 Periode Terpilih: <strong>{manualYear}-{manualMonth}</strong>
                    </div>
                  </div>

                  <div>
                    <label style={{ fontSize: '0.9rem', fontWeight: 600, display: 'block', marginBottom: '6px' }}>Nominal Pembayaran (Rp) *</label>
                    <input 
                      type="text" 
                      value={manualForm.nominal !== '' && manualForm.nominal !== null && manualForm.nominal !== undefined ? (Math.round(parseFloat(String(manualForm.nominal).replace(/[^0-9]/g, '')) || 0) > 0 ? Math.round(parseFloat(String(manualForm.nominal).replace(/[^0-9]/g, '')) || 0).toLocaleString('id-ID') : '') : ''} 
                      onChange={(e) => {
                        const rawDigits = e.target.value.replace(/[^0-9]/g, '');
                        setManualForm({ ...manualForm, nominal: rawDigits });
                      }}
                      placeholder="0" 
                      required 
                      style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontWeight: 700, fontSize: '1.05rem', color: '#0f172a' }}
                    />
                  </div>

                  <div>
                    <label style={{ fontSize: '0.9rem', fontWeight: 600, display: 'block', marginBottom: '6px' }}>No. Bukti Transfer / Ref Bank</label>
                    <input 
                      type="text" 
                      value={manualForm.reference_no} 
                      onChange={(e) => setManualForm({ ...manualForm, reference_no: e.target.value })}
                      placeholder="Contoh: TRF-BCA-987123" 
                      style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    />
                  </div>

                  <div>
                    <label style={{ fontSize: '0.9rem', fontWeight: 600, display: 'block', marginBottom: '6px' }}>Catatan Pelunasan / Karyawan Resign</label>
                    <input 
                      type="text" 
                      value={manualForm.notes} 
                      onChange={(e) => setManualForm({ ...manualForm, notes: e.target.value })}
                      placeholder="Catatan pelunasan mandiri / pesangon" 
                      style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    />
                  </div>

                  {/* Clarified Box Pink: Early Full Settlement */}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', padding: '12px 16px', background: '#fef2f2', borderRadius: '8px', border: '1px solid #fca5a5' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                      <input 
                        type="checkbox" 
                        id="isFull" 
                        checked={manualForm.is_full_settlement}
                        onChange={(e) => setManualForm({ ...manualForm, is_full_settlement: e.target.checked })}
                        style={{ width: '20px', height: '20px', cursor: 'pointer' }}
                      />
                      <label htmlFor="isFull" style={{ fontSize: '0.95rem', fontWeight: 700, color: '#991b1b', cursor: 'pointer' }}>
                        🔥 Pelunasan Lunas Sekaligus (Lunas Total / Early Full Settlement)
                      </label>
                    </div>
                    <div style={{ fontSize: '0.78rem', color: '#7f1d1d', marginLeft: '30px', lineHeight: '1.4' }}>
                      💡 <em>Beri tanda centang jika pembayaran ini melunasi seluruh sisa hutang pinjaman secara permanen (misal: Karyawan Resign / Pelunasan Dipercepat). Status pinjaman akan otomatis diubah menjadi <strong>CLOSED</strong>.</em>
                    </div>
                  </div>

                  <button 
                    type="submit" 
                    style={{ marginTop: '12px', padding: '14px 24px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '8px', fontWeight: 'bold', cursor: 'pointer', fontSize: '1.05rem', boxShadow: '0 4px 6px -1px rgba(2, 132, 199, 0.3)' }}
                  >
                    💳 Proses
                  </button>
                </form>
              </div>
            </div>

            {/* Modal Kuitansi Bukti Pelunasan Manual */}
            {receiptModalOpen && receiptData && (
              <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1050, backdropFilter: 'blur(4px)' }}>
                <div style={{ backgroundColor: 'white', borderRadius: '12px', padding: '32px', width: '90%', maxWidth: '650px', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)', border: '2px solid #0284c7' }}>
                  
                  <div style={{ textAlign: 'center', borderBottom: '2px solid #e2e8f0', paddingBottom: '16px', marginBottom: '20px' }}>
                    <h3 style={{ margin: 0, color: '#0369a1', fontSize: '1.1rem', textTransform: 'uppercase', letterSpacing: '1px' }}>KOPERASI KARYAWAN (KOPKARA) - LMS SYSTEM</h3>
                    <h2 style={{ margin: '6px 0 0 0', color: '#0f172a', fontSize: '1.4rem', fontWeight: 800 }}>🧾 KUITANSI BUKTI PELUNASAN PINJAMAN</h2>
                    <div style={{ fontSize: '0.85rem', color: '#64748b', marginTop: '4px', fontFamily: 'monospace' }}>No: {receiptData.kwtNo} | Tanggal: {receiptData.date}</div>
                  </div>

                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem', marginBottom: '24px' }}>
                    <tbody>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b', width: '38%' }}>No. Pinjaman</td>
                        <td style={{ padding: '8px 0', fontWeight: 'bold', color: '#0f172a' }}>#{receiptData.loanNo}</td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>NIK / No. Anggota</td>
                        <td style={{ padding: '8px 0', fontWeight: 'bold', color: '#0f172a' }}>{receiptData.memberNo}</td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>Nama Anggota / Pemohon</td>
                        <td style={{ padding: '8px 0', fontWeight: 'bold', color: '#0f172a' }}>{receiptData.memberName}</td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>Metode / Sumber Dana</td>
                        <td style={{ padding: '8px 0', fontWeight: 600, color: '#2563eb' }}>{receiptData.paymentTypeLabel}</td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>No. Referensi / Bukti</td>
                        <td style={{ padding: '8px 0', fontFamily: 'monospace', color: '#475569' }}>{receiptData.referenceNo}</td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>Jenis Pelunasan</td>
                        <td style={{ padding: '8px 0', fontWeight: 'bold', color: receiptData.isFullSettlement ? '#047857' : '#0284c7' }}>
                          {receiptData.isFullSettlement ? '🔥 Pelunasan Lunas Sekaligus (CLOSED)' : 'Angsuran / Pelunasan Partial'}
                        </td>
                      </tr>
                      <tr style={{ borderBottom: '1px solid #f1f5f9' }}>
                        <td style={{ padding: '8px 0', color: '#64748b' }}>Catatan Petugas</td>
                        <td style={{ padding: '8px 0', color: '#475569' }}>{receiptData.notes}</td>
                      </tr>
                      <tr style={{ background: '#f0fdf4', borderTop: '2px solid #16a34a' }}>
                        <td style={{ padding: '12px 10px', fontSize: '1rem', fontWeight: 800, color: '#166534' }}>JUMLAH DIBAYAR</td>
                        <td style={{ padding: '12px 10px', fontSize: '1.25rem', fontWeight: 800, color: '#15803d', textAlign: 'left' }}>
                          Rp {receiptData.nominal.toLocaleString('id-ID')}
                        </td>
                      </tr>
                    </tbody>
                  </table>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: '16px', borderTop: '1px solid #cbd5e1' }}>
                    <div style={{ fontSize: '0.8rem', color: '#64748b' }}>
                      Petugas Operasional: <strong>{receiptData.createdUser}</strong>
                    </div>
                    <div style={{ display: 'flex', gap: '12px' }}>
                      <button
                        onClick={() => {
                          const originalTitle = document.title;
                          document.title = "Kopkara LMS - Pelunasan Manual";
                          window.print();
                          setTimeout(() => {
                            document.title = originalTitle;
                          }, 1000);
                        }}
                        style={{ padding: '8px 18px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '6px' }}
                      >
                        🖨️ Cetak Kuitansi
                      </button>
                      <button
                        onClick={() => setReceiptModalOpen(false)}
                        style={{ padding: '8px 18px', background: '#64748b', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.85rem' }}
                      >
                        Tutup
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
          );
        })()}

        {activeTab === 'products' && (() => {
          const canManageProducts = !['anggota'].includes(String(realRoleName || '').toLowerCase());
          return (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <div>
                <h2 style={{ margin: 0, color: '#0f172a', display: 'flex', alignItems: 'center', gap: '8px' }}>
                  📦 Katalog & Pengaturan Produk Pinjaman Kopkara
                </h2>
                <p style={{ margin: '4px 0 0 0', color: '#64748b', fontSize: '0.9rem' }}>
                  Kelola jenis produk pinjaman, aturan suku bunga, max tenor, dan batas plafon peminjaman.
                </p>
              </div>
              {canManageProducts && (
                <button 
                  onClick={() => {
                    setEditingProduct(null);
                    setProductForm({
                      name: '',
                      loan_type: 'FLAT',
                      max_tenor_months: 24,
                      submission_period_start: 1,
                      submission_period_end: 25,
                      max_percentage_salary: 40.0,
                      interest_rate: 1.5,
                      status: 'ACTIVE'
                    });
                    setProductModalOpen(true);
                  }}
                  style={{ padding: '10px 18px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '6px', fontWeight: 'bold', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}
                >
                  ➕ Tambah Produk Baru
                </button>
              )}
            </div>

            {/* Product Summary Statistics Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px', marginBottom: '24px' }}>
              <div className="card" style={{ background: '#eff6ff', border: '1px solid #bfdbfe', padding: '16px' }}>
                <div style={{ fontSize: '0.8rem', color: '#1e40af', fontWeight: 600 }}>Total Produk Pinjaman</div>
                <div style={{ fontSize: '1.6rem', fontWeight: 'bold', color: '#1e3a8a', marginTop: '4px' }}>{products.length} Produk</div>
              </div>
              <div className="card" style={{ background: '#ecfdf5', border: '1px solid #a7f3d0', padding: '16px' }}>
                <div style={{ fontSize: '0.8rem', color: '#065f46', fontWeight: 600 }}>Produk Aktif</div>
                <div style={{ fontSize: '1.6rem', fontWeight: 'bold', color: '#047857', marginTop: '4px' }}>{products.filter(p => p.status === 'ACTIVE').length} Aktif</div>
              </div>
              <div className="card" style={{ background: '#fefce8', border: '1px solid #fef08a', padding: '16px' }}>
                <div style={{ fontSize: '0.8rem', color: '#854d0e', fontWeight: 600 }}>Rata-Rata Suku Bunga</div>
                <div style={{ fontSize: '1.6rem', fontWeight: 'bold', color: '#a16207', marginTop: '4px' }}>
                  {products.length > 0 ? (products.reduce((acc, p) => acc + (p.interest_rate || 0), 0) / products.length).toFixed(1) : 0}% / bln
                </div>
              </div>
              <div className="card" style={{ background: '#f3e8ff', border: '1px solid #e9d5ff', padding: '16px' }}>
                <div style={{ fontSize: '0.8rem', color: '#6b21a8', fontWeight: 600 }}>Tenor Maksimal</div>
                <div style={{ fontSize: '1.6rem', fontWeight: 'bold', color: '#7e22ce', marginTop: '4px' }}>
                  {products.length > 0 ? Math.max(...products.map(p => p.max_tenor_months || 0)) : 0} Bulan
                </div>
              </div>
            </div>

            {/* Filter & Search Bar */}
            <div className="card" style={{ padding: '12px 16px', marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <input 
                type="text" 
                value={productSearchQuery}
                onChange={(e) => setProductSearchQuery(e.target.value)}
                placeholder="🔍 Cari nama produk / tipe pinjaman (FLAT / SLIDING / ANUITAS)..."
                style={{ width: '350px', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1', fontSize: '0.85rem' }}
              />
              <span style={{ fontSize: '0.85rem', color: '#64748b' }}>
                Menampilkan <strong>{products.filter(p => (p.name || '').toLowerCase().includes(productSearchQuery.toLowerCase()) || (p.loan_type || '').toLowerCase().includes(productSearchQuery.toLowerCase())).length}</strong> dari {products.length} produk
              </span>
            </div>

            {/* Products Table */}
            <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
                <thead>
                  <tr style={{ background: '#f8fafc', borderBottom: '2px solid #e2e8f0', color: '#475569' }}>
                    <th style={{ padding: '12px 16px', textAlign: 'left' }}>ID</th>
                    <th style={{ padding: '12px 16px', textAlign: 'left' }}>Nama Produk Pinjaman</th>
                    <th style={{ padding: '12px 16px', textAlign: 'center' }}>Tipe Angsuran</th>
                    <th style={{ padding: '12px 16px', textAlign: 'center' }}>Max Tenor</th>
                    <th style={{ padding: '12px 16px', textAlign: 'center' }}>Periode Pengajuan</th>
                    <th style={{ padding: '12px 16px', textAlign: 'right' }}>Max % Gaji (DSR)</th>
                    <th style={{ padding: '12px 16px', textAlign: 'right' }}>Bunga / Bulan</th>
                    <th style={{ padding: '12px 16px', textAlign: 'center' }}>Status</th>
                    {canManageProducts && <th style={{ padding: '12px 16px', textAlign: 'center' }}>Aksi</th>}
                  </tr>
                </thead>
                <tbody>
                  {products.filter(p => (p.name || '').toLowerCase().includes(productSearchQuery.toLowerCase()) || (p.loan_type || '').toLowerCase().includes(productSearchQuery.toLowerCase())).length === 0 ? (
                    <tr><td colSpan="9" style={{ textAlign: 'center', padding: '32px', color: '#94a3b8' }}>Tidak ada data produk pinjaman ditemukan.</td></tr>
                  ) : (
                    products.filter(p => (p.name || '').toLowerCase().includes(productSearchQuery.toLowerCase()) || (p.loan_type || '').toLowerCase().includes(productSearchQuery.toLowerCase())).map(p => (
                      <tr key={p.id} style={{ borderBottom: '1px solid #e2e8f0' }}>
                        <td style={{ padding: '12px 16px', fontWeight: 'bold', color: '#64748b' }}>#{p.id}</td>
                        <td style={{ padding: '12px 16px', fontWeight: 600, color: '#0f172a' }}>
                          <div>{p.name}</div>
                        </td>
                        <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                          <span style={{ padding: '3px 8px', borderRadius: '4px', background: '#e0f2fe', color: '#0369a1', fontWeight: 600, fontSize: '0.75rem' }}>
                            {p.loan_type || 'FLAT'}
                          </span>
                        </td>
                        <td style={{ padding: '12px 16px', textAlign: 'center', fontWeight: 600 }}>{p.max_tenor_months} Bulan</td>
                        <td style={{ padding: '12px 16px', textAlign: 'center', color: '#475569' }}>Tgl {p.submission_period_start} - {p.submission_period_end}</td>
                        <td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: 600, color: '#166534' }}>{p.max_percentage_salary}%</td>
                        <td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: 'bold', color: '#0284c7' }}>{p.interest_rate}%</td>
                        <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                          <span className={p.status === 'ACTIVE' ? 'badge badge-success' : 'badge badge-danger'}>
                            {p.status || 'ACTIVE'}
                          </span>
                        </td>
                        {canManageProducts && (
                          <td style={{ padding: '12px 16px', textAlign: 'center' }}>
                            <div style={{ display: 'flex', gap: '6px', justifyContent: 'center' }}>
                              <button 
                                onClick={() => {
                                  setEditingProduct(p);
                                  setProductForm({
                                    name: p.name || '',
                                    loan_type: p.loan_type || 'FLAT',
                                    max_tenor_months: p.max_tenor_months || 24,
                                    submission_period_start: p.submission_period_start || 1,
                                    submission_period_end: p.submission_period_end || 25,
                                    max_percentage_salary: p.max_percentage_salary || 40.0,
                                    interest_rate: p.interest_rate || 1.5,
                                    status: p.status || 'ACTIVE'
                                  });
                                  setProductModalOpen(true);
                                }}
                                style={{ padding: '4px 10px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}
                              >
                                ✏️ Edit
                              </button>
                              <button 
                                onClick={async () => {
                                  if (window.confirm(`Hapus produk pinjaman "${p.name}"?`)) {
                                    try {
                                      await axios.delete(`${API_BASE_URL}/api/products/${p.id}`);
                                      alert("✅ Produk pinjaman berhasil dihapus!");
                                      fetchProducts();
                                    } catch (err) {
                                      alert("❌ Gagal menghapus produk: " + (err.response?.data?.error || err.message));
                                    }
                                  }
                                }}
                                style={{ padding: '4px 10px', background: '#ef4444', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.75rem', fontWeight: 600 }}
                              >
                                🗑️ Hapus
                              </button>
                            </div>
                          </td>
                        )}
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            {/* Modal Form Tambah / Edit Produk Pinjaman */}
            {productModalOpen && (
              <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
                <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '90%', maxWidth: '600px', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                    <h2 style={{ margin: 0, fontSize: '1.15rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                      📦 {editingProduct ? `Edit Produk Pinjaman #${editingProduct.id}` : 'Tambah Produk Pinjaman Baru'}
                    </h2>
                    <button onClick={() => setProductModalOpen(false)} style={{ background: '#f1f5f9', border: 'none', borderRadius: '50%', width: '32px', height: '32px', cursor: 'pointer', fontWeight: 'bold', color: '#64748b' }}>✕</button>
                  </div>

                  <form onSubmit={async (e) => {
                    e.preventDefault();
                    try {
                      const payload = {
                        ...productForm,
                        id: editingProduct ? editingProduct.id : 0,
                        max_tenor_months: parseInt(productForm.max_tenor_months) || 12,
                        submission_period_start: parseInt(productForm.submission_period_start) || 1,
                        submission_period_end: parseInt(productForm.submission_period_end) || 25,
                        max_percentage_salary: parseFloat(String(productForm.max_percentage_salary).replace(',', '.')) || 40.0,
                        interest_rate: parseFloat(String(productForm.interest_rate).replace(',', '.')) || 1.5
                      };

                      if (editingProduct) {
                        await axios.put(`${API_BASE_URL}/api/products/${editingProduct.id}`, payload);
                        alert("✅ Produk pinjaman berhasil diperbarui!");
                      } else {
                        await axios.post(`${API_BASE_URL}/api/products`, payload);
                        alert("✅ Produk pinjaman baru berhasil ditambahkan!");
                      }
                      setProductModalOpen(false);
                      fetchProducts();
                    } catch (err) {
                      alert("❌ Gagal menyimpan data produk: " + (err.response?.data?.error || err.message));
                    }
                  }} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                    
                    <div>
                      <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Nama Produk Pinjaman *</label>
                      <input 
                        type="text"
                        value={productForm.name}
                        onChange={(e) => setProductForm({ ...productForm, name: e.target.value })}
                        placeholder="Contoh: Pinjaman Multiguna Fluktuatif"
                        required
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                      />
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Tipe Angsuran *</label>
                        <select 
                          value={productForm.loan_type}
                          onChange={(e) => setProductForm({ ...productForm, loan_type: e.target.value })}
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        >
                          <option value="FLAT">FLAT (Pokok & Bunga Tetap)</option>
                          <option value="SLIDING">SLIDING (Bunga Menurun)</option>
                          <option value="ANUITAS">ANUITAS (Angsuran Efektif)</option>
                        </select>
                      </div>

                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Tenor Maksimal (Bulan) *</label>
                        <input 
                          type="number"
                          value={productForm.max_tenor_months}
                          onChange={(e) => setProductForm({ ...productForm, max_tenor_months: e.target.value })}
                          required
                          min="1"
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        />
                      </div>
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Periode Pengajuan (Tgl Awal) *</label>
                        <input 
                          type="number"
                          value={productForm.submission_period_start}
                          onChange={(e) => setProductForm({ ...productForm, submission_period_start: e.target.value })}
                          min="1" max="31" required
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        />
                      </div>

                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Periode Pengajuan (Tgl Akhir) *</label>
                        <input 
                          type="number"
                          value={productForm.submission_period_end}
                          onChange={(e) => setProductForm({ ...productForm, submission_period_end: e.target.value })}
                          min="1" max="31" required
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        />
                      </div>
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Max % Gaji (DSR %) *</label>
                        <input 
                          type="number"
                          step="0.1"
                          value={productForm.max_percentage_salary}
                          onChange={(e) => setProductForm({ ...productForm, max_percentage_salary: e.target.value })}
                          required
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        />
                      </div>

                      <div>
                        <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Suku Bunga (% / Bulan) *</label>
                        <input 
                          type="number"
                          step="0.01"
                          value={productForm.interest_rate}
                          onChange={(e) => setProductForm({ ...productForm, interest_rate: e.target.value })}
                          required
                          style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                        />
                      </div>
                    </div>

                    <div>
                      <label style={{ fontSize: '0.85rem', fontWeight: 600, display: 'block', marginBottom: '4px' }}>Status Produk *</label>
                      <select 
                        value={productForm.status}
                        onChange={(e) => setProductForm({ ...productForm, status: e.target.value })}
                        style={{ width: '100%', padding: '8px 12px', borderRadius: '4px', border: '1px solid #cbd5e1' }}
                      >
                        <option value="ACTIVE">ACTIVE (Aktif Dapat Dipilih)</option>
                        <option value="INACTIVE">INACTIVE (Nonaktifkan sementara)</option>
                      </select>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '12px', paddingTop: '12px', borderTop: '1px solid #e2e8f0' }}>
                      <button type="button" onClick={() => setProductModalOpen(false)} style={{ padding: '8px 16px', background: '#64748b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Batal</button>
                      <button type="submit" style={{ padding: '8px 18px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>💾 Simpan Produk Pinjaman</button>
                    </div>
                  </form>
                </div>
              </div>
            )}
          </div>
          );
        })()}

        {isReportTab() && (() => {
          const pageSize = 10;
          const safeApps = Array.isArray(applications) ? applications : [];

          const curYear = new Date().getFullYear();
          const reportYearOptions = [curYear - 3, curYear - 2, curYear - 1, curYear, curYear + 1, curYear + 2, curYear + 3];
          const reportMonthOptions = [
            { code: '01', name: '01 - Januari' },
            { code: '02', name: '02 - Februari' },
            { code: '03', name: '03 - Maret' },
            { code: '04', name: '04 - April' },
            { code: '05', name: '05 - Mei' },
            { code: '06', name: '06 - Juni' },
            { code: '07', name: '07 - Juli' },
            { code: '08', name: '08 - Agustus' },
            { code: '09', name: '09 - September' },
            { code: '10', name: '10 - Oktober' },
            { code: '11', name: '11 - November' },
            { code: '12', name: '12 - Desember' }
          ];

          const monthNameMap = {
            '01': 'Januari', '02': 'Februari', '03': 'Maret', '04': 'April',
            '05': 'Mei', '06': 'Juni', '07': 'Juli', '08': 'Agustus',
            '09': 'September', '10': 'Oktober', '11': 'November', '12': 'Desember'
          };
          const selectedMonthName = monthNameMap[reportMonthFilter.month] || reportMonthFilter.month;
          const dynamicReportTitle = `Laporan loan periode ${selectedMonthName} ${reportMonthFilter.year}`;

          const filteredReportApps = safeApps.filter(app => {
            if (!reportMemberSearchQuery) return true;
            const q = reportMemberSearchQuery.trim().replace(/^['"]|['"]$/g, '').toLowerCase();
            if (!q) return true;
            return String(app.ApplicationNo || '').toLowerCase().includes(q) ||
                   String(app.MemberNo || '').toLowerCase().includes(q) ||
                   String(app.Status || '').toLowerCase().includes(q);
          });
          const totalReportPages = Math.ceil(filteredReportApps.length / pageSize) || 1;
          const startIndex = (reportPage - 1) * pageSize;
          const paginatedReportApps = filteredReportApps.slice(startIndex, startIndex + pageSize);

          // Metrics Summary Calculation
          const totalCount = safeApps.length;
          const totalRequested = safeApps.reduce((sum, a) => sum + (parseFloat(a.RequestedAmount) || 0), 0);
          const approvedOrDisbursedApps = safeApps.filter(a => a.Status === 'APPROVED' || a.Status === 'DISBURSED');
          const approvedCount = approvedOrDisbursedApps.length;
          const totalApprovedNominal = approvedOrDisbursedApps.reduce((sum, a) => sum + (parseFloat(a.ApprovedAmount || a.RequestedAmount) || 0), 0);
          const rejectedCount = safeApps.filter(a => a.Status === 'REJECTED').length;
          const submittedCount = safeApps.filter(a => a.Status === 'SUBMITTED').length;

          const paramLmsTitle = getParamVal('LMS_Title', 'kopkara.jfif');
          const logoFileName = (paramLmsTitle && (paramLmsTitle.endsWith('.jfif') || paramLmsTitle.endsWith('.png') || paramLmsTitle.endsWith('.jpg') || paramLmsTitle.endsWith('.jpeg') || paramLmsTitle.endsWith('.svg')))
            ? paramLmsTitle
            : 'kopkara.jfif';
          const logoPath = '/' + logoFileName;

          const printNow = new Date();
          const monthNamesIndo = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
          const formattedPrintDateTime = `${printNow.getDate()} ${monthNamesIndo[printNow.getMonth()]} ${printNow.getFullYear()} ${String(printNow.getHours()).padStart(2, '0')}:${String(printNow.getMinutes()).padStart(2, '0')}:${String(printNow.getSeconds()).padStart(2, '0')}`;

          return (
            <div>
              {/* Header Cetak PDF (Tampak hanya saat Cetak / Print) - Logo & H1 Judul Periode Rata Kiri */}
              <div className="print-only" style={{ marginBottom: '20px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingBottom: '10px' }}>
                  {/* Left: Logo & H1 Judul Periode */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                    <img 
                      src={logoPath} 
                      alt="Logo" 
                      style={{ height: '54px', width: 'auto', objectFit: 'contain' }}
                      onError={(e) => {
                        e.target.onerror = null;
                        e.target.src = '/kopkara.jfif';
                      }}
                    />
                    <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 900, color: '#0A2540', lineHeight: 1.2 }}>
                      {dynamicReportTitle}
                    </h1>
                  </div>

                  {/* Right: Tanggal Cetak Dengan Jam */}
                  <div style={{ textAlign: 'right', flexShrink: 0 }}>
                    <div style={{ fontSize: '0.85rem', color: '#64748b', fontWeight: 600 }}>
                      Tanggal Cetak: {formattedPrintDateTime}
                    </div>
                  </div>
                </div>

                {/* Garis Pembatas Tebal */}
                <div style={{ borderBottom: '3.5px solid #0A2540', width: '100%', marginTop: '4px', marginBottom: '16px' }}></div>
              </div>

              {/* Kotak Merah - Disembunyikan Saat Cetak (no-print) */}
              <div className="no-print">
                {/* Header Title & Export Buttons */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
                  <div>
                    <h2 style={{ margin: 0, color: '#0f172a', display: 'flex', alignItems: 'center', gap: '8px' }}>
                      📊 Laporan Pengajuan Pinjaman
                    </h2>
                    <p style={{ margin: '4px 0 0 0', color: '#64748b', fontSize: '0.9rem' }}>
                      Rekapitulasi dan laporan detail status seluruh pengajuan pinjaman anggota Kopkara per periode.
                    </p>
                  </div>
                  <div style={{ display: 'flex', gap: '10px' }}>
                    <button
                      onClick={() => {
                        const originalTitle = document.title;
                        document.title = dynamicReportTitle;
                        window.print();
                        setTimeout(() => {
                          document.title = originalTitle;
                        }, 1000);
                      }}
                      style={{ padding: '9px 16px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.88rem', display: 'flex', alignItems: 'center', gap: '6px' }}
                    >
                      🖨️ Cetak PDF
                    </button>
                    <button
                      onClick={() => {
                        if (!safeApps || safeApps.length === 0) {
                          alert("Tidak ada data laporan untuk diekspor.");
                          return;
                        }
                        const headers = ["No.", "No. Pengajuan", "Tanggal", "NIK / Employee ID", "Nominal Pengajuan (Rp)", "Nominal Disetujui (Rp)", "Tenor (Bln)", "Angsuran/Bln (Rp)", "Status"];
                        const rows = safeApps.map((app, idx) => [
                          idx + 1,
                          app.ApplicationNo,
                          app.SubmissionDate ? new Date(app.SubmissionDate).toLocaleDateString('id-ID') : '',
                          app.MemberNo,
                          app.RequestedAmount || 0,
                          app.ApprovedAmount || app.RequestedAmount || 0,
                          app.Tenor || 0,
                          app.TotalInstallment || 0,
                          app.Status || ''
                        ]);
                        const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
                        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
                        const url = URL.createObjectURL(blob);
                        const link = document.createElement('a');
                        link.href = url;
                        link.setAttribute('download', `${dynamicReportTitle.replace(/\s+/g, '_')}.csv`);
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                      }}
                      style={{ padding: '9px 16px', background: '#10b981', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.88rem', display: 'flex', alignItems: 'center', gap: '6px' }}
                    >
                      📊 Ekspor CSV
                    </button>
                  </div>
                </div>

                {/* Metrics Summary Cards */}
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '16px', marginBottom: '24px' }}>
                  <div style={{ background: '#ffffff', padding: '18px', borderRadius: '10px', border: '1px solid #cbd5e1', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
                    <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Total Pengajuan</div>
                    <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#0f172a', margin: '4px 0' }}>{totalCount} Record</div>
                    <div style={{ fontSize: '0.85rem', color: '#0284c7', fontWeight: 700 }}>Rp {Math.round(totalRequested).toLocaleString('id-ID')}</div>
                  </div>

                  <div style={{ background: '#ffffff', padding: '18px', borderRadius: '10px', border: '1px solid #cbd5e1', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
                    <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Disetujui / Dicairkan</div>
                    <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#059669', margin: '4px 0' }}>{approvedCount} Record</div>
                    <div style={{ fontSize: '0.85rem', color: '#10b981', fontWeight: 700 }}>Rp {Math.round(totalApprovedNominal).toLocaleString('id-ID')}</div>
                  </div>

                  <div style={{ background: '#ffffff', padding: '18px', borderRadius: '10px', border: '1px solid #cbd5e1', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
                    <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Dalam Proses (Submitted)</div>
                    <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#2563eb', margin: '4px 0' }}>{submittedCount} Record</div>
                    <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Menunggu Review HRD</div>
                  </div>

                  <div style={{ background: '#ffffff', padding: '18px', borderRadius: '10px', border: '1px solid #cbd5e1', boxShadow: '0 2px 6px rgba(0,0,0,0.03)' }}>
                    <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600, textTransform: 'uppercase' }}>Ditolak (Rejected)</div>
                    <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#dc2626', margin: '4px 0' }}>{rejectedCount} Record</div>
                    <div style={{ fontSize: '0.8rem', color: '#64748b' }}>Tidak Memenuhi Syarat</div>
                  </div>
                </div>

                {/* Filter Controls Card */}
                <div className="card" style={{ background: '#ffffff', padding: '20px', borderRadius: '10px', border: '1px solid #cbd5e1', marginBottom: '24px' }}>
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '14px', alignItems: 'end' }}>
                    
                    {/* Dropdown Periode */}
                    <div>
                      <label style={{ fontSize: '0.85rem', fontWeight: 700, color: '#334155', display: 'block', marginBottom: '6px' }}>Periode (Tahun & Bulan)</label>
                      <div style={{ display: 'flex', gap: '6px' }}>
                        <select 
                          value={reportMonthFilter.year}
                          onChange={(e) => {
                            const newY = e.target.value;
                            setReportMonthFilter(prev => ({ ...prev, year: newY }));
                            fetchApplications(newY, reportMonthFilter.month);
                          }}
                          style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.88rem', flex: 1 }}
                        >
                          {reportYearOptions.map(y => <option key={y} value={y}>{y}</option>)}
                        </select>

                        <select 
                          value={reportMonthFilter.month}
                          onChange={(e) => {
                            const newM = e.target.value;
                            setReportMonthFilter(prev => ({ ...prev, month: newM }));
                            fetchApplications(reportMonthFilter.year, newM);
                          }}
                          style={{ padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.88rem', flex: 1 }}
                        >
                          {reportMonthOptions.map(m => <option key={m.code} value={m.code}>{m.name}</option>)}
                        </select>
                      </div>
                    </div>

                    {/* Filter Status */}
                    <div>
                      <label style={{ fontSize: '0.85rem', fontWeight: 700, color: '#334155', display: 'block', marginBottom: '6px' }}>Status Pengajuan</label>
                      <select 
                        value={reportStatusFilter}
                        onChange={(e) => {
                          setReportStatusFilter(e.target.value);
                        }}
                        style={{ width: '100%', padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.88rem' }}
                      >
                        <option value="ALL">-- Semua Status --</option>
                        <option value="SUBMITTED">SUBMITTED (Diajukan)</option>
                        <option value="APPROVED">APPROVED (Disetujui)</option>
                        <option value="DISBURSED">DISBURSED (Dicairkan)</option>
                        <option value="REJECTED">REJECTED (Ditolak)</option>
                      </select>
                    </div>

                    {/* Input Search NIK / Employee ID */}
                    <div>
                      <label style={{ fontSize: '0.85rem', fontWeight: 700, color: '#334155', display: 'block', marginBottom: '6px' }}>Cari NIK / Member No</label>
                      <div style={{ display: 'flex', gap: '6px' }}>
                        <input 
                          type="text"
                          value={reportMemberSearchQuery}
                          onChange={(e) => setReportMemberSearchQuery(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              e.preventDefault();
                              setReportPage(1);
                              fetchApplications(reportMonthFilter.year, reportMonthFilter.month, reportMemberSearchQuery);
                            }
                          }}
                          placeholder="Ketik NIK / App No..."
                          style={{ width: '100%', padding: '8px 10px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.88rem' }}
                        />
                        <button
                          type="button"
                          onClick={() => {
                            setReportPage(1);
                            fetchApplications(reportMonthFilter.year, reportMonthFilter.month, reportMemberSearchQuery);
                          }}
                          style={{ padding: '8px 14px', background: '#0284c7', color: 'white', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 'bold', fontSize: '0.88rem' }}
                        >
                          🔍 Cari
                        </button>
                      </div>
                    </div>

                  </div>
                </div>
              </div>

              {/* Data Table */}
              <div className="card" style={{ background: '#ffffff', padding: '20px', borderRadius: '10px', border: '1px solid #cbd5e1' }}>
                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
                    <thead>
                      <tr style={{ background: '#f8fafc', borderBottom: '2px solid #e2e8f0' }}>
                        <th style={{ padding: '10px', textAlign: 'center', color: '#475569', whiteSpace: 'nowrap', width: '50px' }}>No.</th>
                        <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>No. Pengajuan</th>
                        <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>Tanggal</th>
                        <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>NIK / Employee ID</th>
                        <th style={{ padding: '10px', textAlign: 'right', color: '#475569', whiteSpace: 'nowrap' }}>Nominal Pengajuan</th>
                        <th style={{ padding: '10px', textAlign: 'right', color: '#475569', whiteSpace: 'nowrap' }}>Nominal Disetujui</th>
                        <th style={{ padding: '10px', textAlign: 'center', color: '#475569', whiteSpace: 'nowrap' }}>Tenor</th>
                        <th style={{ padding: '10px', textAlign: 'right', color: '#475569', whiteSpace: 'nowrap' }}>Angsuran / Bln</th>
                        <th style={{ padding: '10px', textAlign: 'center', color: '#475569', whiteSpace: 'nowrap' }}>Status</th>
                        <th className="no-print" style={{ padding: '10px', textAlign: 'center', color: '#475569', whiteSpace: 'nowrap' }}>Aksi</th>
                      </tr>
                    </thead>
                    <tbody>
                      {paginatedReportApps.length === 0 ? (
                        <tr>
                          <td colSpan="10" style={{ padding: '32px', textAlign: 'center', color: '#64748b' }}>
                            📭 Tidak ada data pengajuan pinjaman untuk periode dan filter yang dipilih.
                          </td>
                        </tr>
                      ) : (
                        paginatedReportApps.map((app, idx) => {
                          const statusColor = 
                            app.Status === 'DISBURSED' ? { bg: '#dcfce7', fg: '#15803d' } :
                            app.Status === 'APPROVED' ? { bg: '#e0f2fe', fg: '#0369a1' } :
                            app.Status === 'SUBMITTED' ? { bg: '#fef3c7', fg: '#b45309' } :
                            { bg: '#fee2e2', fg: '#b91c1c' };

                          return (
                            <tr key={app.ApplicationNo} style={{ borderBottom: '1px solid #f1f5f9' }}>
                              <td style={{ padding: '10px', textAlign: 'center', color: '#475569', fontWeight: 600 }}>
                                {startIndex + idx + 1}
                              </td>
                              <td style={{ padding: '10px', fontWeight: 'bold', fontFamily: 'monospace', color: '#0f172a' }}>
                                #{app.ApplicationNo}
                              </td>
                              <td style={{ padding: '10px', color: '#475569', whiteSpace: 'nowrap' }}>
                                {app.SubmissionDate ? new Date(app.SubmissionDate).toLocaleDateString('id-ID') : '-'}
                              </td>
                              <td style={{ padding: '10px', fontWeight: 600, color: '#1e293b' }}>
                                {app.MemberNo}
                              </td>
                              <td style={{ padding: '10px', textAlign: 'right', fontWeight: 600, color: '#0f172a' }}>
                                Rp {Math.round(app.RequestedAmount || 0).toLocaleString('id-ID')}
                              </td>
                              <td style={{ padding: '10px', textAlign: 'right', fontWeight: 700, color: '#047857' }}>
                                Rp {Math.round(app.ApprovedAmount || app.RequestedAmount || 0).toLocaleString('id-ID')}
                              </td>
                              <td style={{ padding: '10px', textAlign: 'center', color: '#475569' }}>
                                {app.Tenor} Bln
                              </td>
                              <td style={{ padding: '10px', textAlign: 'right', fontWeight: 600, color: '#2563eb' }}>
                                Rp {Math.round(app.TotalInstallment || 0).toLocaleString('id-ID')}
                              </td>
                              <td style={{ padding: '10px', textAlign: 'center' }}>
                                <span style={{ padding: '4px 10px', borderRadius: '12px', background: statusColor.bg, color: statusColor.fg, fontWeight: 'bold', fontSize: '0.75rem' }}>
                                  {app.Status}
                                </span>
                              </td>
                              <td className="no-print" style={{ padding: '10px', textAlign: 'center' }}>
                                <div style={{ display: 'flex', gap: '4px', justifyContent: 'center' }}>
                                  {(app.Status === 'APPROVED' || app.Status === 'DISBURSED') && (
                                    <button 
                                      onClick={() => handlePrintContract(app)}
                                      style={{ padding: '3px 8px', background: '#0f172a', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                      title="Cetak Kontrak"
                                    >
                                      📄 Kontrak
                                    </button>
                                  )}
                                  <button 
                                    onClick={() => handleOpenTracking(app.ApplicationNo)}
                                    style={{ padding: '3px 8px', background: '#3b82f6', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600, fontSize: '0.75rem' }}
                                    title="Lihat Tracking Status"
                                  >
                                    📜 Track
                                  </button>
                                </div>
                              </td>
                            </tr>
                          );
                        })
                      )}
                    </tbody>
                  </table>
                </div>

                {/* Pagination Controls - Disembunyikan saat cetak */}
                <div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '16px', fontSize: '0.85rem', color: '#475569' }}>
                  <span>Halaman <strong>{reportPage}</strong> dari <strong>{totalReportPages}</strong> ({filteredReportApps.length} Data)</span>
                  <div style={{ display: 'flex', gap: '6px' }}>
                    <button
                      disabled={reportPage <= 1}
                      onClick={() => setReportPage(prev => Math.max(prev - 1, 1))}
                      style={{ padding: '5px 12px', background: reportPage <= 1 ? '#e2e8f0' : '#0284c7', color: reportPage <= 1 ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: reportPage <= 1 ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                    >
                      ◀ Prev
                    </button>
                    <button
                      disabled={reportPage >= totalReportPages}
                      onClick={() => setReportPage(prev => Math.min(prev + 1, totalReportPages))}
                      style={{ padding: '5px 12px', background: reportPage >= totalReportPages ? '#e2e8f0' : '#0284c7', color: reportPage >= totalReportPages ? '#94a3b8' : 'white', border: 'none', borderRadius: '4px', cursor: reportPage >= totalReportPages ? 'not-allowed' : 'pointer', fontWeight: 'bold' }}
                    >
                      Next ▶
                    </button>
                  </div>
                </div>
              </div>
            </div>
          );
        })()}

          {!isReportTab() && activeTab !== 'dashboard' && activeTab !== 'pengajuan' && activeTab !== 'pinjaman' && activeTab !== 'parameters' && activeTab !== 'master' && activeTab !== 'approval' && activeTab !== 'disbursement' && activeTab !== 'payroll' && activeTab !== 'payroll-reconciliation' && activeTab !== 'manual-repayment' && activeTab !== 'products' && activeTab !== 'report-loan-applications' && !activeTab.startsWith('master-') && (
            <div className="card">
              <h2>Module: {visibleMenus.find(m => m.path === activeTab)?.title || activeTab}</h2>
              <p style={{ marginTop: '16px', color: '#64748B' }}>
                Halaman ini masih dalam tahap pengembangan. Anda mengakses halaman ini menggunakan role: <strong>{realRoleName}</strong>.
              </p>
            </div>
          )}
        {/* Modal UI Form Konfirmasi Pencairan Dana Pinjaman (Treasury Disbursement) */}
        {disbursementModalOpen && selectedDisburseApp && (
          <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1100, backdropFilter: 'blur(4px)' }}>
            <div style={{ backgroundColor: 'white', borderRadius: '12px', padding: '28px', width: '90%', maxWidth: '560px', maxHeight: '90vh', overflowY: 'auto', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)', border: '1px solid #cbd5e1' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
                <h3 style={{ margin: 0, fontSize: '1.2rem', color: '#0f172a', display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 700 }}>
                  💵 Form Konfirmasi Pencairan Dana
                </h3>
                <button type="button" onClick={() => setDisbursementModalOpen(false)} style={{ background: 'none', border: 'none', fontSize: '1.5rem', cursor: 'pointer', color: '#64748b', lineHeight: 1 }}>&times;</button>
              </div>

              {/* Rincian Ringkasan Pengajuan Pinjaman */}
              <div style={{ backgroundColor: '#f8fafc', borderRadius: '8px', padding: '16px', marginBottom: '20px', border: '1px solid #e2e8f0' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', fontSize: '0.875rem' }}>
                  <span style={{ color: '#64748b' }}>No. Pengajuan:</span>
                  <span style={{ fontWeight: 700, color: '#0f172a' }}>#{selectedDisburseApp.application_no}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', fontSize: '0.875rem' }}>
                  <span style={{ color: '#64748b' }}>Pemohon / Member:</span>
                  <span style={{ fontWeight: 600, color: '#0f172a' }}>Member #{selectedDisburseApp.member_no}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', fontSize: '0.875rem' }}>
                  <span style={{ color: '#64748b' }}>Tenor Jangka Waktu:</span>
                  <span style={{ fontWeight: 600, color: '#0f172a' }}>{selectedDisburseApp.tenor} Bulan</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', paddingTop: '8px', borderTop: '1px dashed #cbd5e1', fontSize: '0.95rem' }}>
                  <span style={{ fontWeight: 600, color: '#334155' }}>Nominal Plafon Diterima:</span>
                  <span style={{ fontWeight: 800, color: '#15803d', fontSize: '1.15rem' }}>Rp {Number(selectedDisburseApp.approved_amount).toLocaleString('id-ID')}</span>
                </div>
              </div>

              {/* Form Input Seluruh Data Pencairan dalam 1 Modal */}
              <form onSubmit={submitDisbursement}>
                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.875rem', color: '#334155' }}>
                    1. Nama Bank Tujuan Pencairan <span style={{ color: '#dc2626' }}>*</span>
                  </label>
                  <select
                    value={disburseForm.bank_name}
                    onChange={e => setDisburseForm({ ...disburseForm, bank_name: e.target.value })}
                    style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', backgroundColor: 'white', fontSize: '0.9rem', fontWeight: 500 }}
                  >
                    <option value="BCA">Bank BCA</option>
                    <option value="Mandiri">Bank Mandiri</option>
                    <option value="BNI">Bank BNI</option>
                    <option value="BRI">Bank BRI</option>
                    <option value="CIMB Niaga">Bank CIMB Niaga</option>
                    <option value="BSI">Bank Syariah Indonesia (BSI)</option>
                    <option value="Permata">Bank Permata</option>
                    <option value="Danamon">Bank Danamon</option>
                    <option value="Lainnya">-- Bank Lainnya (Ketik Manual) --</option>
                  </select>
                </div>

                {disburseForm.bank_name === 'Lainnya' && (
                  <div style={{ marginBottom: '16px' }}>
                    <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.875rem', color: '#334155' }}>
                      Nama Bank Lainnya <span style={{ color: '#dc2626' }}>*</span>
                    </label>
                    <input 
                      type="text" required
                      placeholder="Masukkan nama bank transfer..."
                      value={disburseForm.custom_bank}
                      onChange={e => setDisburseForm({ ...disburseForm, custom_bank: e.target.value })}
                      style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                    />
                  </div>
                )}

                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.875rem', color: '#334155' }}>
                    2. No. Rekening Tujuan Transfer <span style={{ color: '#dc2626' }}>*</span>
                  </label>
                  <input 
                    type="text" required
                    placeholder="Contoh: 1234567890"
                    value={disburseForm.bank_account_no}
                    onChange={e => setDisburseForm({ ...disburseForm, bank_account_no: e.target.value })}
                    style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem', fontWeight: 600, letterSpacing: '0.5px' }}
                  />
                </div>

                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.875rem', color: '#334155' }}>
                    3. Nama Pemilik Rekening / Atas Nama <span style={{ color: '#dc2626' }}>*</span>
                  </label>
                  <input 
                    type="text" required
                    placeholder="Nama pemilik rekening sesuai buku tabungan..."
                    value={disburseForm.account_holder_name}
                    onChange={e => setDisburseForm({ ...disburseForm, account_holder_name: e.target.value })}
                    style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '0.9rem' }}
                  />
                </div>

                <div style={{ marginBottom: '24px' }}>
                  <label style={{ display: 'block', marginBottom: '6px', fontWeight: 600, fontSize: '0.875rem', color: '#334155' }}>
                    4. Catatan Persetujuan / Bukti Referensi Treasury:
                  </label>
                  <textarea 
                    value={disburseForm.notes}
                    onChange={e => setDisburseForm({ ...disburseForm, notes: e.target.value })}
                    style={{ width: '100%', padding: '10px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', minHeight: '60px', fontSize: '0.85rem' }}
                  />
                </div>

                <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end', paddingTop: '16px', borderTop: '1px solid #e2e8f0' }}>
                  <button type="button" onClick={() => setDisbursementModalOpen(false)} style={{ padding: '10px 20px', background: '#f1f5f9', color: '#475569', border: '1px solid #cbd5e1', borderRadius: '6px', cursor: 'pointer', fontWeight: 600, fontSize: '0.875rem' }}>
                    Batal
                  </button>
                  <button type="submit" disabled={disburseLoading} style={{ padding: '10px 24px', background: disburseLoading ? '#94a3b8' : '#10b981', color: 'white', border: 'none', borderRadius: '6px', cursor: disburseLoading ? 'not-allowed' : 'pointer', fontWeight: 700, fontSize: '0.875rem', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    {disburseLoading ? 'Memproses...' : '💵 Proses Pencairan Dana'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

      {/* LIMS Style Riwayat Status / Loan Tracking Modal */}
      {trackingModalOpen && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(15, 23, 42, 0.65)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000, backdropFilter: 'blur(4px)' }}>
          <div style={{ backgroundColor: 'white', borderRadius: '8px', padding: '24px', width: '95%', maxWidth: '1200px', maxHeight: '85vh', overflowY: 'auto', boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.2)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', paddingBottom: '12px', borderBottom: '2px solid #e2e8f0' }}>
              <h2 style={{ margin: 0, fontSize: '1.25rem', color: '#1e293b', display: 'flex', alignItems: 'center', gap: '8px' }}>
                📜 Riwayat Status Pengajuan Pinjaman <span style={{ color: 'var(--primary-blue)', fontFamily: 'monospace' }}>#{trackingAppNo}</span>
              </h2>
              <button onClick={() => setTrackingModalOpen(false)} style={{ background: '#f1f5f9', border: 'none', borderRadius: '50%', width: '32px', height: '32px', cursor: 'pointer', fontWeight: 'bold', color: '#64748b' }}>✕</button>
            </div>

            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.85rem' }}>
              <thead>
                <tr style={{ background: '#f8fafc', borderBottom: '2px solid #e2e8f0' }}>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>Tanggal</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>Status</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>Employee ID</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>Nama</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>SLA (Durasi)</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', minWidth: '260px' }}>Catatan</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>IP Address</th>
                  <th style={{ padding: '10px', textAlign: 'left', color: '#475569', whiteSpace: 'nowrap' }}>User Agent</th>
                </tr>
              </thead>
              <tbody>
                {trackingList.length === 0 ? (
                  <tr><td colSpan="8" style={{ textAlign: 'center', padding: '24px', color: '#94a3b8' }}>Belum ada catatan riwayat status.</td></tr>
                ) : (
                  trackingList.map(t => (
                    <tr key={t.ID} style={{ borderBottom: '1px solid #f1f5f9' }}>
                      <td style={{ padding: '8px 10px', color: '#334155', whiteSpace: 'nowrap' }}>{formatDate(t.ActionDate)}</td>
                      <td style={{ padding: '8px 10px', whiteSpace: 'nowrap' }}>
                        <span className={getStatusBadge(t.Status)} style={{ backgroundColor: t.Status === 'REVISION_REQUIRED' ? '#f59e0b' : undefined, fontSize: '0.75rem' }}>
                          {t.Status === 'REVISION_REQUIRED' ? 'REVISI' : t.Status}
                        </span>
                      </td>
                      <td style={{ padding: '8px 10px', fontWeight: 600, color: '#0f172a', whiteSpace: 'nowrap' }}>{t.UserID || t.user_id || '-'}</td>
                      <td style={{ padding: '8px 10px', fontWeight: 600, color: '#1e293b', whiteSpace: 'nowrap' }}>{t.UserName || t.user_name || '-'}</td>
                      <td style={{ padding: '8px 10px', fontFamily: 'monospace', color: '#2563eb', fontWeight: 600, whiteSpace: 'nowrap' }}>{t.SLADuration || '-'}</td>
                      <td style={{ padding: '8px 10px', color: '#475569', minWidth: '260px', maxWidth: '450px' }}>
                        <div style={{
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          lineHeight: '1.35',
                          wordBreak: 'break-word'
                        }}>
                          {t.Notes || '-'}
                        </div>
                      </td>
                      <td style={{ padding: '8px 10px', fontFamily: 'monospace', fontSize: '0.8rem', color: '#64748b', whiteSpace: 'nowrap' }}>{t.IPAddress || '127.0.0.1'}</td>
                      <td style={{ padding: '8px 10px', fontSize: '0.8rem', color: '#64748b', maxWidth: '140px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.UserAgent || 'Chrome'}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>

            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '20px', paddingTop: '12px', borderTop: '1px solid #e2e8f0' }}>
              <button onClick={() => setTrackingModalOpen(false)} style={{ padding: '8px 20px', background: '#64748b', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 600 }}>Tutup</button>
            </div>
          </div>
        </div>
      )}
        </div>
      </main>
    </div>
  )
}

export default App
