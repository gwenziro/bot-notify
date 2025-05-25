/**
 * Header Component JavaScript
 * Menangani fungsi dropdown, notifikasi, dan aksi lainnya di header
 */

document.addEventListener('DOMContentLoaded', function() {
    initializeHeader();
});

function initializeHeader() {
    // Setup dropdown toggles
    setupDropdowns();
    
    // Setup refresh button
    setupRefreshButton();
    
    // Setup notifications
    setupNotifications();
    
    // Update WhatsApp status indicator
    updateConnectionIndicator();
}

function setupDropdowns() {
    // User dropdown
    const userDropdownToggle = document.getElementById('user-dropdown-toggle');
    const userDropdownMenu = document.getElementById('user-dropdown-menu');
    
    if (userDropdownToggle && userDropdownMenu) {
        userDropdownToggle.addEventListener('click', function(e) {
            e.stopPropagation();
            toggleDropdown(userDropdownMenu);
            
            // Close other dropdowns
            const notificationsMenu = document.getElementById('notifications-menu');
            if (notificationsMenu && notificationsMenu.classList.contains('show')) {
                notificationsMenu.classList.remove('show');
            }
        });
    }
    
    // Notifications dropdown
    const notificationsBtn = document.getElementById('notifications-btn');
    const notificationsMenu = document.getElementById('notifications-menu');
    
    if (notificationsBtn && notificationsMenu) {
        notificationsBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            toggleDropdown(notificationsMenu);
            
            // Close other dropdowns
            if (userDropdownMenu && userDropdownMenu.classList.contains('show')) {
                userDropdownMenu.classList.remove('show');
            }
        });
        
        // Clear all notifications
        const clearNotificationsBtn = document.getElementById('clear-notifications');
        if (clearNotificationsBtn) {
            clearNotificationsBtn.addEventListener('click', function(e) {
                e.stopPropagation();
                clearNotifications();
            });
        }
    }
    
    // Close all dropdowns when clicking elsewhere
    document.addEventListener('click', function() {
        const dropdowns = document.querySelectorAll('.dropdown-menu');
        dropdowns.forEach(menu => {
            if (menu && menu.classList.contains('show')) {
                menu.classList.remove('show');
            }
        });
    });
    
    // Prevent closing when clicking inside dropdown menu
    const dropdownMenus = document.querySelectorAll('.dropdown-menu');
    dropdownMenus.forEach(menu => {
        menu.addEventListener('click', function(e) {
            e.stopPropagation();
        });
    });
}

function toggleDropdown(dropdownMenu) {
    if (dropdownMenu) {
        dropdownMenu.classList.toggle('show');
    }
}

function setupRefreshButton() {
    const refreshBtn = document.getElementById('refresh-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            // Add animation
            this.classList.add('fa-spin');
            
            // Refresh data or reload page after delay
            setTimeout(() => {
                location.reload();
            }, 500);
        });
    }
}

function setupNotifications() {
    // Load notifications from storage
    loadNotifications();
    
    // Update notification badge
    updateNotificationBadge();
}

function loadNotifications() {
    const notificationsList = document.getElementById('notifications-list');
    if (!notificationsList) return;
    
    // Get notifications from local storage
    const notifications = JSON.parse(localStorage.getItem('notifications') || '[]');
    
    // Clear current notifications
    notificationsList.innerHTML = '';
    
    // Show empty state if no notifications
    if (notifications.length === 0) {
        notificationsList.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-bell-slash"></i>
                <p>Tidak ada notifikasi baru</p>
            </div>
        `;
        return;
    }
    
    // Add notifications to the list
    notifications.forEach((notification, index) => {
        const notificationItem = document.createElement('div');
        notificationItem.className = 'dropdown-item notification-item';
        notificationItem.dataset.id = index;
        
        let iconClass = 'fas fa-info-circle';
        let iconColor = 'var(--color-info)';
        
        // Set icon based on type
        switch (notification.type) {
            case 'success':
                iconClass = 'fas fa-check-circle';
                iconColor = 'var(--color-success)';
                break;
            case 'warning':
                iconClass = 'fas fa-exclamation-triangle';
                iconColor = 'var(--color-warning)';
                break;
            case 'error':
                iconClass = 'fas fa-times-circle';
                iconColor = 'var(--color-danger)';
                break;
        }
        
        notificationItem.innerHTML = `
            <div class="notification-icon" style="color: ${iconColor}">
                <i class="${iconClass}"></i>
            </div>
            <div class="notification-content">
                <div class="notification-text">${notification.message}</div>
                <div class="notification-time">${formatTimeAgo(notification.time)}</div>
            </div>
            <button class="notification-dismiss" data-id="${index}">
                <i class="fas fa-times"></i>
            </button>
        `;
        
        notificationsList.appendChild(notificationItem);
        
        // Add dismiss handler
        const dismissBtn = notificationItem.querySelector('.notification-dismiss');
        if (dismissBtn) {
            dismissBtn.addEventListener('click', function(e) {
                e.stopPropagation();
                dismissNotification(this.dataset.id);
            });
        }
    });
}

function updateNotificationBadge() {
    const badge = document.getElementById('notifications-count');
    if (!badge) return;
    
    const notifications = JSON.parse(localStorage.getItem('notifications') || '[]');
    
    if (notifications.length > 0) {
        badge.textContent = notifications.length;
        badge.style.display = 'flex';
    } else {
        badge.style.display = 'none';
    }
}

function dismissNotification(id) {
    // Get notifications from storage
    let notifications = JSON.parse(localStorage.getItem('notifications') || '[]');
    
    // Remove notification
    notifications = notifications.filter((_, index) => index !== parseInt(id));
    
    // Save back to storage
    localStorage.setItem('notifications', JSON.stringify(notifications));
    
    // Reload notifications and update badge
    loadNotifications();
    updateNotificationBadge();
}

function clearNotifications() {
    // Clear notifications in storage
    localStorage.removeItem('notifications');
    
    // Reload notifications and update badge
    loadNotifications();
    updateNotificationBadge();
}

function formatTimeAgo(timestamp) {
    const now = new Date();
    const date = new Date(timestamp);
    const seconds = Math.floor((now - date) / 1000);
    
    if (seconds < 60) {
        return 'baru saja';
    }
    
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return `${minutes} menit yang lalu`;
    }
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
        return `${hours} jam yang lalu`;
    }
    
    const days = Math.floor(hours / 24);
    if (days < 7) {
        return `${days} hari yang lalu`;
    }
    
    // Format as date for older notifications
    return date.toLocaleDateString('id-ID', { 
        day: 'numeric', 
        month: 'short', 
        year: 'numeric' 
    });
}

function updateConnectionIndicator() {
    const indicator = document.getElementById('whatsapp-status-indicator');
    const statusText = document.getElementById('whatsapp-status-text');
    
    if (!indicator || !statusText) return;
    
    // Check connection status (can be replaced with real API call)
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            if (data.status === 'connected') {
                indicator.className = 'status-indicator online';
                statusText.textContent = 'Terhubung';
                statusText.className = 'status-text';
            } else if (data.status === 'connecting') {
                indicator.className = 'status-indicator connecting';
                statusText.textContent = 'Menghubungkan...';
                statusText.className = 'status-text';
            } else {
                indicator.className = 'status-indicator offline';
                statusText.textContent = 'Terputus';
                statusText.className = 'status-text';
            }
        })
        .catch(error => {
            console.error('Error checking connection status:', error);
            indicator.className = 'status-indicator offline';
            statusText.textContent = 'Error';
            statusText.className = 'status-text';
        });
}

// Add a notification (can be called from other components)
window.addNotification = function(message, type = 'info') {
    let notifications = JSON.parse(localStorage.getItem('notifications') || '[]');
    
    notifications.unshift({
        message,
        type,
        time: new Date().toISOString()
    });
    
    // Limit to 20 notifications
    if (notifications.length > 20) {
        notifications = notifications.slice(0, 20);
    }
    
    localStorage.setItem('notifications', JSON.stringify(notifications));
    
    // Update UI if needed
    if (document.getElementById('notifications-list')) {
        loadNotifications();
        updateNotificationBadge();
    }
};
