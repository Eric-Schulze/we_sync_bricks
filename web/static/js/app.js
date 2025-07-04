/**
 * WE Sync Bricks - Application JavaScript
 * 
 * This file contains custom JavaScript functionality for the WE Sync Bricks application.
 * It works alongside HTMX and Alpine.js to provide interactive features.
 */

document.addEventListener('DOMContentLoaded', function() {
    console.log('WE Sync Bricks app initialized');

    // Initialize notification system
    initializeNotifications();
    
    // Initialize modal handlers
    initializeModals();
    
    // Initialize form enhancements
    initializeForms();
});

/**
 * Notification System
 */
function initializeNotifications() {
    // Create notification container if it doesn't exist
    if (!document.getElementById('notifications')) {
        const notificationContainer = document.createElement('div');
        notificationContainer.id = 'notifications';
        notificationContainer.className = 'fixed top-4 right-4 z-50 space-y-2';
        document.body.appendChild(notificationContainer);
    }
}

/**
 * Show a notification message
 * @param {string} message - The message to display
 * @param {string} type - The type of notification (success, error, warning, info)
 * @param {number} duration - Duration in milliseconds (default: 5000)
 */
function showNotification(message, type = 'info', duration = 5000) {
    const container = document.getElementById('notifications');
    if (!container) return;

    const notification = document.createElement('div');
    notification.className = `
        p-4 rounded-lg shadow-lg max-w-sm transform transition-all duration-300 translate-x-full opacity-0
        ${getNotificationClasses(type)}
    `;
    
    notification.innerHTML = `
        <div class="flex items-center justify-between">
            <span class="text-sm font-medium">${message}</span>
            <button class="ml-3 text-white hover:text-gray-200" onclick="dismissNotification(this)">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                </svg>
            </button>
        </div>
    `;

    container.appendChild(notification);

    // Animate in
    setTimeout(() => {
        notification.classList.remove('translate-x-full', 'opacity-0');
    }, 10);

    // Auto-dismiss
    if (duration > 0) {
        setTimeout(() => {
            dismissNotification(notification.querySelector('button'));
        }, duration);
    }
}

/**
 * Get CSS classes for notification types
 */
function getNotificationClasses(type) {
    const classes = {
        success: 'bg-green-500 text-white',
        error: 'bg-red-500 text-white',
        warning: 'bg-yellow-500 text-white',
        info: 'bg-blue-500 text-white'
    };
    return classes[type] || classes.info;
}

/**
 * Dismiss a notification
 */
function dismissNotification(button) {
    const notification = button.closest('div[class*="rounded-lg"]');
    if (notification) {
        notification.classList.add('translate-x-full', 'opacity-0');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }
}

/**
 * Modal System
 */
function initializeModals() {
    // Handle modal close on backdrop click
    document.addEventListener('click', function(e) {
        if (e.target.classList.contains('modal-overlay')) {
            closeModal();
        }
    });

    // Handle ESC key to close modals
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeModal();
        }
    });
}

/**
 * Close modal
 */
function closeModal() {
    const modalContainer = document.getElementById('modal-container');
    if (modalContainer) {
        modalContainer.innerHTML = '';
    }
}

/**
 * Form Enhancements
 */
function initializeForms() {
    // Auto-focus first input in modals
    document.addEventListener('htmx:afterSettle', function(e) {
        const modal = e.target.querySelector('.modal');
        if (modal) {
            const firstInput = modal.querySelector('input, textarea, select');
            if (firstInput) {
                firstInput.focus();
            }
        }
    });

    // Form validation feedback
    document.addEventListener('htmx:responseError', function(e) {
        showNotification('There was an error processing your request. Please try again.', 'error');
    });

    document.addEventListener('htmx:sendError', function(e) {
        showNotification('Network error. Please check your connection and try again.', 'error');
    });
}

/**
 * Global utility functions
 */
window.WESyncBricks = {
    showNotification,
    dismissNotification,
    closeModal
};