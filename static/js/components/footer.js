/**
 * Footer Component JavaScript
 * Handles footer functionality including server time and status updates
 */

class FooterComponent {
    constructor() {
        // Elements
        this.footerElement = document.querySelector('.dashboard-footer');
        this.serverTimeElement = document.getElementById('server-time');
        this.serverStatusElement = document.getElementById('server-status');
        
        // Initialize component
        this.init();
    }
    
    init() {
        // Start the clock
        this.startClock();
        
        // Initial server status check
        this.checkServerStatus();
        
        // Setup interval for status checks
        setInterval(() => this.checkServerStatus(), 30000); // Check every 30 seconds
    }
    
    startClock() {
        if (!this.serverTimeElement) return;
        
        const updateClock = () => {
            const now = new Date();
            const hours = String(now.getHours()).padStart(2, '0');
            const minutes = String(now.getMinutes()).padStart(2, '0');
            const seconds = String(now.getSeconds()).padStart(2, '0');
            
            this.serverTimeElement.textContent = `${hours}:${minutes}:${seconds}`;
        };
        
        // Update immediately
        updateClock();
        
        // Then update every second
        setInterval(updateClock, 1000);
    }
    
    checkServerStatus() {
        if (!this.serverStatusElement) return;
        
        // In a real application, you would fetch the status from the server
        // For demonstration, we'll simulate with a random status (95% chance of being online)
        const online = Math.random() > 0.05;
        
        this.updateServerStatus(online ? 'online' : 'offline');
    }
    
    updateServerStatus(status) {
        if (!this.serverStatusElement) return;
        
        // Remove existing status classes
        this.serverStatusElement.className = 'status-text';
        
        // Add appropriate class and text
        if (status === 'online') {
            this.serverStatusElement.classList.add('online');
            this.serverStatusElement.textContent = 'Server: Online';
        } else {
            this.serverStatusElement.classList.add('offline');
            this.serverStatusElement.textContent = 'Server: Offline';
        }
    }
    
    // Method to manually update server status from outside
    setServerStatus(status) {
        this.updateServerStatus(status);
    }
}

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    window.footerComponent = new FooterComponent();
});
