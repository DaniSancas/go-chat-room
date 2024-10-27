// Constants
var API_URL = 'http://localhost:8080';
var WS_URL = 'ws://localhost:8080/stream';

// Global state
var state = {
    username: null,
    token: null,
    ws: null
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
        success: function (data) {
            console.log(data);
            parsed = $.parseJSON(data);
            state.username = username;
            // Parse { "token": "mytoken" } from data and store in state.token
            state.token = parsed.token;
            $('#logout-username').text(username);
            $('#login-form').hide();
            $('#logout-form').show();

            // Start websocket connection
            state.ws = connect();
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
        success: function (data) {
            console.log(data);
            state.username = null;
            state.token = null;
            $('#login-form').show();
            $('#logout-form').hide();

            // Close websocket connection
            disconnect(state.ws);
            state.ws = null;
        }
    });
}

function connect() {
    var ws = new WebSocket(WS_URL);

    ws.onopen = function () {
        console.log('Connected to server');
        handshake(ws);
    };
    ws.onmessage = function (event) {
        listenWs(event);
    };
    ws.onclose = function () {
        console.log('Disconnected from server');
    };

    return ws;
}

// First message to websocket server
function handshake(ws) {
    message = { "username": state.username, "token": state.token }
    console.log('Sending ws handshake: ' + JSON.stringify(message));
    ws.send(JSON.stringify(message));
}

function disconnect(ws) {
    ws.close();
}

function listenWs(event) {
    console.log('Received message: ' + event.data);
}

// Bindings on document.ready
$(document).ready(function () {
    $("#login-button").on("click", function () {
        login();
    });
    $("#logout-button").on("click", function () {
        logout();
    });
});
