/**
 * Login Page JavaScript
 * Mengelola fungsionalitas halaman login
 */

document.addEventListener('DOMContentLoaded', function() {
    // Setup toggle password visibility
    setupPasswordToggle();
    
    // Animasi masuk untuk login card
    animateLoginCard();
    
    // Form submission handling
    setupFormSubmission();
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
    const loginForm = document.querySelector('.login-form');
    if (!loginForm) return;
    
    loginForm.addEventListener('submit', function(e) {
        const tokenField = document.getElementById('token');
        
        // Validasi sederhana - pastikan token tidak kosong
        if (!tokenField.value.trim()) {
            e.preventDefault();
            tokenField.classList.add('error');
            
            // Animasi shake untuk feedback error
            tokenField.parentElement.classList.add('shake');
            setTimeout(() => {
                tokenField.parentElement.classList.remove('shake');
            }, 500);
            
            return false;
        }
    });
}
