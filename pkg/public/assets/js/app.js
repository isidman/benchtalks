//DOM
const createSection = document.getElementById('createSection');
const loadingSection = document.getElementById('loadingSection');
const errorSection = document.getElementById('errorSection');
const createRoomBtn = document.getElementById('createRoomBtn');
const retryBtn = document.getElementById('retryBtn');
const errorMessage = document.getElementById('errorMessage');

//states
let currentRoomId = null;
let currentShareableURL = null;
let currentAdminURL = null;

//Throttle - max 2 calls/sec
let lastCreateTime = 0;
const CREATE_THROTTLE_MS = 500;

function isThrottled() {
    const now = Date.now();
    if (now - lastCreateTime < CREATE_THROTTLE_MS) return true;
    lastCreateTime = now;
    return false;
}

//ui helpers
function showSection(section) {
    //hide all sections
    createSection.classList.add('hidden');
    loadingSection.classList.add('hidden');
    errorSection.classList.add('hidden');

    //show requested
    section.classList.remove('hidden');
}

function showError(message) {
    errorMessage.textContent = message;
    showSection(errorSection);
}

// room creation
async function createRoom() {
    try {
        showSection(loadingSection); // show the loading state
        const roomId = generateId(16); // 32 chars hex string
        const encryptionKey = generateEncryptionKey(); // 32 random bytes
        const adminToken = generateAdminToken(); // base64 32 random bytes

        const keyString = keyToString(encryptionKey);
        const adminTokenHash = await hashAdminToken(adminToken);

        // No POST to server needed anymore.
        // The room is created automatically when the first WebSocket
        // connection arrives in room.html with a "join" message.
        // The adminTokenHash travels with that first join, not here.

        // Build URLs
        let shareableURL;
        let adminURL;
        try {
            shareableURL = buildRoomURL(roomId, encryptionKey);
            adminURL = buildAdminURL(roomId, encryptionKey, adminToken);
        } catch (urlError) {
            console.error('Error building URLs:', urlError);
            throw urlError;
        }

        // store the state
        currentRoomId = roomId;
        currentShareableURL = shareableURL;
        currentAdminURL = adminURL;

        // open the admin URL directly — the room gets created when
        // the WebSocket join message arrives in room.html, not here.
        // We redirect immediately so the admin lands in the room straight away.
        window.location.href = adminURL;

    } catch (error) {
        console.error('Error creating room', error);
        showError('Failed to create room. Please check your internet connection and try again.');
    }
}

//helper function for random id generation
function generateId(length) {
    //Generate random bytes
    const bytes = nacl.randomBytes(length);

    //Conver to hex string
    let hex = '';
    for (let i = 0; i < bytes.length; i++) {
        hex += bytes[i].toString(16).padStart(2, '0');
    }
    return hex;
}
//clippy copy desu.
async function copyToClipboard(text, button) {
    try {
        await navigator.clipboard.writeText(text);

        const originalText = button.textContent;
        button.textContent = 'Copied!';
        button.classList.add('success');

        setTimeout(() => {
            button.textContent = originalText;
            button.classList.remove('success');
        }, 2000);
    }catch (error) {
        console.error('Failed to copy:', error); //fallback to that: text selection for user to copy manually
        const input = button.previousElementSibling;
        input.select();
        alert('Please copy the link manually (Ctrl+C or Cmd+C)');

    }
}

//event listeners
createRoomBtn.addEventListener('click', () => {
    if (isThrottled()) return;
    createRoom();
});

retryBtn.addEventListener('click', () => {
    showSection(createSection);
});

//first thing user sees is the creation section by default
showSection(createSection);
