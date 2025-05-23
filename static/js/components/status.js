/**
 * Status Component
 * Handles connection status indicators and updates
 */

document.addEventListener('DOMContentLoaded', function() {
    initConnectionStatus();
    
    // Connection status check interval
    setInterval(checkConnectionStatus, 30000); // Refresh every 30 seconds
});

// Initialize connection status indicators
function initConnectionStatus() {
    // Immediately check connection status on page load
    checkConnectionStatus();
}

// Check WhatsApp connection status via API
function checkConnectionStatus() {
    getConnectionStatus()
        .then(data => {
            updateConnectionUI(data);
        })
        .catch(error => {
            console.error('Error checking connection status:', error);
        });
}

// Update page title based on current page
function updatePageTitle() {
    const currentPath = window.location.pathname;
    const pageTitleElement = document.getElementById('current-page-title');
    
    if (!pageTitleElement) return;
    
    let title = 'Dashboard';
    
    // Map paths to titles
    const pageTitles = {
        '/': 'Beranda',
        '/dashboard': 'Dashboard',
        '/connectivity': 'Konektivitas',
        '/messages': 'Pesan',
        '/groups': 'Grup',
        '/settings': 'Pengaturan',
        '/help': 'Bantuan'
    };
    
    if (pageTitles[currentPath]) {
        title = pageTitles[currentPath];
    }
    
    pageTitleElement.textContent = title;
    document.title = `${title} - Bot Notify`;
}

// Update server time display in footer
function updateServerTime() {
    const serverTimeElement = document.getElementById('server-time');
    if (serverTimeElement) {
        const now = new Date();
        const timeString = formatTime(now);
        serverTimeElement.textContent = timeString;
    }
}
