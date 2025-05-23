/**
 * Login Page JavaScript
 * Manages login page functionality
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize login page
    initLoginPage();
});

// Initialize login page functionality
function initLoginPage() {
    // Setup password toggle visibility
    setupPasswordToggle();
    
    // Focus on token input
    focusTokenInput();
    
    // Setup animations
    setupAnimations();
    
    // Handle form submission
    setupFormSubmission();
}

// Toggle password/token visibility
function setupPasswordToggle() {
    const togglePasswordButton = document.querySelector('.toggle-password');
    const passwordInput = document.querySelector('#token');
    
    if (!togglePasswordButton || !passwordInput) return;
    
    togglePasswordButton.addEventListener('click', function() {
        const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
        passwordInput.setAttribute('type', type);
        
        // Toggle icon
        const icon = togglePasswordButton.querySelector('i');
        if (icon) {
            icon.classList.toggle('fa-eye');
            icon.classList.toggle('fa-eye-slash');
        }
    });
}

// Focus on token input automatically
function focusTokenInput() {
    const passwordInput = document.querySelector('#token');
    if (passwordInput) {
        setTimeout(() => {
            passwordInput.focus();
        }, 500); // Delay focus for animation to complete
    }
}

// Setup animations for login elements
function setupAnimations() {
    // Add fade-in animation to login container
    const loginContainer = document.querySelector('.login-container');
    if (loginContainer) {
        loginContainer.classList.add('fade-in');
    }
    
    // Animate login card with slight delay
    setTimeout(() => {
        const loginCard = document.querySelector('.login-card');
        if (loginCard) {
            loginCard.style.opacity = '1';
            loginCard.style.transform = 'translateY(0)';
        }
    }, 300);
}

// Handle form submission
function setupFormSubmission() {
    const loginForm = document.querySelector('.login-form');
    if (!loginForm) return;
    
    loginForm.addEventListener('submit', function(e) {
        const tokenInput = document.querySelector('#token');
        
        if (!tokenInput || !tokenInput.value.trim()) {
            e.preventDefault();
            showError('Token tidak boleh kosong');
            return;
        }
        
        // Disable the submit button to prevent double submission
        const submitButton = this.querySelector('button[type="submit"]');
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memproses...';
        }
    });
}

// Show error message
function showError(message) {
    // Check if error alert already exists
    let errorAlert = document.querySelector('.alert.alert-danger');
    
    // If not, create a new one
    if (!errorAlert) {
        errorAlert = document.createElement('div');
        errorAlert.className = 'alert alert-danger';
        errorAlert.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
        
        // Insert after the text-muted paragraph
        const textMuted = document.querySelector('.text-muted');
        if (textMuted && textMuted.parentNode) {
            textMuted.parentNode.insertBefore(errorAlert, textMuted.nextSibling);
        }
    } else {
        // Update existing error message
        errorAlert.innerHTML = `<i class="fas fa-exclamation-circle"></i> ${message}`;
    }
    
    // Shake the input to provide visual feedback
    const tokenInput = document.querySelector('#token');
    if (tokenInput) {
        tokenInput.classList.add('shake');
        setTimeout(() => {
            tokenInput.classList.remove('shake');
        }, 500);
    }
}
