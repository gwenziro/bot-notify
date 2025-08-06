/**
 * Landing Page JavaScript
 * Handles animations and interactions for the landing page
 */

document.addEventListener('DOMContentLoaded', function() {
    // Animate code typing and border glow
    animateCodeTyping();
    
    // Floating notification pulsing effect
    const notification = document.querySelector('.floating-notification');
    if (notification) {
        setInterval(() => {
            notification.classList.toggle('pulse');
        }, 2000);
    }
});

/**
 * Animate code typing with a typewriter effect and add glowing border after completion
 */
function animateCodeTyping() {
    const codeElement = document.querySelector('.code-content code');
    const codePreview = document.querySelector('.code-preview');
    
    if (!codeElement || !codePreview) return;
    
    const originalText = codeElement.innerHTML;
    let currentText = '';
    let currentIndex = 0;
    
    // Clear initial text
    codeElement.innerHTML = '';
    
    // Typing animation
    const typingInterval = setInterval(() => {
        if (currentIndex < originalText.length) {
            currentText += originalText[currentIndex];
            codeElement.innerHTML = currentText;
            currentIndex++;
            
            // Auto-scroll to show the latest typed text
            codeElement.scrollTop = codeElement.scrollHeight;
        } else {
            clearInterval(typingInterval);
            
            // Add glowing border after typing animation completes
            setTimeout(() => {
                codePreview.classList.add('animated');
                addGlowingBorderEffect(codePreview);
            }, 500);
        }
    }, 30); // Adjust typing speed here
}

/**
 * Add a pulsing glow effect to the code preview border
 * @param {HTMLElement} element - The element to add the effect to
 */
function addGlowingBorderEffect(element) {
    if (!element) return;
    
    // Create keyframe animation dynamically
    const styleSheet = document.createElement('style');
    styleSheet.innerHTML = `
        @keyframes borderGlow {
            0% { 
                box-shadow: 0 0 5px rgba(10, 147, 159, 0.5);
                border-color: rgba(10, 147, 159, 0.7);
            }
            50% { 
                box-shadow: 0 0 20px rgba(10, 147, 159, 0.7);
                border-color: rgba(10, 147, 159, 1);
            }
            100% { 
                box-shadow: 0 0 5px rgba(10, 147, 159, 0.5);
                border-color: rgba(10, 147, 159, 0.7);
            }
        }
        
        .code-preview.animated {
            animation: borderGlow 3s infinite ease-in-out;
        }
    `;
    document.head.appendChild(styleSheet);
}
