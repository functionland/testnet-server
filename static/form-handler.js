// form-handler.js
document.addEventListener("DOMContentLoaded", function() {
    var form = document.getElementById('registerForm');
    var submitButton = form.querySelector('button[type="submit"]');
    var successMessage = document.getElementById('successMessage');
    var errorMessage = document.getElementById('errorMessage');
    var verifyingMessage = document.getElementById('verifyingMessage'); // Add an element for this in your HTML

    // Function to parse URL search parameters
    function getSearchParams(k) {
        var p = {};
        location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi, function(s, k, v) { p[k] = v });
        return k ? p[k] : p;
    }
    // Automatically fill the tokenAccountId field if accountId is present in the URL
    var accountId = getSearchParams('accountId');
    if (accountId) {
        tokenAccountId.value = accountId.startsWith('5') ? accountId : '';
    }
    form.addEventListener('submit', function(event) {
        event.preventDefault();

        // Clear existing messages
        successMessage.style.display = 'none';
        errorMessage.style.display = 'none';
        
        // Show verifying message
        verifyingMessage.innerText = "Verifying the Request. This may take up to 1 minute";
        verifyingMessage.style.display = 'block';

        // Disable the button
        submitButton.disabled = true;

        var formData = new FormData(form);
        fetch('/register', {
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
});
