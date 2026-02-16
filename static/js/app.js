// SSHmasher custom JS

// Theme toggle functions
function applyTheme(theme) {
    const html = document.documentElement;
    const icon = document.getElementById('theme-icon');
    
    if (theme === 'auto') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        html.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
        if (icon) {
            icon.className = prefersDark ? 'fa-solid fa-moon' : 'fa-solid fa-sun';
        }
    } else {
        html.setAttribute('data-theme', theme);
        if (icon) {
            icon.className = theme === 'dark' ? 'fa-solid fa-moon' : 'fa-solid fa-sun';
        }
    }
    localStorage.setItem('theme', theme);
}

function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const newTheme = current === 'dark' ? 'light' : 'dark';
    applyTheme(newTheme);
}

// Open modal after HTMX request for modal content
document.body.addEventListener('htmx:afterSwap', function(e) {
    // Check if the swap target is or is inside config-modal-content
    const target = e.detail.target;
    if (!target) return;
    
    const modalContent = document.getElementById('config-modal-content');
    const modal = document.getElementById('config-modal');
    
    if (modalContent && modal && (target === modalContent || modalContent.contains(target))) {
        if (modalContent.innerHTML.trim().length > 0) {
            modal.showModal();
        }
    }
});

// Copy to clipboard for public keys
function copyPublicKey(btn) {
    const publicKey = btn.getAttribute('data-public-key');
    navigator.clipboard.writeText(publicKey).then(function() {
        const icon = btn.querySelector('i');
        icon.classList.replace('fa-copy', 'fa-check');
        setTimeout(function() {
            icon.classList.replace('fa-check', 'fa-copy');
        }, 2000);
    });
}

// Listen for HTMX response errors and show alert
document.addEventListener("htmx:responseError", function(evt) {
    const msg = evt.detail.xhr.responseText || "An error occurred";
    showAlert(msg, "error");
});

// Show a temporary alert message
function showAlert(message, type) {
    const container = document.getElementById("alerts");
    if (!container) return;

    const div = document.createElement("div");
    div.className = "alert alert-" + type;
    div.role = "alert";
    div.textContent = message;
    container.appendChild(div);

    setTimeout(function() {
        div.remove();
    }, 5000);
}

// Listen for custom events from HTMX responses
document.addEventListener("showAlert", function(evt) {
    showAlert(evt.detail.message, evt.detail.type || "success");
});
