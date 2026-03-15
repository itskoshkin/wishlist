// ============================================
// API CLIENT
// ============================================

const API_BASE = '/api/v1'; // API base URL

// Get auth headers
function getAuthHeaders() {
    const token = localStorage.getItem('access_token'); // Get access token
    return {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
    };
}

// Handle API errors
async function handleAPIError(response) {
    if (response.status === 401) {
        // Unauthorized - try to refresh token
        const refreshed = await refreshAccessToken();
        if (!refreshed) {
            // Refresh failed, redirect to login
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/';
            return null;
        }
        return 'retry'; // Signal to retry the request
    }

    const error = await response.json();
    throw new Error(translateApiErrorMessage(error.message, 'common.error'));
}

// Refresh access token
async function refreshAccessToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return false;

    try {
        const response = await fetch(`${API_BASE}/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        if (!response.ok) return false;

        const data = await response.json();
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        return true;
    } catch (error) {
        console.error('Token refresh error:', error);
        return false;
    }
}

// API request wrapper
async function apiRequest(endpoint, options = {}) {
    let response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers: {
            ...getAuthHeaders(),
            ...options.headers
        }
    });

    if (!response.ok) {
        const result = await handleAPIError(response);
        if (result === 'retry') {
            // Retry with new token
            response = await fetch(`${API_BASE}${endpoint}`, {
                ...options,
                headers: {
                    ...getAuthHeaders(),
                    ...options.headers
                }
            });
            if (!response.ok) {
                return null;
            }
        } else {
            return null;
        }
    }

    // Handle 204 No Content
    if (response.status === 204) {
        return {};
    }

    return response.json();
}

// ============================================
// USER API
// ============================================

// Get current user
async function getCurrentUser() {
    return await apiRequest('/users/me');
}

// Update current user
async function updateCurrentUser(data) {
    return await apiRequest('/users/me', {
        method: 'PATCH',
        body: JSON.stringify(data)
    });
}

// Get public user by username
async function getUserByUsername(username) {
    return await apiRequest(`/users/by-username/${encodeURIComponent(username)}`);
}

async function searchUsers(query) {
    return await apiRequest(`/users/search?query=${encodeURIComponent(query)}`);
}

// Update password
async function updatePassword(oldPassword, newPassword) {
    return await apiRequest('/users/me/update-password', {
        method: 'PATCH',
        body: JSON.stringify({
            current_password: oldPassword,
            new_password: newPassword
        })
    });
}

// Upload avatar
async function uploadAvatar(file) {
    const formData = new FormData();
    formData.append('avatar', file);

    const token = localStorage.getItem('access_token');
    const response = await fetch(`${API_BASE}/users/me/avatar`, {
        method: 'PUT',
        headers: {
            'Authorization': token ? `Bearer ${token}` : ''
        },
        body: formData
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(translateApiErrorMessage(error.message, 'profile.avatar.uploadFailed'));
    }

    return response.json();
}

// Delete avatar
async function deleteUserAvatar() {
    return await apiRequest('/users/me/avatar', {
        method: 'DELETE'
    });
}

async function requestPasswordReset(email) {
    return await apiRequest('/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email })
    });
}

async function setNewPassword(token, newPassword) {
    return await apiRequest('/auth/set-new-password', {
        method: 'POST',
        body: JSON.stringify({
            token,
            new_password: newPassword
        })
    });
}

// Upload wish image
async function uploadWishImage(listId, wishId, file) {
    const formData = new FormData();
    formData.append('image', file);

    const token = localStorage.getItem('access_token');
    const response = await fetch(`${API_BASE}/lists/${listId}/wishes/${wishId}/image`, {
        method: 'PUT',
        headers: {
            'Authorization': token ? `Bearer ${token}` : ''
        },
        body: formData
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(translateApiErrorMessage(error.message, 'wish.image.updateFailed'));
    }

    return response.json();
}

// ============================================
// LIST API
// ============================================

// Get all lists for current user
async function getUserLists() {
    return await apiRequest('/lists');
}

// Get public lists for selected user
async function getPublicListsByUserId(userId) {
    return await apiRequest(`/users/${encodeURIComponent(userId)}/lists`);
}

// Get list by ID
async function getListById(listId) {
    return await apiRequest(`/lists/${listId}`);
}

// Get list by shared slug
async function getListBySlug(slug) {
    return await apiRequest(`/lists/shared/${slug}`);
}

// Create list
async function createList(title) {
    return await apiRequest('/lists', {
        method: 'POST',
        body: JSON.stringify({ title })
    });
}

// Update list
async function updateList(listId, data) {
    return await apiRequest(`/lists/${listId}`, {
        method: 'PATCH',
        body: JSON.stringify(data)
    });
}

// Rotate shared link
async function rotateSharedLink(listId) {
    return await apiRequest(`/lists/${listId}/rotate-share-link`, {
        method: 'POST'
    });
}

// Delete list
async function deleteList(listId) {
    return await apiRequest(`/lists/${listId}`, {
        method: 'DELETE'
    });
}

// ============================================
// WISH API
// ============================================

// Create wish
async function createWish(listId, data) {
    return await apiRequest(`/lists/${listId}/wishes`, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

// Update wish
async function updateWish(listId, wishId, data) {
    return await apiRequest(`/lists/${listId}/wishes/${wishId}`, {
        method: 'PATCH',
        body: JSON.stringify(data)
    });
}

// Delete wish
async function deleteWish(listId, wishId) {
    return await apiRequest(`/lists/${listId}/wishes/${wishId}`, {
        method: 'DELETE'
    });
}

// Reserve wish
async function reserveWish(listId, wishId) {
    return await apiRequest(`/lists/${listId}/wishes/${wishId}/reserve`, {
        method: 'POST'
    });
}

// Release wish
async function releaseWish(listId, wishId) {
    return await apiRequest(`/lists/${listId}/wishes/${wishId}/reserve`, {
        method: 'DELETE'
    });
}
