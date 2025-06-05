/**
 * Dashboard JS - Handles all dashboard interactions
 */

document.addEventListener('DOMContentLoaded', function() {
    // Elements - ubah dari const menjadi let untuk variabel yang perlu dimodifikasi
    const refreshBtn = document.getElementById('refresh-btn');
    const disconnectBtn = document.getElementById('disconnect-btn');
    const connectionBadge = document.getElementById('connection-badge');
    let qrContainer = document.getElementById('qr-container');
    let qrCode = document.getElementById('qr-code');
    let qrLoading = document.getElementById('qr-loading');
    let connectedView = document.getElementById('connected-view');
    let qrView = document.getElementById('qr-view');
    let refreshQrBtn = document.getElementById('refresh-qr-btn');
    const serverTime = document.getElementById('server-time');
    
    // Inisialisasi sistem notifikasi jika belum ada
    if (!window.notificationSystem) {
        console.log('Initializing notification system...');
        window.notificationSystem = new NotificationSystem();
    }
    
    // Update server time every second
    function updateServerTime() {
        const now = new Date();
        const timeString = now.toLocaleTimeString('id-ID', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        });
        serverTime.textContent = timeString;
    }
    
    // Initialize server time and update every second
    updateServerTime();
    setInterval(updateServerTime, 1000);
    
    // Setup sticky header
    const header = document.querySelector('.dashboard-header');
    const headerObserver = new IntersectionObserver(
        ([entry]) => {
            if (!entry.isIntersecting) {
                header.classList.add('scrolled');
            } else {
                header.classList.remove('scrolled');
            }
        },
        { threshold: 0.1 } // Trigger ketika 10% header tidak terlihat
    );
    
    // Buat dummy element sebagai "trigger point"
    const headerTrigger = document.createElement('div');
    headerTrigger.style.height = '1px';
    headerTrigger.style.width = '100%';
    headerTrigger.style.position = 'absolute';
    headerTrigger.style.top = '0';
    headerTrigger.style.left = '0';
    headerTrigger.style.zIndex = '-1';
    document.body.prepend(headerTrigger);
    
    // Observe trigger element
    headerObserver.observe(headerTrigger);
    
    // Function to check QR status
    function checkQRStatus() {
        fetch('/api/qr/status')
        .then(response => response.json())
        .then(data => {
            console.log('QR status:', data);
            
            if (data.available) {
                // QR code is available, update the image with cache-busting
                if (qrContainer) {
                    // Pastikan elemen qrCode ada
                    if (!qrCode) {
                        // Jika tidak ada, buat elemen img baru
                        qrCode = document.createElement('img');
                        qrCode.id = 'qr-code';
                        qrCode.className = 'qr-image';
                        qrCode.alt = 'QR Code WhatsApp';
                        qrContainer.appendChild(qrCode);
                        
                        // Notifikasi QR code baru dibuat
                        notificationSystem.info('QR Code baru telah dibuat');
                    }
                    
                    // Update src dengan timestamp untuk mencegah cache
                    const qrUrl = `/api/qr/image?t=${Date.now()}`;
                    qrCode.src = qrUrl;
                    qrCode.style.display = 'block';
                    
                    // Sembunyikan loading indicator
                    if (qrLoading) {
                        qrLoading.style.display = 'none';
                    }
                    
                    // Log info untuk debugging
                    console.log('QR code displayed with URL:', qrUrl);
                } else {
                    console.error('QR container not found');
                    notificationSystem.error('Container QR Code tidak ditemukan');
                }
            } else {
                // QR code is not available or expired
                if (qrLoading) {
                    if (data.expired) {
                        qrLoading.innerHTML = '<i class="fas fa-exclamation-circle"></i><span>QR Code kedaluwarsa, memuat ulang...</span>';
                        qrLoading.style.display = 'flex';
                        
                        // Sembunyikan QR code yang mungkin sudah tidak valid
                        if (qrCode) {
                            qrCode.style.display = 'none';
                        }
                        
                        // Jika QR code kedaluwarsa, otomatis refresh QR code
                        // Delay sedikit untuk mencegah terlalu banyak request berturut-turut
                        setTimeout(() => {
                            refreshQRCode();
                        }, 1000);
                    } else {
                        qrLoading.innerHTML = '<i class="fas fa-clock"></i><span>Menunggu QR Code...</span>';
                        qrLoading.style.display = 'flex';
                        
                        // Sembunyikan QR code yang mungkin sudah tidak valid
                        if (qrCode) {
                            qrCode.style.display = 'none';
                        }
                    }
                }
            }
        })
        .catch(error => {
            console.error('Error checking QR status:', error);
            notificationSystem.error('Gagal memeriksa status QR: ' + error.message);
        });
    }
    
    // Tambahkan flag untuk mencegah terlalu banyak pemanggilan refreshQRCode bersamaan
    let isRefreshingQR = false;
    
    // Function to refresh the QR code
    function refreshQRCode() {
        // Mencegah multiple refresh requests
        if (isRefreshingQR) {
            console.log('QR refresh already in progress, skipping');
            return;
        }
        
        isRefreshingQR = true;
        
        if (refreshQrBtn) {
            refreshQrBtn.disabled = true;
            refreshQrBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>Memuat...</span>';
        }
        
        // Show loading state
        if (qrContainer && qrLoading) {
            qrCode && (qrCode.style.display = 'none');
            qrLoading.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>Memuat Kode QR...</span>';
            qrLoading.style.display = 'flex';
        }
        
        // Call API to refresh QR code
        fetch('/dashboard/refresh-qr', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => response.json())
        .then(data => {
            console.log('QR refresh response:', data);
            
            if (data.success) {
                // Show success notification
                notificationSystem.success('QR Code berhasil disegarkan');
                
                // Check if QR is available immediately
                checkQRStatus();
                
                // Schedule regular QR checks
                startQRStatusChecks();
            } else {
                // Show error message
                if (qrLoading) {
                    qrLoading.innerHTML = `<i class="fas fa-exclamation-triangle"></i><span>${data.message || 'Gagal memuat QR code'}</span>`;
                }
                notificationSystem.error(data.message || 'Gagal menyegarkan QR Code');
            }
            
            // Re-enable the button
            if (refreshQrBtn) {
                refreshQrBtn.disabled = false;
                refreshQrBtn.innerHTML = '<i class="fas fa-sync"></i><span>Segarkan QR</span>';
            }
            
            // Reset flag
            isRefreshingQR = false;
        })
        .catch(error => {
            console.error('Error refreshing QR code:', error);
            notificationSystem.error('Terjadi kesalahan: ' + error.message);
            
            // Re-enable the button
            if (refreshQrBtn) {
                refreshQrBtn.disabled = false;
                refreshQrBtn.innerHTML = '<i class="fas fa-sync"></i><span>Segarkan QR</span>';
            }
            
            // Reset flag
            isRefreshingQR = false;
        });
    }
    
    // Modifikasi fungsi startQRStatusChecks untuk mendukung auto-refresh
    let qrCheckInterval;
    
    function startQRStatusChecks() {
        // Clear any existing interval
        if (qrCheckInterval) {
            clearInterval(qrCheckInterval);
        }
        
        // Check every 3 seconds
        qrCheckInterval = setInterval(() => {
            // Also check if we've already connected
            fetchConnectionStatus(true);
            
            // Only check QR if we're still on the QR view
            if (qrView && qrView.style.display !== 'none') {
                checkQRStatus();
            } else {
                // Stop checking if we're no longer on QR view
                clearInterval(qrCheckInterval);
            }
        }, 3000);
    }
    
    // Function to update UI directly after disconnecting
    function updateUIAfterDisconnect() {
        console.log('Updating UI after disconnect');
        
        // Update badge status
        if (connectionBadge) {
            connectionBadge.textContent = 'Terputus';
            connectionBadge.classList.remove('connected');
            connectionBadge.classList.add('disconnected');
        }
        
        // Temukan elemen yang perlu diupdate
        const connectedInfo = document.getElementById('connected-view');
        const qrConnectionView = document.getElementById('qr-view');
        
        // Debug info
        console.log('Elements found:', {
            connectedView: connectedInfo ? true : false,
            qrView: qrConnectionView ? true : false
        });
        
        // Hide connected view
        if (connectedInfo) {
            connectedInfo.style.display = 'none';
        }
        
        // Create QR view if it doesn't exist
        if (!qrConnectionView) {
            console.log('QR view not found, creating it');
            const newQrView = createQRView();
            
            // Find the parent container to insert QR view
            const cardBody = document.querySelector('.contact-info-section .card-body');
            if (cardBody) {
                cardBody.innerHTML = ''; // Clear current content
                cardBody.appendChild(newQrView);
                
                // Get references to new elements - tetap menggunakan variabel global
                qrContainer = document.getElementById('qr-container');
                qrLoading = document.getElementById('qr-loading');
                qrCode = null; // Reset qrCode reference - penting!
                refreshQrBtn = document.getElementById('refresh-qr-btn');
                qrView = document.getElementById('qr-view');
                
                // Attach event listener to new refresh button
                if (refreshQrBtn) {
                    refreshQrBtn.addEventListener('click', refreshQRCode);
                }
            }
        } else {
            // Show existing QR view
            qrConnectionView.style.display = 'flex';
            
            // Update QR container to loading state
            if (qrContainer) {
                if (qrLoading) {
                    qrLoading.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>Memuat Kode QR...</span>';
                    qrLoading.style.display = 'flex';
                }
                
                if (qrCode) {
                    qrCode.style.display = 'none';
                }
            }
        }
        
        // Start refreshing QR code
        setTimeout(refreshQRCode, 500);
    }
    
    // Helper function to create QR view if it doesn't exist
    function createQRView() {
        const qrView = document.createElement('div');
        qrView.id = 'qr-view';
        qrView.className = 'qr-connection';
        qrView.style.display = 'flex';
        
        qrView.innerHTML = `
            <p class="qr-instructions">
                Silakan pindai kode QR berikut dengan WhatsApp di ponsel Anda untuk menghubungkan Bot:
            </p>
            <div class="qr-container" id="qr-container">
                <div class="qr-loading" id="qr-loading">
                    <i class="fas fa-spinner fa-spin"></i>
                    <span>Memuat Kode QR...</span>
                </div>
                <!-- QR code image akan dibuat secara dinamis dengan JavaScript -->
            </div>
            <div class="qr-actions">
                <button id="refresh-qr-btn" class="btn btn-primary">
                    <i class="fas fa-sync"></i>
                    <span>Segarkan QR</span>
                </button>
            </div>
        `;
        
        return qrView;
    }
    
    // Function to disconnect WhatsApp
    function disconnectWhatsApp() {
        if (disconnectBtn) {
            disconnectBtn.disabled = true;
            disconnectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>Memutuskan...</span>';
        }
        
        fetch('/dashboard/disconnect', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(data => {
            console.log('Disconnect response:', data);
            
            if (data.success) {
                // Tampilkan notifikasi sukses
                notificationSystem.success('Koneksi WhatsApp berhasil diputuskan');
                
                // Langsung update UI tanpa reload
                updateUIAfterDisconnect();
                
                // Reload halaman jika UI gagal diupdate setelah beberapa detik
                setTimeout(() => {
                    // Cek badge status untuk memastikan UI diperbarui
                    if (connectionBadge && connectionBadge.classList.contains('connected')) {
                        console.log('UI not updated properly, reloading page');
                        window.location.reload();
                    }
                    
                    // Cek juga untuk memastikan QR view sudah muncul
                    const qrView = document.getElementById('qr-view');
                    if (!qrView || qrView.style.display === 'none') {
                        console.log('QR view not visible, reloading page');
                        window.location.reload();
                    }
                }, 3000);
            } else {
                // Tampilkan pesan error
                notificationSystem.error(data.message || 'Gagal memutuskan koneksi');
                
                // Kembalikan tombol ke state awal
                if (disconnectBtn) {
                    disconnectBtn.disabled = false;
                    disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i><span>Putuskan</span>';
                }
            }
        })
        .catch(error => {
            console.error('Error disconnecting WhatsApp:', error);
            notificationSystem.error('Terjadi kesalahan: ' + error.message);
            
            // Kembalikan tombol ke state awal
            if (disconnectBtn) {
                disconnectBtn.disabled = false;
                disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i><span>Putuskan</span>';
            }
        });
    }
    
    // Function to check connection status
    function fetchConnectionStatus(silent = false) {
        fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            if (!silent) {
                console.log('Connection status:', data);
            }
            
            // If connection state changed, reload the page to update all elements
            const isCurrentlyConnected = connectionBadge && connectionBadge.classList.contains('connected');
            
            if (data.isConnected !== isCurrentlyConnected) {
                window.location.reload();
            }
        })
        .catch(error => {
            if (!silent) {
                console.error('Error fetching connection status:', error);
            }
        });
    }
    
    // Fungsi untuk syntax highlighting sederhana pada code blocks
    function highlightCodeSyntax() {
      const codeBlocks = document.querySelectorAll('.code-block code');
      
      codeBlocks.forEach(block => {
        // Hanya highlight property JSON untuk kesederhanaan
        block.innerHTML = block.innerHTML.replace(
          /"([^"]+)":/g, 
          '"<span class="property">$1</span>":'
        );
      });
    }
    
    // Fungsi untuk menangani tombol copy kode
    function setupCodeCopyButtons() {
        const copyButtons = document.querySelectorAll('.copy-btn');
        
        copyButtons.forEach(btn => {
            // Simpan ikon asli dalam data attribute untuk referensi yang lebih konsisten
            const iconElement = btn.querySelector('i');
            if (iconElement) {
                btn.setAttribute('data-original-icon', iconElement.className);
            }
            
            btn.addEventListener('click', function() {
                const codeType = this.getAttribute('data-code');
                const codeElement = document.getElementById(`${codeType}-code`);
                
                if (!codeElement) {
                    window.notificationSystem.error('Elemen kode tidak ditemukan');
                    return;
                }
                
                // Dapatkan teks yang akan disalin
                const codeText = codeElement.textContent;
                
                // Ambil original icon class dari data attribute
                const originalIconClass = this.getAttribute('data-original-icon') || 'fas fa-copy';
                
                // Gunakan clipboard API untuk menyalin teks
                navigator.clipboard.writeText(codeText)
                    .then(() => {
                        // Tampilkan efek sukses pada tombol
                        this.classList.add('copied');
                        
                        // Ganti ikon dengan ikon centang
                        this.innerHTML = '<i class="fas fa-check"></i>';
                        
                        // Tampilkan notifikasi sukses
                        window.notificationSystem.success('Kode telah disalin ke clipboard');
                        
                        // Kembalikan tombol ke keadaan semula setelah 2 detik
                        setTimeout(() => {
                            this.classList.remove('copied');
                            // Kembalikan ikon asli menggunakan data attribute
                            this.innerHTML = `<i class="${originalIconClass}"></i>`;
                        }, 2000);
                    })
                    .catch(err => {
                        console.error('Gagal menyalin teks: ', err);
                        window.notificationSystem.error('Tidak dapat menyalin kode');
                    });
            });
        });
    }

    // Tambahkan dukungan copy untuk token API
    function setupTokenCopy() {
        const tokenCopyBtn = document.querySelector('[data-code="api-token"]');
        if (tokenCopyBtn) {
            // Simpan ikon asli dalam data attribute
            const iconElement = tokenCopyBtn.querySelector('i');
            if (iconElement) {
                tokenCopyBtn.setAttribute('data-original-icon', iconElement.className);
            }
            
            tokenCopyBtn.addEventListener('click', function() {
                const tokenElement = document.getElementById('api-token-code');
                if (!tokenElement) {
                    window.notificationSystem.error('Elemen token tidak ditemukan');
                    return;
                }
                
                // Dapatkan teks token
                const tokenText = tokenElement.textContent;
                
                // Ambil original icon class dari data attribute
                const originalIconClass = this.getAttribute('data-original-icon') || 'fas fa-copy';
                
                // Salin ke clipboard
                navigator.clipboard.writeText(tokenText)
                    .then(() => {
                        // Tampilkan efek sukses
                        this.classList.add('copied');
                        this.innerHTML = '<i class="fas fa-check"></i>';
                        
                        // Notifikasi sukses
                        window.notificationSystem.success('Token API telah disalin ke clipboard');
                        
                        // Kembalikan tombol ke keadaan semula setelah 2 detik
                        setTimeout(() => {
                            this.classList.remove('copied');
                            this.innerHTML = `<i class="${originalIconClass}"></i>`;
                        }, 2000);
                    })
                    .catch(err => {
                        console.error('Gagal menyalin token: ', err);
                        window.notificationSystem.error('Tidak dapat menyalin token API');
                    });
            });
        }
    }

    // Attach event listeners - PENTING: hanya satu event listener per tombol
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            window.location.reload();
        });
    }
    
    if (refreshQrBtn) {
        refreshQrBtn.addEventListener('click', refreshQRCode);
    }
    
    // Create dialog instance
    const dialog = new Dialog();
    
    // Handler untuk tombol disconnect
    if (disconnectBtn) {
        disconnectBtn.addEventListener('click', function() {
            console.log('Disconnect button clicked');
            // Gunakan dialog modular
            dialog.show({
                title: 'Konfirmasi Pemutusan',
                message: 'Apakah Anda yakin ingin memutuskan koneksi WhatsApp?',
                description: 'Pemutusan koneksi akan menghapus sesi dan memerlukan pemindaian QR code lagi untuk terhubung kembali.',
                confirmText: 'Putuskan Koneksi',
                cancelText: 'Batal',
                type: 'danger',
                onConfirm: function() {
                    console.log('Disconnect confirmed');
                    // Ubah teks tombol untuk indikasi proses
                    disconnectBtn.disabled = true;
                    disconnectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memutuskan...';
                    
                    // Kirim permintaan disconnect ke server
                    fetch('/dashboard/disconnect', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        }
                    })
                    .then(response => response.json())
                    .then(data => {
                        console.log('Disconnect response:', data);
                        
                        if (data.success) {
                            // Tampilkan notifikasi sukses - perbaiki pemanggilan API
                            console.log('Showing success notification');
                            notificationSystem.success('WhatsApp berhasil diputuskan');
                            
                            // Tunggu sebentar agar notifikasi terlihat, lalu refresh halaman
                            setTimeout(function() {
                                console.log('Reloading page after disconnect');
                                window.location.reload();
                            }, 1500); // Refresh setelah 1.5 detik
                        } else {
                            // Jika gagal, kembalikan tombol ke keadaan semula
                            disconnectBtn.disabled = false;
                            disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i> Putuskan';
                            notificationSystem.error(data.message || 'Gagal memutuskan koneksi');
                        }
                    })
                    .catch(error => {
                        console.error('Error disconnecting:', error);
                        disconnectBtn.disabled = false;
                        disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i> Putuskan';
                        notificationSystem.error('Terjadi kesalahan: ' + error.message);
                    });
                },
                onCancel: function() {
                    console.log('Disconnect cancelled');
                    notificationSystem.info('Pemutusan koneksi dibatalkan');
                }
            });
        });
    }
    
    // Start QR status checks if we're on QR view
    if (qrView && qrView.style.display !== 'none') {
        startQRStatusChecks();
    }

    // Periodically check connection status (every 5 seconds)
    setInterval(() => {
        fetchConnectionStatus(true);
    }, 30000);
    
    // Auto-refresh QR jika berada di halaman QR view dan QR tidak tersedia/kedaluwarsa
    if (qrView && qrView.style.display !== 'none' && qrCode && qrCode.style.display === 'none') {
        console.log('Auto-refreshing QR on page load');
        setTimeout(refreshQRCode, 500);
    }
    
    // Tambahkan syntax highlighting ke code blocks
    highlightCodeSyntax();
    
    // Gunakan hanya setupCodeCopyButtons yang sudah diperbaiki
    setupCodeCopyButtons();
});
