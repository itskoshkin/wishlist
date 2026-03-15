// ============================================
// CUSTOM ALERT/CONFIRM DIALOGS
// ============================================

const TOAST_DURATION_MS = 4000;
const TOAST_EXIT_ANIMATION_MS = 300;
const FLASH_TOAST_STORAGE_KEY = 'wishlist_flash_toast';
const FLASH_TOAST_MAX_AGE_MS = 15000;

function i18nOr(key, fallback) {
    if (typeof t === 'function') {
        const value = t(key);
        if (value && value !== key) return value;
    }
    return fallback;
}

// Show custom alert
function showAlert(message, title = i18nOr('dialog.alertTitle', 'Notification')) {
    return new Promise((resolve) => {
        // Create overlay
        const overlay = document.createElement('div'); // Create overlay element
        overlay.className = 'custom-alert-overlay'; // Add overlay class

        // Create alert dialog
        const alert = document.createElement('div'); // Create alert element
        alert.className = 'custom-alert'; // Add alert class

        // Build alert HTML
        alert.innerHTML = `
            <div class="custom-alert-title">${title}</div>
            <div class="custom-alert-message">${message}</div>
            <div class="custom-alert-buttons">
                <button class="custom-alert-button primary">${i18nOr('common.ok', 'OK')}</button>
            </div>
        `;

        // Append to overlay
        overlay.appendChild(alert); // Add alert to overlay
        document.body.appendChild(overlay); // Add overlay to body

        // Show overlay
        setTimeout(() => overlay.classList.add('active'), 10); // Trigger animation

        // Handle button click
        const okButton = alert.querySelector('.custom-alert-button'); // Get OK button
        okButton.addEventListener('click', () => {
            overlay.classList.remove('active'); // Hide overlay
            setTimeout(() => {
                document.body.removeChild(overlay); // Remove from DOM
                resolve(); // Resolve promise
            }, 300); // Wait for animation
        });
    });
}

// Show custom confirm dialog
function showConfirm(message, title = i18nOr('dialog.confirmTitle', 'Confirmation')) {
    return new Promise((resolve) => {
        // Create overlay
        const overlay = document.createElement('div'); // Create overlay element
        overlay.className = 'custom-alert-overlay'; // Add overlay class

        // Create confirm dialog
        const confirm = document.createElement('div'); // Create confirm element
        confirm.className = 'custom-alert'; // Add alert class

        // Build confirm HTML
        confirm.innerHTML = `
            <div class="custom-alert-title">${title}</div>
            <div class="custom-alert-message">${message}</div>
            <div class="custom-alert-buttons">
                <button class="custom-alert-button secondary" data-action="cancel">${i18nOr('common.cancel', 'Cancel')}</button>
                <button class="custom-alert-button danger" data-action="confirm">${i18nOr('dialog.confirm', 'Confirm')}</button>
            </div>
        `;

        // Append to overlay
        overlay.appendChild(confirm); // Add confirm to overlay
        document.body.appendChild(overlay); // Add overlay to body

        // Show overlay
        setTimeout(() => overlay.classList.add('active'), 10); // Trigger animation

        let isClosed = false;
        const closeWithResult = (result) => {
            if (isClosed) return;
            isClosed = true;
            overlay.classList.remove('active');
            setTimeout(() => {
                if (overlay.parentNode) {
                    document.body.removeChild(overlay);
                }
                resolve(result);
            }, 300);
        };

        // Handle button clicks
        const buttons = confirm.querySelectorAll('.custom-alert-button'); // Get all buttons
        buttons.forEach(button => {
            button.addEventListener('click', () => {
                const action = button.getAttribute('data-action'); // Get button action
                closeWithResult(action === 'confirm');
            });
        });

        overlay.addEventListener('click', (event) => {
            if (event.target === overlay) {
                closeWithResult(false);
            }
        });
    });
}

// Show custom prompt dialog
function showPrompt(message, defaultValue = '', title = i18nOr('dialog.promptTitle', 'Input')) {
    return new Promise((resolve) => {
        // Create overlay
        const overlay = document.createElement('div'); // Create overlay element
        overlay.className = 'custom-alert-overlay'; // Add overlay class

        // Create prompt dialog
        const prompt = document.createElement('div'); // Create prompt element
        prompt.className = 'custom-alert'; // Add alert class

        // Build prompt HTML
        prompt.innerHTML = `
            <div class="custom-alert-title">${title}</div>
            <div class="custom-alert-message">${message}</div>
            <input type="text" class="form-input" value="${defaultValue}" style="margin-bottom: 1.5rem;">
            <div class="custom-alert-buttons">
                <button class="custom-alert-button secondary" data-action="cancel">${i18nOr('common.cancel', 'Cancel')}</button>
                <button class="custom-alert-button primary" data-action="confirm">${i18nOr('common.ok', 'OK')}</button>
            </div>
        `;

        // Append to overlay
        overlay.appendChild(prompt); // Add prompt to overlay
        document.body.appendChild(overlay); // Add overlay to body

        // Show overlay
        setTimeout(() => overlay.classList.add('active'), 10); // Trigger animation

        // Focus input
        const input = prompt.querySelector('.form-input'); // Get input element
        setTimeout(() => input.focus(), 100); // Focus after animation

        // Handle button clicks
        const buttons = prompt.querySelectorAll('.custom-alert-button'); // Get all buttons
        buttons.forEach(button => {
            button.addEventListener('click', () => {
                const action = button.getAttribute('data-action'); // Get button action
                overlay.classList.remove('active'); // Hide overlay
                setTimeout(() => {
                    document.body.removeChild(overlay); // Remove from DOM
                    resolve(action === 'confirm' ? input.value : null); // Resolve with value or null
                }, 300); // Wait for animation
            });
        });

        // Handle Enter key
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                overlay.classList.remove('active'); // Hide overlay
                setTimeout(() => {
                    document.body.removeChild(overlay); // Remove from DOM
                    resolve(input.value); // Resolve with value
                }, 300); // Wait for animation
            }
        });
    });
}

// Show toast notification
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;

    document.body.appendChild(toast);

    setTimeout(() => toast.classList.add('active'), 10);

    setTimeout(() => {
        toast.classList.remove('active');
        setTimeout(() => {
            if (toast.parentNode) {
                document.body.removeChild(toast);
            }
        }, TOAST_EXIT_ANIMATION_MS);
    }, TOAST_DURATION_MS);
}

function scheduleFlashToast(message, type = 'success') {
    try {
        sessionStorage.setItem(
            FLASH_TOAST_STORAGE_KEY,
            JSON.stringify({
                message,
                type,
                timestamp: Date.now(),
            }),
        );
    } catch (error) {
        console.warn('Failed to store flash toast:', error);
    }
}

function consumeFlashToast() {
    try {
        const raw = sessionStorage.getItem(FLASH_TOAST_STORAGE_KEY);
        if (!raw) return;

        sessionStorage.removeItem(FLASH_TOAST_STORAGE_KEY);
        const payload = JSON.parse(raw);
        if (!payload || !payload.message) return;

        const timestamp = Number(payload.timestamp || 0);
        if (timestamp && Date.now() - timestamp > FLASH_TOAST_MAX_AGE_MS) {
            return;
        }

        showToast(payload.message, payload.type || 'success');
    } catch (error) {
        console.warn('Failed to consume flash toast:', error);
        sessionStorage.removeItem(FLASH_TOAST_STORAGE_KEY);
    }
}

document.addEventListener('DOMContentLoaded', consumeFlashToast);
