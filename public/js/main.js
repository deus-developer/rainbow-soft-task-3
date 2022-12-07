(function () {
    function onGenerateButtonSubmit(event) {
        event.preventDefault();
        const data = new FormData(event.target);
        const countNumbers = parseInt(data.get("countNumbers"));
        const countThreads = parseInt(data.get("countThreads"));

        if (!countNumbers || !countThreads) {
            return alert("Ошибка ввода");
        }

        const params = new URLSearchParams({
            "countNumbers": countNumbers,
            "countThreads": countThreads
        });
        const url = "ws://127.0.0.1:8080/generator?" + params.toString();
        if (window.websocketGenerator) {
            window.websocketGenerator.close();
        }

        document.getElementById("numbersOutput").innerHTML = "";

        // Open websocket connection with backend
        window.websocketGenerator = new WebSocket(url);
        window.websocketGenerator.addEventListener("message", (event) => {
            // Show random numbers from websocket
            document.getElementById("numbersOutput").innerHTML += " " + event.data;
        });
    }

    // Wait page load
    document.addEventListener("DOMContentLoaded", function(event) {
        var formElement = document.getElementById("generator");
        if (!formElement) {
            return console.error("Form element not found");
        }
        
        // Add hook to form submit then request backend
        formElement.addEventListener("submit", onGenerateButtonSubmit);
    });
})();
