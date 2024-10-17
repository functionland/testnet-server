document.addEventListener("DOMContentLoaded", function() {
    let streamrForm = document.getElementById('streamrForm');
    let submitButton = streamrForm.querySelector('button[type="submit"]');
    let successMessage = document.getElementById('successMessage');
    let errorMessage = document.getElementById('errorMessage');
    let verifyingMessage = document.getElementById('verifyingMessage');

    
    if (streamrForm) {
        streamrForm.addEventListener('submit', function(event) {
            event.preventDefault();

            // Clear existing messages
            successMessage.style.display = 'none';
            errorMessage.style.display = 'none';

            // Show verifying message
            verifyingMessage.innerText = "Processing your Streamr node request...";
            verifyingMessage.style.display = 'block';

            // Disable the button
            submitButton.disabled = true;

            let formData = new FormData(streamrForm);
            fetch('/streamr', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                verifyingMessage.style.display = 'none'; // Hide verifying message

                if (data.status === 'success') {
                    successMessage.innerText = data.message;
                    successMessage.style.display = 'block';
                } else {
                    errorMessage.innerText = data.message;
                    errorMessage.style.display = 'block';
                    submitButton.disabled = false; // Re-enable the button on error
                }
            })
            .catch(error => {
                verifyingMessage.style.display = 'none'; // Hide verifying message

                errorMessage.innerText = 'Error submitting form: ' + error.message;
                errorMessage.style.display = 'block';
                submitButton.disabled = false; // Re-enable the button on fetch failure
            });
        });
    }
});