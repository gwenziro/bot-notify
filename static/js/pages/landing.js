/**
 * Landing Page JavaScript - Versi Sederhana
 * Dengan efek minimal yang lebih elegan
 */

document.addEventListener('DOMContentLoaded', function() {
    // Tambahkan class ke body setelah halaman dimuat untuk efek fade-in
    setTimeout(() => {
        document.body.classList.add('loaded');
    }, 100);
    
    // Animasi typing code
    initTypeCodeAnimation();
    
    // Efek hover yang lebih halus untuk kartu
    initCardHover();
    
    // Animasi fade-in sederhana untuk elemen tanpa kompleksitas berlebihan
    initSimpleFadeIn();
});

/**
 * Inisialisasi animasi pengetikan kode
 */
function initTypeCodeAnimation() {
    const codeContent = document.querySelector('.code-content code');
    if (!codeContent) return;
    
    // Simpan konten asli
    const originalContent = codeContent.innerHTML;
    
    // Kosongkan konten untuk efek typing
    codeContent.innerHTML = '';
    
    // Parse konten HTML untuk diketik
    const htmlContent = parseHTMLForTyping(originalContent);
    
    // Mulai animasi setelah sedikit delay
    setTimeout(() => {
        typeCode(codeContent, htmlContent, 0, 30);
    }, 800);
}

/**
 * Parse HTML menjadi array untuk typing animation
 * @param {string} html - HTML string untuk diparse
 * @returns {Array} Array elemen HTML dan text untuk typing
 */
function parseHTMLForTyping(html) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(`<div>${html}</div>`, 'text/html');
    const parentNode = doc.body.firstChild;
    
    const result = [];
    
    function traverseNode(node) {
        if (node.nodeType === Node.TEXT_NODE) {
            // Untuk text node, tambahkan sebagai karakter individual
            const text = node.textContent;
            for (let i = 0; i < text.length; i++) {
                result.push({
                    type: 'char',
                    content: text[i],
                    html: false
                });
            }
        } else if (node.nodeType === Node.ELEMENT_NODE) {
            // Untuk element node, simpan tag pembuka
            const tagName = node.tagName.toLowerCase();
            const attributes = Array.from(node.attributes)
                .map(attr => `${attr.name}="${attr.value}"`)
                .join(' ');
                
            const openTag = attributes 
                ? `<${tagName} ${attributes}>`
                : `<${tagName}>`;
                
            result.push({
                type: 'tag',
                content: openTag,
                html: true
            });
            
            // Traverse anak-anak node
            Array.from(node.childNodes).forEach(traverseNode);
            
            // Tambahkan tag penutup
            result.push({
                type: 'tag',
                content: `</${tagName}>`,
                html: true
            });
        }
    }
    
    Array.from(parentNode.childNodes).forEach(traverseNode);
    return result;
}

/**
 * Mengetik kode secara bertahap
 * @param {HTMLElement} element - Element untuk mengetik konten
 * @param {Array} content - Konten yang akan diketik
 * @param {number} index - Indeks karakter saat ini
 * @param {number} speed - Kecepatan typing (ms)
 */
function typeCode(element, content, index, speed) {
    // Tambahkan cursor class ke element
    if (index === 0) {
        element.classList.add('typing-cursor');
    }
    
    // Jika sudah mencapai akhir, selesaikan animasi
    if (index >= content.length) {
        element.classList.remove('typing-cursor');
        initCodeFloatAnimation(); // Mulai animasi floating setelah typing selesai
        return;
    }
    
    // Ketik karakter saat ini
    const item = content[index];
    
    if (item.html) {
        element.innerHTML += item.content;
    } else {
        element.innerHTML += item.content;
    }
    
    // Hitung kecepatan ketik berikutnya (lebih cepat untuk tag HTML)
    let nextSpeed = speed;
    if (item.type === 'tag') {
        nextSpeed = 0; // Render tag HTML langsung
    } else if (item.content === ' ' || item.content === '\n') {
        nextSpeed = speed / 2; // Lebih cepat untuk spasi dan newline
    }
    
    // Schedule next character
    setTimeout(() => {
        typeCode(element, content, index + 1, speed);
    }, nextSpeed);
}

/**
 * Inisialisasi animasi mengambang untuk container kode
 */
function initCodeFloatAnimation() {
    const codePreview = document.querySelector('.code-preview');
    if (!codePreview) return;
    
    // Tambahkan class untuk floating animation
    setTimeout(() => {
        codePreview.classList.add('code-float-animation');
    }, 300);
}

/**
 * Inisialisasi efek hover yang halus pada kartu
 */
function initCardHover() {
    const visual = document.querySelector('.hero-visual');
    
    if (!visual) return;
    // Hover effect pada visual
    visual.addEventListener('mouseenter', () => {
        visual.style.transform = 'translateY(-5px)';
    });
    
    visual.addEventListener('mouseleave', () => {
        visual.style.transform = 'translateY(0)';
    });
}

/**
 * Inisialisasi animasi fade-in sederhana tanpa kompleksitas
 */
function initSimpleFadeIn() {
    const fadeElements = [
        '.hero-description',
        '.hero-title',
        '.hero-actions',
        '.tech-badges'
    ];
    
    // Tambahkan CSS untuk animasi sederhana
    const style = document.createElement('style');
    style.textContent = `
        .fade-in {
            opacity: 0;
            transform: translateY(10px);
            animation: fadeIn 0.8s forwards;
        }

        @keyframes fadeIn {
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .typing-cursor::after {
            content: '|';
            color: var(--accent-light);
            animation: blinkCursor 0.8s infinite;
        }

        @keyframes blinkCursor {
            from, to { opacity: 1; }
            50% { opacity: 0; }
        }

        .code-float-animation {
            animation: codeFloat 4s ease-in-out infinite alternate;
            transform-origin: center center;
        }

        @keyframes codeFloat {
            0% {
                transform: translateY(0) rotate(0deg);
                box-shadow: 0 15px 25px -15px rgba(0, 0, 0, 0.3), 0 0 15px var(--accent-glow);
            }
            100% {
                transform: translateY(-10px) rotate(0.5deg);
                box-shadow: 0 20px 30px -15px rgba(0, 0, 0, 0.4), 0 0 20px var(--accent-glow);
            }
        }

        .delay-1 { animation-delay: 0.1s; }
        .delay-2 { animation-delay: 0.2s; }
        .delay-3 { animation-delay: 0.3s; }
        .delay-4 { animation-delay: 0.4s; }
        .delay-5 { animation-delay: 0.5s; }
    `;
    document.head.appendChild(style);
    
    // Terapkan kelas animasi ke elemen
    fadeElements.forEach((selector, index) => {
        const element = document.querySelector(selector);
        if (element) {
            element.classList.add('fade-in', `delay-${index + 1}`);
        }
    });
}
