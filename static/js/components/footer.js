/**
 * Footer Component JavaScript
 * Menangani fitur seperti server time dan uptime display
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize server time
    updateServerTime();
    
    // Start server time updater
    setInterval(updateServerTime, 1000);
    
    // Initialize system uptime
    updateSystemUptime();
    
    // Start system uptime updater
    setInterval(updateSystemUptime, 1000);
});

// Global variables to track uptime
let startTime = sessionStorage.getItem('systemStartTime') ? 
    new Date(sessionStorage.getItem('systemStartTime')) : new Date();

// Initialize start time if not already set
if (!sessionStorage.getItem('systemStartTime')) {
    sessionStorage.setItem('systemStartTime', startTime.toISOString());
}

/**
 * Update server time display in footer
 */
function updateServerTime() {
    const serverTimeElement = document.getElementById('server-time');
    if (!serverTimeElement) return;
    
    const now = new Date();
    const formattedTime = formatTime(now);
    
    serverTimeElement.textContent = formattedTime;
}

/**
 * Update system uptime display in footer
 */
function updateSystemUptime() {
    const uptimeElement = document.getElementById('system-uptime');
    if (!uptimeElement) return;
    
    const uptime = getUptime();
    uptimeElement.textContent = `Uptime: ${uptime}`;
}

/**
 * Format date as time string (HH:MM:SS)
 */
function formatTime(date) {
    return date.toLocaleTimeString('id-ID', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
}

/**
 * Calculate and format uptime
 */
function getUptime() {
    const now = new Date();
    const diff = now - startTime;
    
    // Convert to seconds
    let seconds = Math.floor(diff / 1000);
    
    // Calculate hours, minutes, and remaining seconds
    const hours = Math.floor(seconds / 3600);
    seconds %= 3600;
    const minutes = Math.floor(seconds / 60);
    seconds %= 60;
    
    // Format with leading zeros
    const formattedHours = String(hours).padStart(2, '0');
    const formattedMinutes = String(minutes).padStart(2, '0');
    const formattedSeconds = String(seconds).padStart(2, '0');
    
    return `${formattedHours}:${formattedMinutes}:${formattedSeconds}`;
}

/**
 * Format date as ISO string
 */
function formatISODate(date) {
    return date.toISOString().slice(0, 19).replace('T', ' ');
}
