/**
 * Settings page JavaScript
 */
document.addEventListener('DOMContentLoaded', function() {
    // Handle URL parameters for success messages
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('updated') === 'true') {
        showSuccessMessage('update-success');
    }
    if (urlParams.get('token_updated') === 'true') {
        showSuccessMessage('token-update-success');
    }
    
    // Initialize the Bootstrap tabs
    initTabs();
    
    // Form validation and submission
    setupFormValidation();
    
    // Token visibility toggle
    setupTokenVisibility();
    
    // Copy token to clipboard
    setupTokenCopy();
    
    // Setup danger zone actions
    setupDangerActions();
    
    // Setup connection test
    setupConnectionTest();
});

// Initialize Bootstrap tabs with history support
function initTabs() {
    // If hash exists in URL, activate that tab
    const hash = window.location.hash;
    if (hash) {
        const tabId = hash.replace('#', '');
        const tab = document.querySelector(`.nav-link[data-bs-target="#${tabId}"]`);
        if (tab) {
            // Use Bootstrap's tab API
            const bsTab = new bootstrap.Tab(tab);
            bsTab.show();
        }
    }
    
    // Update URL hash when tab changes
    document.querySelectorAll('.nav-link').forEach(tab => {
        tab.addEventListener('shown.bs.tab', function(e) {
            const targetId = e.target.getAttribute('data-bs-target').replace('#', '');
            history.replaceState(null, null, `#${targetId}`);
        });
    });
}

// Form validation and submission
function setupFormValidation() {
    // General settings form
    const generalForm = document.getElementById('general-settings-form');
    if (generalForm) {
        generalForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Validate form fields
            const serverHost = document.getElementById('server-host').value;
            const serverPort = document.getElementById('server-port').value;
            
            if (!serverHost || !serverPort) {
                alert('Semua field harus diisi');
                return;
            }
            
            // Submit the form
            this.submit();
        });
    }
    
    // Connection settings form
    const connectionForm = document.getElementById('connection-settings-form');
    if (connectionForm) {
        connectionForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Validate form fields
            const maxRetry = document.getElementById('max-retry').value;
            const retryDelay = document.getElementById('retry-delay').value;
            
            if (!maxRetry || !retryDelay || maxRetry < 0 || retryDelay < 0) {
                alert('Nilai tidak valid');
                return;
            }
            
            // Submit the form
            this.submit();
        });
    }
    
    // Session settings form
    const sessionForm = document.getElementById('session-settings-form');
    if (sessionForm) {
        sessionForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Validate form fields
            const tokenExpiry = document.getElementById('token-expiry').value;
            
            if (!tokenExpiry || tokenExpiry < 0) {
                alert('Nilai tidak valid');
                return;
            }
            
            // Submit the form
            this.submit();
        });
    }
    
    // Storage settings form
    const storageForm = document.getElementById('storage-settings-form');
    if (storageForm) {
        storageForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            // No specific validation needed
            
            // Submit the form
            this.submit();
        });
    }
}

// Setup token visibility toggle
function setupTokenVisibility() {
    const toggleBtn = document.getElementById('toggle-token-btn');
    if (toggleBtn) {
        toggleBtn.addEventListener('click', function() {
            const tokenInput = document.getElementById('api-token');
            const type = tokenInput.getAttribute('type') === 'password' ? 'text' : 'password';
            tokenInput.setAttribute('type', type);
            
            // Update icon
            const icon = this.querySelector('i');
            if (type === 'text') {
                icon.className = 'fas fa-eye-slash';
            } else {
                icon.className = 'fas fa-eye';
            }
        });
    }
}

// Setup copy token to clipboard
function setupTokenCopy() {
    const copyBtn = document.getElementById('copy-token-btn');
    if (copyBtn) {
        copyBtn.addEventListener('click', function() {
            const tokenInput = document.getElementById('api-token');
            
            // Copy to clipboard
            tokenInput.select();
            document.execCommand('copy');
            
            // Visual feedback
            const originalText = this.innerHTML;
            this.innerHTML = '<i class="fas fa-check"></i>';
            
            setTimeout(() => {
                this.innerHTML = originalText;
            }, 2000);
        });
    }
}

// Setup danger zone actions with confirmation
function setupDangerActions() {
    // Clear sessions button
    const clearSessionsBtn = document.getElementById('clear-sessions-btn');
    if (clearSessionsBtn) {
        clearSessionsBtn.addEventListener('click', function() {
            if (confirm('Apakah Anda yakin ingin menghapus semua sesi WhatsApp? Anda harus scan QR code lagi untuk terhubung.')) {
                // Send request to clear sessions endpoint
                fetch('/api/sessions/clear', {
                    method: 'POST',
                    headers: {
                        'X-Access-Token': localStorage.getItem('access_token') || ''
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.sukses) {
                        alert('Sesi berhasil dihapus');
                        window.location.reload();
                    } else {
                        alert('Gagal menghapus sesi: ' + data.pesan);
                    }
                })
                .catch(error => {
                    alert('Error: ' + error.message);
                });
            }
        });
    }
    
    // Clear storage button
    const clearStorageBtn = document.getElementById('clear-storage-btn');
    if (clearStorageBtn) {
        clearStorageBtn.addEventListener('click', function() {
            if (confirm('Apakah Anda yakin ingin menghapus semua data storage? Semua data tersimpan akan hilang.')) {
                // Send request to clear storage endpoint
                fetch('/api/storage/clear', {
                    method: 'POST',
                    headers: {
                        'X-Access-Token': localStorage.getItem('access_token') || ''
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.sukses) {
                        alert('Data storage berhasil dihapus');
                        window.location.reload();
                    } else {
                        alert('Gagal menghapus data storage: ' + data.pesan);
                    }
                })
                .catch(error => {
                    alert('Error: ' + error.message);
                });
            }
        });
    }
    
    // Reset config button
    const resetConfigBtn = document.getElementById('reset-config-btn');
    if (resetConfigBtn) {
        resetConfigBtn.addEventListener('click', function() {
            if (confirm('Apakah Anda yakin ingin mengembalikan semua konfigurasi ke pengaturan default?')) {
                // Send request to reset config endpoint
                fetch('/api/config/reset', {
                    method: 'POST',
                    headers: {
                        'X-Access-Token': localStorage.getItem('access_token') || ''
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.sukses) {
                        alert('Konfigurasi berhasil direset ke default');
                        window.location.reload();
                    } else {
                        alert('Gagal mereset konfigurasi: ' + data.pesan);
                    }
                })
                .catch(error => {
                    alert('Error: ' + error.message);
                });
            }
        });
    }
}

// Setup connection test
function setupConnectionTest() {
    const testConnectionBtn = document.getElementById('test-connection-btn');
    if (testConnectionBtn) {
        testConnectionBtn.addEventListener('click', function() {
            this.disabled = true;
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Tes Koneksi';
            
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    alert('Status koneksi: ' + data.status);
                })
                .catch(error => {
                    alert('Gagal menguji koneksi: ' + error.message);
                })
                .finally(() => {
                    this.disabled = false;
                    this.innerHTML = '<i class="fas fa-vial"></i> Tes Koneksi';
                });
        });
    }
    
    const reconnectBtn = document.getElementById('reconnect-btn');
    if (reconnectBtn) {
        reconnectBtn.addEventListener('click', function() {
            this.disabled = true;
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Menghubungkan...';
            
            fetch('/api/reconnect', {
                method: 'POST',
                headers: {
                    'X-Access-Token': localStorage.getItem('access_token') || ''
                }
            })
            .then(response => response.json())
            .then(data => {
                alert('Permintaan koneksi ulang telah dikirim');
                setTimeout(() => window.location.reload(), 3000);
            })
            .catch(error => {
                alert('Gagal menghubungkan: ' + error.message);
                this.disabled = false;
                this.innerHTML = '<i class="fas fa-sync"></i> Hubungkan Ulang';
            });
        });
    }
}

// Show success message and auto-hide after delay
function showSuccessMessage(id, delay = 5000) {
    const element = document.getElementById(id);
    if (element) {
        element.style.display = 'block';
        setTimeout(() => {
            element.style.opacity = '0';
            setTimeout(() => {
                element.style.display = 'none';
                element.style.opacity = '1';
            }, 500);
        }, delay);
    }
}
