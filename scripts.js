let socket;
let isHost = false;

// Store connection info when establishing connection
function connectWebSocket(username, role) {
    const wsUrl = `ws://${window.location.hostname}:6969/ws`;
    console.log("Attempting WebSocket connection to:", wsUrl, "for user:", username, "role:", role);
    
    try {
        socket = new WebSocket(wsUrl);
        
        socket.onopen = function() {
            console.log("WebSocket Connected Successfully!");
            
            // Send initial connection message with proper role
            const msg = {
                username: username,
                message: role || (localStorage.getItem('isObserver') === 'true' ? 'observer' : 'player'),
                type: "JOIN_SESSION"
            };
            console.log("Sending join message:", msg);
            socket.send(JSON.stringify(msg));

            // Add automatic navigation after sending join message
            if (window.location.pathname === "/" || window.location.pathname === "/index.html") {
                window.location.href = '/game.html';
            }
        };
        
        socket.onmessage = function(event) {
            console.log('Raw message received:', event.data);
            try {
                const response = JSON.parse(event.data);
                
                if (response.username === "Server") {
                    const header = document.querySelector('h1');
                    if (header) header.textContent = `Planning Poker - ${response.message}`;
                    const serverMessage = document.getElementById('serverMessage');
                    if (serverMessage) serverMessage.textContent = response.message;

                    // Handle player list updates
                    if (response.type === "PLAYER_LIST") {
                        updatePlayersList(response.players);
                    }
                }
            } catch (e) {
                console.error('Error parsing message:', e);
            }
        };
        
        socket.onerror = function(error) {
            console.error('WebSocket Error:', error);
            console.error('WebSocket State:', socket.readyState);
            const serverMessage = document.getElementById('serverMessage');
            serverMessage.textContent = "Error connecting to server";
        };

        socket.onclose = () => {
            console.log("WebSocket Disconnected! State:", socket.readyState);
            const serverMessage = document.getElementById('serverMessage');
            serverMessage.textContent = "Disconnected from server. Reconnecting...";
            // Attempt to reconnect after a delay
            setTimeout(() => connectWebSocket(username, role), 1000);
        };
    } catch (error) {
        console.error("WebSocket connection error:", error);
    }
}

function handleSubmit(event) {
    event.preventDefault();
    
    const playerName = document.getElementById('playerName').value;
    const isObserver = document.querySelector('input[name="role"][value="observer"]').checked;
    console.log("Joining as observer:", isObserver);
    
    localStorage.setItem('isObserver', isObserver);
    localStorage.setItem('username', playerName);
    
    // Use the existing connectWebSocket function instead of creating a new connection
    connectWebSocket(playerName, isObserver ? "observer" : "player");
    
    return false;
}

function startGame() {
    if (!isHost) return;
    
    socket.send(JSON.stringify({
        username: localStorage.getItem('username'),
        type: "START_GAME"
    }));
}

function submitEstimate(value) {
    socket.send(JSON.stringify({
        username: localStorage.getItem('username'),
        message: value,
        type: "GUESS"
    }));
}

function updatePlayersList(players) {
    console.log("Updating players list with data:", players);  // Debug log
    
    const playersList = document.getElementById('playersList');
    const observersList = document.getElementById('observersList');
    if (!playersList || !observersList) return;

    // Separate players and observers
    const activePlayers = players.filter(p => !p.isObserver);
    const observers = players.filter(p => p.isObserver);

    // Update players list
    const playersHtml = activePlayers.map(player => {
        let voteDisplay = '...';
        if (player.hasVoted) {
            // Only show actual vote if it's explicitly provided and not null/undefined
            voteDisplay = player.vote !== undefined && player.vote !== null ? player.vote : 'VOTED';
            console.log(`Player ${player.name} vote status:`, {  // Debug log
                hasVoted: player.hasVoted,
                vote: player.vote,
                display: voteDisplay
            });
        }
        
        return `
            <div class="player-item ${player.hasVoted ? 'has-voted' : 'not-voted'}">
                <span>${player.name}</span>
                <span>${voteDisplay}</span>
            </div>
        `;
    }).join('');
    
    // Update observers list
    const observersHtml = observers.map(observer => `
        <div class="observer-item">
            <span>${observer.name}</span>
        </div>
    `).join('');
    
    playersList.innerHTML = playersHtml || '<div class="no-players">No active players</div>';
    observersList.innerHTML = observersHtml || '<div class="no-observers">No observers</div>';
}

function clearVotes() {
    socket.send(JSON.stringify({
        username: localStorage.getItem('username'),
        message: "clear_votes",
        type: "CLEAR_VOTES"
    }));
}

function showVotes() {
    socket.send(JSON.stringify({
        username: localStorage.getItem('username'),
        message: "show_votes",
        type: "SHOW_VOTES"
    }));
}

// Initialize connection when page loads
window.onload = function() {
    console.log("Page loaded");
    const serverMessage = document.getElementById('serverMessage');
    serverMessage.textContent = "Waiting for connection...";
    
    // Restore connection if username exists
    const username = localStorage.getItem('username');
    if (username) {
        console.log("Restoring connection for user:", username);
        isHost = localStorage.getItem('isHost') === 'true';
        
        // Hide voting cards if observer
        const isObserver = localStorage.getItem('isObserver') === 'true';
        const cardSelection = document.querySelector('.card-selection');
        if (cardSelection && isObserver) {
            cardSelection.style.display = 'none';
        }
        
        // Pass the role when reconnecting
        const role = isObserver ? 'observer' : 'player';
        connectWebSocket(username, role);
    } else {
        console.log("No stored username found");
    }
};
