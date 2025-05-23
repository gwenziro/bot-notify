/**
 * Dashboard Layout JavaScript
 * Handles the layout functionality of the dashboard
 */

// Wait for DOM to be fully loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize dashboard layout components
    initSidebar();
    initDropdowns();
    initActiveNav();
    initServerTime();
    setupSearchFunctionality();
    listenForConnectionStatus();
});

/**
 * Initialize sidebar functionality
 */
function initSidebar() {
    const sidebarToggle = document.getElementById('sidebar-toggle');
    const sidebarCollapse = document.getElementById('sidebar-collapse');
    const dashboardContainer = document.querySelector('.dashboard-container');
    
    // Toggle sidebar on mobile
    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', function() {
            dashboardContainer.classList.toggle('sidebar-open');
        });
    }
    
    // Collapse/expand sidebar
    if (sidebarCollapse) {
        sidebarCollapse.addEventListener('click', function() {
            dashboardContainer.classList.toggle('sidebar-collapsed');
            
            // Save preference to localStorage
            const isCollapsed = dashboardContainer.classList.contains('sidebar-collapsed');
            localStorage.setItem('sidebar-collapsed', isCollapsed);
        });
    }
    
    // Check for saved sidebar state
    const savedSidebarState = localStorage.getItem('sidebar-collapsed');
    if (savedSidebarState === 'true') {
        dashboardContainer.classList.add('sidebar-collapsed');
    }
    
    // Handle submenu toggles
    const submenuToggles = document.querySelectorAll('[data-toggle="collapse"]');
    submenuToggles.forEach(toggle => {
        toggle.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('data-target');
            const targetElement = document.querySelector(targetId);
            
            if (targetElement) {
                targetElement.classList.toggle('show');
                this.setAttribute('aria-expanded', 
                    targetElement.classList.contains('show'));
            }
        });
    });
    
    // Close sidebar when clicking outside on mobile
    document.addEventListener('click', function(e) {
        if (window.innerWidth < 992 && 
            dashboardContainer.classList.contains('sidebar-open')) {
            
            const sidebar = document.getElementById('sidebar');
            const sidebarToggle = document.getElementById('sidebar-toggle');
            
            if (!sidebar.contains(e.target) && 
                e.target !== sidebarToggle && 
                !sidebarToggle.contains(e.target)) {
                dashboardContainer.classList.remove('sidebar-open');
            }
        }
    });
}

/**
 * Initialize dropdown functionality
 */
function initDropdowns() {
    const dropdownToggles = document.querySelectorAll('.dropdown');
    
    dropdownToggles.forEach(dropdown => {
        const toggle = dropdown.querySelector('.btn-icon, .user-profile');
        const menu = dropdown.querySelector('.dropdown-menu');
        
        if (toggle && menu) {
            toggle.addEventListener('click', function(e) {
                e.stopPropagation();
                menu.classList.toggle('show');
                
                // Close other dropdowns
                dropdownToggles.forEach(otherDropdown => {
                    if (otherDropdown !== dropdown) {
                        const otherMenu = otherDropdown.querySelector('.dropdown-menu');
                        if (otherMenu) {
                            otherMenu.classList.remove('show');
                        }
                    }
                });
            });
        }
    });
    
    // Close dropdowns when clicking outside
    document.addEventListener('click', function() {
        dropdownToggles.forEach(dropdown => {
            const menu = dropdown.querySelector('.dropdown-menu');
            if (menu) {
                menu.classList.remove('show');
            }
        });
    });
}

/**
 * Highlight active navigation item based on current URL
 */
function initActiveNav() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-link');
    
    navLinks.forEach(link => {
        // Skip submenu toggles
        if (link.hasAttribute('data-toggle')) return;
        
        const linkPath = link.getAttribute('href');
        
        // Mark as active if paths match
        if (linkPath && (linkPath === currentPath || 
            (linkPath !== '/' && currentPath.startsWith(linkPath)))) {
            link.classList.add('active');
            
            // If inside submenu, expand the parent
            const submenu = link.closest('.nav-submenu');
            if (submenu) {
                submenu.classList.add('show');
                const toggle = document.querySelector(`[data-target="#${submenu.id}"]`);
                if (toggle) {
                    toggle.setAttribute('aria-expanded', 'true');
                }
            }
            
            // Update page title
            const pageTitle = link.querySelector('.nav-text')?.textContent || '';
            const currentPageTitle = document.getElementById('current-page-title');
            if (currentPageTitle && pageTitle) {
                currentPageTitle.textContent = pageTitle;
            }
        }
    });
}

/**
 * Initialize and update server time
 */
function initServerTime() {
    const serverTimeElement = document.getElementById('server-time');
    if (!serverTimeElement) return;
    
    function updateTime() {
        const now = new Date();
        const hours = now.getHours().toString().padStart(2, '0');
        const minutes = now.getMinutes().toString().padStart(2, '0');
        const seconds = now.getSeconds().toString().padStart(2, '0');
        
        serverTimeElement.textContent = `${hours}:${minutes}:${seconds}`;
    }
    
    // Update immediately and then every second
    updateTime();
    setInterval(updateTime, 1000);
}

/**
 * Setup search functionality
 */
function setupSearchFunctionality() {
    const searchInput = document.getElementById('sidebar-search');
    const clearSearchBtn = document.getElementById('clear-search');
    const navItems = document.querySelectorAll('.nav-item');
    
    if (!searchInput || !clearSearchBtn) return;
    
    searchInput.addEventListener('input', function() {
        const searchTerm = this.value.toLowerCase().trim();
        
        // Toggle the clear button visibility
        clearSearchBtn.style.display = searchTerm ? 'block' : 'none';
        
        // Filter navigation items
        navItems.forEach(item => {
            // Skip section headers
            if (item.classList.contains('nav-section')) return;
            
            const navLink = item.querySelector('.nav-link');
            if (!navLink) return;
            
            const navText = navLink.querySelector('.nav-text')?.textContent || '';
            const isMatch = navText.toLowerCase().includes(searchTerm);
            
            item.style.display = isMatch || searchTerm === '' ? 'block' : 'none';
            
            // Special handling for submenus
            if (item.classList.contains('dropdown-nav')) {
                const submenu = item.querySelector('.nav-submenu');
                const submenuItems = submenu ? submenu.querySelectorAll('.nav-item') : [];
                
                let hasSubMatch = false;
                
                submenuItems.forEach(subItem => {
                    const subNavText = subItem.querySelector('.nav-text')?.textContent || '';
                    const isSubMatch = subNavText.toLowerCase().includes(searchTerm);
                    
                    subItem.style.display = isSubMatch || searchTerm === '' ? 'block' : 'none';
                    
                    if (isSubMatch) {
                        hasSubMatch = true;
                    }
                });
                
                // Show parent if any child matches
                if (hasSubMatch) {
                    item.style.display = 'block';
                    if (submenu && searchTerm) {
                        submenu.classList.add('show');
                    }
                }
                
                // Reset submenu when search is cleared
                if (searchTerm === '' && submenu) {
                    submenu.classList.remove('show');
                }
            }
        });
    });
    
    clearSearchBtn.addEventListener('click', function() {
        searchInput.value = '';
        searchInput.dispatchEvent(new Event('input'));
        searchInput.focus();
    });
}

/**
 * Listen for and update the connection status indicators
 */
function listenForConnectionStatus() {
    // Placeholder function to be connected to your actual API
    function checkConnectionStatus() {
        fetch('/api/status')
            .then(response => response.json())
            .then(data => {
                updateConnectionIndicators(data.status);
            })
            .catch(error => {
                console.error('Error checking connection status:', error);
                updateConnectionIndicators('disconnected');
            });
    }
    
    // Update all connection status indicators
    function updateConnectionIndicators(status) {
        const indicators = document.querySelectorAll('.status-indicator');
        const headerStatusText = document.getElementById('whatsapp-status-text');
        const sidebarConnectionStatus = document.getElementById('sidebar-connection-status');
        
        indicators.forEach(indicator => {
            indicator.className = 'status-indicator';
            
            if (status === 'connected') {
                indicator.classList.add('online');
            } else if (status === 'connecting') {
                indicator.classList.add('connecting');
            }
        });
        
        if (headerStatusText) {
            if (status === 'connected') {
                headerStatusText.textContent = 'Terhubung';
            } else if (status === 'connecting') {
                headerStatusText.textContent = 'Menghubungkan...';
            } else {
                headerStatusText.textContent = 'Terputus';
            }
        }
        
        if (sidebarConnectionStatus) {
            if (status === 'connected') {
                sidebarConnectionStatus.textContent = 'Terhubung';
            } else if (status === 'connecting') {
                sidebarConnectionStatus.textContent = 'Menghubungkan...';
            } else {
                sidebarConnectionStatus.textContent = 'Terputus';
            }
        }
    }
    
    // Check status initially and then every 30 seconds
    checkConnectionStatus();
    setInterval(checkConnectionStatus, 30000);
}
