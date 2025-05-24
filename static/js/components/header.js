/**
 * Header Component JavaScript
 * Handles header interactivity including search, dropdowns, notifications, and connection status
 */

class HeaderComponent {
    constructor() {
        // Elements
        this.headerElement = document.querySelector('.dashboard-header');
        this.statusIndicator = document.getElementById('whatsapp-status-indicator');
        this.statusText = document.getElementById('whatsapp-status-text');
        this.refreshBtn = document.getElementById('refresh-btn');
        this.notificationsBtn = document.getElementById('notifications-btn');
        this.notificationsCount = document.getElementById('notifications-count');
        this.notificationsMenu = document.getElementById('notifications-menu');
        this.notificationsList = document.getElementById('notifications-list');
        this.clearNotificationsBtn = document.getElementById('clear-notifications');
        this.userProfileBtn = document.querySelector('.user-profile');
        this.userDropdown = document.querySelector('.user-dropdown .dropdown-menu');
        
        // Initialize component
        this.init();
    }
    
    init() {
        // Setup event listeners
        this.setupEventListeners();
        
        // Setup header search
        this.setupHeaderSearch();
        
        // Load existing notifications
        this.loadNotifications();
        
        // Initialize scroll handling
        this.handleHeaderScroll();
    }
    
    setupEventListeners() {
        // Refresh button
        if (this.refreshBtn) {
            this.refreshBtn.addEventListener('click', () => {
                this.refreshBtn.classList.add('fa-spin');
                setTimeout(() => {
                    window.location.reload();
                }, 600);
            });
        }
        
        // Notifications button
        if (this.notificationsBtn && this.notificationsMenu) {
            this.notificationsBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.toggleDropdown(this.notificationsMenu);
            });
        }
        
        // Clear notifications button
        if (this.clearNotificationsBtn) {
            this.clearNotificationsBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.clearNotifications();
            });
        }
        
        // User profile dropdown
        if (this.userProfileBtn && this.userDropdown) {
            this.userProfileBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.toggleDropdown(this.userDropdown);
            });
        }
        
        // Close dropdowns when clicking outside
        document.addEventListener('click', () => {
            this.closeAllDropdowns();
        });
        
        // Prevent clicks inside dropdowns from closing them
        const dropdowns = document.querySelectorAll('.dropdown-menu');
        dropdowns.forEach(dropdown => {
            dropdown.addEventListener('click', (e) => {
                e.stopPropagation();
            });
        });
        
        // Handle window scroll
        window.addEventListener('scroll', () => {
            this.handleHeaderScroll();
        });
    }
    
    toggleDropdown(dropdown) {
        // Close all other dropdowns first
        this.closeAllDropdowns();
        
        // Toggle the target dropdown
        dropdown.classList.toggle('show');
    }
    
    closeAllDropdowns() {
        const dropdowns = document.querySelectorAll('.dropdown-menu');
        dropdowns.forEach(dropdown => {
            dropdown.classList.remove('show');
        });
    }
    
    handleHeaderScroll() {
        if (!this.headerElement) return;
        
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        
        if (scrollTop > 10) {
            this.headerElement.classList.add('scrolled');
        } else {
            this.headerElement.classList.remove('scrolled');
        }
    }
    
    updateConnectionStatus(status) {
        if (!this.statusIndicator || !this.statusText) return;
        
        // Remove existing classes
        this.statusIndicator.className = 'status-indicator';
        
        // Set appropriate class and text based on status
        switch (status) {
            case 'connected':
                this.statusIndicator.classList.add('online');
                this.statusText.textContent = 'Terhubung';
                break;
            case 'connecting':
                this.statusIndicator.classList.add('connecting');
                this.statusText.textContent = 'Menghubungkan...';
                break;
            case 'disconnected':
                this.statusText.textContent = 'Terputus';
                break;
            default:
                this.statusText.textContent = 'Status Tidak Diketahui';
        }
    }
    
    addNotification(title, message, time) {
        if (!this.notificationsList) return;
        
        // Check if the empty state exists and remove it
        const emptyState = this.notificationsList.querySelector('.empty-state');
        if (emptyState) {
            this.notificationsList.removeChild(emptyState);
        }
        
        // Create notification element
        const notificationItem = document.createElement('div');
        notificationItem.className = 'notification-item';
        
        // Current time if not provided
        const displayTime = time || new Date().toLocaleTimeString();
        
        // Set content
        notificationItem.innerHTML = `
            <div class="notification-icon">
                <i class="fas fa-bell"></i>
            </div>
            <div class="notification-content">
                <p class="notification-title">${title}</p>
                <p class="notification-message">${message}</p>
                <span class="notification-time">${displayTime}</span>
            </div>
        `;
        
        // Add to the list
        this.notificationsList.prepend(notificationItem);
        
        // Update count
        this.updateNotificationCount(1);
        
        // Save to storage
        this.saveNotifications();
    }
    
    updateNotificationCount(increment = 0) {
        if (!this.notificationsCount) return;
        
        const currentCount = parseInt(this.notificationsCount.textContent) || 0;
        const newCount = currentCount + increment;
        
        this.notificationsCount.textContent = newCount;
        
        // Toggle visibility based on count
        if (newCount > 0) {
            this.notificationsCount.style.display = 'flex';
        } else {
            this.notificationsCount.style.display = 'none';
        }
    }
    
    clearNotifications() {
        if (!this.notificationsList) return;
        
        // Clear the list
        this.notificationsList.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-bell-slash"></i>
                <p>Tidak ada notifikasi baru</p>
            </div>
        `;
        
        // Reset count
        this.notificationsCount.textContent = '0';
        this.notificationsCount.style.display = 'none';
        
        // Clear from storage
        localStorage.removeItem('notifications');
    }
    
    saveNotifications() {
        if (!this.notificationsList) return;
        
        // Get all notification items
        const notifications = Array.from(this.notificationsList.querySelectorAll('.notification-item'))
            .map(item => {
                return {
                    title: item.querySelector('.notification-title').textContent,
                    message: item.querySelector('.notification-message').textContent,
                    time: item.querySelector('.notification-time').textContent
                };
            });
        
        // Save to local storage (limiting to 10 items)
        if (notifications.length > 0) {
            localStorage.setItem('notifications', JSON.stringify(notifications.slice(0, 10)));
        }
    }
    
    loadNotifications() {
        if (!this.notificationsList) return;
        
        // Get from local storage
        const saved = localStorage.getItem('notifications');
        if (!saved) {
            // Set empty state
            this.notificationsList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-bell-slash"></i>
                    <p>Tidak ada notifikasi baru</p>
                </div>
            `;
            return;
        }
        
        try {
            const notifications = JSON.parse(saved);
            
            // Clear list first
            this.notificationsList.innerHTML = '';
            
            // Add each notification
            notifications.forEach(notif => {
                const notificationItem = document.createElement('div');
                notificationItem.className = 'notification-item';
                
                notificationItem.innerHTML = `
                    <div class="notification-icon">
                        <i class="fas fa-bell"></i>
                    </div>
                    <div class="notification-content">
                        <p class="notification-title">${notif.title}</p>
                        <p class="notification-message">${notif.message}</p>
                        <span class="notification-time">${notif.time}</span>
                    </div>
                `;
                
                this.notificationsList.appendChild(notificationItem);
            });
            
            // Update count
            this.notificationsCount.textContent = notifications.length;
            this.notificationsCount.style.display = notifications.length > 0 ? 'flex' : 'none';
            
        } catch (error) {
            console.error('Error loading notifications:', error);
            // Set empty state on error
            this.notificationsList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-bell-slash"></i>
                    <p>Tidak ada notifikasi baru</p>
                </div>
            `;
        }
    }
    
    // Menambahkan setup untuk search di header
    setupHeaderSearch() {
        const searchInput = document.getElementById('header-search');
        const clearSearchBtn = document.getElementById('clear-search');
        
        if (!searchInput || !clearSearchBtn) return;
        
        // Add event listener to input for showing/hiding clear button
        searchInput.addEventListener('input', () => {
            const term = searchInput.value.trim();
            if (term) {
                clearSearchBtn.style.display = 'block';
                this.searchMenuItems(term);
            } else {
                clearSearchBtn.style.display = 'none';
                this.resetMenuItems();
            }
        });
        
        // Add event listener to clear button
        clearSearchBtn.addEventListener('click', () => {
            searchInput.value = '';
            clearSearchBtn.style.display = 'none';
            this.resetMenuItems();
            searchInput.focus();
        });
        
        // Add event listener for Enter key
        searchInput.addEventListener('keyup', (e) => {
            if (e.key === 'Enter' && searchInput.value.trim()) {
                // Find first matched menu item and navigate to it
                const firstMatch = document.querySelector('.nav-link.search-match');
                if (firstMatch) {
                    window.location.href = firstMatch.getAttribute('href');
                }
            }
        });
    }
    
    // Search through menu items
    searchMenuItems(term) {
        const navLinks = document.querySelectorAll('.nav-link');
        term = term.toLowerCase();
        
        let matches = 0;
        
        navLinks.forEach(link => {
            const text = link.querySelector('.nav-text')?.textContent.toLowerCase() || '';
            if (text.includes(term)) {
                link.classList.add('search-match');
                link.closest('.nav-item').style.display = 'block';
                matches++;
            } else {
                link.classList.remove('search-match');
                link.closest('.nav-item').style.display = 'none';
            }
        });
        
        // Show "no results" message if necessary
        const noResults = document.getElementById('search-no-results');
        if (noResults) {
            noResults.style.display = matches === 0 ? 'block' : 'none';
        }
    }
    
    // Reset menu items display
    resetMenuItems() {
        const navItems = document.querySelectorAll('.nav-item');
        const navLinks = document.querySelectorAll('.nav-link');
        
        navItems.forEach(item => {
            item.style.display = 'block';
        });
        
        navLinks.forEach(link => {
            link.classList.remove('search-match');
        });
        
        // Hide "no results" message
        const noResults = document.getElementById('search-no-results');
        if (noResults) {
            noResults.style.display = 'none';
        }
    }
}

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    window.headerComponent = new HeaderComponent();
});
