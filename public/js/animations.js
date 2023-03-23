// animations.js is where we store functions used for animating the interface.

// expandSidebar() is used on mobile to expand the sidebar.tmpl component, 
// which is hidden by default on mobile. Triggered by pressing the hamburger
// menu found at the top of the page.
function expandSidebar() {
        var bg = document.getElementById("bgBlur")
        bg.style.display = "unset";
        var sb = document.getElementById("sidebar")
        sb.style.width = 0;
        sb.style.fontSize = 0;
        sb.style.display = "unset";
        var i = 0;
        var expand = setInterval(function() {
                sb.style.width = i++ + "%";
                sb.style.fontSize = i/30 + "em";
                if (i >= 60) {
                        clearInterval(expand);
                }
        }, 10);
}

// hideSidebar() hides the sidebar.tmpl component. Only used for mobile. 
// Triggered by pressing the darkened background.
function hideSidebar() {
        hideAuth();
        var bg = document.getElementById("bgBlur")
        bg.style.display = "none";
        var sb = document.getElementById("sidebar")
        var i = 40;
        var expand = setInterval(function() {
                sb.style.width = i-- + "%";
                sb.style.fontSize = i*0.025 + "em";
                if (i >= 40) {
                        sb.style.display = "none";
                        clearInterval(expand);
                }
        }, 10);
}

// showAuth is used to show the signin/signup form
function showAuth(isLoggedIn) {
        document.getElementById("closeAuth").style.display = "unset";
        document.getElementById("authButt_").style.display = "none";
        if (isLoggedIn) {
                auth("logout"); 
        } else {
                document.getElementById("hiddenAuth").style.display = "unset"; 
        }
}

// hideAuth is used to hide the signin/signup form
function hideAuth() {
        document.getElementById("hiddenAuth").style.display = "none"; 
        document.getElementById("authButt_").style.display = "unset";
        document.getElementById("closeAuth").style.display = "none";
}


