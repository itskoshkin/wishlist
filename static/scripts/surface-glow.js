// Unified pointer glow engine for all glass surfaces:
// header, top FABs, controls/buttons on lists page, feature cards, list cards, wish cards.
// A single RAF loop avoids duplicate per-page handlers (especially expensive in Safari).
window.__surfaceGlowUnified = true;

document.addEventListener('DOMContentLoaded', () => {
    if (window.matchMedia('(pointer: coarse)').matches) return;

    const isSafari = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
    const minFrameInterval = isSafari ? 24 : 0;

    const staticProfiles = [
        { selector: '.header', maxDistance: 320, borderDistance: 56 },
        { selector: '.fab-left, .fab-right', maxDistance: 190, borderDistance: 44 },
        { selector: '.fab .popup-menu.active, .popup-menu.glow-popup-menu.active', maxDistance: 240, borderDistance: 48 },
        { selector: '#headerButton, .empty-state .btn', maxDistance: 240, borderDistance: 48 },
        { selector: '.page-wishlists .controls, .page-list-detail .controls', maxDistance: 300, borderDistance: 56 },
        {
            selector: '.page-wishlists .header-right .profile-btn, .page-wishlists .header-right .logout-btn, .page-list-detail .header-right .profile-btn, .page-list-detail .header-right .logout-btn, .page-wishlists #sortMenuButton, .page-wishlists #filterMenuButton, .page-list-detail #sortMenuButton, .page-list-detail #filterMenuButton',
            maxDistance: 190,
            borderDistance: 42,
        },
        {
            selector: '.page-list-detail .list-actions .btn-secondary',
            maxDistance: 220,
            borderDistance: 44,
        },
        {
            selector: '.page-wishlists .fab-create, .page-list-detail .fab-create',
            maxDistance: 210,
            borderDistance: 44,
        },
        {
            selector: '#profileModal .profile-inline-actions .icon-button, #profileModal .profile-avatar-edit-indicator, #profileModal .profile-avatar-delete-btn',
            maxDistance: 170,
            borderDistance: 36,
        },
        {
            selector: '#profileModal .profile-password-btn',
            maxDistance: 260,
            borderDistance: 52,
        },
        {
            selector: '#profileModal .modal-close',
            maxDistance: 180,
            borderDistance: 40,
        },
        {
            selector: '#changePasswordModal .profile-password-btn',
            maxDistance: 260,
            borderDistance: 52,
        },
        {
            selector: '#createListModal .form-submit, #addWishModal .form-submit, #findUserModal .form-submit, #forgotPasswordModal .form-submit, #resetPasswordPageModal .form-submit',
            maxDistance: 260,
            borderDistance: 52,
        },
        {
            selector: '#createListModal .modal-close, #addWishModal .modal-close, #findUserModal .modal-close, #forgotPasswordModal .modal-close, #resetPasswordPageModal .modal-close',
            maxDistance: 180,
            borderDistance: 40,
        },
        {
            selector: '#authModal .form-submit, #authModal .segmented-option, #authModal .modal-close',
            maxDistance: 260,
            borderDistance: 52,
        },
        {
            selector: '#wishDetailModal .modal-close, #wishDetailModal .icon-button, #wishDetailModal .btn-danger, #wishDetailModal .wish-detail-image-edit-btn, #wishDetailModal .wish-detail-image-delete-btn',
            maxDistance: 230,
            borderDistance: 46,
        },
        {
            selector: '.custom-alert .custom-alert-button',
            maxDistance: 230,
            borderDistance: 44,
        },
    ];

    const cardProfiles = [
        { selector: '.feature-card', maxDistance: 240, borderDistance: 44 },
        { selector: '.list-card', maxDistance: 240, borderDistance: 48 },
        { selector: '.wish-card', maxDistance: 240, borderDistance: 48 },
    ];

    let staticTargets = [];

    function collectStaticTargets() {
        const targets = [];
        staticProfiles.forEach((profile) => {
            document.querySelectorAll(profile.selector).forEach((element) => {
                targets.push({
                    element,
                    maxDistance: profile.maxDistance,
                    borderDistance: profile.borderDistance,
                    allowWhenModal: profile.selector.includes('Modal') || profile.selector.includes('.custom-alert'),
                    rect: null,
                });
            });
        });
        return targets;
    }

    staticTargets = collectStaticTargets();

    const cardSelectorUnion = cardProfiles.map((profile) => profile.selector).join(', ');
    const hasCardTargets = Boolean(document.querySelector(cardSelectorUnion));
    if (!staticTargets.length && !hasCardTargets) return;

    const stateCache = new WeakMap();
    let activeCard = null;
    let mouseX = -9999;
    let mouseY = -9999;
    let pointerInWindow = false;
    let rectsDirty = true;
    let staticTargetsDirty = false;
    let rafId = 0;
    let lastFrameAt = 0;

    function distanceToRect(x, y, rect) {
        const dx = Math.max(rect.left - x, 0, x - rect.right);
        const dy = Math.max(rect.top - y, 0, y - rect.bottom);
        return Math.sqrt(dx * dx + dy * dy);
    }

    function distanceToRectSurface(x, y, rect) {
        if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
            const distToLeft = x - rect.left;
            const distToRight = rect.right - x;
            const distToTop = y - rect.top;
            const distToBottom = rect.bottom - y;
            return Math.min(distToLeft, distToRight, distToTop, distToBottom);
        }

        const dx = Math.max(rect.left - x, 0, x - rect.right);
        const dy = Math.max(rect.top - y, 0, y - rect.bottom);
        return Math.sqrt(dx * dx + dy * dy);
    }

    function getCardProfile(element) {
        if (!element) return null;
        return cardProfiles.find((profile) => element.matches(profile.selector)) || null;
    }

    function writeGlowState(element, x, y, glowOpacity, borderGlow) {
        const nextState = {
            x: Number(x.toFixed(2)),
            y: Number(y.toFixed(2)),
            glow: Number(glowOpacity.toFixed(3)),
            border: Number(borderGlow.toFixed(3)),
        };
        const prevState = stateCache.get(element);

        if (
            prevState &&
            Math.abs(prevState.x - nextState.x) < 0.5 &&
            Math.abs(prevState.y - nextState.y) < 0.5 &&
            Math.abs(prevState.glow - nextState.glow) < 0.01 &&
            Math.abs(prevState.border - nextState.border) < 0.01
        ) {
            return;
        }

        element.style.setProperty('--mouse-x', `${nextState.x}px`);
        element.style.setProperty('--mouse-y', `${nextState.y}px`);
        element.style.setProperty('--glow-opacity', String(nextState.glow));
        element.style.setProperty('--border-glow-intensity', String(nextState.border));
        stateCache.set(element, nextState);
    }

    function resetElementGlow(element) {
        const prevState = stateCache.get(element);
        if (prevState && prevState.glow === 0 && prevState.border === 0) return;

        element.style.setProperty('--glow-opacity', '0');
        element.style.setProperty('--border-glow-intensity', '0');
        stateCache.set(element, { x: -9999, y: -9999, glow: 0, border: 0 });
    }

    function applyGlowToElement(element, rect, maxDistance, borderDistance) {
        const relativeX = mouseX - rect.left;
        const relativeY = mouseY - rect.top;
        const distance = distanceToRect(mouseX, mouseY, rect);
        const surfaceDistance = distanceToRectSurface(mouseX, mouseY, rect);
        const glowOpacity = Math.max(0, 1 - (distance / maxDistance));
        const borderGlow = surfaceDistance <= borderDistance
            ? Math.max(0, 1 - (surfaceDistance / borderDistance))
            : 0;

        writeGlowState(element, relativeX, relativeY, glowOpacity, borderGlow);
    }

    function resetAllGlow() {
        staticTargets.forEach(({ element }) => resetElementGlow(element));
        if (activeCard) {
            resetElementGlow(activeCard);
            activeCard = null;
        }
    }

    function getActiveOverlay() {
        const overlays = document.querySelectorAll('.modal-overlay.active, .custom-alert-overlay.active');
        if (!overlays.length) return null;
        return overlays[overlays.length - 1];
    }

    function scheduleFrame() {
        if (rafId) return;
        rafId = window.requestAnimationFrame(updateFrame);
    }

    function updateFrame(now) {
        rafId = 0;

        if (minFrameInterval && now - lastFrameAt < minFrameInterval) {
            scheduleFrame();
            return;
        }
        lastFrameAt = now;

        if (!pointerInWindow) {
            resetAllGlow();
            return;
        }

        const activeOverlay = getActiveOverlay();
        const modalActive = Boolean(activeOverlay);

        if (staticTargetsDirty) {
            staticTargets = collectStaticTargets();
            staticTargetsDirty = false;
            rectsDirty = true;
        }

        staticTargets.forEach((target) => {
            if (rectsDirty || !target.rect || target.rect.width === 0 || target.rect.height === 0) {
                target.rect = target.element.getBoundingClientRect();
            }

            // When any modal/dialog is open, keep glow only inside the topmost active overlay.
            if (modalActive && !activeOverlay.contains(target.element)) {
                resetElementGlow(target.element);
                return;
            }

            // Prevent hidden modal/dialog controls from glowing when nothing is open.
            if (!modalActive && target.allowWhenModal) {
                resetElementGlow(target.element);
                return;
            }

            applyGlowToElement(
                target.element,
                target.rect,
                target.maxDistance,
                target.borderDistance,
            );
        });
        rectsDirty = false;

        if (modalActive) {
            if (activeCard) {
                resetElementGlow(activeCard);
                activeCard = null;
            }
            return;
        }

        const hoveredElement = document.elementFromPoint(mouseX, mouseY);
        const hoveredCard = hoveredElement ? hoveredElement.closest(cardSelectorUnion) : null;

        if (hoveredCard !== activeCard) {
            if (activeCard) {
                resetElementGlow(activeCard);
            }
            activeCard = hoveredCard;
        }

        if (activeCard) {
            const profile = getCardProfile(activeCard);
            if (profile) {
                const cardRect = activeCard.getBoundingClientRect();
                applyGlowToElement(activeCard, cardRect, profile.maxDistance, profile.borderDistance);
            }
        }
    }

    function handlePointerMove(event) {
        mouseX = event.clientX;
        mouseY = event.clientY;
        pointerInWindow = true;
        scheduleFrame();
    }

    if ('onpointermove' in window) {
        document.addEventListener('pointermove', handlePointerMove, { passive: true });
    } else {
        document.addEventListener('mousemove', handlePointerMove, { passive: true });
    }

    window.addEventListener('resize', () => {
        rectsDirty = true;
        scheduleFrame();
    }, { passive: true });

    window.addEventListener('scroll', () => {
        rectsDirty = true;
        if (pointerInWindow) {
            scheduleFrame();
        }
    }, { passive: true, capture: true });

    document.addEventListener('mouseleave', () => {
        pointerInWindow = false;
        resetAllGlow();
    }, { passive: true });

    window.addEventListener('blur', () => {
        pointerInWindow = false;
        resetAllGlow();
    });

    const observer = new MutationObserver(() => {
        staticTargetsDirty = true;
        scheduleFrame();
    });

    observer.observe(document.body, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ['class', 'style'],
    });
});
