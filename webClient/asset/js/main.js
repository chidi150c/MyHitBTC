//Function to toggle cards
function myFunction() {
    var x = document.getElementById("Demo");
    var y = document.getElementById("ichat");
    
    if (x.className.indexOf("w3-show") == -1) {
        x.className += " w3-show";
       y.className = y.className.replace("fa fa-comments", "");
        y.innerHTML = " &times "
    } else {
        x.className = x.className.replace(" w3-show", "");
        y.className = y.className.replace("", "fa fa-comments");
        y.innerHTML = ""
        
    }
}