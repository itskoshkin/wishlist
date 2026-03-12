// ============================================
// CUSTOM FILE INPUT
// ============================================

// Handle file input change
function handleFileInput(input, previewId) {
    const file = input.files[0]; // Get selected file
    if (!file) return; // No file selected

    const reader = new FileReader(); // Create file reader
    reader.onload = (e) => {
        const preview = document.getElementById(previewId); // Get preview element
        if (preview) {
            preview.src = e.target.result; // Set preview image
            preview.style.display = 'block'; // Show preview
            preview.parentElement.querySelector('.file-input-placeholder').style.display = 'none'; // Hide placeholder
        }
    };
    reader.readAsDataURL(file); // Read file as data URL
}

// ============================================
// CUSTOM SELECT
// ============================================

// Toggle custom select
function toggleCustomSelect(selectId) {
    const select = document.getElementById(selectId); // Get select element
    select.classList.toggle('active'); // Toggle active class

    // Close other selects
    document.querySelectorAll('.custom-select').forEach(s => {
        if (s.id !== selectId) {
            s.classList.remove('active'); // Close other selects
        }
    });
}

// Select custom option
function selectCustomOption(selectId, value, label) {
    const select = document.getElementById(selectId); // Get select element
    const trigger = select.querySelector('.custom-select-trigger span'); // Get trigger text
    trigger.textContent = label; // Set trigger text
    select.setAttribute('data-value', value); // Store selected value

    // Update selected option
    select.querySelectorAll('.custom-select-option').forEach(opt => {
        opt.classList.remove('selected'); // Remove selected class
    });
    event.target.classList.add('selected'); // Add selected class to clicked option

    select.classList.remove('active'); // Close select
}

// Close custom selects on outside click
document.addEventListener('click', (e) => {
    if (!e.target.closest('.custom-select')) {
        document.querySelectorAll('.custom-select').forEach(s => {
            s.classList.remove('active'); // Close all selects
        });
    }
});

// ============================================
// MODAL CHANGE TRACKING
// ============================================

let modalHasChanges = false; // Track if modal has changes

// Track form changes
function trackModalChanges(modalId) {
    const modal = document.getElementById(modalId); // Get modal element
    if (!modal) return; // Modal not found

    const form = modal.querySelector('form'); // Get form element
    if (!form) return; // Form not found

    const inputs = form.querySelectorAll('input, textarea, select'); // Get all inputs
    const closeButton = modal.querySelector('.modal-close'); // Get close button

    // Store original values
    const originalValues = {};
    inputs.forEach(input => {
        originalValues[input.name || input.id] = input.value; // Store original value
    });

    // Listen for input changes
    inputs.forEach(input => {
        input.addEventListener('input', () => {
            // Check if any value changed
            let hasChanges = false;
            inputs.forEach(inp => {
                if (inp.value !== originalValues[inp.name || inp.id]) {
                    hasChanges = true; // Value changed
                }
            });

            modalHasChanges = hasChanges; // Update global flag

            // Update close button style
            if (hasChanges) {
                closeButton.classList.add('has-changes'); // Add red style
            } else {
                closeButton.classList.remove('has-changes'); // Remove red style
            }
        });
    });
}

// Override closeModal to check for changes
const originalCloseModal = window.closeModal;
window.closeModal = async function(modalId) {
    if (modalHasChanges) {
        const confirmed = await showConfirm(t('form.unsaved.message'), t('form.unsaved.title'));
        if (!confirmed) return; // User cancelled
    }

    modalHasChanges = false; // Reset flag
    const modal = document.getElementById(modalId);
    if (modal) {
        const closeButton = modal.querySelector('.modal-close');
        if (closeButton) {
            closeButton.classList.remove('has-changes'); // Remove red style
        }
    }

    originalCloseModal(modalId); // Call original function
};
