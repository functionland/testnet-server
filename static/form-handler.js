// form-handler.js
document.addEventListener("DOMContentLoaded", function() {
    var form = document.getElementById('registerForm');
    var submitButton = form.querySelector('button[type="submit"]');

    form.addEventListener('submit', function(event) {
        event.preventDefault();
        submitButton.disabled = true; // Disable the button on submit

        var formData = new FormData(form);
        fetch('/register', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                document.getElementById('successMessage').innerText = data.message;
                document.getElementById('successMessage').style.display = 'block';
                document.getElementById('errorMessage').style.display = 'none';
                // Do not re-enable the button on success
            } else {
                // If the server responds with a non-success status
                document.getElementById('errorMessage').innerText = data.message;
                document.getElementById('errorMessage').style.display = 'block';
                document.getElementById('successMessage').style.display = 'none';
                submitButton.disabled = false; // Re-enable the button on error
            }
        })
        .catch(error => {
            document.getElementById('errorMessage').innerText = 'Error submitting form: ' + error.message;
            document.getElementById('errorMessage').style.display = 'block';
            document.getElementById('successMessage').style.display = 'none';
            submitButton.disabled = false; // Re-enable the button on fetch failure
        });
    });
});
