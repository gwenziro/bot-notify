/**
 * Dashboard JS - Handles all dashboard interactions
 */

document.addEventListener('DOMContentLoaded', function() {
    // Elements
    const refreshBtn = document.getElementById('refresh-btn');
    const refreshQrBtn = document.getElementById('refresh-qr-btn');
    const disconnectBtn = document.getElementById('disconnect-btn');
    const copyApiBtn = document.getElementById('copy-api-btn');
    const connectionBadge = document.getElementById('connection-badge');
    const qrContainer = document.getElementById('qr-container');
    const qrCode = document.getElementById('qr-code');
    const qrLoading = document.getElementById('qr-loading');
    const connectedView = document.getElementById('connected-view');
    const qrView = document.getElementById('qr-view');
    const serverTime = document.getElementById('server-time');
    
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
    
    // Function to check QR status
    function checkQRStatus() {
        fetch('/api/qr/status')
        .then(response => response.json())
        .then(data => {
            console.log('QR status:', data);
            
            if (data.available) {
                // QR code is available, update the image with cache-busting
                if (qrCode) {
                    const qrUrl = `/api/qr/image?t=${Date.now()}`;
                    qrCode.src = qrUrl;
                    qrCode.style.display = 'block';
                    qrLoading && (qrLoading.style.display = 'none');
                }
            } else {
                // QR code is not available or expired
                if (qrLoading) {
                    if (data.expired) {
                        qrLoading.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>QR Code kedaluwarsa, memuat ulang...</span>';
                        qrLoading.style.display = 'flex';
                        qrCode && (qrCode.style.display = 'none');
                        
                        // Jika QR code kedaluwarsa, otomatis refresh QR code
                        // Delay sedikit untuk mencegah terlalu banyak request berturut-turut
                        setTimeout(() => {
                            refreshQRCode();
                        }, 1000);
                    } else {
                        qrLoading.innerHTML = '<i class="fas fa-clock"></i><span>Menunggu QR Code...</span>';
                        qrLoading.style.display = 'flex';
                        qrCode && (qrCode.style.display = 'none');
                    }
                }
            }
        })
        .catch(error => {
            console.error('Error checking QR status:', error);
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
                // Check if QR is available immediately
                checkQRStatus();
                
                // Schedule regular QR checks
                startQRStatusChecks();
            } else {
                // Show error message
                if (qrLoading) {
                    qrLoading.innerHTML = `<i class="fas fa-exclamation-triangle"></i><span>${data.message || 'Gagal memuat QR code'}</span>`;
                }
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
            
            // Show error message
            if (qrLoading) {
                qrLoading.innerHTML = '<i class="fas fa-exclamation-triangle"></i><span>Terjadi kesalahan saat memuat QR code</span>';
            }
            
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
    let autoRefreshEnabled = true; // Flag untuk mengaktifkan auto-refresh
    
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
        .then(response => response.json())
        .then(data => {
            console.log('Disconnect response:', data);
            
            if (data.success) {
                // Update UI to show disconnected state
                if (connectionBadge) {
                    connectionBadge.textContent = 'Terputus';
                    connectionBadge.classList.remove('connected');
                    connectionBadge.classList.add('disconnected');
                }
                
                // Switch to QR view
                if (connectedView && qrView) {
                    connectedView.style.display = 'none';
                    qrView.style.display = 'flex';
                }
                
                // Start QR refresh
                refreshQRCode();
            } else {
                // Show error
                alert('Gagal memutuskan WhatsApp: ' + (data.message || 'Error tidak diketahui'));
                
                // Re-enable the button
                if (disconnectBtn) {
                    disconnectBtn.disabled = false;
                    disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i><span>Putuskan Koneksi</span>';
                }
            }
        })
        .catch(error => {
            console.error('Error disconnecting WhatsApp:', error);
            alert('Terjadi kesalahan saat memutuskan WhatsApp');
            
            // Re-enable the button
            if (disconnectBtn) {
                disconnectBtn.disabled = false;
                disconnectBtn.innerHTML = '<i class="fas fa-power-off"></i><span>Putuskan Koneksi</span>';
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
    
    // Function to copy API example to clipboard
    function copyApiExample() {
        const codeBlock = document.querySelector('.code-block code');
        if (!codeBlock) return;
        
        const text = codeBlock.textContent;
        
        // Create temporary textarea to copy from
        const textarea = document.createElement('textarea');
        textarea.value = text;
        document.body.appendChild(textarea);
        textarea.select();
        
        try {
            document.execCommand('copy');
            // Change button icon temporarily
            if (copyApiBtn) {
                copyApiBtn.innerHTML = '<i class="fas fa-check"></i>';
                setTimeout(() => {
                    copyApiBtn.innerHTML = '<i class="fas fa-copy"></i>';
                }, 2000);
            }
        } catch (err) {
            console.error('Could not copy text: ', err);
        }
        
        document.body.removeChild(textarea);
    }
    
    // Attach event listeners
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            window.location.reload();
        });
    }
    
    if (refreshQrBtn) {
        refreshQrBtn.addEventListener('click', refreshQRCode);
    }
    
    if (disconnectBtn) {
        disconnectBtn.addEventListener('click', disconnectWhatsApp);
    }
    
    if (copyApiBtn) {
        copyApiBtn.addEventListener('click', copyApiExample);
    }
    
    // Start QR status checks if we're on QR view
    if (qrView && qrView.style.display !== 'none') {
        startQRStatusChecks();
    }
    
    // Periodically check connection status (every 10 seconds)
    setInterval(() => {
        fetchConnectionStatus(true);
    }, 10000);
    
    // Auto-refresh QR jika berada di halaman QR view dan QR tidak tersedia/kedaluwarsa
    // Perbaiki error dengan menambahkan pengecekan null untuk qrCode
    if (qrView && qrView.style.display !== 'none') {
        // Hanya coba akses style jika qrCode tidak null
        if ((qrCode && !qrCode.style.display) || (qrCode && qrCode.style.display === 'none')) {
            console.log('Auto-refreshing QR on page load');
            setTimeout(refreshQRCode, 500);
        } else if (!qrCode) {
            // Jika qrCode tidak ada sama sekali, juga refresh
            console.log('QR element not found, refreshing QR');
            setTimeout(refreshQRCode, 500);
        }
    }
    
    // Tambahkan event listener untuk sticky header
    document.addEventListener('DOMContentLoaded', function() {
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
        
        // Setup copy buttons for API examples
        const copyButtons = document.querySelectorAll('.copy-btn');
        copyButtons.forEach(button => {
            button.addEventListener('click', function() {
                const codeType = this.getAttribute('data-code');
                const codeElement = document.getElementById(`${codeType}-code`);
                const codeText = codeElement.textContent;
                
                navigator.clipboard.writeText(codeText).then(() => {
                    // Temporarily change icon to show success
                    const icon = this.querySelector('i');
                    icon.classList.remove('fa-copy');
                    icon.classList.add('fa-check');
                    
                    // Change back after 2 seconds
                    setTimeout(() => {
                        icon.classList.remove('fa-check');
                        icon.classList.add('fa-copy');
                    }, 2000);
                });
            });
        });
        
        // Penanganan tombol disconnect
        const disconnectBtn = document.getElementById('disconnect-btn');
        if (disconnectBtn) {
            disconnectBtn.addEventListener('click', function() {
                // Ubah tampilan tombol menjadi loading state
                const originalText = disconnectBtn.innerHTML;
                disconnectBtn.disabled = true;
                disconnectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> <span>Memutuskan...</span>';
                
                // Panggil endpoint disconnect
                fetch('/dashboard/disconnect', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Gagal memutuskan koneksi');
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('Disconnect response:', data);
                    if (data.success) {
                        // Tampilkan notifikasi sukses
                        showNotification('Koneksi WhatsApp berhasil diputuskan', 'success');
                        
                        // Reload halaman setelah 1 detik
                        setTimeout(() => {
                            window.location.reload();
                        }, 1000);
                    } else {
                        // Tampilkan pesan error
                        showNotification(data.message || 'Gagal memutuskan koneksi', 'error');
                        
                        // Kembalikan tombol ke state awal
                        disconnectBtn.disabled = false;
                        disconnectBtn.innerHTML = originalText;
                    }
                })
                .catch(error => {
                    console.error('Error disconnecting:', error);
                    showNotification('Terjadi kesalahan: ' + error.message, 'error');
                    
                    // Kembalikan tombol ke state awal
                    disconnectBtn.disabled = false;
                    disconnectBtn.innerHTML = originalText;
                });
            });
        }
    });
    
    // Fungsi untuk menampilkan notifikasi
    function showNotification(message, type = 'info') {
        // Cek apakah container notifikasi sudah ada
        let notifContainer = document.getElementById('notification-container');
        
        // Jika belum ada, buat container baru
        if (!notifContainer) {
            notifContainer = document.createElement('div');
            notifContainer.id = 'notification-container';
            notifContainer.style.position = 'fixed';
            notifContainer.style.top = '20px';
            notifContainer.style.right = '20px';
            notifContainer.style.zIndex = '1000';
            document.body.appendChild(notifContainer);
        }
        
        // Buat elemen notifikasi
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.padding = '12px 16px';
        notification.style.marginBottom = '10px';
        notification.style.borderRadius = '4px';
        notification.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
        notification.style.display = 'flex';
        notification.style.alignItems = 'center';
        notification.style.justifyContent = 'space-between';
        notification.style.minWidth = '300px';
        notification.style.animation = 'slideInRight 0.3s ease-out forwards';
        
        // Set warna berdasarkan tipe
        if (type === 'success') {
            notification.style.backgroundColor = 'rgba(14, 204, 141, 0.15)';
            notification.style.color = 'var(--success)';
            notification.style.border = '1px solid rgba(14, 204, 141, 0.3)';
        } else if (type === 'error') {
            notification.style.backgroundColor = 'rgba(229, 62, 62, 0.15)';
            notification.style.color = 'var(--danger)';
            notification.style.border = '1px solid rgba(229, 62, 62, 0.3)';
        } else {
            notification.style.backgroundColor = 'rgba(49, 130, 206, 0.15)';
            notification.style.color = 'var(--info)';
            notification.style.border = '1px solid rgba(49, 130, 206, 0.3)';
        }
        
        // Isi konten notifikasi
        notification.innerHTML = `
            <div style="display: flex; align-items: center; gap: 8px;">
                <i class="fas ${type === 'success' ? 'fa-check-circle' : type === 'error' ? 'fa-exclamation-circle' : 'fa-info-circle'}"></i>
                <span>${message}</span>
            </div>
            <button class="close-notification" style="background: none; border: none; cursor: pointer; color: inherit;">
                <i class="fas fa-times"></i>
            </button>
        `;
        
        // Tambahkan notifikasi ke container
        notifContainer.appendChild(notification);
        
        // Tambahkan event listener untuk tombol close
        notification.querySelector('.close-notification').addEventListener('click', function() {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(20px)';
            setTimeout(() => {
                notifContainer.removeChild(notification);
            }, 300);
        });
        
        // Hapus notifikasi setelah 5 detik
        setTimeout(() => {
            if (notification.parentNode === notifContainer) {
                notification.style.opacity = '0';
                notification.style.transform = 'translateX(20px)';
                setTimeout(() => {
                    if (notification.parentNode === notifContainer) {
                        notifContainer.removeChild(notification);
                    }
                }, 300);
            }
        }, 5000);
    }
    
    // Tambahkan event listener untuk tombol Putuskan
    document.addEventListener('DOMContentLoaded', function() {
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
        
        // Setup copy buttons for API examples
        const copyButtons = document.querySelectorAll('.copy-btn');
        copyButtons.forEach(button => {
            button.addEventListener('click', function() {
                const codeType = this.getAttribute('data-code');
                const codeElement = document.getElementById(`${codeType}-code`);
                const codeText = codeElement.textContent;
                
                navigator.clipboard.writeText(codeText).then(() => {
                    // Temporarily change icon to show success
                    const icon = this.querySelector('i');
                    icon.classList.remove('fa-copy');
                    icon.classList.add('fa-check');
                    
                    // Change back after 2 seconds
                    setTimeout(() => {
                        icon.classList.remove('fa-check');
                        icon.classList.add('fa-copy');
                    }, 2000);
                });
            });
        });
        
        // Penanganan tombol disconnect
        const disconnectBtn = document.getElementById('disconnect-btn');
        if (disconnectBtn) {
            disconnectBtn.addEventListener('click', function() {
                // Ubah tampilan tombol menjadi loading state
                const originalText = disconnectBtn.innerHTML;
                disconnectBtn.disabled = true;
                disconnectBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> <span>Memutuskan...</span>';
                
                // Panggil endpoint disconnect
                fetch('/dashboard/disconnect', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Gagal memutuskan koneksi');
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('Disconnect response:', data);
                    if (data.success) {
                        // Tampilkan notifikasi sukses
                        showNotification('Koneksi WhatsApp berhasil diputuskan', 'success');
                        
                        // Reload halaman setelah 1 detik
                        setTimeout(() => {
                            window.location.reload();
                        }, 1000);
                    } else {
                        // Tampilkan pesan error
                        showNotification(data.message || 'Gagal memutuskan koneksi', 'error');
                        
                        // Kembalikan tombol ke state awal
                        disconnectBtn.disabled = false;
                        disconnectBtn.innerHTML = originalText;
                    }
                })
                .catch(error => {
                    console.error('Error disconnecting:', error);
                    showNotification('Terjadi kesalahan: ' + error.message, 'error');
                    
                    // Kembalikan tombol ke state awal
                    disconnectBtn.disabled = false;
                    disconnectBtn.innerHTML = originalText;
                });
            });
        }
    });
});
