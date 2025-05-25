/**
 * API Utility Functions
 * Handles API calls with proper authentication
 */

// Get authentication token from session/storage
function getAuthToken() {
    // Try to get from localStorage first (for API calls)
    let token = localStorage.getItem('access_token');
    
    // If not found, try to get from session storage
    if (!token) {
        token = sessionStorage.getItem('access_token');
    }
    
    // If still not found, try to get from cookie
    if (!token) {
        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'auto_login') {
                token = value;
                break;
            }
        }
    }
    
    return token;
}

// Set authentication token
function setAuthToken(token) {
    localStorage.setItem('access_token', token);
    sessionStorage.setItem('access_token', token);
}

// Make authenticated API request
async function apiRequest(url, options = {}) {
    const token = getAuthToken();
    
    // Default headers
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    // Add authentication header if token is available
    if (token) {
        headers['X-Access-Token'] = token;
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    // Merge options
    const requestOptions = {
        ...options,
        headers
    };
    
    try {
        const response = await fetch(url, requestOptions);
        
        // Handle authentication errors
        if (response.status === 401) {
            console.warn('API request failed: Authentication required');
            
            // If this is not a login page, redirect to login
            if (!window.location.pathname.includes('/login')) {
                window.location.href = '/login?redirect=' + encodeURIComponent(window.location.pathname);
                return null;
            }
        }
        
        return response;
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

// API helper functions
async function getConnectionStatus() {
    try {
        const response = await apiRequest('/api/status');
        if (response && response.ok) {
            return await response.json();
        }
        throw new Error('Failed to get connection status');
    } catch (error) {
        console.error('Error getting connection status:', error);
        throw error;
    }
}

async function reconnectWhatsApp() {
    try {
        const response = await apiRequest('/api/reconnect', {
            method: 'POST'
        });
        if (response && response.ok) {
            return await response.json();
        }
        throw new Error('Failed to reconnect WhatsApp');
    } catch (error) {
        console.error('Error reconnecting WhatsApp:', error);
        throw error;
    }
}

async function disconnectWhatsApp() {
    try {
        const response = await apiRequest('/api/disconnect', {
            method: 'POST'
        });
        if (response && response.ok) {
            return await response.json();
        }
        throw new Error('Failed to disconnect WhatsApp');
    } catch (error) {
        console.error('Error disconnecting WhatsApp:', error);
        throw error;
    }
}

async function getGroups() {
    try {
        const response = await apiRequest('/api/groups');
        if (response && response.ok) {
            return await response.json();
        }
        throw new Error('Failed to get groups');
    } catch (error) {
        console.error('Error getting groups:', error);
        throw error;
    }
}

async function sendMessage(type, recipient, message) {
    try {
        const endpoint = type === 'group' ? '/api/send/group' : '/api/send/personal';
        const response = await apiRequest(endpoint, {
            method: 'POST',
            body: JSON.stringify({
                recipient: recipient,
                message: message
            })
        });
        
        if (response && response.ok) {
            return await response.json();
        }
        throw new Error('Failed to send message');
    } catch (error) {
        console.error('Error sending message:', error);
        throw error;
    }
}

// Initialize authentication token from session when page loads
document.addEventListener('DOMContentLoaded', function() {
    // Try to get token from current session and store it for API calls
    const sessionAuthenticated = sessionStorage.getItem('authenticated');
    if (sessionAuthenticated === 'true') {
        // Session exists, try to get token from cookie or storage
        const token = getAuthToken();
        if (token) {
            setAuthToken(token);
        }
    }
});
