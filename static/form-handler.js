document.addEventListener("DOMContentLoaded", function() {
    let form = document.getElementById('registerForm');
    let submitButton = form.querySelector('button[type="submit"]');
    let successMessage = document.getElementById('successMessage');
    let errorMessage = document.getElementById('errorMessage');
    let verifyingMessage = document.getElementById('verifyingMessage');
    let appIdSelect = document.getElementById('appId'); // Get the appId select element

    // Function to parse URL search parameters
    function getSearchParams(k) {
        let p = {};
        location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi, function(s, k, v) { p[k] = v });
        return k ? p[k] : p;
    }

    // Automatically set the appId field based on URL parameter or default to "main"
    let appIdParam = getSearchParams('appId');
    let validAppIds = ['land.fx.fotos', 'land.fx.blox', 'main']; // List of valid appIds
    if (appIdParam && validAppIds.includes(appIdParam)) {
        appIdSelect.value = appIdParam;
    } else {
        appIdSelect.value = 'main'; // Default to 'main' if not valid or not present
    }

    // Automatically fill the tokenAccountId field if accountId is present in the URL
    let accountId = getSearchParams('accountId');
    if (accountId) {
        form.tokenAccountId.value = accountId.startsWith('5') ? accountId : '';
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

        let formData = new FormData(form);
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