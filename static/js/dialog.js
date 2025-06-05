/**
 * Dialog Component - Reusable, accessible dialog functionality
 * Provides a consistent UX for confirmation dialogs
 */
class Dialog {
  constructor() {
    this.visible = false;
    this.overlay = null;
    this.box = null;
    this.title = null;
    this.message = null;
    this.description = null;
    this.cancelBtn = null;
    this.confirmBtn = null;
    this.closeBtn = null;
    
    // Event callbacks
    this.onConfirm = null;
    this.onCancel = null;
    this.onClose = null;
    
    // Track the previously focused element
    this.previouslyFocused = null;
    
    // Event handler references for cleanup
    this.keydownHandler = this.handleKeydown.bind(this);
    
    // Create container if needed
    this.initialize();
  }
  
  initialize() {
    // Create container if it doesn't exist
    if (!document.getElementById('dialog-container')) {
      const container = document.createElement('div');
      container.id = 'dialog-container';
      document.body.appendChild(container);
    }
  }
  
  show(options) {
    // Merge options with defaults
    const {
      title = 'Konfirmasi',
      message = 'Apakah Anda yakin?',
      description = '',
      confirmText = 'Konfirmasi',
      cancelText = 'Batal',
      type = 'default', // default, danger, warning, info
      onConfirm = () => {},
      onCancel = () => {},
      onClose = () => {}
    } = options;
    
    // Store callbacks
    this.onConfirm = onConfirm;
    this.onCancel = onCancel;
    this.onClose = onClose;
    
    // Create dialog elements if they don't exist yet
    if (!this.overlay) {
      this.createDialogElements();
    }
    
    // Store currently focused element
    this.previouslyFocused = document.activeElement;
    
    // Set dialog content
    this.title.textContent = title;
    this.message.textContent = message;
    this.description.textContent = description;
    this.confirmBtn.textContent = confirmText;
    this.cancelBtn.textContent = cancelText;
    
    // Apply type-specific styling
    this.box.className = `dialog-box dialog-${type}`;
    
    // Add to DOM
    document.getElementById('dialog-container').appendChild(this.overlay);
    
    // Prevent body scrolling
    document.body.style.overflow = 'hidden';
    
    // Show dialog with animation
    requestAnimationFrame(() => {
      this.overlay.classList.add('visible');
      this.box.classList.add('visible');
      
      // Set focus to the confirm button
      setTimeout(() => this.confirmBtn.focus(), 100);
    });
    
    // Set visible state
    this.visible = true;
    
    // Add event listeners
    this.attachEventListeners();
    
    // Return instance for chaining
    return this;
  }
  
  createDialogElements() {
    // Create overlay
    this.overlay = document.createElement('div');
    this.overlay.className = 'dialog-overlay';
    this.overlay.setAttribute('role', 'dialog');
    this.overlay.setAttribute('aria-modal', 'true');
    
    // Create box
    this.box = document.createElement('div');
    this.box.className = 'dialog-box';
    this.box.setAttribute('role', 'document');
    
    // Create header
    const header = document.createElement('div');
    header.className = 'dialog-header';
    
    this.title = document.createElement('h3');
    this.title.className = 'dialog-title';
    this.title.id = 'dialog-title-' + Math.random().toString(36).substr(2, 9);
    
    this.closeBtn = document.createElement('button');
    this.closeBtn.className = 'dialog-close';
    this.closeBtn.setAttribute('aria-label', 'Tutup dialog');
    this.closeBtn.innerHTML = '<i class="fas fa-times"></i>';
    
    header.appendChild(this.title);
    header.appendChild(this.closeBtn);
    
    // Create body
    const body = document.createElement('div');
    body.className = 'dialog-body';
    
    this.message = document.createElement('p');
    this.message.className = 'dialog-message';
    
    this.description = document.createElement('p');
    this.description.className = 'dialog-description';
    
    body.appendChild(this.message);
    body.appendChild(this.description);
    
    // Create footer
    const footer = document.createElement('div');
    footer.className = 'dialog-footer';
    
    this.cancelBtn = document.createElement('button');
    this.cancelBtn.className = 'dialog-btn dialog-cancel';
    this.cancelBtn.type = 'button';
    
    this.confirmBtn = document.createElement('button');
    this.confirmBtn.className = 'dialog-btn dialog-confirm';
    this.confirmBtn.type = 'button';
    
    footer.appendChild(this.cancelBtn);
    footer.appendChild(this.confirmBtn);
    
    // Set ARIA attributes for accessibility
    this.overlay.setAttribute('aria-labelledby', this.title.id);
    
    // Assemble dialog
    this.box.appendChild(header);
    this.box.appendChild(body);
    this.box.appendChild(footer);
    this.overlay.appendChild(this.box);
  }
  
  attachEventListeners() {
    // Close button handler
    this.closeBtn.addEventListener('click', () => this.handleClose());
    
    // Cancel button handler
    this.cancelBtn.addEventListener('click', () => this.handleCancel());
    
    // Confirm button handler
    this.confirmBtn.addEventListener('click', () => this.handleConfirm());
    
    // Close on outside click
    this.overlay.addEventListener('click', (e) => {
      if (e.target === this.overlay) {
        this.handleCancel();
      }
    });
    
    // Global keyboard handling
    document.addEventListener('keydown', this.keydownHandler);
  }
  
  detachEventListeners() {
    document.removeEventListener('keydown', this.keydownHandler);
  }
  
  handleKeydown(e) {
    if (!this.visible) return;
    
    // Close on Escape
    if (e.key === 'Escape') {
      e.preventDefault();
      this.handleCancel();
    }
    
    // Confirm on Enter when not in a form context
    if (e.key === 'Enter' && e.target.tagName !== 'TEXTAREA' && !e.target.closest('form')) {
      e.preventDefault();
      this.handleConfirm();
    }
    
    // Trap focus within dialog
    if (e.key === 'Tab') {
      // Get all focusable elements
      const focusableEls = this.box.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
      const firstFocusableEl = focusableEls[0];
      const lastFocusableEl = focusableEls[focusableEls.length - 1];
      
      // If shift+tab and on first element, move to last element
      if (e.shiftKey && document.activeElement === firstFocusableEl) {
        e.preventDefault();
        lastFocusableEl.focus();
      }
      // If tab and on last element, move to first element
      else if (!e.shiftKey && document.activeElement === lastFocusableEl) {
        e.preventDefault();
        firstFocusableEl.focus();
      }
    }
  }
  
  handleClose() {
    if (typeof this.onClose === 'function') {
      this.onClose();
    }
    this.close();
  }
  
  handleCancel() {
    if (typeof this.onCancel === 'function') {
      this.onCancel();
    }
    this.close();
  }
  
  handleConfirm() {
    if (typeof this.onConfirm === 'function') {
      this.onConfirm();
    }
    this.close();
  }
  
  close() {
    if (!this.visible) return;
    
    // Hide with animation
    this.overlay.classList.remove('visible');
    this.box.classList.remove('visible');
    
    // Set state
    this.visible = false;
    
    // Re-enable body scrolling
    document.body.style.overflow = '';
    
    // Remove from DOM after animation
    setTimeout(() => {
      if (this.overlay.parentNode) {
        this.overlay.parentNode.removeChild(this.overlay);
      }
      
      // Clean up event listeners
      this.detachEventListeners();
      
      // Restore focus to previously focused element
      if (this.previouslyFocused && typeof this.previouslyFocused.focus === 'function') {
        this.previouslyFocused.focus();
      }
    }, 300);
  }
  
  // Public API for programmatically closing the dialog
  forceClose() {
    this.close();
  }
}

// Export as global if needed
window.Dialog = Dialog;
