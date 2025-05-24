/**
 * Login Page JavaScript
 * Mengelola fungsionalitas halaman login
 */

document.addEventListener('DOMContentLoaded', function() {
    // Setup toggle password visibility
    setupPasswordToggle();
    
    // Animasi masuk untuk login card
    animateLoginCard();
    
    // Form submission handling with validation
    setupFormSubmission();
    
    // Setup throttling and security
    setupSecurityMeasures();
});

// Toggle password visibility
function setupPasswordToggle() {
    const toggleBtn = document.querySelector('.toggle-password');
    const tokenInput = document.getElementById('token');
    
    if (!toggleBtn || !tokenInput) return;
    
    // Tambahkan event langsung saat halaman load untuk mencegah ikon bawaan
    setTimeout(() => {
        // Force tampilan bawaan browser untuk tersembunyi
        const currentCss = tokenInput.style.cssText;
        tokenInput.style.cssText = currentCss + '; --password-reveal-display: none !important;';
    }, 100);
    
    toggleBtn.addEventListener('click', function() {
        // Toggle input type antara 'password' dan 'text'
        const currentType = tokenInput.getAttribute('type');
        const newType = currentType === 'password' ? 'text' : 'password';
        tokenInput.setAttribute('type', newType);
        
        // Perbarui ikon berdasarkan type
        const icon = this.querySelector('i');
        if (newType === 'text') {
            icon.className = 'fas fa-eye-slash'; // Ganti ke ikon tutup mata
        } else {
            icon.className = 'fas fa-eye'; // Kembali ke ikon mata normal
        }
    });
}

// Animasi login card dengan fade in
function animateLoginCard() {
    const loginCard = document.querySelector('.login-card');
    if (!loginCard) return;
    
    // Tambahkan animasi masuk
    setTimeout(() => {
        loginCard.style.opacity = '1';
        loginCard.style.transform = 'translateY(0)';
    }, 100);
}

// Handling form submission & validasi
function setupFormSubmission() {
    const loginForm = document.getElementById('login-form');
    const tokenField = document.getElementById('token');
    
    if (!loginForm || !tokenField) return;
    
    loginForm.addEventListener('submit', function(e) {
        // Validasi token
        if (!tokenField.value.trim()) {
            e.preventDefault();
            showValidationError(tokenField, 'Token harus diisi');
            return false;
        }
        
        // Validasi minimal length
        if (tokenField.value.trim().length < 8) {
            e.preventDefault();
            showValidationError(tokenField, 'Token tidak valid (terlalu pendek)');
            return false;
        }
        
        // Simpan timestamp login attempt untuk deteksi brute force
        const attempts = getLoginAttempts();
        attempts.push(Date.now());
        localStorage.setItem('login_attempts', JSON.stringify(attempts));
        
        // Disable button to prevent double-submit
        document.getElementById('login-btn').disabled = true;
        document.getElementById('login-btn').innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memproses...';
    });
    
    // Reset error state on input
    tokenField.addEventListener('input', function() {
        this.classList.remove('error');
        const errorElem = loginForm.querySelector('.validation-error');
        if (errorElem) {
            errorElem.remove();
        }
    });
}

// Setup security measures
function setupSecurityMeasures() {
    // Throttle login attempts jika terlalu banyak
    const attempts = getLoginAttempts();
    
    // Hapus percobaan yang lebih lama dari 5 menit
    const now = Date.now();
    const recentAttempts = attempts.filter(time => (now - time) < 5 * 60 * 1000);
    
    // Jika banyak percobaan dalam waktu singkat, tambahkan delay
    if (recentAttempts.length > 5) {
        const loginBtn = document.getElementById('login-btn');
        if (loginBtn) {
            loginBtn.disabled = true;
            loginBtn.innerHTML = '<i class="fas fa-hourglass"></i> Menunggu...';
            
            setTimeout(() => {
                loginBtn.disabled = false;
                loginBtn.innerHTML = '<i class="fas fa-sign-in-alt"></i> Login';
            }, 3000);
        }
    }
    
    // Update storage
    localStorage.setItem('login_attempts', JSON.stringify(recentAttempts));
}

// Helper untuk mendapatkan login attempts dari localStorage
function getLoginAttempts() {
    try {
        return JSON.parse(localStorage.getItem('login_attempts') || '[]');
    } catch(e) {
        return [];
    }
}

// Show validation error message
function showValidationError(inputElement, message) {
    // Tambahkan class error ke input
    inputElement.classList.add('error');
    
    // Animasi shake untuk feedback error
    const parentElement = inputElement.parentElement;
    parentElement.classList.add('shake');
    setTimeout(() => {
        parentElement.classList.remove('shake');
    }, 500);
    
    // Tampilkan pesan error
    const errorElement = document.createElement('div');
    errorElement.className = 'validation-error';
    errorElement.textContent = message;
    
    // Insert after input group
    const formGroup = parentElement.parentElement;
    formGroup.appendChild(errorElement);
    
    // Focus input
    inputElement.focus();
}
