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

        getIconClass(level) {
            switch (level) {
                case 'success':
                    return 'heroicons:check-circle-solid';
                case 'error':
                    return 'heroicons:x-circle-solid';
                case 'warning':
                    return 'heroicons:exclamation-triangle-solid';
                case 'info':
                    return 'heroicons:information-circle-solid';
                default:
                    return 'heroicons:information-circle-solid';
            }
        },

        getToastClass(level) {
            switch (level) {
                case 'success':
                    return "bg-green-50 border border-green-200 text-green-800";
                case 'error':
                    return "bg-red-50 border border-red-200 text-red-800";
                case 'warning':
                    return "bg-amber-50 border border-amber-200 text-amber-800";
                case 'info':
                    return "bg-blue-50 border border-blue-200 text-blue-800";
                default:
                    return "bg-blue-50 border border-blue-200 text-blue-800";
            }
        }
    }));
});
