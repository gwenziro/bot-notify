/**
 * Connectivity Page JavaScript
 * Handles QR Code display and WhatsApp connection management
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize page
    initConnectivityPage();
    
    // Setup event listeners
    setupConnectivityEvents();
    
    // Check for QR Code
    checkForQRCode();
});

// Initialize connectivity page
function initConnectivityPage() {
    // Update page title
    updatePageTitle();
    
    // Check connection status immediately
    checkConnectionStatus();
}

// Setup connectivity page event listeners
function setupConnectivityEvents() {
    // Refresh QR Code button
    document.getElementById('refresh-qr-btn')?.addEventListener('click', function() {
        window.location.reload();
    });
    
    // Reconnect button
    document.getElementById('reconnect-btn')?.addEventListener('click', function() {
        this.disabled = true;
        this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Menghubungkan...';
        
        reconnectWhatsApp()
            .then(data => {
                showTemporaryMessage(
                    document.getElementById('status-message'), 
                    'Permintaan koneksi ulang berhasil dikirim. Mohon tunggu...',
                    'info'
                );
                setTimeout(() => {
                    checkForQRCode();
                    checkConnectionStatus();
                }, 2000);
            })
            .catch(error => {
                showTemporaryMessage(
                    document.getElementById('status-message'), 
                    'Error saat menghubungkan WhatsApp: ' + error.message,
                    'danger'
                );
            })
            .finally(() => {
                this.disabled = false;
                this.innerHTML = '<i class="fas fa-sync"></i> Hubungkan Ulang WhatsApp';
            });
    });
    
    // Logout/Disconnect button
    document.getElementById('logout-btn')?.addEventListener('click', function() {
        confirmAction('Apakah Anda yakin ingin logout dari WhatsApp?', () => {
            this.disabled = true;
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memutuskan...';
            
            disconnectWhatsApp()
                .then(data => {
                    showTemporaryMessage(
                        document.getElementById('status-message'), 
                        'WhatsApp berhasil diputuskan. Silakan scan QR code lagi.',
                        'success'
                    );
                    setTimeout(() => {
                        window.location.reload();
                    }, 2000);
                })
                .catch(error => {
                    showTemporaryMessage(
                        document.getElementById('status-message'), 
                        'Error saat memutuskan WhatsApp: ' + error.message,
                        'danger'
                    );
                })
                .finally(() => {
                    this.disabled = false;
                    this.innerHTML = '<i class="fas fa-power-off"></i> Logout';
                });
        });
    });
    
    // Start QR Code refresh countdown if applicable
    startQRCountdown();
}

// Check for QR Code and update UI
function checkForQRCode() {
    const qrContainer = document.getElementById('qrcode-container');
    if (!qrContainer) return;
    
    getQRCodeStatus()
        .then(data => {
            if (data.available) {
                // QR Code available, update image
                const qrImg = document.getElementById('qrcode-img');
                if (qrImg) {
                    qrImg.src = 'data:image/png;base64,' + data.image;
                    
                    // Update timestamp if available
                    const timeElement = document.getElementById('qr-timestamp');
                    if (timeElement && data.timestamp) {
                        timeElement.textContent = formatDateTime(data.timestamp);
                    }
                    
                    // Start countdown if applicable
                    startQRCountdown();
                }
            } else if (data.isConnected) {
                // Already connected, show connected message
                qrContainer.innerHTML = `
                    <div class="alert alert-success">
                        <i class="fas fa-check-circle fa-2x mb-3"></i>
                        <h4>WhatsApp Sudah Terhubung!</h4>
                        <p>Anda telah berhasil terhubung dengan WhatsApp. Tidak perlu scan QR code lagi.</p>
                        <div class="mt-3">
                            <a href="/status" class="btn btn-primary">Lihat Status</a>
                            <button id="logout-btn" class="btn btn-danger">Logout</button>
                        </div>
                    </div>
                `;
                
                // Re-attach event listener for the new button
                document.getElementById('logout-btn')?.addEventListener('click', function() {
                    confirmAction('Apakah Anda yakin ingin logout dari WhatsApp?', () => {
                        disconnectWhatsApp()
                            .then(() => window.location.reload());
                    });
                });
            } else {
                // QR Code not available, show waiting message
                qrContainer.innerHTML = `
                    <div class="alert alert-warning">
                        <i class="fas fa-exclamation-triangle fa-2x mb-3"></i>
                        <h4>QR Code Tidak Tersedia</h4>
                        <p>Server sedang mencoba membuat QR code baru. Silakan tunggu beberapa saat atau klik tombol refresh.</p>
                        <div class="mt-3">
                            <button id="reconnect-btn" class="btn btn-primary">Hubungkan WhatsApp</button>
                            <button id="refresh-qr-btn" class="btn btn-secondary">Refresh</button>
                        </div>
                    </div>
                `;
                
                // Re-attach event listeners for the new buttons
                setupConnectivityEvents();
            }
        })
        .catch(error => {
            console.error('Error checking QR code:', error);
            qrContainer.innerHTML = `
                <div class="alert alert-danger">
                    <i class="fas fa-times-circle fa-2x mb-3"></i>
                    <h4>Error</h4>
                    <p>Gagal memuat QR code: ${error.message}</p>
                    <div class="mt-3">
                        <button id="refresh-qr-btn" class="btn btn-primary">Coba Lagi</button>
                    </div>
                </div>
            `;
            
            // Re-attach event listener for the refresh button
            document.getElementById('refresh-qr-btn')?.addEventListener('click', function() {
                window.location.reload();
            });
        });
}

// Start QR Code countdown timer for auto refresh
function startQRCountdown() {
    const timerElement = document.getElementById('time-remaining');
    const interval = parseInt(timerElement?.dataset.interval || '60');
    
    if (!timerElement) return;
    
    let secondsLeft = interval;
    const countdownInterval = setInterval(() => {
        secondsLeft--;
        timerElement.textContent = `(akan diperbarui dalam ${secondsLeft} detik)`;
        
        if (secondsLeft <= 0) {
            clearInterval(countdownInterval);
            window.location.reload();
        }
    }, 1000);
}
