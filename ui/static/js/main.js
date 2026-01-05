document.addEventListener('alpine:init', () => {
    Alpine.data('toastManager', () => ({
        toast: null,
        timeout: null,

        init() {
            // Create bound event handlers
            this.handleShowToast = (event) => {
                // console.log("showToast event received:", event.detail);
                this.showToast(event.detail.level, event.detail.message);
            };


            // Add event listeners with properly bound handlers
            document.body.addEventListener('showToast', this.handleShowToast);
            // document.body.addEventListener('htmx:trigger', this.handleTrigger);

            // console.log("Toast manager initialized - listening for events");
        },

        showToast(level, message) {
            // console.log(`Showing toast: ${level} - ${message}`);

            // Clear existing timeout if there is one
            if (this.timeout) {
                clearTimeout(this.timeout);
                this.timeout = null;
            }

            // Create the new toast
            this.toast = {
                id: Date.now(),
                level,
                message,
                visible: true,
                style: this.getToastStyle(level),
                icon: this.getIconClass(level)
            };

            // Set auto-hide timeout
            this.timeout = setTimeout(() => {
                this.hideToast();
            }, 5000);
        },

        hideToast() {
            if (!this.toast) return;

            // console.log(`Hiding toast: ${this.toast.id}`);
            this.toast.visible = false;

            // Clear after animation completes
            setTimeout(() => {
                this.toast = null;
            }, 300);
        },

        getToastStyle(level) {
            switch (level) {
                case 'success':
                    return "background-color: #f0fdf4; border-color: #86efac;";
                case 'error':
                    return "background-color: #fef2f2; border-color: #fecaca;";
                case 'warning':
                    return "background-color: #fffbeb; border-color: #fde68a;";
                case 'info':
                    return "background-color: #eff6ff; border-color: #bfdbfe;";
                default:
                    return "background-color: #ffffff; border-color: #e5e7eb;";
            }
        },

        getIconClass(level) {
            switch (level) {
                case 'success':
                    return 'heroicons:check-circle';
                case 'error':
                    return 'heroicons:x-circle';
                case 'warning':
                    return 'heroicons:exclamation-triangle';
                case 'info':
                    return 'heroicons:information-circle';
                default:
                    return 'heroicons:information-circle';
            }
        },

        getToastClass(level) {
            switch (level) {
                case 'success':
                    return "bg-accent-50 border border-accent-200 text-primary-800";
                case 'error':
                    return "bg-secondary-50 border border-secondary-200 text-primary-800";
                case 'warning':
                    return "bg-amber-50 border border-amber-200 text-primary-800";
                case 'info':
                    return "bg-blue-50 border border-blue-200 text-primary-800";
                default:
                    return "bg-blue-50 border border-blue-200 text-primary-800";
            }
        }
    }));
});
