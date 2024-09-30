// Constants
var API_URL = 'http://localhost:8080';

// Global state
var state = {
    username: null,
    token: null,
};

// Function to get the username from #username-login 
// input and do an AJAX call to /login with a POST request in JSON format
// Example JSON: { "username": "myusername" }
function login() {
    var username = $('#login-username').val();
    $.ajax({
        url: API_URL + '/login',
        type: 'POST',
        data: JSON.stringify({ "username": username }),
        contentType: 'application/json',
        success: function(data) {
            console.log(data);
            parsed = $.parseJSON(data);
            state.username = username;
            // Parse { "token": "mytoken" } from data and store in state.token
            state.token = parsed.token;
            $('#logout-username').text(username);
            $('#login-form').hide();
            $('#logout-form').show();
        }
    });
}

// Function to do an AJAX call to /logout with a POST request in JSON format
// Example JSON: {"username": "myusername", "token": "mytoken"}
function logout() {
    $.ajax({
        url: API_URL + '/logout',
        type: 'POST',
        data: JSON.stringify({ "username": state.username, "token": state.token }),
        contentType: 'application/json',
        success: function(data) {
            console.log(data);
            state.username = null;
            state.token = null;
            $('#login-form').show();
            $('#logout-form').hide();
        }
    });
}

// Bindings on document.ready
$(document).ready(function() {
    $("#login-button").on("click", function() {
        login();
    });
    $("#logout-button").on("click", function() {
        logout();
    });
});
