/**
 * Sidebar Component
 * Handles sidebar functionality
 */

document.addEventListener('DOMContentLoaded', function() {
    initSidebar();
});

function initSidebar() {
    const sidebarToggle = document.getElementById('sidebar-toggle');
    const body = document.body;
    const sidebarCollapseBtn = document.getElementById('sidebar-collapse');
    
    if (sidebarToggle) {
        sidebarToggle.addEventListener('click', function() {
            body.classList.toggle('sidebar-open');
        });
    }
    
    if (sidebarCollapseBtn) {
        sidebarCollapseBtn.addEventListener('click', function() {
            body.classList.toggle('sidebar-collapsed');
            
            // Update icon
            const icon = sidebarCollapseBtn.querySelector('i');
            if (body.classList.contains('sidebar-collapsed')) {
                icon.classList.remove('fa-chevron-left');
                icon.classList.add('fa-chevron-right');
                sidebarCollapseBtn.setAttribute('title', 'Perluas');
            } else {
                icon.classList.remove('fa-chevron-right');
                icon.classList.add('fa-chevron-left');
                sidebarCollapseBtn.setAttribute('title', 'Ciutkan');
            }
        });
    }
    
    // Set active menu item based on current URL
    setActiveMenuItem();
    
    // Handle sidebar menu clicks on mobile (auto close)
    const navLinks = document.querySelectorAll('.sidebar-nav .nav-link');
    if (window.innerWidth < 768) {
        navLinks.forEach(link => {
            link.addEventListener('click', function() {
                body.classList.remove('sidebar-open');
            });
        });
    }
}

// Set active menu item based on current URL path
function setActiveMenuItem() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-link');
    
    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        }
    });
}
