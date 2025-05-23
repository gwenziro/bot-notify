/**
 * API Utilities
 * Handles common API calls to the backend
 */

// Get access token from localStorage or cookie
function getAccessToken() {
    return localStorage.getItem('access_token') || '';
}

// Fetch with common headers and error handling
async function fetchAPI(endpoint, options = {}) {
    // Set default options
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'X-Access-Token': getAccessToken()
        }
    };
    
    // Merge options
    const fetchOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...(options.headers || {})
        }
    };
    
    try {
        const response = await fetch(endpoint, fetchOptions);
        
        // Handle unauthorized
        if (response.status === 401) {
            // If not on login page, redirect to login
            if (!window.location.pathname.includes('/login')) {
                window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
                return { error: 'Unauthorized' };
            }
        }
        
        // Parse JSON response
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('API fetch error:', error);
        return { error: error.message };
    }
}

// Check WhatsApp connection status
async function getConnectionStatus() {
    return fetchAPI('/api/status');
}

// Send reconnect request
async function reconnectWhatsApp() {
    return fetchAPI('/api/reconnect', { method: 'POST' });
}

// Send disconnect request
async function disconnectWhatsApp() {
    return fetchAPI('/api/disconnect', { method: 'POST' });
}

// Get available groups
async function getGroups() {
    return fetchAPI('/api/groups');
}

// Get QR code status
async function getQRCodeStatus() {
    return fetchAPI('/api/qr/status');
}
