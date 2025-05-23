/**
 * Date Time Utilities
 * Functions for working with dates and times
 */

// Format time for display (HH:MM:SS)
function formatTime(time) {
    if (!time) return '';
    
    const date = time instanceof Date ? time : new Date(time);
    
    return date.toLocaleTimeString('id-ID', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// Format date time for display (DD MMM YYYY HH:MM:SS)
function formatDateTime(time) {
    if (!time) return '';
    
    const date = time instanceof Date ? time : new Date(time);
    
    return date.toLocaleDateString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// Format date for display (DD MMM YYYY)
function formatDate(time) {
    if (!time) return '';
    
    const date = time instanceof Date ? time : new Date(time);
    
    return date.toLocaleDateString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric'
    });
}

// Calculate time difference in human readable format
function timeDifference(current, previous) {
    const msPerMinute = 60 * 1000;
    const msPerHour = msPerMinute * 60;
    const msPerDay = msPerHour * 24;
    const msPerMonth = msPerDay * 30;
    const msPerYear = msPerDay * 365;
    
    const elapsed = current - previous;
    
    if (elapsed < msPerMinute) {
        return Math.round(elapsed/1000) + ' detik yang lalu';   
    }
    else if (elapsed < msPerHour) {
        return Math.round(elapsed/msPerMinute) + ' menit yang lalu';   
    }
    else if (elapsed < msPerDay) {
        return Math.round(elapsed/msPerHour) + ' jam yang lalu';   
    }
    else if (elapsed < msPerMonth) {
        return Math.round(elapsed/msPerDay) + ' hari yang lalu';   
    }
    else if (elapsed < msPerYear) {
        return Math.round(elapsed/msPerMonth) + ' bulan yang lalu';   
    }
    else {
        return Math.round(elapsed/msPerYear) + ' tahun yang lalu';   
    }
}

// Calculate uptime from start time
function calculateUptime(startTime) {
    const now = new Date();
    const start = new Date(startTime);
    const diff = Math.floor((now - start) / 1000); // seconds
    
    const hours = Math.floor(diff / 3600);
    const minutes = Math.floor((diff % 3600) / 60);
    const seconds = diff % 60;
    
    return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}
