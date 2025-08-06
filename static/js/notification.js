/**
 * Notification System - Reusable notification component
 * Provides a consistent way to show notifications across the application
 */
class NotificationSystem {
  constructor(options = {}) {
    // Default configuration
    this.config = {
      position: 'top-right',
      duration: 2500,
      maxCount: 5,
      ...options
    };
    
    // Container for notifications
    this.container = null;
    
    // Active notifications
    this.notifications = [];
    
    // Initialize the system
    this.initialize();
    
    // Log for debugging
    console.log('NotificationSystem initialized');
  }
  
  initialize() {
    // Create container if it doesn't exist
    this.container = document.getElementById('notification-container');
    
    if (!this.container) {
      console.log('Creating notification container');
      this.container = document.createElement('div');
      this.container.id = 'notification-container';
      this.container.className = 'notification-container';
      document.body.appendChild(this.container);
    } else {
      console.log('Using existing notification container');
    }
  }
   
  /**
   * Show a notification
   * @param {string} message - The message to display
   * @param {string} type - Type of notification: 'success', 'error', 'warning', 'info'
   * @param {object} options - Custom options for this specific notification
   * @returns {HTMLElement} The notification element
   */
  show(message, type = 'info', options = {}) {
    console.log(`Showing ${type} notification: ${message}`);
    
    // Make sure container exists
    if (!this.container || !document.body.contains(this.container)) {
      console.log('Container not in DOM, reinitializing');
      this.initialize();
    }
    
    // Get notification options
    const settings = { ...this.config, ...options };
    
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.setAttribute('role', 'alert');
    
    // Get appropriate icon
    const icon = this.getIconForType(type);
    
    // Get appropriate title based on type
    const title = this.getTitleForType(type);
    
    // Set notification content with title and message
    notification.innerHTML = `
      <div class="notification-icon">
        <i class="fas ${icon}"></i>
      </div>
      <div class="notification-content">
        <div class="notification-title">${title}</div>
        <div class="notification-message">${message}</div>
      </div>
      <button class="notification-close" aria-label="Close notification">
        <i class="fas fa-times"></i>
      </button>
    `;
    
    // Attach close handler
    const closeBtn = notification.querySelector('.notification-close');
    closeBtn.addEventListener('click', () => this.close(notification));
    
    // Add to container
    this.container.appendChild(notification);
    
    // Add to tracking array
    this.notifications.push(notification);
    
    // Enforce maximum number of notifications
    this.enforceMaxCount();
    
    // Trigger animation after a small delay
    setTimeout(() => {
      notification.classList.add('visible');
    }, 10);
    
    // Auto-dismiss after duration (if not 0)
    if (settings.duration > 0) {
      setTimeout(() => {
        this.close(notification);
      }, settings.duration);
    }
    
    return notification;
  }
  
  /**
   * Close a specific notification
   * @param {HTMLElement} notification - The notification to close
   */
  close(notification) {
    // Start closing animation dengan menambahkan kelas fade-out
    notification.classList.add('fade-out');
    
    // Remove after animation completes
    setTimeout(() => {
      if (notification.parentNode === this.container) {
        this.container.removeChild(notification);
        
        // Remove from tracking array
        const index = this.notifications.indexOf(notification);
        if (index !== -1) {
          this.notifications.splice(index, 1);
        }
      }
    }, 300);
  }
  
  /**
   * Close all notifications
   */
  closeAll() {
    [...this.notifications].forEach(notification => {
      this.close(notification);
    });
  }
  
  /**
   * Enforce maximum number of notifications
   */
  enforceMaxCount() {
    if (this.notifications.length > this.config.maxCount) {
      const excess = this.notifications.length - this.config.maxCount;
      
      // Remove oldest notifications first
      for (let i = 0; i < excess; i++) {
        this.close(this.notifications[i]);
      }
    }
  }
  
  /**
   * Get FontAwesome icon class for notification type
   */
  getIconForType(type) {
    switch (type) {
      case 'success': return 'fa-check-circle';
      case 'error': return 'fa-exclamation-circle';
      case 'warning': return 'fa-exclamation-triangle';
      case 'info':
      default: return 'fa-info-circle';
    }
  }
  
  /**
   * Get title text based on notification type
   */
  getTitleForType(type) {
    switch (type) {
      case 'success': return 'Berhasil';
      case 'error': return 'Gagal';
      case 'warning': return 'Perhatian';
      case 'info':
      default: return 'Informasi';
    }
  }
  
  /**
   * Show success notification
   */
  success(message, options = {}) {
    return this.show(message, 'success', options);
  }
  
  /**
   * Show error notification
   */
  error(message, options = {}) {
    return this.show(message, 'error', options);
  }
  
  /**
   * Show warning notification
   */
  warning(message, options = {}) {
    return this.show(message, 'warning', options);
  }
  
  /**
   * Show info notification
   */
  info(message, options = {}) {
    return this.show(message, 'info', options);
  }
}

// Create global instance
console.log('Creating global notification system instance');
window.notificationSystem = new NotificationSystem();

// Backwards compatibility function
window.showNotification = function(message, type = 'info') {
  console.log(`Backward compatibility call: ${message} (${type})`);
  return window.notificationSystem.show(message, type);
};
