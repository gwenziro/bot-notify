/**
 * Dashboard JavaScript
 * Handles all dashboard functionality
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize dashboard
    initDashboard();
    
    // Setup event listeners
    setupEventListeners();
    
    // Update data periodically
    startDataUpdates();
    
    // Update server time
    updateServerTime();
    setInterval(updateServerTime, 1000);
    
    // If QR view is showing, load QR code
    if (document.getElementById('qr-view')) {
        loadQRCode();
    }
    
    // Animate components entry
    animateComponentsEntry();
});

/**
 * Initialize dashboard components
 */
function initDashboard() {
    console.log('Dashboard initialized');
    
    // Fetch initial data for connected view
    if (document.getElementById('connected-view')) {
        updateConnectionStatus();
        // Initialize uptime counter
        startUptimeCounter();
    }
}

/**
 * Setup all event listeners
 */
function setupEventListeners() {
    // Refresh button
    const refreshBtn = document.getElementById('refresh-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            // Add spinner to indicate refresh
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
            // Reload the page after a short delay
            setTimeout(() => {
                location.reload();
            }, 500);
        });
    }
    
    // Disconnect button
    const disconnectBtn = document.getElementById('disconnect-btn');
    if (disconnectBtn) {
        disconnectBtn.addEventListener('click', function() {
            showConfirmation(
                'Konfirmasi Pemutusan',
                'Apakah Anda yakin ingin memutuskan koneksi WhatsApp? Bot tidak akan dapat mengirim pesan hingga Anda menghubungkan kembali.',
                () => {
                    disconnectWhatsApp(this);
                }
            );
        });
    }
    
    // Refresh QR button
    const refreshQrBtn = document.getElementById('refresh-qr-btn');
    if (refreshQrBtn) {
        refreshQrBtn.addEventListener('click', function() {
            loadQRCode();
        });
    }
    
    // Copy API example button
    const copyApiBtn = document.getElementById('copy-api-btn');
    if (copyApiBtn) {
        copyApiBtn.addEventListener('click', function() {
            const codeBlock = document.querySelector('.code-block code');
            if (codeBlock) {
                copyToClipboard(codeBlock.innerText);
                
                // Visual feedback
                this.innerHTML = '<i class="fas fa-check"></i>';
                setTimeout(() => {
                    this.innerHTML = '<i class="fas fa-copy"></i>';
                }, 2000);
                
                // Show floating notification
                showNotification('Kode contoh berhasil disalin!', 'success');
            }
        });
    }
    
    // Clear activity button
    const clearActivityBtn = document.getElementById('clear-activity-btn');
    if (clearActivityBtn) {
        clearActivityBtn.addEventListener('click', function() {
            clearActivityList();
        });
    }
}

/**
 * Start periodic data updates
 */
function startDataUpdates() {
    // Update connection status every 30 seconds
    setInterval(() => {
        updateConnectionStatus();
    }, 30000);
}

/**
 * Update server time display
 */
function updateServerTime() {
    const serverTimeEl = document.getElementById('server-time');
    if (serverTimeEl) {
        const now = new Date();
        serverTimeEl.textContent = now.toLocaleTimeString();
    }
}

/**
 * Start uptime counter that increases every second
 */
function startUptimeCounter() {
    const uptimeEl = document.getElementById('uptime-value');
    if (!uptimeEl) return;
    
    // Get initial uptime from element
    let uptimeText = uptimeEl.textContent;
    let uptime = parseUptime(uptimeText);
    
    // Update every second
    setInterval(() => {
        uptime += 1; // Add one second
        uptimeEl.textContent = formatUptime(uptime);
        
        // Update progress bar if exists
        const progressFill = document.querySelector('.progress-fill');
        if (progressFill) {
            // Calculate percentage for 24 hours max (86400 seconds)
            const percentage = Math.min(100, (uptime / 86400) * 100);
            progressFill.style.width = `${percentage}%`;
            
            // Update percentage text
            const uptimePercent = document.getElementById('uptime-percent');
            if (uptimePercent) {
                uptimePercent.textContent = `${Math.round(percentage)}%`;
            }
        }
    }, 1000);
}

/**
 * Parse uptime string into seconds
 * @param {string} uptimeText - Uptime in format "XXh XXm XXs" or "XXd XXh XXm XXs"
 * @returns {number} - Total seconds
 */
function parseUptime(uptimeText) {
    let totalSeconds = 0;
    
    // Parse days if present
    const daysMatch = uptimeText.match(/(\d+)d/);
    if (daysMatch) {
        totalSeconds += parseInt(daysMatch[1], 10) * 86400;
    }
    
    // Parse hours
    const hoursMatch = uptimeText.match(/(\d+)h/);
    if (hoursMatch) {
        totalSeconds += parseInt(hoursMatch[1], 10) * 3600;
    }
    
    // Parse minutes
    const minutesMatch = uptimeText.match(/(\d+)m/);
    if (minutesMatch) {
        totalSeconds += parseInt(minutesMatch[1], 10) * 60;
    }
    
    // Parse seconds
    const secondsMatch = uptimeText.match(/(\d+)s/);
    if (secondsMatch) {
        totalSeconds += parseInt(secondsMatch[1], 10);
    }
    
    return totalSeconds;
}

/**
 * Format seconds into readable uptime string
 * @param {number} seconds - Total seconds
 * @returns {string} - Formatted uptime
 */
function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (days > 0) {
        return `${days}d ${hours}h ${minutes}m ${secs}s`;
    }
    
    return `${hours}h ${minutes}m ${secs}s`;
}

/**
 * Update connection status from API
 */
function updateConnectionStatus() {
    fetch('/api/status', {
        headers: {
            'X-Access-Token': getAuthToken(),
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        updateUIWithConnectionData(data);
    })
    .catch(error => {
        console.error('Error fetching connection status:', error);
        showNotification('Gagal memperbarui status koneksi', 'error');
    });
}

/**
 * Update UI elements with connection data
 * @param {Object} data - Connection data from API
 */
function updateUIWithConnectionData(data) {
    const connectionBadge = document.getElementById('connection-badge');
    
    if (data && data.is_connected) {
        // Update connected info elements if they exist
        updateElementText('connection-status', data.status || 'Unknown');
        updateElementText('connected-since', formatDateTime(data.connected_since) || 'Unknown');
        updateElementText('device-name', data.device_name || 'Unknown');
        updateElementText('messages-sent', data.messages_sent || '0');
        updateElementText('groups-count', data.groups_count || '0');
        updateElementText('retry-count', data.connection_retries || '0');
        
        // Update connection badge
        if (connectionBadge) {
            connectionBadge.className = 'status-badge connected';
            connectionBadge.textContent = 'Terhubung';
        }
        
        // Update groups info text
        const groupsInfo = document.getElementById('groups-info');
        if (groupsInfo) {
            groupsInfo.textContent = data.groups_count > 0 
                ? `${data.groups_count} grup siap untuk pengiriman` 
                : 'Tidak ada grup';
        }
        
        // Update retry info text
        const retryInfo = document.getElementById('retry-info');
        if (retryInfo) {
            retryInfo.textContent = data.connection_retries > 3 
                ? 'Koneksi tidak stabil' 
                : 'Koneksi stabil';
        }
    } else {
        // If we're showing the connected view but we're actually disconnected
        // we should reload the page to show QR code
        const connectedInfo = document.getElementById('connected-view');
        if (connectedInfo && !connectedInfo.classList.contains('hidden')) {
            location.reload();
        }
        
        // Update connection badge if we're just showing disconnect status
        if (connectionBadge) {
            connectionBadge.className = 'status-badge disconnected';
            connectionBadge.textContent = 'Terputus';
        }
    }
}

/**
 * Load QR Code from API
 */
function loadQRCode() {
    const qrContainer = document.getElementById('qr-container');
    const qrImage = document.getElementById('qr-code');
    const qrLoading = document.getElementById('qr-loading');
    
    if (!qrContainer || !qrImage || !qrLoading) return;
    
    // Show loading
    qrImage.classList.add('hidden');
    qrLoading.classList.remove('hidden');
    
    // Get QR code from API with cache-busting
    fetch('/api/qr/image?t=' + new Date().getTime(), {
        headers: {
            'X-Access-Token': getAuthToken()
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to load QR code');
        }
        return response.blob();
    })
    .then(blob => {
        const imageUrl = URL.createObjectURL(blob);
        qrImage.src = imageUrl;
        
        // Hide loading, show QR
        qrLoading.classList.add('hidden');
        qrImage.classList.remove('hidden');
        
        // Add QR loaded to activity
        addActivity('QR Code dimuat', 'system');
    })
    .catch(error => {
        console.error('Error loading QR code:', error);
        qrLoading.innerHTML = `
            <i class="fas fa-exclamation-circle"></i>
            <span>Gagal memuat QR. <br>Silakan coba lagi.</span>
        `;
        
        showNotification('Gagal memuat kode QR', 'error');
    });
}

/**
 * Disconnect WhatsApp
 * @param {HTMLElement} button - The disconnect button
 */
function disconnectWhatsApp(button) {
    // Disable button and show loading
    if (button) {
        button.disabled = true;
        button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memutuskan...';
    }
    
    fetch('/api/disconnect', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Access-Token': getAuthToken()
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to disconnect');
        }
        return response.json();
    })
    .then(data => {
        showNotification('Berhasil memutuskan koneksi WhatsApp', 'success');
        
        // Reload page to show QR code after short delay
        setTimeout(() => {
            location.reload();
        }, 1000);
    })
    .catch(error => {
        console.error('Error disconnecting:', error);
        showNotification('Gagal memutuskan koneksi', 'error');
        
        // Reset button
        if (button) {
            button.disabled = false;
            button.innerHTML = '<i class="fas fa-power-off"></i> Putuskan Koneksi';
        }
    });
}

/**
 * Copy text to clipboard
 * @param {string} text - Text to copy
 */
function copyToClipboard(text) {
    // Use modern clipboard API if available
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
            console.log('Text copied to clipboard');
        }).catch(err => {
            console.error('Failed to copy text: ', err);
            fallbackCopyToClipboard(text);
        });
    } else {
        fallbackCopyToClipboard(text);
    }
}

/**
 * Fallback method to copy text to clipboard
 * @param {string} text - Text to copy
 */
function fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        const successful = document.execCommand('copy');
        console.log(successful ? 'Success' : 'Failed');
    } catch (err) {
        console.error('Fallback: Oops, unable to copy', err);
    }
    
    document.body.removeChild(textArea);
}

/**
 * Format date/time for display
 * @param {string|number|Date} timestamp - Date to format
 * @returns {string} Formatted date
 */
function formatDateTime(timestamp) {
    if (!timestamp) return 'Unknown';
    
    const date = new Date(timestamp);
    return date.toLocaleString();
}

/**
 * Update element text if element exists
 * @param {string} elementId - Element ID
 * @param {string} text - Text content
 */
function updateElementText(elementId, text) {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = text;
    }
}

/**
 * Show confirmation dialog
 * @param {string} title - Dialog title
 * @param {string} message - Dialog message
 * @param {Function} onConfirm - Callback for confirmation
 */
function showConfirmation(title, message, onConfirm) {
    // Check if we already have a modal container
    let modalContainer = document.querySelector('.modal-container');
    if (!modalContainer) {
        // Create modal container
        modalContainer = document.createElement('div');
        modalContainer.className = 'modal-container';
        document.body.appendChild(modalContainer);
        
        // Add styles for modal
        const style = document.createElement('style');
        style.textContent = `
            .modal-container {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.5);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 1000;
                animation: fadeIn 0.3s ease-out;
            }
            .modal-content {
                background: var(--glass-bg);
                backdrop-filter: blur(12px);
                -webkit-backdrop-filter: blur(12px);
                border: 1px solid var(--glass-border);
                border-radius: var(--radius-lg);
                width: 90%;
                max-width: 400px;
                padding: 0;
                box-shadow: 0 15px 40px rgba(0, 0, 0, 0.3);
                animation: scaleIn 0.3s ease-out;
            }
            .modal-header {
                padding: var(--space-4);
                border-bottom: 1px solid var(--glass-border);
            }
            .modal-header h3 {
                margin: 0;
                font-size: 1.25rem;
            }
            .modal-body {
                padding: var(--space-4);
            }
            .modal-footer {
                padding: var(--space-3) var(--space-4);
                display: flex;
                justify-content: flex-end;
                gap: var(--space-2);
                border-top: 1px solid var(--glass-border);
            }
            @keyframes fadeIn {
                from { opacity: 0; }
                to { opacity: 1; }
            }
            @keyframes scaleIn {
                from { transform: scale(0.95); opacity: 0; }
                to { transform: scale(1); opacity: 1; }
            }
        `;
        document.head.appendChild(style);
    }
    
    // Clear any existing modals
    modalContainer.innerHTML = '';
    
    // Create modal content
    const modalContent = document.createElement('div');
    modalContent.className = 'modal-content';
    modalContent.innerHTML = `
        <div class="modal-header">
            <h3>${title}</h3>
        </div>
        <div class="modal-body">
            <p>${message}</p>
        </div>
        <div class="modal-footer">
            <button class="btn btn-outline" id="modal-cancel">Batal</button>
            <button class="btn btn-danger" id="modal-confirm">Konfirmasi</button>
        </div>
    `;
    
    modalContainer.appendChild(modalContent);
    
    // Show modal
    modalContainer.style.display = 'flex';
    
    // Add event listeners
    document.getElementById('modal-cancel').addEventListener('click', function() {
        modalContainer.style.display = 'none';
    });
    
    document.getElementById('modal-confirm').addEventListener('click', function() {
        modalContainer.style.display = 'none';
        onConfirm();
    });
    
    // Close when clicking outside
    modalContainer.addEventListener('click', function(e) {
        if (e.target === modalContainer) {
            modalContainer.style.display = 'none';
        }
    });
}

/**
 * Show notification toast
 * @param {string} message - Notification message
 * @param {string} type - Notification type (success, error, info)
 */
function showNotification(message, type = 'info') {
    // Check if we already have notifications container
    let notificationsContainer = document.querySelector('.notifications-container');
    if (!notificationsContainer) {
        // Create notifications container
        notificationsContainer = document.createElement('div');
        notificationsContainer.className = 'notifications-container';
        document.body.appendChild(notificationsContainer);
        
        // Add styles for notifications
        const style = document.createElement('style');
        style.textContent = `
            .notifications-container {
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 1000;
                display: flex;
                flex-direction: column;
                gap: 10px;
            }
            .notification {
                padding: 12px 20px;
                border-radius: 8px;
                background: var(--glass-bg);
                backdrop-filter: blur(12px);
                -webkit-backdrop-filter: blur(12px);
                border: 1px solid var(--glass-border);
                color: white;
                min-width: 250px;
                box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
                display: flex;
                align-items: center;
                gap: 12px;
                animation: slideIn 0.3s ease-out forwards;
                opacity: 0;
                transform: translateX(50px);
            }
            .notification.success {
                border-left: 4px solid var(--success);
            }
            .notification.error {
                border-left: 4px solid var(--danger);
            }
            .notification.info {
                border-left: 4px solid var(--info);
            }
            .notification-icon {
                font-size: 18px;
            }
            .notification.success .notification-icon {
                color: var(--success);
            }
            .notification.error .notification-icon {
                color: var(--danger);
            }
            .notification.info .notification-icon {
                color: var(--info);
            }
            .notification-message {
                flex: 1;
            }
            @keyframes slideIn {
                to {
                    opacity: 1;
                    transform: translateX(0);
                }
            }
            @keyframes slideOut {
                to {
                    opacity: 0;
                    transform: translateX(100px);
                }
            }
        `;
        document.head.appendChild(style);
    }
    
    // Create notification
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    
    // Set icon based on type
    let icon = 'info-circle';
    if (type === 'success') icon = 'check-circle';
    if (type === 'error') icon = 'exclamation-circle';
    
    notification.innerHTML = `
        <div class="notification-icon">
            <i class="fas fa-${icon}"></i>
        </div>
        <div class="notification-message">${message}</div>
    `;
    
    // Add to container
    notificationsContainer.appendChild(notification);
    
    // Remove after 5 seconds
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease-out forwards';
        setTimeout(() => {
            if (notification.parentNode === notificationsContainer) {
                notificationsContainer.removeChild(notification);
            }
        }, 300);
    }, 5000);
}

/**
 * Add activity to activity list
 * @param {string} text - Activity description
 * @param {string} type - Activity type (connected, disconnected, message, system)
 */
function addActivity(text, type = 'system') {
    const activityList = document.getElementById('activity-list');
    if (!activityList) return;
    
    // Create activity item
    const activityItem = document.createElement('li');
    activityItem.className = 'activity-item fade-in';
    
    // Set icon based on type
    let icon = 'info-circle';
    if (type === 'connected') icon = 'plug';
    if (type === 'disconnected') icon = 'power-off';
    if (type === 'message') icon = 'paper-plane';
    
    activityItem.innerHTML = `
        <div class="activity-icon ${type}">
            <i class="fas fa-${icon}"></i>
        </div>
        <div class="activity-content">
            <span class="activity-text">${text}</span>
            <span class="activity-time">Baru saja</span>
        </div>
    `;
    
    // Add to list at the beginning
    activityList.insertBefore(activityItem, activityList.firstChild);
    
    // Limit to 10 items
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
        // Add system activity about clearing
        const clearItem = document.createElement('li');
        clearItem.className = 'activity-item fade-in';
        clearItem.innerHTML = `
            <div class="activity-icon system">
                <i class="fas fa-eraser"></i>
            </div>
            <div class="activity-content">
                <span class="activity-text">Aktivitas dibersihkan</span>
                <span class="activity-time">Baru saja</span>
            </div>
        `;
        
        // Clear list and add clear message
        activityList.innerHTML = '';
        activityList.appendChild(clearItem);
        
        showNotification('Daftar aktivitas telah dibersihkan', 'success');
    }
}

/**
 * Get auth token from storage
 * @returns {string} Auth token
 */
function getAuthToken() {
    return localStorage.getItem('access_token') || 
           sessionStorage.getItem('access_token') || 
           '';
}

/**
 * Animate components entry
 */
function animateComponentsEntry() {
    const elements = [
        '.dashboard-header',
        '.status-card',
        '.stat-card',
        '.activity-card',
        '.info-card',
        '.dashboard-footer'
    ];
    
    elements.forEach((selector, index) => {
        const elements = document.querySelectorAll(selector);
        elements.forEach((el) => {
            el.style.opacity = '0';
            el.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                el.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                el.style.opacity = '1';
                el.style.transform = 'translateY(0)';
            }, 100 + (index * 150));
        });
    });
}
