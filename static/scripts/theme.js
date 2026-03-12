// ============================================
// THEME MANAGEMENT
// ============================================

// Available themes
const THEMES = {
    LIGHT: 'light',
    DARK: 'dark',
    SYSTEM: 'system'
};

// Current theme (default: system)
let currentTheme = localStorage.getItem('wishlist_theme') || THEMES.SYSTEM;

// Get system theme preference
function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? THEMES.DARK : THEMES.LIGHT;
}

// Apply theme to body
function applyTheme(theme) {
    const body = document.body; // Get body element

    // Remove all theme classes
    body.classList.remove('theme-light', 'theme-dark');

    // Determine actual theme to apply
    const actualTheme = theme === THEMES.SYSTEM ? getSystemTheme() : theme;

    // Add theme class
    body.classList.add(`theme-${actualTheme}`);
}

// Set theme
function setTheme(theme) {
    if (!Object.values(THEMES).includes(theme)) return; // Invalid theme

    currentTheme = theme; // Update current theme
    localStorage.setItem('wishlist_theme', theme); // Save to localStorage
    applyTheme(theme); // Apply theme
}

// Cycle through themes (light -> system -> dark -> light)
function cycleTheme() {
    let nextTheme;

    // Determine next theme in cycle
    if (currentTheme === THEMES.LIGHT) {
        nextTheme = THEMES.SYSTEM;
    } else if (currentTheme === THEMES.SYSTEM) {
        nextTheme = THEMES.DARK;
    } else {
        nextTheme = THEMES.LIGHT;
    }

    setTheme(nextTheme); // Apply next theme
}

// Listen for system theme changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (currentTheme === THEMES.SYSTEM) {
        applyTheme(THEMES.SYSTEM); // Re-apply system theme
    }
});

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', () => {
    applyTheme(currentTheme); // Apply saved theme
});
