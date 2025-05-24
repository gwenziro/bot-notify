/**
 * Logs Page JavaScript
 * Enhanced version with modern data loading and UI interactions
 */

class LogsManager {
    constructor() {
        // State
        this.logs = [];
        this.currentPage = 1;
        this.totalPages = 0;
        this.totalLogs = 0;
        this.limit = 25;
        this.loading = false;
        this.filters = {
            level: 'all',
            source: 'all',
            search: '',
            dateFrom: '',
            dateTo: ''
        };

        // Elements
        this.logsTable = document.getElementById('logs-table');
        this.logsTableBody = document.getElementById('log-entries');
        this.logCount = document.getElementById('log-count');
        this.currentPageEl = document.getElementById('current-page');
        this.totalPagesEl = document.getElementById('total-pages');
        this.logDetailModal = document.getElementById('log-detail-modal');

        // Filter elements
        this.levelFilter = document.getElementById('log-level');
        this.sourceFilter = document.getElementById('log-source');
        this.searchInput = document.getElementById('log-search');
        this.dateFromInput = document.getElementById('log-date-from');
        this.dateToInput = document.getElementById('log-date-to');

        // Buttons
        this.refreshBtn = document.getElementById('refresh-logs');
        this.clearLogsBtn = document.getElementById('clear-logs');
        this.exportCsvBtn = document.getElementById('export-csv');
        this.exportJsonBtn = document.getElementById('export-json');
        this.prevPageBtn = document.getElementById('prev-page');
        this.nextPageBtn = document.getElementById('next-page');

        // Initialize
        this.init();
    }

    init() {
        // Setup event listeners
        this.setupFilterListeners();
        this.setupButtonListeners();
        this.setupModalListeners();

        // Initial load
        this.loadLogs();
    }

    setupFilterListeners() {
        // Level filter
        if (this.levelFilter) {
            this.levelFilter.addEventListener('change', () => {
                this.filters.level = this.levelFilter.value;
                this.currentPage = 1;
                this.loadLogs();
            });
        }

        // Source filter
        if (this.sourceFilter) {
            this.sourceFilter.addEventListener('change', () => {
                this.filters.source = this.sourceFilter.value;
                this.currentPage = 1;
                this.loadLogs();
            });
        }

        // Search input (with debounce)
        if (this.searchInput) {
            let debounceTimeout;
            this.searchInput.addEventListener('input', () => {
                clearTimeout(debounceTimeout);
                debounceTimeout = setTimeout(() => {
                    this.filters.search = this.searchInput.value;
                    this.currentPage = 1;
                    this.loadLogs();
                }, 500);
            });
        }

        // Date filters
        if (this.dateFromInput) {
            this.dateFromInput.addEventListener('change', () => {
                this.filters.dateFrom = this.dateFromInput.value;
                this.currentPage = 1;
                this.loadLogs();
            });
        }

        if (this.dateToInput) {
            this.dateToInput.addEventListener('change', () => {
                this.filters.dateTo = this.dateToInput.value;
                this.currentPage = 1;
                this.loadLogs();
            });
        }
    }

    setupButtonListeners() {
        // Refresh button
        if (this.refreshBtn) {
            this.refreshBtn.addEventListener('click', () => {
                this.refreshLogs();
            });
        }

        // Clear logs button
        if (this.clearLogsBtn) {
            this.clearLogsBtn.addEventListener('click', () => {
                this.confirmAndClearLogs();
            });
        }

        // Export buttons
        if (this.exportCsvBtn) {
            this.exportCsvBtn.addEventListener('click', () => {
                this.exportLogs('csv');
            });
        }

        if (this.exportJsonBtn) {
            this.exportJsonBtn.addEventListener('click', () => {
                this.exportLogs('json');
            });
        }

        // Pagination buttons
        if (this.prevPageBtn) {
            this.prevPageBtn.addEventListener('click', () => {
                if (this.currentPage > 1) {
                    this.currentPage--;
                    this.loadLogs();
                }
            });
        }

        if (this.nextPageBtn) {
            this.nextPageBtn.addEventListener('click', () => {
                if (this.currentPage < this.totalPages) {
                    this.currentPage++;
                    this.loadLogs();
                }
            });
        }
    }

    setupModalListeners() {
        // Show log details when clicking on a log row or detail button
        document.addEventListener('click', (e) => {
            const detailBtn = e.target.closest('.log-detail-btn');
            if (detailBtn) {
                const logId = detailBtn.dataset.id;
                this.showLogDetail(logId);
            }
        });

        // Close modal with close button or clicking outside
        if (this.logDetailModal) {
            this.logDetailModal.addEventListener('click', (e) => {
                if (e.target === this.logDetailModal || e.target.classList.contains('modal-close')) {
                    this.hideLogDetail();
                }
            });

            // Close with ESC key
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && this.logDetailModal.classList.contains('active')) {
                    this.hideLogDetail();
                }
            });
        }
    }

    // Load logs from API with current filters and pagination
    async loadLogs() {
        this.setLoading(true);

        try {
            const params = new URLSearchParams({
                page: this.currentPage,
                limit: this.limit
            });

            // Add filters if not default
            if (this.filters.level !== 'all') params.append('level', this.filters.level);
            if (this.filters.source !== 'all') params.append('source', this.filters.source);
            if (this.filters.search) params.append('search', this.filters.search);
            if (this.filters.dateFrom) params.append('from', this.filters.dateFrom);
            if (this.filters.dateTo) params.append('to', this.filters.dateTo);

            const response = await fetch(`/api/logs?${params.toString()}`);

            if (!response.ok) {
                throw new Error(`Error fetching logs: ${response.statusText}`);
            }

            const data = await response.json();

            if (data.success) {
                this.logs = data.logs || [];
                this.totalPages = data.totalPages || 0;
                this.totalLogs = data.totalLogs || 0;

                this.renderLogs();
                this.updatePagination();
            } else {
                throw new Error(data.error || 'Failed to load logs');
            }
        } catch (error) {
            this.showError('Failed to load logs', error.message);
        } finally {
            this.setLoading(false);
        }
    }

    // Render logs to table
    renderLogs() {
        if (!this.logsTableBody) return;

        // Clear table first
        this.logsTableBody.innerHTML = '';

        if (this.logs.length === 0) {
            // No logs found
            const emptyRow = document.createElement('tr');
            emptyRow.innerHTML = `
                <td colspan="5" class="text-center p-4">
                    <div class="empty-state">
                        <i class="fas fa-clipboard-list text-muted"></i>
                        <p class="mt-2">Tidak ada log yang ditemukan</p>
                    </div>
                </td>
            `;
            this.logsTableBody.appendChild(emptyRow);
        } else {
            // Add each log to table
            this.logs.forEach(log => {
                const row = document.createElement('tr');
                row.className = `log-entry log-level-${log.level.toLowerCase()}`;
                row.dataset.id = log.id;

                const timestamp = new Date(log.timestamp);
                const formattedTime = this.formatDateTime(timestamp);

                row.innerHTML = `
                    <td class="log-time">${formattedTime}</td>
                    <td class="log-level">
                        <span class="badge badge-${this.getLevelClass(log.level)}">${log.level}</span>
                    </td>
                    <td class="log-source">${log.source}</td>
                    <td class="log-message">${this.escapeHtml(log.message)}</td>
                    <td class="log-actions">
                        <button class="btn-icon btn-sm log-detail-btn" data-id="${log.id}" title="Detail">
                            <i class="fas fa-info-circle"></i>
                        </button>
                    </td>
                `;

                this.logsTableBody.appendChild(row);
            });
        }

        // Update count in UI
        if (this.logCount) {
            this.logCount.textContent = this.totalLogs;
        }
    }

    // Update pagination UI
    updatePagination() {
        if (this.currentPageEl) {
            this.currentPageEl.textContent = this.currentPage;
        }
        if (this.totalPagesEl) {
            this.totalPagesEl.textContent = this.totalPages;
        }

        // Enable/disable prev/next buttons
        if (this.prevPageBtn) {
            this.prevPageBtn.disabled = this.currentPage <= 1;
        }
        if (this.nextPageBtn) {
            this.nextPageBtn.disabled = this.currentPage >= this.totalPages;
        }
    }

    // Refresh logs with animation
    refreshLogs() {
        this.refreshBtn.querySelector('i').classList.add('fa-spin');
        this.loadLogs().then(() => {
            setTimeout(() => {
                this.refreshBtn.querySelector('i').classList.remove('fa-spin');
            }, 500);
        });
    }

    // Clear all logs with confirmation
    confirmAndClearLogs() {
        const confirmed = window.confirm('Apakah Anda yakin ingin menghapus semua log? Tindakan ini tidak dapat dibatalkan.');
        if (confirmed) {
            this.clearLogs();
        }
    }

    // Clear logs API call
    async clearLogs() {
        this.setLoading(true);

        try {
            const response = await fetch('/api/logs/clear', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`Error clearing logs: ${response.statusText}`);
            }

            const data = await response.json();

            if (data.success) {
                this.showSuccess('Logs cleared successfully');
                this.loadLogs();  // Reload to show empty state
            } else {
                throw new Error(data.error || 'Failed to clear logs');
            }
        } catch (error) {
            this.showError('Failed to clear logs', error.message);
        } finally {
            this.setLoading(false);
        }
    }

    // Export logs (open in new tab/download)
    exportLogs(format) {
        const params = new URLSearchParams({
            format: format
        });

        // Add filters if not default
        if (this.filters.level !== 'all') params.append('level', this.filters.level);
        if (this.filters.source !== 'all') params.append('source', this.filters.source);
        if (this.filters.search) params.append('search', this.filters.search);
        if (this.filters.dateFrom) params.append('from', this.filters.dateFrom);
        if (this.filters.dateTo) params.append('to', this.filters.dateTo);

        // Create export URL and open in new tab/download
        const exportUrl = `/api/logs/export?${params.toString()}`;
        window.open(exportUrl, '_blank');
    }

    // Show log detail in modal
    showLogDetail(logId) {
        const log = this.logs.find(l => l.id === logId);
        if (!log || !this.logDetailModal) return;

        // Format timestamp
        const timestamp = new Date(log.timestamp);
        const formattedTime = this.formatDateTime(timestamp, true);

        // Format data as JSON
        let dataHtml = '';
        if (log.data && Object.keys(log.data).length > 0) {
            const formattedJson = JSON.stringify(log.data, null, 2);
            dataHtml = `<pre class="code-block">${this.escapeHtml(formattedJson)}</pre>`;
        } else {
            dataHtml = '<p class="text-muted">No additional data</p>';
        }

        // Update modal content
        const modalContent = this.logDetailModal.querySelector('.modal-body');
        if (modalContent) {
            modalContent.innerHTML = `
                <div class="log-detail-content">
                    <div class="mb-4">
                        <label class="text-muted">Timestamp</label>
                        <p class="mb-2">${formattedTime}</p>
                        
                        <label class="text-muted">Level</label>
                        <p class="mb-2">
                            <span class="badge badge-${this.getLevelClass(log.level)}">${log.level}</span>
                        </p>
                        
                        <label class="text-muted">Source</label>
                        <p class="mb-2">${log.source}</p>
                    </div>
                    
                    <div class="mb-4">
                        <label class="text-muted">Message</label>
                        <p class="mb-2">${this.escapeHtml(log.message)}</p>
                    </div>
                    
                    <div>
                        <label class="text-muted">Additional Data</label>
                        ${dataHtml}
                    </div>
                </div>
            `;
        }

        // Set title if exists
        const modalTitle = this.logDetailModal.querySelector('.modal-title');
        if (modalTitle) {
            modalTitle.textContent = `Log Detail (${log.level})`;
        }

        // Show modal
        this.logDetailModal.classList.add('active');
        document.body.classList.add('modal-open');
    }

    // Hide log detail modal
    hideLogDetail() {
        if (this.logDetailModal) {
            this.logDetailModal.classList.remove('active');
            document.body.classList.remove('modal-open');
        }
    }

    // Set loading state
    setLoading(isLoading) {
        this.loading = isLoading;

        // Toggle loading state on buttons
        if (this.refreshBtn) {
            this.refreshBtn.disabled = isLoading;
            if (!isLoading) {
                this.refreshBtn.querySelector('i')?.classList.remove('fa-spin');
            }
        }

        if (this.clearLogsBtn) {
            this.clearLogsBtn.disabled = isLoading;
        }

        // Show loading indicator on table
        if (this.logsTable) {
            if (isLoading) {
                this.logsTable.classList.add('loading');
            } else {
                this.logsTable.classList.remove('loading');
            }
        }
    }

    // Show success message
    showSuccess(message) {
        // Use toast notification or alert
        if (window.toast) {
            window.toast.success(message);
        } else {
            alert(message);
        }
    }

    // Show error message
    showError(title, details = '') {
        // Use toast notification or alert
        if (window.toast) {
            window.toast.error(`${title}: ${details}`);
        } else {
            console.error(`${title}: ${details}`);
            alert(`${title}: ${details}`);
        }
    }

    // Format date and time
    formatDateTime(date, includeMs = false) {
        if (!(date instanceof Date)) {
            date = new Date(date);
        }

        const options = {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        };

        let formattedDate = date.toLocaleString('id-ID', options);
        
        // Add milliseconds if requested
        if (includeMs) {
            formattedDate += `.${date.getMilliseconds().toString().padStart(3, '0')}`;
        }
        
        return formattedDate;
    }

    // Get CSS class for log level
    getLevelClass(level) {
        level = level.toLowerCase();
        switch (level) {
            case 'debug': return 'info';
            case 'info': return 'success';
            case 'warning': return 'warning';
            case 'error': 
            case 'fatal': return 'danger';
            default: return 'secondary';
        }
    }

    // Escape HTML to prevent XSS
    escapeHtml(unsafe) {
        if (typeof unsafe !== 'string') return unsafe;
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    window.logsManager = new LogsManager();
});
