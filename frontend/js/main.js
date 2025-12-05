/**
 * Research Data Analysis - Main Application
 * Built with JSCroot Framework
 */

// JSCroot imports from CDN
import { 
    setInner, 
    addInner, 
    getValue, 
    setValue, 
    onClick, 
    hide, 
    show, 
    addChild 
} from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/element.js';

import { 
    postJSON, 
    getJSON, 
    postFileWithHeader 
} from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/api.js';

import { 
    setCookieWithExpireHour, 
    getCookie, 
    deleteCookie 
} from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/cookie.js';

import { 
    redirect, 
    getHash, 
    setHash, 
    onHashChange 
} from 'https://cdn.jsdelivr.net/gh/jscroot/lib@0.0.3/url.js';

// ===== Configuration =====
const API_BASE_URL = 'https://research-data-backend-118539834796.asia-southeast1.run.app';

// ===== State Management =====
let currentUser = null;
let currentProject = null;
let currentAnalysis = null;
let selectedMethods = [];

// ===== Initialization =====
document.addEventListener('DOMContentLoaded', () => {
    initApp();
    initHeroChart();
    setupUploadDropzone();
    setupHashNavigation();
});

function initApp() {
    // Check if user is logged in
    const token = getCookie('token');
    if (token) {
        loadUserProfile();
    } else {
        updateAuthUI(false);
    }
    
    // Navigate based on hash or default
    const hash = getHash();
    if (hash) {
        navigateTo(hash);
    } else {
        navigateTo('home');
    }
}

function setupHashNavigation() {
    onHashChange((event) => {
        const hash = getHash();
        if (hash) {
            navigateTo(hash);
        }
    });
}

// ===== Navigation =====
window.navigateTo = function(page) {
    // Hide all pages
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    
    // Show requested page
    const pageEl = document.getElementById(`page-${page}`);
    if (pageEl) {
        pageEl.classList.add('active');
    }
    
    // Update nav links
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
        if (link.dataset.page === page) {
            link.classList.add('active');
        }
    });
    
    // Update hash without triggering navigation
    if (getHash() !== page) {
        setHash(page);
    }
    
    // Protected pages - require login
    const protectedPages = ['dashboard', 'projects', 'new-project', 'project', 'profile'];
    if (protectedPages.includes(page) && !getCookie('token')) {
        showToast('Silakan login terlebih dahulu', 'warning');
        navigateTo('login');
        return;
    }
    
    // Page-specific actions
    if (page === 'dashboard') {
        loadDashboard();
    }
};

// ===== Auth Functions =====
window.handleLogin = async function(event) {
    event.preventDefault();
    
    const email = getValue('login-email');
    const password = getValue('login-password');
    
    if (!email || !password) {
        showToast('Mohon isi semua field', 'error');
        return;
    }
    
    showLoading();
    
    postJSON(
        `${API_BASE_URL}/auth/login`,
        { email, password },
        (response) => {
            hideLoading();
            if (response.status === 200) {
                const data = response.data;
                setCookieWithExpireHour('token', data.data.token, 24);
                currentUser = data.data.user;
                updateAuthUI(true);
                showToast('Login berhasil!', 'success');
                navigateTo('dashboard');
            } else {
                showToast(response.data.message || 'Login gagal', 'error');
            }
        }
    );
};

window.handleRegister = async function(event) {
    event.preventDefault();
    
    const fullName = getValue('register-name');
    const email = getValue('register-email');
    const password = getValue('register-password');
    const institution = getValue('register-institution');
    const researchField = getValue('register-field');
    
    if (!fullName || !email || !password) {
        showToast('Mohon isi semua field yang diperlukan', 'error');
        return;
    }
    
    if (password.length < 8) {
        showToast('Password minimal 8 karakter', 'error');
        return;
    }
    
    showLoading();
    
    postJSON(
        `${API_BASE_URL}/auth/register`,
        {
            email,
            password,
            fullName: fullName,
            institution,
            researchField: researchField
        },
        (response) => {
            hideLoading();
            if (response.status === 201) {
                const data = response.data;
                setCookieWithExpireHour('token', data.data.token, 24);
                currentUser = data.data.user;
                updateAuthUI(true);
                showToast('Registrasi berhasil!', 'success');
                navigateTo('dashboard');
            } else {
                showToast(response.data.message || 'Registrasi gagal', 'error');
            }
        }
    );
};

window.logout = function() {
    deleteCookie('token');
    currentUser = null;
    updateAuthUI(false);
    showToast('Berhasil logout', 'success');
    navigateTo('home');
};

function loadUserProfile() {
    const token = getCookie('token');
    if (!token) return;
    
    getJSON(
        `${API_BASE_URL}/auth/profile`,
        (response) => {
            if (response.status === 200) {
                currentUser = response.data.data;
                updateAuthUI(true);
            } else {
                deleteCookie('token');
                updateAuthUI(false);
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
}

function updateAuthUI(isLoggedIn) {
    const authButtons = document.getElementById('auth-buttons');
    const userMenu = document.getElementById('user-menu');
    const userName = document.getElementById('user-name');
    
    if (isLoggedIn && currentUser) {
        hide('auth-buttons');
        show('user-menu');
        setInner('user-name', currentUser.full_name || currentUser.email);
    } else {
        show('auth-buttons');
        hide('user-menu');
    }
}

window.toggleUserMenu = function() {
    const dropdown = document.getElementById('user-dropdown');
    dropdown.classList.toggle('hidden');
};

// ===== Dashboard Functions =====
function loadDashboard() {
    const token = getCookie('token');
    if (!token) return;
    
    showLoading();
    
    getJSON(
        `${API_BASE_URL}/api/project`,
        (response) => {
            hideLoading();
            if (response.status === 200) {
                const projects = response.data.data || [];
                renderProjects(projects);
                updateStats(projects);
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
}

function renderProjects(projects) {
    const container = document.getElementById('projects-list');
    
    if (!projects || projects.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
                </svg>
                <p>Belum ada proyek. Buat proyek pertama Anda!</p>
                <button class="btn btn-primary" onclick="navigateTo('new-project')">Buat Proyek</button>
            </div>
        `;
        return;
    }
    
    container.innerHTML = projects.map(project => `
        <div class="project-card" onclick="openProject('${project._id}')">
            <div class="project-info">
                <h3>${escapeHtml(project.title)}</h3>
                <div class="project-meta">
                    <span>${getResearchTypeName(project.research_type)}</span>
                    <span>${formatDate(project.created_at)}</span>
                </div>
            </div>
            <span class="status-badge ${project.status}">${getStatusName(project.status)}</span>
        </div>
    `).join('');
}

function updateStats(projects) {
    const total = projects.length;
    const completed = projects.filter(p => p.status === 'completed').length;
    const analyzing = projects.filter(p => p.status === 'analyzing').length;
    
    setInner('stat-projects', total.toString());
    setInner('stat-completed', completed.toString());
    setInner('stat-analyzing', analyzing.toString());
    setInner('stat-uploads', '0'); // Will be updated when we have upload counts
}

// ===== Project Functions =====
window.handleCreateProject = async function(event) {
    event.preventDefault();
    
    const title = getValue('project-title');
    const description = getValue('project-description');
    const researchType = getValue('project-type');
    const hypothesis = getValue('project-hypothesis');
    const varIndependent = getValue('var-independent');
    const varDependent = getValue('var-dependent');
    
    if (!title || !researchType) {
        showToast('Mohon isi field yang diperlukan', 'error');
        return;
    }
    
    const token = getCookie('token');
    showLoading();
    
    postJSON(
        `${API_BASE_URL}/api/project`,
        {
            title,
            description,
            research_type: researchType,
            hypothesis,
            variables: {
                independent: varIndependent ? varIndependent.split(',').map(v => v.trim()) : [],
                dependent: varDependent ? varDependent.split(',').map(v => v.trim()) : []
            }
        },
        (response) => {
            hideLoading();
            if (response.status === 201) {
                showToast('Proyek berhasil dibuat!', 'success');
                openProject(response.data.data._id);
            } else {
                showToast(response.data.message || 'Gagal membuat proyek', 'error');
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
};

window.openProject = function(projectId) {
    const token = getCookie('token');
    showLoading();
    
    getJSON(
        `${API_BASE_URL}/api/project/${projectId}`,
        (response) => {
            hideLoading();
            if (response.status === 200) {
                currentProject = response.data.data;
                renderProjectDetail(currentProject);
                navigateTo('project');
            } else {
                showToast('Gagal memuat proyek', 'error');
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
};

function renderProjectDetail(project) {
    setInner('project-detail-title', escapeHtml(project.title));
    
    const statusBadge = document.getElementById('project-status');
    statusBadge.textContent = getStatusName(project.status);
    statusBadge.className = `status-badge ${project.status}`;
    
    // Overview details
    setInner('detail-type', getResearchTypeName(project.research_type));
    setInner('detail-hypothesis', project.hypothesis || '-');
    setInner('detail-var-ind', project.variables?.independent?.join(', ') || '-');
    setInner('detail-var-dep', project.variables?.dependent?.join(', ') || '-');
    
    // Update next steps based on status
    updateNextSteps(project.status);
}

function updateNextSteps(status) {
    const steps = document.querySelectorAll('.step-item');
    const statusOrder = ['draft', 'uploaded', 'analyzing', 'completed'];
    const currentIndex = statusOrder.indexOf(status);
    
    steps.forEach((step, index) => {
        step.classList.remove('active', 'completed');
        if (index < currentIndex) {
            step.classList.add('completed');
        } else if (index === currentIndex) {
            step.classList.add('active');
        }
    });
}

window.switchProjectTab = function(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
        if (btn.dataset.tab === tabName) {
            btn.classList.add('active');
        }
    });
    
    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`tab-${tabName}`).classList.add('active');
};

// ===== Upload Functions =====
function setupUploadDropzone() {
    const dropzone = document.getElementById('upload-dropzone');
    const fileInput = document.getElementById('file-input');
    
    if (!dropzone || !fileInput) return;
    
    dropzone.addEventListener('click', () => fileInput.click());
    
    dropzone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropzone.classList.add('dragover');
    });
    
    dropzone.addEventListener('dragleave', () => {
        dropzone.classList.remove('dragover');
    });
    
    dropzone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropzone.classList.remove('dragover');
        
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            handleFileUpload(files[0]);
        }
    });
    
    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            handleFileUpload(e.target.files[0]);
        }
    });
}

function handleFileUpload(file) {
    if (!currentProject) {
        showToast('Tidak ada proyek yang dipilih', 'error');
        return;
    }
    
    // Validate file type
    const validTypes = ['.csv', '.xlsx', '.xls', '.json'];
    const fileExt = '.' + file.name.split('.').pop().toLowerCase();
    if (!validTypes.includes(fileExt)) {
        showToast('Format file tidak didukung. Gunakan CSV, Excel, atau JSON.', 'error');
        return;
    }
    
    // Validate file size (50MB max)
    if (file.size > 50 * 1024 * 1024) {
        showToast('Ukuran file maksimal 50MB', 'error');
        return;
    }
    
    const token = getCookie('token');
    showLoading();
    
    // Create form data
    const formData = new FormData();
    formData.append('file', file);
    
    fetch(`${API_BASE_URL}/api/upload/${currentProject._id}`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        },
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        hideLoading();
        if (data.status === 'success') {
            showToast('File berhasil diupload!', 'success');
            renderUploadPreview(data.data);
        } else {
            showToast(data.message || 'Upload gagal', 'error');
        }
    })
    .catch(error => {
        hideLoading();
        showToast('Terjadi kesalahan saat upload', 'error');
        console.error('Upload error:', error);
    });
}

function renderUploadPreview(upload) {
    show('upload-preview');
    show('data-preview');
    
    // Show uploaded file info
    const filesList = document.getElementById('uploaded-files-list');
    filesList.innerHTML = `
        <div class="uploaded-file">
            <span class="file-name">${escapeHtml(upload.file_name)}</span>
            <span class="file-size">${formatFileSize(upload.file_size)}</span>
        </div>
    `;
    
    // Show data summary
    const summary = upload.data_summary;
    const summaryEl = document.getElementById('data-summary');
    summaryEl.innerHTML = `
        <div class="summary-item">
            <span class="value">${summary.rows}</span>
            <span class="label">Baris</span>
        </div>
        <div class="summary-item">
            <span class="value">${summary.columns}</span>
            <span class="label">Kolom</span>
        </div>
        <div class="summary-item">
            <span class="value">${Object.values(summary.column_types).filter(t => t === 'numeric').length}</span>
            <span class="label">Numerik</span>
        </div>
        <div class="summary-item">
            <span class="value">${Object.values(summary.column_types).filter(t => t === 'categorical').length}</span>
            <span class="label">Kategorikal</span>
        </div>
    `;
    
    // Load data preview
    loadDataPreview(upload._id);
}

function loadDataPreview(uploadId) {
    const token = getCookie('token');
    
    getJSON(
        `${API_BASE_URL}/api/preview/${uploadId}`,
        (response) => {
            if (response.status === 200) {
                const preview = response.data.data.preview;
                renderPreviewTable(preview);
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
}

function renderPreviewTable(data) {
    if (!data || data.length === 0) return;
    
    const thead = document.getElementById('preview-thead');
    const tbody = document.getElementById('preview-tbody');
    
    // Render header
    if (Array.isArray(data[0])) {
        // CSV format
        thead.innerHTML = `<tr>${data[0].map(h => `<th>${escapeHtml(h)}</th>`).join('')}</tr>`;
        tbody.innerHTML = data.slice(1, 11).map(row => 
            `<tr>${row.map(cell => `<td>${escapeHtml(cell)}</td>`).join('')}</tr>`
        ).join('');
    } else {
        // JSON format
        const headers = Object.keys(data[0]);
        thead.innerHTML = `<tr>${headers.map(h => `<th>${escapeHtml(h)}</th>`).join('')}</tr>`;
        tbody.innerHTML = data.slice(0, 10).map(row => 
            `<tr>${headers.map(h => `<td>${escapeHtml(row[h] || '')}</td>`).join('')}</tr>`
        ).join('');
    }
}

// ===== Analysis Functions =====
window.getRecommendations = function() {
    if (!currentProject) {
        showToast('Tidak ada proyek yang dipilih', 'error');
        return;
    }
    
    const token = getCookie('token');
    showLoading();
    
    postJSON(
        `${API_BASE_URL}/api/recommend/${currentProject._id}`,
        {},
        (response) => {
            hideLoading();
            if (response.status === 200) {
                const data = response.data.data;
                currentAnalysis = { id: data.analysis_id };
                renderRecommendations(data.recommendations);
                show('recommendations-container');
            } else {
                showToast(response.data.message || 'Gagal mendapatkan rekomendasi', 'error');
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
};

function renderRecommendations(recommendations) {
    const container = document.getElementById('recommendations-list');
    selectedMethods = [];
    
    if (!recommendations || recommendations.length === 0) {
        container.innerHTML = '<p>Tidak ada rekomendasi yang tersedia.</p>';
        return;
    }
    
    container.innerHTML = recommendations.map((rec, index) => `
        <div class="recommendation-card" onclick="toggleMethodSelection('${rec.method}', this)">
            <div class="recommendation-checkbox">
                <input type="checkbox" id="method-${index}" value="${rec.method}">
            </div>
            <div class="recommendation-content">
                <h4>${escapeHtml(rec.method)}</h4>
                <p>${escapeHtml(rec.reasoning)}</p>
                <div class="recommendation-meta">
                    <span>Kategori: ${escapeHtml(rec.category)}</span>
                    <span>Prioritas: ${rec.priority}</span>
                </div>
            </div>
        </div>
    `).join('');
}

window.toggleMethodSelection = function(method, element) {
    const checkbox = element.querySelector('input[type="checkbox"]');
    checkbox.checked = !checkbox.checked;
    element.classList.toggle('selected', checkbox.checked);
    
    if (checkbox.checked) {
        if (!selectedMethods.includes(method)) {
            selectedMethods.push(method);
        }
    } else {
        selectedMethods = selectedMethods.filter(m => m !== method);
    }
    
    // Enable/disable process button
    const processBtn = document.getElementById('btn-process-analysis');
    processBtn.disabled = selectedMethods.length === 0;
};

window.processAnalysis = function() {
    if (selectedMethods.length === 0) {
        showToast('Pilih minimal satu metode analisis', 'warning');
        return;
    }
    
    if (!currentAnalysis || !currentAnalysis.id) {
        showToast('Tidak ada analisis yang aktif', 'error');
        return;
    }
    
    const token = getCookie('token');
    showLoading();
    
    postJSON(
        `${API_BASE_URL}/api/process`,
        {
            analysis_id: currentAnalysis.id,
            selected_methods: selectedMethods
        },
        (response) => {
            hideLoading();
            if (response.status === 200) {
                const data = response.data.data;
                showToast('Analisis selesai!', 'success');
                renderResults(data.results, data.summary);
                switchProjectTab('results');
            } else {
                showToast(response.data.message || 'Analisis gagal', 'error');
            }
        },
        'Authorization',
        `Bearer ${token}`
    );
};

function renderResults(results, summary) {
    hide('results-empty');
    show('results-content');
    
    const resultsList = document.getElementById('results-list');
    resultsList.innerHTML = results.map(result => `
        <div class="result-card">
            <h4>${escapeHtml(result.method)}</h4>
            <div class="result-interpretation">
                <p>${escapeHtml(result.interpretation)}</p>
            </div>
            <div class="result-conclusion">
                <strong>Kesimpulan:</strong> ${escapeHtml(result.conclusion)}
            </div>
        </div>
    `).join('');
    
    // Render summary
    const summaryEl = document.getElementById('results-summary');
    if (summary) {
        summaryEl.innerHTML = `
            <h3>Ringkasan Analisis</h3>
            <p>${escapeHtml(summary)}</p>
        `;
    }
}

// ===== Export Functions =====
window.exportResults = function(format) {
    if (!currentAnalysis || !currentAnalysis.id) {
        showToast('Tidak ada hasil untuk diekspor', 'warning');
        return;
    }
    
    const token = getCookie('token');
    const url = `${API_BASE_URL}/api/export/${currentAnalysis.id}?format=${format}`;
    
    // Open download in new window/tab
    window.open(url, '_blank');
};

// ===== Hero Chart =====
function initHeroChart() {
    const ctx = document.getElementById('hero-chart');
    if (!ctx) return;
    
    new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: ['Kuantitatif', 'Kualitatif', 'Mixed Methods'],
            datasets: [{
                data: [45, 30, 25],
                backgroundColor: [
                    'rgba(59, 130, 246, 0.8)',
                    'rgba(99, 102, 241, 0.8)',
                    'rgba(16, 185, 129, 0.8)'
                ],
                borderWidth: 0
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    position: 'bottom'
                }
            },
            cutout: '60%'
        }
    });
}

// ===== Utility Functions =====
function showLoading() {
    show('loading-overlay');
}

function hideLoading() {
    hide('loading-overlay');
}

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.remove();
    }, 5000);
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('id-ID', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

function formatFileSize(bytes) {
    if (!bytes) return '0 B';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
}

function getResearchTypeName(type) {
    const types = {
        'quantitative': 'Kuantitatif',
        'qualitative': 'Kualitatif',
        'mixed': 'Mixed Methods'
    };
    return types[type] || type || '-';
}

function getStatusName(status) {
    const statuses = {
        'draft': 'Draft',
        'uploaded': 'Data Diupload',
        'analyzing': 'Sedang Dianalisis',
        'completed': 'Selesai'
    };
    return statuses[status] || status || 'Draft';
}

// Export functions for global access
window.showLoading = showLoading;
window.hideLoading = hideLoading;
window.showToast = showToast;