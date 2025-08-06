/**
 * Dialog System - Reusable dialog component
 * Provides a consistent way to show dialogs across the application
 */
class Dialog {
  constructor() {
    // Container reference
    this.container = document.getElementById('dialog-container');
    
    // Buat container jika belum ada
    if (!this.container) {
      this.container = document.createElement('div');
      this.container.id = 'dialog-container';
      document.body.appendChild(this.container);
    }
    
    // Dialog aktif saat ini
    this.activeDialog = null;
    
    // Callback references - penting untuk mencegah multiple execution
    this.onConfirmCallback = null;
    this.onCancelCallback = null;
    this.onCloseCallback = null;
  }
  
  /**
   * Show a dialog
   * @param {Object} options Dialog options
   */
  show(options) {
    // Bersihkan dialog yang ada sebelum menampilkan yang baru
    this.cleanup();
    
    const defaults = {
      title: 'Konfirmasi',
      message: 'Apakah Anda yakin?',
      description: '',
      confirmText: 'Konfirmasi',
      cancelText: 'Batal',
      type: 'primary', // primary, danger, warning
      onConfirm: null,
      onCancel: null,
      onClose: null
    };
    
    const settings = { ...defaults, ...options };
    
    // Simpan callback untuk digunakan nanti - penting untuk menghindari kebocoran memori
    this.onConfirmCallback = settings.onConfirm;
    this.onCancelCallback = settings.onCancel;
    this.onCloseCallback = settings.onClose;
    
    // Buat elemen dialog
    const dialog = document.createElement('div');
    dialog.className = `dialog-overlay dialog-${settings.type}`;
    
    // Tambahkan konten HTML - mempertahankan struktur original
    dialog.innerHTML = `
      <div class="dialog">
        <div class="dialog-header">
          <h3 class="dialog-title">${settings.title}</h3>
          <button class="dialog-close" id="dialog-close" aria-label="Close dialog">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="dialog-body">
          <p class="dialog-message">${settings.message}</p>
          ${settings.description ? `<p class="dialog-description">${settings.description}</p>` : ''}
        </div>
        <div class="dialog-footer">
          <button class="btn btn-${settings.type}" id="dialog-confirm">${settings.confirmText}</button>
          <button class="btn btn-secondary" id="dialog-cancel">${settings.cancelText}</button>
        </div>
      </div>
    `;
    
    // Tambahkan dialog ke container
    this.container.appendChild(dialog);
    this.activeDialog = dialog;
    
    // Tambahkan event listeners
    this.attachEventListeners();
    
    // Tambahkan kelas untuk animasi
    setTimeout(() => {
      dialog.classList.add('visible');
    }, 10);
    
    return dialog;
  }
  
  /**
   * Attach event listeners untuk dialog yang aktif
   */
  attachEventListeners() {
    if (!this.activeDialog) return;
    
    // Temukan tombol dialog
    const confirmBtn = this.activeDialog.querySelector('#dialog-confirm');
    const cancelBtn = this.activeDialog.querySelector('#dialog-cancel');
    const closeBtn = this.activeDialog.querySelector('#dialog-close');
    
    // Tambahkan event listener dengan binding untuk mencegah duplikat callbacks
    if (confirmBtn) {
      confirmBtn.addEventListener('click', this.handleConfirm.bind(this));
    }
    
    if (cancelBtn) {
      cancelBtn.addEventListener('click', this.handleCancel.bind(this));
    }
    
    if (closeBtn) {
      closeBtn.addEventListener('click', this.handleClose.bind(this));
    }
    
    // Tambahkan click di luar dialog untuk menutupnya
    this.activeDialog.addEventListener('click', (e) => {
      if (e.target === this.activeDialog) {
        this.handleClose();
      }
    });
  }
  
  /**
   * Handle confirm button click
   */
  handleConfirm() {
    // Simpan referensi callback karena this.onConfirmCallback akan di-reset oleh this.close()
    const callback = this.onConfirmCallback;
    
    // Tutup dialog
    this.close();
    
    // Eksekusi callback jika ada
    if (typeof callback === 'function') {
      callback();
    }
  }
  
  /**
   * Handle cancel button click
   */
  handleCancel() {
    // Simpan referensi callback karena this.onCancelCallback akan di-reset oleh this.close()
    const callback = this.onCancelCallback;
    
    // Tutup dialog
    this.close();
    
    // Eksekusi callback jika ada
    if (typeof callback === 'function') {
      callback();
    }
  }
  
  /**
   * Handle close button click
   */
  handleClose() {
    // Simpan referensi callback karena this.onCloseCallback akan di-reset oleh this.close()
    const callback = this.onCloseCallback || this.onCancelCallback;
    
    // Tutup dialog
    this.close();
    
    // Eksekusi callback jika ada
    if (typeof callback === 'function') {
      callback();
    }
  }
  
  /**
   * Close the dialog dengan animasi
   */
  close() {
    if (!this.activeDialog) return;
    
    // Tambahkan kelas untuk animasi keluar
    this.activeDialog.classList.remove('visible');
    
    // Hapus dialog setelah animasi selesai
    setTimeout(() => {
      this.cleanup();
    }, 300);
  }
  
  /**
   * Bersihkan dialog dan hapus event listener
   * Penting untuk mencegah memory leaks dan event handler duplikat
   */
  cleanup() {
    // Reset callback references terlebih dahulu untuk mencegah pemanggilan yang tidak diinginkan
    this.onConfirmCallback = null;
    this.onCancelCallback = null;
    this.onCloseCallback = null;
    
    // Hapus dialog dari DOM jika ada
    if (this.activeDialog && this.container) {
      this.container.removeChild(this.activeDialog);
      this.activeDialog = null;
    }
    
    // Bersihkan container untuk keamanan
    if (this.container) {
      this.container.innerHTML = '';
    }
  }
}

// Export as global if needed
window.Dialog = Dialog;
