/**
 * Documentation Page JavaScript
 * Handles documentation page specific functionality
 */

document.addEventListener('DOMContentLoaded', function() {
    // Initialize documentation page
    initDocumentation();
});

/**
 * Initialize documentation page functionality
 */
function initDocumentation() {
    // Setup scroll spy
    setupScrollSpy();
    
    // Setup copy buttons
    setupCopyButtons();
    
    // Setup smooth scrolling for navigation links
    setupSmoothScrolling();
    
    // Make code examples look nice
    formatCodeExamples();
}

/**
 * Set up scroll spy to highlight active navigation items
 */
function setupScrollSpy() {
    const sections = document.querySelectorAll('.docs-section, .endpoint-section');
    const navLinks = document.querySelectorAll('.docs-nav-link');
    
    // Highlight the first link by default
    if (navLinks.length > 0) {
        navLinks[0].classList.add('active');
    }
    
    // Add scroll event listener
    window.addEventListener('scroll', function() {
        let current = '';
        
        // Find the current section in view
        sections.forEach(section => {
            const sectionTop = section.offsetTop - 100;
            const sectionHeight = section.offsetHeight;
            
            if (window.pageYOffset >= sectionTop && window.pageYOffset < sectionTop + sectionHeight) {
                current = section.getAttribute('id');
            }
        });
        
        // Remove active class from all links
        navLinks.forEach(link => {
            link.classList.remove('active');
        });
        
        // Add active class to current link
        if (current) {
            const activeLink = document.querySelector(`.docs-nav-link[href="#${current}"]`);
            if (activeLink) {
                activeLink.classList.add('active');
            }
        }
    });
}

/**
 * Set up copy buttons for code examples
 */
function setupCopyButtons() {
    // Create clipboard.js instance for each copy button
    document.querySelectorAll('.copy-btn').forEach(button => {
        button.addEventListener('click', function() {
            const targetId = this.getAttribute('data-clipboard-target');
            const text = document.querySelector(targetId).textContent;
            
            copyToClipboard(text)
                .then(() => {
                    // Show success feedback
                    const originalIcon = this.innerHTML;
                    this.innerHTML = '<i class="fas fa-check"></i>';
                    setTimeout(() => {
                        this.innerHTML = originalIcon;
                    }, 2000);
                })
                .catch(err => {
                    console.error('Failed to copy', err);
                });
        });
    });
}

/**
 * Copy text to clipboard
 * @param {string} text - Text to copy
 * @returns {Promise} - Promise that resolves when text is copied
 */
function copyToClipboard(text) {
    // Use the modern Clipboard API if available
    if (navigator.clipboard && navigator.clipboard.writeText) {
        return navigator.clipboard.writeText(text);
    }
    
    // Fallback to the older approach
    return new Promise((resolve, reject) => {
        try {
            const textArea = document.createElement('textarea');
            textArea.value = text;
            textArea.style.position = 'fixed';
            textArea.style.left = '-999999px';
            textArea.style.top = '-999999px';
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();
            
            const successful = document.execCommand('copy');
            document.body.removeChild(textArea);
            
            if (successful) {
                resolve();
            } else {
                reject(new Error('Failed to copy'));
            }
        } catch (err) {
            reject(err);
        }
    });
}

/**
 * Set up smooth scrolling for navigation links
 */
function setupSmoothScrolling() {
    document.querySelectorAll('.docs-nav-link').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href');
            const targetElement = document.querySelector(targetId);
            
            if (targetElement) {
                window.scrollTo({
                    top: targetElement.offsetTop - 80,
                    behavior: 'smooth'
                });
                
                // Update URL without reloading
                history.pushState(null, null, targetId);
                
                // Set active class
                document.querySelectorAll('.docs-nav-link').forEach(navLink => {
                    navLink.classList.remove('active');
                });
                this.classList.add('active');
            }
        });
    });
    
    // Handle initial hash in URL
    if (window.location.hash) {
        const targetElement = document.querySelector(window.location.hash);
        if (targetElement) {
            // Small delay to ensure page is fully loaded
            setTimeout(() => {
                window.scrollTo({
                    top: targetElement.offsetTop - 80,
                    behavior: 'smooth'
                });
                
                // Set active class
                const activeLink = document.querySelector(`.docs-nav-link[href="${window.location.hash}"]`);
                if (activeLink) {
                    document.querySelectorAll('.docs-nav-link').forEach(navLink => {
                        navLink.classList.remove('active');
                    });
                    activeLink.classList.add('active');
                }
            }, 100);
        }
    }
}

/**
 * Format code examples with syntax highlighting
 * This is a simple implementation - for production,
 * consider using libraries like Prism.js or highlight.js
 */
function formatCodeExamples() {
    // Simplistic syntax highlighting for curl commands
    document.querySelectorAll('.code-block').forEach(block => {
        if (block.textContent.includes('curl')) {
            // Highlight parts of curl commands
            let html = block.innerHTML;
            
            // Highlight URLs
            html = html.replace(/"(https?:\/\/[^"]+)"/g, '"<span style="color: #50fa7b;">$1</span>"');
            
            // Highlight headers
            html = html.replace(/-H\s+"([^"]+)"/g, '-H "<span style="color: #ff79c6;">$1</span>"');
            
            // Highlight request body properties
            html = html.replace(/"([a-zA-Z_]+)":/g, '"<span style="color: #ff8c00;">$1</span>":');
            
            // Update the HTML
            block.innerHTML = html;
        } else if (block.textContent.includes('{')) {
            // Basic JSON highlighting
            let html = block.innerHTML;
            
            // Highlight property names
            html = html.replace(/"([a-zA-Z_]+)":/g, '"<span style="color: #ff8c00;">$1</span>":');
            
            // Highlight string values
            html = html.replace(/: "(.*?)"/g, ': "<span style="color: #50fa7b;">$1</span>"');
            
            // Highlight boolean and number values
            html = html.replace(/: (true|false|null)/g, ': <span style="color: #bd93f9;">$1</span>');
            html = html.replace(/: (\d+)/g, ': <span style="color: #bd93f9;">$1</span>');
            
            // Update the HTML
            block.innerHTML = html;
        }
    });
}
