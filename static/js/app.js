// SSHmasher custom JS

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
