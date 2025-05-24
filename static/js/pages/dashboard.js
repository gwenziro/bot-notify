/**
 * Dashboard Page JavaScript
 * Manages dashboard functionality and data fetching
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize dashboard
    initDashboard();
    
    // Set up event listeners for dashboard elements
    setupDashboardEventListeners();
    
    // Initial data fetch
    fetchDashboardData();
    
    // Set up interval for refreshing data
    setInterval(fetchDashboardData, 10000); // Refresh every 10 seconds
    
    // Update uptime counter
    setInterval(updateUptimeDisplay, 1000);
});

// Dashboard initialization
function initDashboard() {
    // Save dashboard start time for uptime calculation if not already set
    if (!sessionStorage.getItem('dashboardStartTime')) {
        sessionStorage.setItem('dashboardStartTime', new Date().toISOString());
        const systemStartTimeEl = document.getElementById('system-start-time');
        if (systemStartTimeEl) {
            systemStartTimeEl.textContent = formatTime(new Date());
        }
    } else {
        const systemStartTimeEl = document.getElementById('system-start-time');
        if (systemStartTimeEl) {
            systemStartTimeEl.textContent = formatTime(new Date(sessionStorage.getItem('dashboardStartTime')));
        }
    }
    
    // Initialize charts
    initCharts();
}

// Setup dashboard event listeners
function setupDashboardEventListeners() {
    // Refresh status
    document.getElementById('refresh-status')?.addEventListener('click', function() {
        fetchDashboardData();
        addActivity('Refresh status sistem dilakukan');
    });
    
    // Connect/disconnect buttons
    document.getElementById('reconnect-btn')?.addEventListener('click', function() {
        this.disabled = true;
        this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Menghubungkan...';
        
        reconnectWhatsApp()
            .then(data => {
                addActivity('Permintaan koneksi ulang berhasil dikirim');
                setTimeout(fetchDashboardData, 2000);
            })
            .catch(error => {
                addActivity('Error saat menghubungkan WhatsApp: ' + error.message, 'error');
            })
            .finally(() => {
                this.disabled = false;
                this.innerHTML = '<i class="fas fa-sync"></i> Hubungkan Ulang WhatsApp';
            });
    });
    
    document.getElementById('disconnect-btn')?.addEventListener('click', function() {
        confirmAction('Apakah Anda yakin ingin memutuskan koneksi WhatsApp?', () => {
            this.disabled = true;
            this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memutuskan...';
            
            disconnectWhatsApp()
                .then(data => {
                    addActivity('WhatsApp berhasil diputuskan');
                    setTimeout(fetchDashboardData, 2000);
                })
                .catch(error => {
                    addActivity('Error saat memutuskan WhatsApp: ' + error.message, 'error');
                })
                .finally(() => {
                    this.disabled = false;
                    this.innerHTML = '<i class="fas fa-power-off"></i> Putuskan WhatsApp';
                });
        });
    });
    
    document.getElementById('scan-qr-btn')?.addEventListener('click', function() {
        window.location.href = '/connectivity';
    });
    
    document.getElementById('check-groups-btn')?.addEventListener('click', function() {
        fetchGroups();
    });
    
    // Chart period selector
    document.getElementById('chart-period')?.addEventListener('change', function() {
        updateMessagesChart(this.value);
    });
    
    // Clear activity log
    document.getElementById('clear-activity')?.addEventListener('click', function() {
        clearActivityLog();
    });
}

// Fetch all dashboard data from APIs
function fetchDashboardData() {
    Promise.all([
        getConnectionStatus(),
        getGroups()
    ]).then(([statusData, groupsData]) => {
        // Update connection status
        updateConnectionUI(statusData);
        
        // Update groups count
        const groupCount = document.getElementById('group-count');
        if (groupCount && groupsData && groupsData.grup) {
            groupCount.textContent = groupsData.grup.length;
        }
        
        // Mock system resources data (replace with real API calls if available)
        updateSystemResources({
            cpu: Math.floor(Math.random() * 100),
            memory: {
                used: Math.floor(Math.random() * 500),
                total: 1024
            },
            disk: {
                used: Math.floor(Math.random() * 50),
                total: 100
            }
        });
        
    }).catch(error => {
        console.error('Error fetching dashboard data:', error);
        addActivity('Error saat mengambil data: ' + error.message, 'error');
    });
}

// Update system resources UI
function updateSystemResources(data) {
    // Update CPU usage
    const cpuUsage = document.getElementById('cpu-usage');
    const cpuBar = document.getElementById('cpu-bar');
    
    if (cpuUsage && cpuBar) {
        cpuUsage.textContent = `${data.cpu}%`;
        cpuBar.style.width = `${data.cpu}%`;
        
        // Set appropriate color based on usage
        if (data.cpu < 50) {
            cpuBar.className = 'progress-bar success';
        } else if (data.cpu < 80) {
            cpuBar.className = 'progress-bar warning';
        } else {
            cpuBar.className = 'progress-bar danger';
        }
    }
    
    // Update Memory usage
    const memoryUsage = document.getElementById('memory-usage');
    const memoryBar = document.getElementById('memory-bar');
    
    if (memoryUsage && memoryBar) {
        const memoryPercentage = (data.memory.used / data.memory.total) * 100;
        
        memoryUsage.textContent = `${data.memory.used} MB / ${data.memory.total} MB`;
        memoryBar.style.width = `${memoryPercentage}%`;
        
        // Set appropriate color based on usage
        if (memoryPercentage < 50) {
            memoryBar.className = 'progress-bar success';
        } else if (memoryPercentage < 80) {
            memoryBar.className = 'progress-bar warning';
        } else {
            memoryBar.className = 'progress-bar danger';
        }
    }
    
    // Update Disk usage
    const diskUsage = document.getElementById('disk-usage');
    const diskBar = document.getElementById('disk-bar');
    
    if (diskUsage && diskBar) {
        const diskPercentage = (data.disk.used / data.disk.total) * 100;
        
        diskUsage.textContent = `${data.disk.used} GB / ${data.disk.total} GB`;
        diskBar.style.width = `${diskPercentage}%`;
        
        // Set appropriate color based on usage
        if (diskPercentage < 50) {
            diskBar.className = 'progress-bar success';
        } else if (diskPercentage < 80) {
            diskBar.className = 'progress-bar warning';
        } else {
            diskBar.className = 'progress-bar danger';
        }
    }
}

// Initialize Charts.js
function initCharts() {
    const ctx = document.getElementById('messages-chart')?.getContext('2d');
    if (!ctx) return;
    
    // Generate some mock data
    const mockData = {
        labels: Array.from({length: 7}, (_, i) => {
            const today = new Date();
            today.setDate(today.getDate() - (6 - i));
            return today.toLocaleDateString('id-ID', {weekday: 'short'});
        }),
        datasets: [
            {
                label: 'Pesan Masuk',
                data: Array.from({length: 7}, () => Math.floor(Math.random() * 50)),
                borderColor: '#1d6a8b',
                backgroundColor: 'rgba(29, 106, 139, 0.2)',
                borderWidth: 2,
                tension: 0.3,
                fill: true
            },
            {
                label: 'Pesan Keluar',
                data: Array.from({length: 7}, () => Math.floor(Math.random() * 30)),
                borderColor: '#0ecc8d',
                backgroundColor: 'rgba(14, 204, 141, 0.2)',
                borderWidth: 2,
                tension: 0.3,
                fill: true
            }
        ]
    };
    
    window.messagesChart = new Chart(ctx, {
        type: 'line',
        data: mockData,
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'top',
                },
                tooltip: {
                    mode: 'index',
                    intersect: false,
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        precision: 0
                    }
                }
            }
        }
    });
}

// Update messages chart based on selected period
function updateMessagesChart(period) {
    if (!window.messagesChart) return;
    
    // Here you would fetch real data based on the period
    // For now, we'll just update the mock data
    
    let days = 7;
    switch(period) {
        case 'day':
            days = 1;
            break;
        case 'week':
            days = 7;
            break;
        case 'month':
            days = 30;
            break;
    }
    
    // Generate labels based on the period
    let labels = [];
    if (period === 'day') {
        // For day, show hours
        labels = Array.from({length: 12}, (_, i) => `${i*2}:00`);
    } else {
        // For week/month, show dates
        labels = Array.from({length: days}, (_, i) => {
            const date = new Date();
            date.setDate(date.getDate() - (days - 1) + i);
            return date.toLocaleDateString('id-ID', {weekday: 'short', day: 'numeric'});
        });
    }
    
    // Update chart data
    window.messagesChart.data.labels = labels;
    window.messagesChart.data.datasets[0].data = Array.from({length: labels.length}, 
        () => Math.floor(Math.random() * 50));
    window.messagesChart.data.datasets[1].data = Array.from({length: labels.length}, 
        () => Math.floor(Math.random() * 30));
    window.messagesChart.update();
    
    addActivity(`Periode chart diubah ke: ${getReadablePeriod(period)}`);
}

// Update uptime counter display
function updateUptimeDisplay() {
    const uptimeElement = document.getElementById('uptime');
    if (!uptimeElement) return;
    
    const startTime = sessionStorage.getItem('dashboardStartTime');
    if (!startTime) return;
    
    uptimeElement.textContent = calculateUptime(startTime);
}

// Fetch groups list
function fetchGroups() {
    const checkGroupsBtn = document.getElementById('check-groups-btn');
    if (checkGroupsBtn) {
        checkGroupsBtn.disabled = true;
        checkGroupsBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Memuat...';
    }
    
    getGroups()
        .then(data => {
            const groupCount = document.getElementById('group-count');
            if (groupCount && data && data.grup) {
                groupCount.textContent = data.grup.length;
                addActivity(`Berhasil memuat ${data.grup.length} grup`);
            } else {
                addActivity('Tidak ada grup yang tersedia');
            }
        })
        .catch(error => {
            addActivity('Gagal memuat daftar grup: ' + error.message, 'error');
        })
        .finally(() => {
            if (checkGroupsBtn) {
                checkGroupsBtn.disabled = false;
                checkGroupsBtn.innerHTML = '<i class="fas fa-users"></i> Cek Grup Tersedia';
            }
        });
}

// Add activity to the activity feed
function addActivity(message, type = 'info') {
    const activityFeed = document.getElementById('activity-feed');
    if (!activityFeed) return;
    
    const now = new Date();
    
    // Create activity item
    const li = document.createElement('li');
    li.className = 'activity-item';
    
    // Create activity icon
    const iconDiv = document.createElement('div');
    iconDiv.className = 'activity-icon';
    
    const icon = document.createElement('i');
    switch (type) {
        case 'error':
            icon.className = 'fas fa-exclamation-circle';
            iconDiv.style.color = 'var(--color-danger)';
            iconDiv.style.backgroundColor = 'rgba(229, 62, 62, 0.1)';
            break;
        case 'warning':
            icon.className = 'fas fa-exclamation-triangle';
            iconDiv.style.color = 'var(--color-warning)';
            iconDiv.style.backgroundColor = 'rgba(246, 173, 85, 0.1)';
            break;
        case 'success':
            icon.className = 'fas fa-check-circle';
            iconDiv.style.color = 'var(--color-success)';
            iconDiv.style.backgroundColor = 'rgba(14, 204, 141, 0.1)';
            break;
        default:
            icon.className = 'fas fa-info-circle';
            break;
    }
    iconDiv.appendChild(icon);
    
    // Create activity content
    const contentDiv = document.createElement('div');
    contentDiv.className = 'activity-content';
    
    const title = document.createElement('p');
    title.className = 'activity-title';
    title.textContent = message;
    
    const timeSpan = document.createElement('span');
    timeSpan.className = 'activity-time';
    timeSpan.textContent = formatTime(now);
    
    contentDiv.appendChild(title);
    contentDiv.appendChild(timeSpan);
    
    // Append elements to the activity item
    li.appendChild(iconDiv);
    li.appendChild(contentDiv);
    
    // Add to the beginning of the feed
    activityFeed.insertBefore(li, activityFeed.firstChild);
    
    // Limit the number of items
    if (activityFeed.children.length > 15) {
        activityFeed.removeChild(activityFeed.lastChild);
    }
    
    // Save to local storage for persistence
    saveActivityLog([message, type, now.toISOString()]);
}

// Clear activity log
function clearActivityLog() {
    const activityFeed = document.getElementById('activity-feed');
    if (!activityFeed) return;
    
    // Keep only the first item (system start)
    while (activityFeed.children.length > 1) {
        activityFeed.removeChild(activityFeed.lastChild);
    }
    
    // Clear saved log
    localStorage.removeItem('activityLog');
    
    addActivity('Log aktivitas dibersihkan');
}

// Save activity log to local storage
function saveActivityLog(activity) {
    const log = JSON.parse(localStorage.getItem('activityLog') || '[]');
    log.unshift(activity);
    
    // Limit to 50 entries
    while (log.length > 50) {
        log.pop();
    }
    
    localStorage.setItem('activityLog', JSON.stringify(log));
}

// Get human-readable period name
function getReadablePeriod(period) {
    switch (period) {
        case 'day': return 'Hari Ini';
        case 'week': return '7 Hari Terakhir';
        case 'month': return '30 Hari Terakhir';
        default: return period;
    }
}
