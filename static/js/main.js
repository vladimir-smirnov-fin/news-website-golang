document.addEventListener('DOMContentLoaded', function() {
    setTimeout(function() {
        var successAlerts = document.querySelectorAll('.alert-success');
        successAlerts.forEach(function(alert) {
            alert.style.transition = 'opacity 0.5s ease';
            alert.style.opacity = '0';
            setTimeout(function() {
                if (alert.parentNode) {
                    alert.remove();
                }
            }, 500);
        });
    }, 3000);
});