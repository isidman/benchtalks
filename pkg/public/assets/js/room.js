//state start

let roomId = null;
let encryptionKey = null;
let adminToken = null;
let displayName = null;
let ws = null;
let isConnected = false;
let pairToken = null;
let isPairingURL = false;
let notificationSound = null;
let soundEnabled = true;
let userCount_n = 0;

//first loads the page, then the sound is enabled
function initSound() {
    try {
        notificationSound = new Audio('/assets/sounds/notification.mp3');
        notificationSound.volume = 0.6;

        //preloading for less latency
        notificationSound.load();

        console.log('Notification sound loaded'); //debugging and testing
    } catch (error) {
        console.error('Failed to load notification sound:', error)
    }
}

//*new* Admin Dropdown

const adminMenu = document.getElementById('adminMenu');
const adminMenuBtn = document.getElementById('adminMenuBtn');
const adminDropdown = document.getElementById('adminDropdown');
const inviteModal = document.getElementById('inviteModal');
const inviteLinkInput = document.getElementById('inviteLinkInput');
const copyInviteBtn = document.getElementById('copyInviteBtn');
const closeModalBtn = document.getElementById('closeModalBtn');
const newBenchBtn = document.getElementById('newBenchBtn');
const deleteBenchBtn = document.getElementById('deleteBenchBtn');
const inviteBtn = document.getElementById('inviteBtn');
const makePublicBtn = document.getElementById('makePublicBtn');

//togglin and boppin
adminMenuBtn.addEventListener('click', (e) => {
    e.stopPropagation(); //this prevents the document click handler from immediately closing it down (SENDING IT TO JAIL)
    adminDropdown.classList.toggle('hidden');
});

//when user clicks outside of it, then it closes c:
document.addEventListener('click', () => {
    adminDropdown.classList.add('hidden');
});

//Now it's going to build the shareable URL from the current room's id and key
//enc.key and room id are already in state + "buildRoomURL" (in crypto.js) builds the URL without admin token sh*t. 
// S U C C E S S
inviteBtn.addEventListener('click', () => {
    adminDropdown.classList.add('hidden');
    const shareableURL = buildRoomURL(roomId, encryptionKey);
    inviteLinkInput.value = shareableURL;
    inviteModal.classList.remove('hidden');
});

//copy *p l a i n* invite link
copyInviteBtn.addEventListener('click', async () => {
    try {
        await navigator.clipboard.writeText(inviteLinkInput.value);
        copyInviteBtn.textContent = 'Copied!';
        setTimeout(() => { copyInviteBtn.textContent = 'Copy';}, 2000);
    } catch {
        inviteLinkInput.select();
    }
});

//closing modal
closeModalBtn.addEventListener('click', () => {
    inviteModal.classList.add('hidden');
});

//closing modal too, but this time when clicking the overlay bg
inviteModal.addEventListener('click', (e) => {
    //closing if only they clicked the dark overlay, not the box inside
    if (e.target === inviteModal) {
        inviteModal.classList.add('hidden');
    }
});

//this creates a new bench and opens it in a new tab
newBenchBtn.addEventListener('click', async () => {
    adminDropdown.classList.add('hidden');

    //generate everything fresh - same as index.html "createRoom"
    //opening result in a new tab c:
    const newRoomId = generateId(16);
    const newKey = generateEncryptionKey();
    const newAdminToken = generateAdminToken();
    const newAdminURL = buildAdminURL(newRoomId, newKey, newAdminToken);

    window.open(newAdminURL, '_blank');
});

//deleting the bench
deleteBenchBtn.addEventListener('click', () => {
    adminDropdown.classList.add('hidden');

    if (!confirm('Are you sure you want to delete this bench? Everyone will be kicked out.')) {
        return;
    }

    ws.send(JSON.stringify({
        type: 'delete',
        roomId: roomId,
        payload: adminToken
    }));
});

// making the room public — one-way, button hides itself after success
// sends the raw admin token as payload; server hashes and verifies it
// same pattern as delete, just a different message type
makePublicBtn.addEventListener('click', () => {
    adminDropdown.classList.add('hidden');

    if (!confirm('Make this room public? Messages will be visible across all connected benches. This cannot be undone.')) {
        return;
    }

    ws.send(JSON.stringify({
        type: 'make_public',
        roomId: roomId,
        payload: adminToken
    }));
});

//doms

const loadingState = document.getElementById('loadingState');
const errorState = document.getElementById('errorState');
const errorText = document.getElementById('errorText');
const roomInterface = document.getElementById('roomInterface');
const messagesContainer = document.getElementById('messagesContainer');
const messageInput = document.getElementById('messageInput');
const sendBtn = document.getElementById('sendBtn');
const userCount = document.getElementById('userCount');
const attachBtn = document.getElementById('attachBtn');
const imageInput = document.getElementById('imageInput');

//room start

function init() {

    //read URL
    const urlData = parseRoomURL();

    if (!urlData.roomId || !urlData.encryptionKey) {
        showError('Invalid room link. Missing room ID or encryption key.');
        return;
    }

    roomId = urlData.roomId;
    encryptionKey = urlData.encryptionKey;
    adminToken = urlData.adminToken; 
    pairToken = urlData.pairToken;
    isPairingURL = pairToken !== null && urlData.claimer !== null;

    if (isPairingURL && !adminToken) {
        showError('This pairing link requires an admin token. Make sure you opened the full pairing URL.');
        return;
    }

    if (adminToken) {
        // show the admin dropdown instead of the old delete button
        adminMenu.classList.remove('hidden');
    }

    displayName = prompt('What name do they know you by?') || 'Anonymous';

    if (!displayName.trim()) {
        displayName = 'Anonymous';
    }

    initSound();

    const soundToggle = document.getElementById('soundToggle');
    if (soundToggle) {
        const savedPref = localStorage.getItem('soundEnabled');
        if (savedPref !== null) {
            soundEnabled = savedPref === 'true';
            soundToggle.checked = soundEnabled;
        }
    } //Check aspects of defensive programming in JS

    // start connection to WS
    connectWebSocket();
}

//ws connection

async function connectWebSocket() {
    // determine WS URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    ws = new WebSocket(wsUrl);

    // must be async because hashAdminToken uses await
    ws.onopen = async () => {
        console.log('WebSocket connected');

        // if this client is the admin, hash their token before sending
        // the server stores the hash — never the real token
        let adminHash = '';
        if (adminToken) {
            adminHash = await hashAdminToken(adminToken);
        }

        // join message — this is what creates the room on the server
        // if the room doesn't exist yet, the server creates it now
        // if it already exists, the client just joins
        ws.send(JSON.stringify({
            type: 'join',
            roomId: roomId,
            adminHash: adminHash,
            payload: displayName  // send username as payload
        }));

        // show the chat interface immediately after sending join
        // the server doesn't echo back to the sender so we don't
        // wait for a confirmation — connection is already established
        handleSelfJoined();
    };

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            handleWebSocketMessage(data);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        showError('Connection error. Please check your internet connection.'); //need to work on error handling
    };

    ws.onclose = () => {
        console.log('WebSocket closed');
        isConnected = false;
        addSystemMessage('Disconnected from bench');
    };
}

// how does ws handle messages?

function handleWebSocketMessage(data) {
    console.log('Received:', data);

    switch (data.type) {
        case 'join':
            // someone joined — could be us (first join) or someone else
            if (!isConnected) {
                // this is our own join confirmation
                handleSelfJoined();
            } else {
                userCount_n++;// someone else joined
                updateUserCount(userCount_n);
                addSystemMessage('Someone joined the bench');
            }
            break;

        case 'leave':
            // someone else left
            addSystemMessage('Someone left the bench');
            userCount_n = Math.max(0, userCount_n - 1);
            updateUserCount(userCount_n);
            break;

        case 'welcome':
            // Server gives info on how many people are in the room at the moment.
            // Better than counting join/leave events.
            userCount_n = parseInt(data.payload);
            updateUserCount(userCount_n);

            // If it was a pairing URL, send the claim now that we've confirmed our seat on this bench
            if (isPairingURL) {
                ws.send(JSON.stringify({
                    type:   'claim_pair',
                    roomId: roomId,
                    payload: pairToken, //raw token from URL
                }))

                // Cleaning the pair token from the URL immidiately after sending.
                // It's single-use - leaving it in the URL invites accidental
                // reuse or sharing. The key and admin stay in the fragment.
                const cleanURL = window.location.pathname + '?room=' + roomId + window.location.hash
                window.history.replaceState(null,'', cleanURL)
            }
            break;

        case 'message':
            handleMessage(data);
            break;

        case 'image':
            handleImage(data);
            break;

        case 'deleted':
            // Admin deleted the room — everyone gets this.
            handleRoomDeleted();
            break;

        case 'made_public':
            // Admin made the room public — everyone gets this including the admin.
            // Hide the button so it can't be clicked twice (it's one-way anyway).
            makePublicBtn.classList.add('hidden');
            addSystemMessage('🌐 This bench is now public — messages are visible across the park');
            break;

        

        case 'error':
            // Server sent us an error (e.g. invalid admin token).
            console.error('Server error:', data.payload);
            alert(data.payload);
            break;
        
        case 'pair_token':
        
            const [rawToken, claimerID] = data.payload.split('|')
            
            const pairingURL = window.location.origin + '/room.html' + '?room=' + roomId + '&pair=' + rawToken + '&claimer=' + claimerID + window.location.hash

            document.getElementById('inviteLinkInput').value = pairingURL
            document.getElementById('inviteModal').classList.remove('hidden')

            addSystemMessage('Pairing link generated - valid for 5 minutes, single use only.')
            break;

        case 'pair_claim_sent':
            addSystemMessage('Pairing request sent, waiting for approval...')
            break;
        
        case 'pair_approved':
            addSystemMessage('Pairing approved! This bench is now federated with bench' + data.payload)
            break;
        
        case 'pair_rejected':
            addSystemMessage('Pairing rejected. Token may be expired or invalid.')
            break;

        default:
            console.log('Unknown message type:', data.type);
    }
}

function handleSelfJoined() {
    // our join was acknowledged — show the chat interface
    isConnected = true;
    loadingState.classList.add('hidden');
    roomInterface.classList.remove('hidden');
    addSystemMessage('You joined the bench');
    messageInput.focus();
}

function handleMessage(data) {
    // data.payload is the encrypted blob
    // data.senderId tells us who sent it
    const decrypted = decryptMessage(data.payload, encryptionKey);

    if (!decrypted) {
        console.error('Failed to decrypt message');
        return;
    }

    // play sound if this came from someone else
    // we identify ourselves by displayName inside the encrypted payload
    if (decrypted.sender !== displayName) {
        playNotificationSound();
    }

    addMessage(decrypted.sender, decrypted.text, decrypted.timestamp);
}

function handleImage(data) {
    // data.payload is the encrypted image blob (base64)
    // decrypt it directly — no HTTP download needed anymore
    const decrypted = decryptImageFromBase64(data.payload, encryptionKey);

    if (!decrypted) {
        console.error('Failed to decrypt image');
        addSystemMessage('❌ Failed to load image');
        return;
    }

    const imageUrl = URL.createObjectURL(decrypted);

    const div = document.createElement('div');
    div.className = 'message';
    div.innerHTML = `
            <div class="message-sender">Image</div>
            <img
                src="${imageUrl}"
                class="message-image"
                alt="image"
                onclick="window.open(this.src)">`;
    messagesContainer.appendChild(div);
    scrollToBottom();
    playNotificationSound();
}

function handleRoomDeleted() {
    showError('This bench has been deleted by the creator.');
    ws.close();
}

//UI helper functions

function updateUserCount(count) {
    const text = count === 1 ? '1 person here' : `${count} people here`;
    userCount.textContent = text;
}

function addSystemMessage(text) {
    const div = document.createElement('div');
    div.className = 'system-message';
    div.textContent = text;
    messagesContainer.appendChild(div);
    scrollToBottom();
}

function addMessage(sender, text, timestamp) {
    const div = document.createElement('div');
    div.className = 'message';

    const time = new Date(timestamp).toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit'
    });

    div.innerHTML = `
            <div class="message-sender">${escapeHtml(sender)}</div>
            <div class="message-content">${escapeHtml(text)}</div>
            <div class="message-time">${time}</div>
        `;

    messagesContainer.appendChild(div);
    scrollToBottom();
}

function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

function playNotificationSound() {
    if (!soundEnabled) {
        return;
    }

    if (!notificationSound) {
        console.warn('Notification sound not loaded');
        return;
    }

    if (!document.hidden) {
        return; //when the tab is active, sounds will not be played.
    }

    notificationSound.currentTime = 0; //from the start
    notificationSound.play().catch(error => {
        console.log('Sound playback blocked:', error);
    });
}

function showError(message) {
    errorText.textContent = message;
    loadingState.classList.add('hidden');
    roomInterface.classList.add('hidden');
    errorState.classList.remove('hidden');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + 'B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1)+' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// sending messages

function sendMessage() {
    const text = messageInput.value.trim();

    if (!text || !isConnected) {
        return;
    }

    // making the message object
    const messageObj = {
        sender: displayName,
        text: text,
        timestamp: Date.now()
    };

    // encryption
    const encrypted = encryptMessage(messageObj, encryptionKey);

    // send to server — server forwards to everyone else
    // field names must match Go's IncomingMessage struct exactly
    ws.send(JSON.stringify({
        type: 'message',
        roomId: roomId,
        payload: encrypted
    }));

    // show our own message immediately — server doesn't echo back to sender
    addMessage(displayName, text, messageObj.timestamp);

    // clear input
    messageInput.value = '';
    messageInput.style.height = 'auto';
}

async function sendImage(file) {
    if (!file.type.startsWith('image/')) {
        alert('Only images are allowed to send in BenchTalks (for the moment at least)!');
        return;
    }

    const MAX_SIZE = 10 * 1024 * 1024; // 10MB
    if (file.size > MAX_SIZE) {
        alert('Image too large. Maximum size is 10MB!');
        return;
    }

    try {
        // encrypt the image client-side
        const encryptedData = await encryptImage(file, encryptionKey);

        // send as WebSocket message — same pattern as text
        // server forwards the encrypted blob to everyone else
        // no /api/upload, no disk storage, server never holds the image
        ws.send(JSON.stringify({
            type: 'image',
            roomId: roomId,
            payload: encryptedData
        }));

        // show our own image immediately — server doesn't echo back to sender
        const imageBlob = await decryptImageFromBase64(encryptedData, encryptionKey, file.type);
        const imageUrl = URL.createObjectURL(imageBlob);

        const div = document.createElement('div');
        div.className = 'message';
        div.innerHTML = `
                <div class="message-sender">You</div>
                <img
                    src="${imageUrl}"
                    class="message-image"
                    alt="${escapeHtml(file.name)}"
                    onclick="window.open(this.src)">
                <div class="message-time">${file.name} (${formatFileSize(file.size)})</div>`;
        messagesContainer.appendChild(div);
        scrollToBottom();

    } catch (error) {
        console.error('Error sending image:', error);
        addSystemMessage('❌ Failed to send image');
    }
}

//event listeners

sendBtn.addEventListener('click', sendMessage);

const soundToggle = document.getElementById('soundToggle');
soundToggle.addEventListener('change', (e) => {
    soundEnabled = e.target.checked;

    localStorage.setItem('soundEnabled', soundEnabled);
    console.log('Sound notifications:', soundEnabled ? 'enabled' : 'disabled');

    if (soundEnabled && notificationSound) {
        notificationSound.currentTime = 0;
        notificationSound.play().catch(() => {});
    }
});

messageInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
    }
});

// auto-resize textarea as user types
messageInput.addEventListener('input', () => {
    messageInput.style.height = 'auto';
    messageInput.style.height = messageInput.scrollHeight + 'px';
});

attachBtn.addEventListener('click', () => {
    imageInput.click();
});

imageInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    if (file) {
        sendImage(file);
    }
    imageInput.value = '';
});

document.getElementById('pairBenchBtn').addEventListener('click', () => {

    const claimerBenchID = prompt('Enter the Bench ID of the bench you want to pair with.\n' + 
        'The other operator can find their Bench ID in their server config (BENCH_ID env var).'
    )

    if (!claimerBenchID || claimerBenchID.trim() === '') {
        return //cancellation or null entry
    }

    ws.send(JSON.stringify({
        type:   'request_pair',
        roomId:     roomId,
        payload:    adminToken,
        adminHash: claimerBenchID.trim(),
    }))
})

init();
