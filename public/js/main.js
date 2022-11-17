(function () {
    // Wait page load
    document.addEventListener("DOMContentLoaded", function(event) {
        const formElement = document.getElementById("generator");
        if (!formElement) {
            return console.error("Form element not found")
        }

        // Add hook to form submit then request backend
        formElement.addEventListener("submit", function(event) {
            event.preventDefault();
            const data = new FormData(event.target);
            
            fetch('/random', {
                method: 'POST',
                headers: {
                    'Accept': 'application/json',
                },
                body: data
            })
            .then(response => response.json())
            .then(function (response) {
                document.getElementById('numbersOutput').value = response.join(" ")
            })
        });
    })
})()
