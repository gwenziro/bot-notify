/**
 * UI Utilities
 * Common UI functions and helpers
 */

// Show a temporary message that automatically disappears after a delay
function showTemporaryMessage(element, message, type = 'success', duration = 3000) {
    if (!element) return;
    
    // Store original content
    const originalContent = element.innerHTML;
    
    // Set classes based on type
    element.className = `alert alert-${type}`;
    element.innerHTML = message;
    element.style.display = 'block';
    
    // Hide after duration
    setTimeout(() => {
        // Fade out
        element.style.opacity = '0';
        setTimeout(() => {
            element.style.display = 'none';
            element.style.opacity = '1';
            // Restore if needed (commented out by default)
            // element.innerHTML = originalContent;
        }, 300);
    }, duration);
}

// Format bytes to human readable format
function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

// Get URL parameter by name
function getUrlParam(name) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(name);
}

// Create confirmation dialog
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// Update connection status UI elements
function updateConnectionUI(data) {
    const isConnected = data.details?.isConnected || false;
    const status = data.status || 'disconnected';
    
    // Elements to update
    const indicators = document.querySelectorAll('.status-indicator');
    const statusTexts = document.querySelectorAll('.status-text');
    const connectivityStatus = document.getElementById('connectivity-status');
    const sidebarStatusText = document.getElementById('sidebar-connection-text');
    
    // Update all status indicators
    indicators.forEach(indicator => {
        indicator.classList.remove('online', 'offline', 'connecting');
        indicator.classList.add(isConnected ? 'online' : 'offline');
    });
    
    // Update status text
    statusTexts.forEach(text => {
        if (text.id === 'whatsapp-status-text') {
            text.textContent = getStatusText(status);
        }
    });
    
    // Update sidebar connection text
    if (sidebarStatusText) {
        sidebarStatusText.textContent = isConnected ? 'WhatsApp Terhubung' : 'WhatsApp Terputus';
    }
    
    // Update connectivity status badge
    if (connectivityStatus) {
        connectivityStatus.classList.remove('online', 'offline', 'connecting');
        connectivityStatus.classList.add(isConnected ? 'online' : 'offline');
    }
}

// Get human-readable status text based on status code
function getStatusText(status) {
    const statusMap = {
        'connected': 'Terhubung',
        'connecting': 'Menghubungkan...',
        'disconnected': 'Terputus',
        'logged_out': 'Keluar'
    };
    
    return statusMap[status] || 'Tidak Diketahui';
}
