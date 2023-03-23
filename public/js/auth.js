// auth is used for signing up and signing in/out. path could be:
// /api/signup
// /api/signin
// /api/logout
function auth(path) {
        var xhr = new XMLHttpRequest();

        xhr.open("POST", "/api/" + path);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.onload = function() {
                if (xhr.status === 200) {
                        var res = JSON.parse(xhr.responseText);
                        if (res.success == "false") {
                                // If we aren't successful we display an error.
                                document.getElementById("errorField").innerHTML = res.error;
                        } else {
                                // Reload the page now that the user is signed in.
                                window.location.reload();
                        }
                }
        };

        // For now, all we're sending is a username and password, but we may start
        // asking for email or mobile number at some point.
        xhr.send(JSON.stringify({
                                password: document.getElementById("password").value,
                                username: document.getElementById("username").value
        }));
}
