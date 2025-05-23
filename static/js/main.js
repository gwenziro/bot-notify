/**
 * Main JavaScript Entry Point
 * Loads and initializes appropriate modules based on current page
 */

document.addEventListener('DOMContentLoaded', function() {
    console.log('WhatsApp Bot Notify loaded');
    
    // Common initialization for all pages
    initCommon();
});

// Initialize common functionality for all pages
function initCommon() {
    // Update server time in footer
    updateServerTime();
    setInterval(updateServerTime, 1000);
    
    // Initialize tooltips
    initTooltips();
    
    // Handle refresh button click
    document.getElementById('refresh-btn')?.addEventListener('click', function() {
        location.reload();
    });
}

// Initialize tooltips
function initTooltips() {
    const tooltips = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    tooltips.forEach(tooltip => {
        if (typeof bootstrap !== 'undefined' && bootstrap.Tooltip) {
            new bootstrap.Tooltip(tooltip);
        }
    });
}

// Initialize status page functionality
function initStatusPage() {
    // Check status immediately
    checkConnectionStatus();
    
    // Setup refresh interval
    setInterval(() => {
        // Use AJAX to refresh only the status information
        fetch('/api/status')
            .then(response => response.json())
            .then(data => {
                // Update status display without reload
                updateConnectionUI(data);
                
                // Update specific status page elements
                const statusBadge = document.querySelector('.status-badge');
                if (statusBadge) {
                    statusBadge.className = 'status-badge status-' + data.status;
                    statusBadge.textContent = getStatusText(data.status);
                }
                
                // Update last activity
                const lastActivity = document.getElementById('last-activity');
                if (lastActivity && data.details) {
                    lastActivity.textContent = formatDateTime(data.details.lastActivity);
                }
            })
            .catch(error => console.error('Error refreshing status:', error));
    }, 30000); // Refresh every 30 seconds
}
