// ============================================
// MODAL MANAGEMENT
// ============================================

// Open modal by ID
function openModal(modalId) {
    const modal = document.getElementById(modalId); // Get modal element
    if (!modal) return; // Modal not found

    modal.classList.add('active'); // Show modal
    document.body.style.overflow = 'hidden'; // Prevent body scroll
}

// Close modal by ID
function closeModal(modalId) {
    const modal = document.getElementById(modalId); // Get modal element
    if (!modal) return; // Modal not found

    modal.classList.remove('active'); // Hide modal
    document.body.style.overflow = ''; // Restore body scroll
}

// Close modal when clicking outside
function setupModalCloseOnOutsideClick() {
    document.querySelectorAll('.modal-overlay').forEach(overlay => {
        overlay.addEventListener('click', (e) => {
            // Check if click was on overlay (not on modal content)
            if (e.target === overlay) {
                closeModal(overlay.id); // Close modal
            }
        });
    });
}

// Initialize modal handlers on page load
document.addEventListener('DOMContentLoaded', () => {
    setupModalCloseOnOutsideClick(); // Setup outside click handlers
});
