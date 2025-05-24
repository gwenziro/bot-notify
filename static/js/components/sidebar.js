/**
 * Sidebar Component JavaScript
 * Handles sidebar functionality including navigation and collapse state
 */

class SidebarComponent {
    constructor() {
        // Elements
        this.sidebar = document.getElementById('sidebar');
        this.dashboardContainer = document.querySelector('.dashboard-container');
        this.sidebarToggle = document.getElementById('sidebar-toggle');
        this.sidebarCollapse = document.getElementById('sidebar-collapse');
        this.navItems = document.querySelectorAll('.nav-item');
        this.navLinks = document.querySelectorAll('.nav-link');
        this.submenuToggles = document.querySelectorAll('[data-toggle="collapse"]');
        this.statusIndicator = document.getElementById('sidebar-status-indicator');
        this.connectionStatus = document.getElementById('sidebar-connection-status');
        
        // Initialize component
        this.init();
    }
    
    init() {
        // Setup event listeners
        this.setupEventListeners();
        
        // Load saved collapse state
        this.loadCollapseState();
        
        // Highlight current page
        this.highlightCurrentPage();
    }
    
    setupEventListeners() {
        // Sidebar toggle on mobile
        if (this.sidebarToggle) {
            this.sidebarToggle.addEventListener('click', () => {
                this.toggleSidebar();
            });
        }
        
        // Sidebar collapse button
        if (this.sidebarCollapse) {
            this.sidebarCollapse.addEventListener('click', () => {
                this.collapseSidebar();
            });
        }
        
        // Submenu toggles
        this.submenuToggles.forEach(toggle => {
            toggle.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleSubmenu(toggle);
            });
        });
        
        // Close sidebar when clicking outside on mobile
        document.addEventListener('click', (e) => {
            if (window.innerWidth < 992 && 
                this.dashboardContainer && 
                this.dashboardContainer.classList.contains('sidebar-open')) {
                
                if (this.sidebar && 
                    !this.sidebar.contains(e.target) && 
                    e.target !== this.sidebarToggle && 
                    !this.sidebarToggle.contains(e.target)) {
                    this.closeSidebar();
                }
            }
        });
        
        // Track window resize
        window.addEventListener('resize', () => {
            this.handleResize();
        });
    }
    
    toggleSidebar() {
        if (!this.dashboardContainer) return;
        
        this.dashboardContainer.classList.toggle('sidebar-open');
    }
    
    closeSidebar() {
        if (!this.dashboardContainer) return;
        
        this.dashboardContainer.classList.remove('sidebar-open');
    }
    
    collapseSidebar() {
        if (!this.dashboardContainer) return;
        
        this.dashboardContainer.classList.toggle('sidebar-collapsed');
        
        // Save state to localStorage
        const isCollapsed = this.dashboardContainer.classList.contains('sidebar-collapsed');
        localStorage.setItem('sidebar-collapsed', isCollapsed);
    }
    
    loadCollapseState() {
        if (!this.dashboardContainer) return;
        
        // Check localStorage for saved collapse state
        const isCollapsed = localStorage.getItem('sidebar-collapsed') === 'true';
        
        if (isCollapsed) {
            this.dashboardContainer.classList.add('sidebar-collapsed');
        }
    }
    
    highlightCurrentPage() {
        const currentPath = window.location.pathname;
        
        this.navLinks.forEach(link => {
            // Skip submenu toggles
            if (link.hasAttribute('data-toggle')) return;
            
            const linkPath = link.getAttribute('href');
            
            // Check if current URL matches link or is a child path
            if (linkPath && (linkPath === currentPath || 
                (linkPath !== '/' && currentPath.startsWith(linkPath)))) {
                
                // Add active class
                link.classList.add('active');
                
                // If inside submenu, expand parent
                const submenu = link.closest('.nav-submenu');
                if (submenu) {
                    submenu.classList.add('show');
                    
                    // Find toggle for this submenu
                    const toggle = document.querySelector(`[data-target="#${submenu.id}"]`);
                    if (toggle) {
                        toggle.setAttribute('aria-expanded', 'true');
                        toggle.classList.add('active');
                    }
                }
                
                // Update page title in breadcrumb
                const pageTitle = link.querySelector('.nav-text')?.textContent || '';
                const currentPageTitle = document.getElementById('current-page-title');
                if (currentPageTitle && pageTitle) {
                    currentPageTitle.textContent = pageTitle;
                }
            }
        });
    }
    
    toggleSubmenu(toggle) {
        const targetId = toggle.getAttribute('data-target');
        const submenu = document.querySelector(targetId);
        
        if (submenu) {
            submenu.classList.toggle('show');
            
            // Update aria-expanded attribute
            const isExpanded = submenu.classList.contains('show');
            toggle.setAttribute('aria-expanded', isExpanded);
        }
    }
    
    handleResize() {
        // Close sidebar on mobile when resizing to larger screens
        if (window.innerWidth >= 992 && this.dashboardContainer) {
            this.dashboardContainer.classList.remove('sidebar-open');
        }
    }
    
    updateConnectionStatus(status) {
        if (!this.statusIndicator || !this.connectionStatus) return;
        
        // Remove existing classes
        this.statusIndicator.className = 'status-indicator';
        
        // Set appropriate class and text based on status
        switch (status) {
            case 'connected':
                this.statusIndicator.classList.add('online');
                this.connectionStatus.textContent = 'Terhubung';
                break;
            case 'connecting':
                this.statusIndicator.classList.add('connecting');
                this.connectionStatus.textContent = 'Menghubungkan...';
                break;
            case 'disconnected':
                this.connectionStatus.textContent = 'Terputus';
                break;
            default:
                this.connectionStatus.textContent = 'Status Tidak Diketahui';
        }
    }
}

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    window.sidebarComponent = new SidebarComponent();
});
