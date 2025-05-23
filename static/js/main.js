// Main JavaScript file for WhatsApp Bot Notify

document.addEventListener('DOMContentLoaded', function() {
    console.log('WhatsApp Bot Notify loaded');
    
    // Initialize tooltips
    const tooltips = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    tooltips.map(function (tooltip) {
        return new bootstrap.Tooltip(tooltip);
    });
    
    // Check if we need to refresh the status page automatically
    if (window.location.pathname === '/status') {
        // Refresh status page every 30 seconds
        setTimeout(function() {
            location.reload();
        }, 30000);
    }
});
