/**
 * Dashboard JavaScript
 * Handles dashboard functionality and real-time updates
 */

console.log('%c🔄 Dashboard.js loaded at ' + new Date().toISOString(), 'color: #0a939f; font-weight: bold; font-size: 14px;');

// Global variables
let statusCheckInterval;
let isPageActive = true;
let connectionStatus = {
    isConnected: false,
    status: '',
    retries: 0
};

// Initialize dashboard on load
document.addEventListener('DOMContentLoaded', function() {
    console.log('Dashboard initialized');
    
    // Setup event listeners
    setupEventListeners();
    
    // Initial status check
    checkStatus();
    
    // Setup polling interval (every 5 seconds)
    statusCheckInterval = setInterval(checkStatus, 5000);
    
    // Setup visibility change detection
    handleVisibilityChange();
    
    // Initialize server time
    updateServerTime();
    setInterval(updateServerTime, 1000);
});

/**
 * Set up all event listeners for dashboard
 */
function setupEventListeners() {
    // Refresh button click
    const refreshBtn = document.getElementById('refresh-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            // Show spinning animation
            this.classList.add('spinning');
            
            // Perform immediate status check
            checkStatus().then(() => {
                // Remove spinning animation after 1 second
                setTimeout(() => {
                    this.classList.remove('spinning');
                }, 1000);
            });
        });
    }
    
    // Disconnect button click
    const disconnectBtn = document.getElementById('disconnect-btn');
    if (disconnectBtn) {
        disconnectBtn.addEventListener('click', function() {
            if (confirm('Apakah Anda yakin ingin memutuskan koneksi WhatsApp?')) {
                disconnectWhatsApp();
            }
        });
    }
    
    // Refresh QR button click
    const refreshQrBtn = document.getElementById('refresh-qr-btn');
    if (refreshQrBtn) {
        refreshQrBtn.addEventListener('click', function() {
            refreshQRCode();
        });
    }
    
    // Clear activity button click
    const clearActivityBtn = document.getElementById('clear-activity-btn');
    if (clearActivityBtn) {
        clearActivityBtn.addEventListener('click', function() {
            clearActivityList();
        });
    }
    
    // Copy API button click
    const copyApiBtn = document.getElementById('copy-api-btn');
    if (copyApiBtn) {
        copyApiBtn.addEventListener('click', function() {
            const codeBlock = document.querySelector('.code-block code');
            if (codeBlock) {
                navigator.clipboard.writeText(codeBlock.textContent)
                    .then(() => {
                        const originalIcon = this.innerHTML;
                        this.innerHTML = '<i class="fas fa-check"></i>';
                        setTimeout(() => {
                            this.innerHTML = originalIcon;
                        }, 2000);
                    })
                    .catch(err => {
                        console.error('Failed to copy: ', err);
                    });
            }
        });
    }
}

/**
 * Handle page visibility changes to optimize polling
 */
function handleVisibilityChange() {
    document.addEventListener('visibilitychange', function() {
        isPageActive = !document.hidden;
        console.log('Page visibility changed, active:', isPageActive);
        
        // If page becomes visible again, check status immediately
        if (isPageActive) {
            checkStatus();
        }
    });
}

/**
 * Check WhatsApp connection status
 * @returns {Promise} Promise that resolves when status check is complete
 */
function checkStatus() {
    console.log('Checking status at ' + new Date().toISOString());
    
    return fetch('/api/status', {
        method: 'GET',
        credentials: 'same-origin',
        headers: {
            'Accept': 'application/json',
            'X-Access-Token': apiToken, // Use the token passed from the server
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
    .then(response => {
        if (!response.ok) {
            if (response.status === 500) {
                throw new Error(`Server error (500) - server may be initializing, will retry later`);
            }
            
            if (response.headers.get('content-type')?.includes('text/html')) {
                throw new Error('Authentication required - received HTML instead of JSON');
            }
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        // Make sure content-type is application/json
        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
            console.warn('Response is not JSON:', contentType);
            throw new Error('Invalid response format: expected JSON');
        }
        
        if (response.redirected) {
            console.warn('Response was redirected to', response.url);
        }
        return response.json();
    })
    .then(data => {
        console.log('Status data received:', data);
        
        // Save previous state to detect changes
        const prevStatus = connectionStatus.status;
        const prevConnected = connectionStatus.isConnected;
        
        // Update connection status
        connectionStatus = {
            isConnected: data.details?.is_connected || false,
            status: data.details?.status || 'disconnected',
            retries: data.details?.connection_retries || 0
        };
        
        // Update UI based on status
        updateStatusUI(connectionStatus);
        
        // If connection status changed, prompt user to reload
        if (prevConnected !== connectionStatus.isConnected) {
            setTimeout(() => {
                if (confirm('Status koneksi berubah. Muat ulang halaman untuk melihat semua fitur?')) {
                    window.location.reload();
                }
            }, 1000);
        }
        
        return data;
    })
    .catch(error => {
        console.error('Error checking status:', error);
        console.error('Error details:', error);
        
        // If we can't reach the server, mark as disconnected
        connectionStatus = {
            isConnected: false,
            status: 'error',
            retries: 0
        };
        updateStatusUI(connectionStatus);
        
        // Return rejected promise to allow retry handling
        return Promise.reject(error);
    });
}

/**
 * Update UI elements based on connection status
 */
function updateStatusUI(status) {
    const connectionBadge = document.getElementById('connection-badge');
    const connectionStatus = document.getElementById('connection-status');
    const qrView = document.getElementById('qr-view');
    const connectedView = document.getElementById('connected-view');
    
    // Only update DOM elements if they exist
    if (connectionBadge) {
        connectionBadge.textContent = status.isConnected ? 'Terhubung' : 'Terputus';
        connectionBadge.className = `status-badge ${status.isConnected ? 'connected' : 'disconnected'}`;
    }
    
    if (connectionStatus) {
        connectionStatus.textContent = status.status;
    }
    
    // Show/hide appropriate view based on connection
    if (qrView && connectedView) {
        if (status.isConnected) {
            qrView.style.display = 'none';
            connectedView.style.display = 'block';
        } else {
            qrView.style.display = 'block';
            connectedView.style.display = 'none';
            // If disconnected, try to load QR code
            loadQRCode();
        }
    }
    
    // Update retry count if element exists
    const retryCount = document.getElementById('retry-count');
    if (retryCount) {
        retryCount.textContent = status.retries;
    }
    
    // Update retry info based on count
    const retryInfo = document.getElementById('retry-info');
    if (retryInfo && status.retries !== undefined) {
        if (status.retries === 0) {
            retryInfo.textContent = 'Koneksi stabil';
        } else if (status.retries < 3) {
            retryInfo.textContent = 'Beberapa masalah koneksi';
        } else {
            retryInfo.textContent = 'Koneksi tidak stabil';
        }
    }
}

/**
 * Add new activity to activity list
 * @param {Object} activity Activity to add
 */
function addActivity(activity) {
    const activityList = document.getElementById('activity-list');
    if (!activityList) return;
    
    // Create activity item
    const item = document.createElement('li');
    item.className = 'activity-item';
    
    // Add animation class
    item.classList.add('fade-in');
    
    // Create icon container
    const iconContainer = document.createElement('div');
    iconContainer.className = `activity-icon ${activity.type}`;
    
    // Create icon
    const icon = document.createElement('i');
    icon.className = `fas fa-${activity.icon}`;
    iconContainer.appendChild(icon);
    
    // Create content container
    const contentContainer = document.createElement('div');
    contentContainer.className = 'activity-content';
    
    // Create text
    const text = document.createElement('span');
    text.className = 'activity-text';
    text.textContent = activity.text;
    contentContainer.appendChild(text);
    
    // Create time
    const time = document.createElement('span');
    time.className = 'activity-time';
    time.textContent = activity.time;
    contentContainer.appendChild(time);
    
    // Add elements to item
    item.appendChild(iconContainer);
    item.appendChild(contentContainer);
    
    // Add item to list at the top
    activityList.insertBefore(item, activityList.firstChild);
    
    // Limit list to 10 items
    while (activityList.children.length > 10) {
        activityList.removeChild(activityList.lastChild);
    }
}

/**
 * Clear activity list
 */
function clearActivityList() {
    const activityList = document.getElementById('activity-list');
    if (activityList) {
        activityList.innerHTML = '';
        
        // Add system message about clearing
        addActivity({
            type: 'system',
            icon: 'trash',
            text: 'Daftar aktivitas dibersihkan',
            time: 'Baru saja'
        });
    }
}

/**
 * Load QR code if disconnected
 */
function loadQRCode() {
    const qrContainer = document.getElementById('qr-container');
    const qrLoading = document.getElementById('qr-loading');
    const qrImage = document.getElementById('qr-code');
    
    if (!qrContainer || !qrLoading || !qrImage) return;
    
    // Show loading
    qrLoading.classList.remove('hidden');
    qrImage.classList.add('hidden');
    
    // Get QR code
    fetch('/api/qr/image', {
        headers: {
            'X-Access-Token': apiToken // Use the token passed from the server
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            if (data.success && data.qr_data) {
                // Update QR image
                qrImage.src = `data:image/png;base64,${data.qr_data}`;
                
                // Hide loading, show QR
                qrLoading.classList.add('hidden');
                qrImage.classList.remove('hidden');
            } else {
                // Show error in loading div
                qrLoading.innerHTML = '<i class="fas fa-exclamation-circle"></i><span>QR Code tidak tersedia</span>';
            }
        })
        .catch(error => {
            console.error('Error loading QR code:', error);
            qrLoading.innerHTML = '<i class="fas fa-exclamation-circle"></i><span>Gagal memuat QR Code</span>';
        });
}

/**
 * Refresh QR code
 */
function refreshQRCode() {
    // First trigger reconnect
    fetch('/api/reconnect', {
        method: 'POST',
        headers: {
            'X-Access-Token': apiToken, // Use the token passed from the server
            'Content-Type': 'application/json'
        }
    })
        .then(response => response.json())
        .then(data => {
            // Wait a moment then try to load the QR code
            setTimeout(loadQRCode, 2000);
        })
        .catch(error => {
            console.error('Error refreshing QR code:', error);
        });
}

/**
 * Disconnect WhatsApp
 */
function disconnectWhatsApp() {
    fetch('/api/disconnect', { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            addActivity({
                type: 'disconnected',
                icon: 'power-off',
                text: 'WhatsApp diputuskan secara manual',
                time: 'Baru saja'
            });
            
            // Check status to update UI
            setTimeout(checkStatus, 1000);
        })
        .catch(error => {
            console.error('Error disconnecting WhatsApp:', error);
        });
}

/**
 * Update server time display
 */
function updateServerTime() {
    const serverTimeElement = document.getElementById('server-time');
    if (!serverTimeElement) return;
    
    const now = new Date();
    const hours = now.getHours().toString().padStart(2, '0');
    const minutes = now.getMinutes().toString().padStart(2, '0');
    const seconds = now.getSeconds().toString().padStart(2, '0');
    
    serverTimeElement.textContent = `${hours}:${minutes}:${seconds}`;
}

// Add spinning animation for refresh button
const style = document.createElement('style');
style.textContent = `
    .spinning {
        animation: spin 1s linear infinite;
    }
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    .fade-in {
        animation: fadeIn 0.5s ease-in-out;
    }
    @keyframes fadeIn {
        0% { opacity: 0; transform: translateY(-10px); }
        100% { opacity: 1; transform: translateY(0); }
    }
`;
document.head.appendChild(style);
