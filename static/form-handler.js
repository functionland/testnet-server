document.addEventListener("DOMContentLoaded", function() {
    let form = document.getElementById('registerForm');
    let submitButton = form.querySelector('button[type="submit"]');
    let successMessage = document.getElementById('successMessage');
    let errorMessage = document.getElementById('errorMessage');
    let verifyingMessage = document.getElementById('verifyingMessage');
    let appIdSelect = document.getElementById('appId'); // Get the appId select element


    // Add event listener to handle appId changes
    appIdSelect.addEventListener('change', function() {
        if (appIdSelect.value === 'land.fx.fotos') {
            form.email.disabled = true;
            form.orderId.disabled = true;
            form.phoneNumber.disabled = true;
        } else {
            form.email.disabled = false;
            form.orderId.disabled = false;
            form.phoneNumber.disabled = false;
        }
    });

    const rotatingMessages = [
        "Initiating blockchain handshake...",
        "Gathering quantum bits...",
        "Encrypting with unbreakable codes...",
        "Deploying smart contracts with style...",
        "Minting your digital assets...",
        "Distributing decentralized dreams...",
        "Cross-verifying ledger integrity...",
        "Summoning blockchain spirits...",
        "Aligning cryptographic stars...",
        "Generating ultra-secure hash...",
        "Propagating transactions to network...",
        "Applying for blockchain citizenship...",
        "Reticulating splines, blockchain style...",
        "Engaging consensus mechanism...",
        "Convincing nodes to agree...",
        "Mining some digital gold...",
        "Crafting pixels into NFTs...",
        "Setting up your digital wallet...",
        "Populating the ledger with magic...",
        "Consulting the oracle...",
        "Performing secret handshakes...",
        "Synchronizing with the metaverse...",
        "Decentralizing the centralized...",
        "Charging quantum flux capacitors...",
        "Finalizing tokenomics equations...",
        "Achieving blockchain enlightenment...",
        "Warming up the crypto engines...",
        "Ensuring immutability...",
        "Broadcasting to the network...",
        "Welcome aboard! Enjoy the decentralized ride!"
    ];
    

    let messageIndex = 0; // To keep track of which message is currently displayed
    let messageInterval; // For clearing the interval when done

    function rotateMessage() {
        if (messageIndex >= rotatingMessages.length) {
            messageIndex = 0; // Reset index if it exceeds the array
        }
        verifyingMessage.innerText = rotatingMessages[messageIndex++];
    }

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

    // Automatically disable fields if appId is 'land.fx.fotos'
    if (appIdSelect.value === 'land.fx.fotos') {
        form.email.disabled = true;
        form.orderId.disabled = true;
        form.phoneNumber.disabled = true;
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
        verifyingMessage.innerText = rotatingMessages[0];
        verifyingMessage.style.display = 'block';
        messageIndex = 1; // Next message to display
        messageInterval = setInterval(rotateMessage, 2000); // Change message every 2 seconds

        // Disable the button
        submitButton.disabled = true;

        let formData = new FormData(form);
        fetch('/register', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            clearInterval(messageInterval); // Stop rotating messages
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