(function () {
    // Open websocket connection with backend
    var webSocket = new WebSocket("ws://127.0.0.1:8080/generator");
    webSocket.addEventListener("message", (event) => {
        // Show random numbers from websocket
        var response = JSON.parse(event.data);
        if (!response) return;
        if (!response.ok) {
            if (response.error) alert(response.error);
            return;
        }
        
        var resultString = response.result.join(" ");
        document.getElementById("numbersOutput").innerHTML = resultString;
    });
    webSocket.addEventListener("close", (event) => {
        // Disable submit button on close websocket
        document.querySelectorAll(".btn").forEach((element) => {
            element.disabled = false;
        });
    });
    
    // Wait page load
    document.addEventListener("DOMContentLoaded", function(event) {
        var formElement = document.getElementById("generator");
        if (!formElement) {
            return console.error("Form element not found");
        }
        
        // Add hook to form submit then request backend
        formElement.addEventListener("submit", function(event) {
            event.preventDefault();
            var data = new FormData(event.target);
            var countNumbers = parseInt(data.get("countNumbers"));
            var countThreads = parseInt(data.get("countThreads"));

            if (!countNumbers || !countThreads) {
                return alert("Ошибка ввода");
            }

            // Send random numbers generator parameters to backend
            webSocket.send(JSON.stringify({
                "countNumbers": countNumbers,
                "countThreads": countThreads,
            }));
        });
    });
})();
