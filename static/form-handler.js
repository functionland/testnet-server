// form-handler.js
document.addEventListener("DOMContentLoaded", function() {
    var form = document.getElementById('registerForm');
    form.addEventListener('submit', function(event) {
        event.preventDefault();

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
            } else {
                // If the server responds with a non-success status
                document.getElementById('errorMessage').innerText = data.message;
                document.getElementById('errorMessage').style.display = 'block';
                document.getElementById('successMessage').style.display = 'none';
            }
        })
        .catch(error => {
            document.getElementById('errorMessage').innerText = 'Error submitting form: ' + error.message;
            document.getElementById('errorMessage').style.display = 'block';
            document.getElementById('successMessage').style.display = 'none';
        });
    });
});
