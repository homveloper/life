// Life Game - Trainer Movement Client
// Real-time character movement synchronization with SSE

// Global state
let authToken = '';
let serverInfo = {
    host: 'localhost',
    port: '8080',
    url: 'http://localhost:8080'
};
let trainerState = {
    position: { x: 15.0, y: 10.0 },
    movement: {
        direction: { x: 0, y: 0 },
        speed: 5.0, // Will be updated from server
        start_time: null,
        start_pos: { x: 15.0, y: 10.0 },
        is_moving: false
    }
};

// Position sync interval
let positionInterval = null;
let interpolationAnimationId = null;

// SSE connection for real-time updates
let eventSource = null;
let otherTrainers = new Map(); // Track other trainers: userId -> {position, movement, nickname}

// Current movement state tracking
let currentDirection = { x: 0, y: 0 };
let isCurrentlyMoving = false;

// Movement update debouncing
let movementUpdateTimeout = null;
const MOVEMENT_UPDATE_DELAY = 100; // 100ms debounce for smoother diagonal movement and reduced server load

// Pressed keys tracking
let pressedKeys = new Set();
const MOVEMENT_KEYS = {
    'ArrowUp': { x: 0, y: -1 },
    'ArrowDown': { x: 0, y: 1 },
    'ArrowLeft': { x: -1, y: 0 },
    'ArrowRight': { x: 1, y: 0 },
    'w': { x: 0, y: -1 },
    's': { x: 0, y: 1 },
    'a': { x: -1, y: 0 },
    'd': { x: 1, y: 0 }
};

// Build API URL using current server info
function buildApiUrl(endpoint) {
    return `${serverInfo.url}${endpoint}`;
}

// Fetch server info dynamically
async function fetchServerInfo() {
    try {
        const response = await fetch('/api/v1/server.Info', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'server.Info',
                params: {},
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            console.error('Failed to fetch server info:', result.error);
            return false;
        }
        
        // Update server info
        serverInfo = result.result;
        console.log('Server info updated:', serverInfo);
        return true;
        
    } catch (error) {
        console.error('Network error fetching server info:', error);
        return false;
    }
}

// Authentication
async function guestLogin() {
    const nickname = document.getElementById('nickname').value.trim();
    
    if (!nickname || nickname.length < 3 || nickname.length > 20) {
        showMessage('Please enter a nickname (3-20 characters)', 'error');
        return;
    }
    
    // Fetch server info first
    const serverInfoFetched = await fetchServerInfo();
    if (!serverInfoFetched) {
        showMessage('Failed to connect to server', 'error');
        return;
    }
    
    // Generate a device ID (or use stored one)
    let deviceId = localStorage.getItem('life-game-device-id');
    if (!deviceId) {
        deviceId = 'device-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
        localStorage.setItem('life-game-device-id', deviceId);
    }
    
    const loginBtn = document.getElementById('login-btn');
    loginBtn.disabled = true;
    loginBtn.textContent = 'Logging in...';
    
    try {
        const response = await fetch(buildApiUrl('/api/v1/auth.GuestLogin'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'auth.GuestLogin',
                params: {
                    device_id: deviceId
                },
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            showMessage(`Login failed: ${result.error.message}`, 'error');
            loginBtn.disabled = false;
            loginBtn.textContent = 'Start Game as Guest';
            return;
        }
        
        // Store auth token
        authToken = result.result.jwt_token;
        const userID = result.result.user_id;
        
        // Store user info
        localStorage.setItem('life-game-jwt', authToken);
        localStorage.setItem('life-game-user-id', userID);
        localStorage.setItem('life-game-nickname', nickname);
        
        // Switch to game UI
        document.getElementById('auth-section').style.display = 'none';
        document.getElementById('game-area').style.display = 'flex';
        
        // Reset trainer state to default before fetching
        resetTrainerState();
        
        // Fetch initial trainer data from server
        await fetchInitialTrainerData();
        
        // Start position sync and SSE connection
        startPositionSync();
        connectSSE();
        
        showMessage(`Welcome ${nickname}! You are logged in as guest.`, 'success');
        
    } catch (error) {
        showMessage(`Network error: ${error.message}`, 'error');
        loginBtn.disabled = false;
        loginBtn.textContent = 'Start Game as Guest';
    }
}

// Movement functions - Event-based with client prediction
async function move(dirX, dirY) {
    if (!authToken) {
        showMessage('Please authenticate first', 'error');
        return;
    }
    
    // Check if direction actually changed
    if (currentDirection.x === dirX && currentDirection.y === dirY && isCurrentlyMoving) {
        console.log('Same direction, skipping API call');
        return;
    }
    
    // Update current direction tracking
    currentDirection = { x: dirX, y: dirY };
    isCurrentlyMoving = true;
    
    // CLIENT PREDICTION: Immediately start moving visually
    const now = new Date();
    
    // If we were already moving, use current interpolated position as new start position
    let newStartPos = { ...trainerState.position };
    if (trainerState.movement.is_moving) {
        // Calculate current interpolated position to maintain smooth transition
        const currentTime = now;
        const startTime = new Date(trainerState.movement.start_time);
        const elapsed = Math.max(0, (currentTime - startTime) / 1000); // seconds
        const speed = trainerState.movement.speed || 5.0;
        const distance = speed * elapsed;
        
        // Validate direction values to prevent NaN
        const dirX = (typeof trainerState.movement.direction.x === 'number' && !isNaN(trainerState.movement.direction.x)) 
            ? trainerState.movement.direction.x : 0;
        const dirY = (typeof trainerState.movement.direction.y === 'number' && !isNaN(trainerState.movement.direction.y)) 
            ? trainerState.movement.direction.y : 0;
        
        newStartPos = {
            x: trainerState.movement.start_pos.x + (dirX * distance),
            y: trainerState.movement.start_pos.y + (dirY * distance)
        };
        
        // Validate calculated position to prevent NaN
        if (isNaN(newStartPos.x) || isNaN(newStartPos.y)) {
            console.warn('Invalid interpolated position calculated, using current position');
            newStartPos = { ...trainerState.position };
        } else {
            // Update position to the interpolated position for smooth transition
            trainerState.position = { ...newStartPos };
        }
    }
    
    trainerState.movement = {
        direction: { x: dirX, y: dirY },
        speed: trainerState.movement.speed || 5.0,
        start_time: now.toISOString(),
        start_pos: newStartPos,
        is_moving: true
    };
    
    // Debug logging for diagonal movement
    if (dirX !== 0 && dirY !== 0) {
        console.log('Starting diagonal movement:', {
            direction: { x: dirX, y: dirY },
            start_pos: newStartPos,
            current_pos: trainerState.position
        });
    }
    updateUI(); // Immediate visual feedback
    
    try {
        const response = await fetch(buildApiUrl('/api/v1/trainer.Move'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'trainer.Move',
                params: {
                    direction_x: dirX,
                    direction_y: dirY,
                    action: 'start'
                },
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            showMessage(`Error: ${result.error.message}`, 'error');
            // If authentication error, force logout
            if (result.error.code === -32602 || result.error.message.includes('authenticated')) {
                console.log('Authentication error, logging out...');
                logout();
            }
            return;
        }
        
        // Apply JSON merge patch
        applyChanges(result.result.changes);
        showMessage(`Started moving ${dirX},${dirY}`, 'success');
        
    } catch (error) {
        showMessage(`Network error: ${error.message}`, 'error');
    }
}

async function stopMovement() {
    if (!authToken) {
        showMessage('Please authenticate first', 'error');
        return;
    }
    
    // Check if already stopped
    if (!isCurrentlyMoving && currentDirection.x === 0 && currentDirection.y === 0) {
        console.log('Already stopped, skipping API call');
        return;
    }
    
    // Update current direction tracking
    currentDirection = { x: 0, y: 0 };
    isCurrentlyMoving = false;
    
    // CLIENT PREDICTION: Immediately stop moving visually
    trainerState.movement.direction = { x: 0, y: 0 };
    trainerState.movement.is_moving = false;
    updateUI(); // Immediate visual feedback
    
    try {
        const response = await fetch(buildApiUrl('/api/v1/trainer.Move'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'trainer.Move',
                params: {
                    direction_x: 0,
                    direction_y: 0,
                    action: 'stop'
                },
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            showMessage(`Error: ${result.error.message}`, 'error');
            // If authentication error, force logout
            if (result.error.code === -32602 || result.error.message.includes('authenticated')) {
                console.log('Authentication error, logging out...');
                logout();
            }
            return;
        }
        
        // Apply JSON merge patch
        applyChanges(result.result.changes);
        showMessage('Movement stopped', 'success');
        
    } catch (error) {
        showMessage(`Network error: ${error.message}`, 'error');
    }
}

async function fetchPosition() {
    if (!authToken) {
        showMessage('Please authenticate first', 'error');
        return;
    }
    
    try {
        const response = await fetch(buildApiUrl('/api/v1/trainer.FetchPosition'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'trainer.FetchPosition',
                params: {},
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            showMessage(`Error: ${result.error.message}`, 'error');
            // If authentication error, force logout
            if (result.error.code === -32602 || result.error.message.includes('authenticated')) {
                console.log('Authentication error, logging out...');
                logout();
            }
            return;
        }
        
        // Full state sync with validation
        const syncData = result.result;
        console.log('Sync data from server:', syncData);
        
        if (syncData.position && 
            typeof syncData.position.x === 'number' && !isNaN(syncData.position.x) &&
            typeof syncData.position.y === 'number' && !isNaN(syncData.position.y)) {
            trainerState.position = syncData.position;
        } else {
            console.warn('Invalid position in sync data:', syncData.position);
        }
        
        if (syncData.movement) {
            // Don't overwrite movement state if we're actively moving via keyboard
            if (pressedKeys.size === 0 || !isCurrentlyMoving) {
                trainerState.movement = syncData.movement;
                // Validate direction, handling both lowercase and uppercase field names
                if (!trainerState.movement.direction) {
                    trainerState.movement.direction = { x: 0, y: 0 };
                } else {
                    // Handle both {x,y} and {X,Y} field names for backward compatibility
                    const dirX = trainerState.movement.direction.x !== undefined ? trainerState.movement.direction.x : trainerState.movement.direction.X;
                    const dirY = trainerState.movement.direction.y !== undefined ? trainerState.movement.direction.y : trainerState.movement.direction.Y;
                    
                    if (isNaN(dirX) || isNaN(dirY)) {
                        trainerState.movement.direction = { x: 0, y: 0 };
                    } else {
                        // Normalize to lowercase field names
                        trainerState.movement.direction = { x: dirX, y: dirY };
                    }
                }
            } else {
                // Only update non-conflicting movement data (like speed) while preserving client prediction
                if (syncData.movement.speed) {
                    trainerState.movement.speed = syncData.movement.speed;
                }
                console.log('Preserving client movement state during keyboard input');
            }
        }
        
        updateUI();
        showMessage('Position synchronized', 'success');
        
    } catch (error) {
        showMessage(`Network error: ${error.message}`, 'error');
    }
}

// JSON merge patch application using standard library
function applyChanges(changes) {
    if (!changes) return;
    
    console.log('Applying changes:', changes);
    
    // Don't apply movement changes if we're actively moving via keyboard
    if (changes.movement && (pressedKeys.size > 0 && isCurrentlyMoving)) {
        // Only update non-conflicting movement data while preserving client prediction
        const filteredChanges = { ...changes };
        delete filteredChanges.movement; // Remove movement to preserve client prediction
        
        // But allow specific movement fields that don't conflict
        if (changes.movement.speed) {
            if (!filteredChanges.movement) filteredChanges.movement = {};
            filteredChanges.movement.speed = changes.movement.speed;
        }
        if (changes.movement.start_pos) {
            if (!filteredChanges.movement) filteredChanges.movement = {};
            filteredChanges.movement.start_pos = changes.movement.start_pos;
        }
        
        console.log('Preserving client movement state during keyboard input, filtered changes:', filteredChanges);
        changes = filteredChanges;
    }
    
    // Use fast-json-patch library for JSON patch operations
    if (window.jsonpatch) {
        try {
            // Create a deep copy of current state
            let newState = JSON.parse(JSON.stringify(trainerState));
            
            // Apply changes using object merge (similar to JSON merge patch)
            function deepMerge(target, source) {
                for (const key in source) {
                    if (source[key] === null) {
                        delete target[key];
                    } else if (typeof source[key] === 'object' && !Array.isArray(source[key]) && source[key] !== null) {
                        target[key] = target[key] || {};
                        deepMerge(target[key], source[key]);
                    } else {
                        target[key] = source[key];
                    }
                }
                return target;
            }
            
            // Apply merge patch logic
            trainerState = deepMerge(newState, changes);
            
            // Validate critical fields after merge
            if (!trainerState.position || isNaN(trainerState.position.x) || isNaN(trainerState.position.y)) {
                console.warn('Invalid position after merge patch, resetting:', trainerState.position);
                trainerState.position = { x: 15.0, y: 10.0 };
            }
            
            if (!trainerState.movement) {
                trainerState.movement = {
                    direction: { x: 0, y: 0 },
                    speed: 5.0,
                    start_time: null,
                    start_pos: trainerState.position,
                    is_moving: false
                };
            } else if (!trainerState.movement.direction) {
                trainerState.movement.direction = { x: 0, y: 0 };
            } else {
                // Handle case sensitivity issues - normalize to lowercase
                if (trainerState.movement.direction.X !== undefined) {
                    trainerState.movement.direction.x = trainerState.movement.direction.X;
                    delete trainerState.movement.direction.X;
                }
                if (trainerState.movement.direction.Y !== undefined) {
                    trainerState.movement.direction.y = trainerState.movement.direction.Y;
                    delete trainerState.movement.direction.Y;
                }
                
                // Validate direction values
                if (isNaN(trainerState.movement.direction.x) || isNaN(trainerState.movement.direction.y)) {
                    console.warn('Invalid direction values after merge, resetting:', trainerState.movement.direction);
                    trainerState.movement.direction = { x: 0, y: 0 };
                }
            }
            
            // Debug log for position issues
            if (trainerState.position.x === 0 || trainerState.position.y === 0) {
                console.warn('Position set to X,0 or 0,Y after merge patch:', {
                    position: trainerState.position,
                    changes: changes
                });
            }
            
        } catch (error) {
            console.error('JSON merge patch failed:', error, 'Changes:', changes);
            // Fallback to manual merging if library fails
            fallbackApplyChanges(changes);
        }
    } else {
        console.warn('fast-json-patch library not loaded, using fallback');
        fallbackApplyChanges(changes);
    }
    
    updateUI();
}

// Fallback manual merge function  
function fallbackApplyChanges(changes) {
    // Apply position changes
    if (changes.position) {
        trainerState.position = { ...trainerState.position, ...changes.position };
    }
    
    // Apply movement changes
    if (changes.movement) {
        trainerState.movement = { ...trainerState.movement, ...changes.movement };
        if (changes.movement.direction) {
            trainerState.movement.direction = { 
                ...trainerState.movement.direction, 
                ...changes.movement.direction 
            };
        }
    }
    
    // Apply other changes
    Object.keys(changes).forEach(key => {
        if (key !== 'position' && key !== 'movement') {
            trainerState[key] = changes[key];
        }
    });
}

// UI Updates with camera following
function updateUI() {
    const trainer = document.getElementById('trainer');
    const mapContainer = trainer.parentElement;
    
    
    // Map container dimensions
    const mapWidth = 600;
    const mapHeight = 400;
    
    // Camera follows player - keep player in center of screen
    const centerX = mapWidth / 2;
    const centerY = mapHeight / 2;
    
    // Position trainer at center of visible area
    trainer.style.left = centerX + 'px';
    trainer.style.top = centerY + 'px';
    trainer.dataset.x = trainerState.position.x.toFixed(2);
    trainer.dataset.y = trainerState.position.y.toFixed(2);
    
    // Move the map background to simulate camera movement
    // Calculate offset to center the player's world position
    const worldOffsetX = -trainerState.position.x * 20; // Scale factor for visual movement
    const worldOffsetY = -trainerState.position.y * 20;
    
    // Apply background offset to simulate world movement
    mapContainer.style.backgroundPosition = `${worldOffsetX + centerX}px ${worldOffsetY + centerY}px`;
    
    
    // Update movement visual state
    if (trainerState.movement.is_moving) {
        trainer.classList.add('moving');
    } else {
        trainer.classList.remove('moving');
    }
    
    // Update status display with NaN protection
    document.getElementById('pos-x').textContent = 
        (typeof trainerState.position.x === 'number' && !isNaN(trainerState.position.x)) 
            ? trainerState.position.x.toFixed(2) : '0.00';
    document.getElementById('pos-y').textContent = 
        (typeof trainerState.position.y === 'number' && !isNaN(trainerState.position.y)) 
            ? trainerState.position.y.toFixed(2) : '0.00';
    document.getElementById('is-moving').textContent = trainerState.movement.is_moving;
    
    // Safe direction display
    const dirX = (typeof trainerState.movement.direction.x === 'number' && !isNaN(trainerState.movement.direction.x)) 
        ? trainerState.movement.direction.x : 0;
    const dirY = (typeof trainerState.movement.direction.y === 'number' && !isNaN(trainerState.movement.direction.y)) 
        ? trainerState.movement.direction.y : 0;
    document.getElementById('direction').textContent = `${dirX}, ${dirY}`;
    
    document.getElementById('speed').textContent = trainerState.movement.speed || 5.0;
    document.getElementById('last-update').textContent = new Date().toLocaleTimeString();
    
    // Update other trainers display when our position changes (for camera following)
    updateOtherTrainersDisplay();
}

// Position interpolation for smooth movement
function interpolatePosition() {
    if (!trainerState.movement.is_moving) return;
    
    const now = new Date();
    const startTime = new Date(trainerState.movement.start_time);
    const elapsed = (now - startTime) / 1000; // seconds
    
    const speed = trainerState.movement.speed || 5.0;
    const distance = speed * elapsed;
    
    // Validate direction values to prevent NaN
    const dirX = (typeof trainerState.movement.direction.x === 'number' && !isNaN(trainerState.movement.direction.x)) 
        ? trainerState.movement.direction.x : 0;
    const dirY = (typeof trainerState.movement.direction.y === 'number' && !isNaN(trainerState.movement.direction.y)) 
        ? trainerState.movement.direction.y : 0;
    
    const newX = trainerState.movement.start_pos.x + (dirX * distance);
    const newY = trainerState.movement.start_pos.y + (dirY * distance);
    
    // Validate calculated position to prevent NaN
    if (!isNaN(newX) && !isNaN(newY)) {
        trainerState.position.x = newX;
        trainerState.position.y = newY;
    } else {
        console.warn('Invalid position calculated during interpolation, stopping movement');
        // Stop movement if invalid position is calculated
        trainerState.movement.is_moving = false;
        trainerState.movement.direction = { x: 0, y: 0 };
    }
    
    updateUI();
}

// Position sync management
function startPositionSync() {
    // Start smooth interpolation using requestAnimationFrame
    startInterpolationLoop();
    
    // Sync with server periodically (less frequent)
    if (positionInterval) clearInterval(positionInterval);
    positionInterval = setInterval(fetchPosition, 2000); // Every 2 seconds
}

// Smooth interpolation loop using requestAnimationFrame
function startInterpolationLoop() {
    function interpolationLoop() {
        interpolatePosition();
        interpolationAnimationId = requestAnimationFrame(interpolationLoop);
    }
    
    if (interpolationAnimationId) {
        cancelAnimationFrame(interpolationAnimationId);
    }
    interpolationAnimationId = requestAnimationFrame(interpolationLoop);
}

function stopPositionSync() {
    if (positionInterval) {
        clearInterval(positionInterval);
        positionInterval = null;
    }
    if (interpolationAnimationId) {
        cancelAnimationFrame(interpolationAnimationId);
        interpolationAnimationId = null;
    }
}

function showMessage(message, type = 'error') {
    const errorDiv = document.getElementById('error-msg');
    errorDiv.innerHTML = `<div class="${type}">${message}</div>`;
    setTimeout(() => {
        errorDiv.innerHTML = '';
    }, 3000);
}

// Handle enter key in nickname input
function handleEnterKey(event) {
    if (event.key === 'Enter') {
        guestLogin();
    }
}

// Calculate movement direction from pressed keys
function calculateMovementDirection() {
    let dirX = 0;
    let dirY = 0;
    
    for (const key of pressedKeys) {
        if (MOVEMENT_KEYS[key]) {
            dirX += MOVEMENT_KEYS[key].x;
            dirY += MOVEMENT_KEYS[key].y;
        }
    }
    
    // Normalize diagonal movement (clamp to -1, 0, 1)
    dirX = Math.max(-1, Math.min(1, dirX));
    dirY = Math.max(-1, Math.min(1, dirY));
    
    return { x: dirX, y: dirY };
}

// Schedule a debounced movement update to handle rapid key combinations
function scheduleMovementUpdate() {
    // Clear any existing timeout
    if (movementUpdateTimeout) {
        clearTimeout(movementUpdateTimeout);
    }
    
    // Schedule new update with debounce
    movementUpdateTimeout = setTimeout(() => {
        const newDirection = calculateMovementDirection();
        
        // Only send movement command if direction actually changed
        if (newDirection.x !== currentDirection.x || newDirection.y !== currentDirection.y) {
            if (newDirection.x === 0 && newDirection.y === 0) {
                stopMovement();
            } else {
                move(newDirection.x, newDirection.y);
            }
        }
        
        movementUpdateTimeout = null;
    }, MOVEMENT_UPDATE_DELAY);
}

// Handle key press events
document.addEventListener('keydown', (e) => {
    if (!authToken) return;
    
    // Handle special keys
    if (e.key === ' ' || e.key === 'Escape') {
        e.preventDefault();
        pressedKeys.clear();
        stopMovement();
        return;
    }
    
    // Handle movement keys
    if (MOVEMENT_KEYS[e.key]) {
        e.preventDefault();
        
        // Prevent key repeat
        if (pressedKeys.has(e.key)) return;
        
        pressedKeys.add(e.key);
        
        // Use debounced movement update to handle rapid key combinations
        scheduleMovementUpdate();
    }
});

// Handle key release events
document.addEventListener('keyup', (e) => {
    if (!authToken) return;
    
    if (MOVEMENT_KEYS[e.key]) {
        e.preventDefault();
        
        pressedKeys.delete(e.key);
        
        // Use debounced movement update to handle rapid key combinations
        scheduleMovementUpdate();
    }
});

// Handle window focus loss (prevent stuck keys)
window.addEventListener('blur', () => {
    if (pressedKeys.size > 0) {
        pressedKeys.clear();
        
        // Clear any pending movement updates
        if (movementUpdateTimeout) {
            clearTimeout(movementUpdateTimeout);
            movementUpdateTimeout = null;
        }
        
        if (authToken) {
            stopMovement();
        }
    }
});

// Initial setup
window.addEventListener('load', async () => {
    updateUI();
    
    // First fetch server info to get current server details
    const serverInfoFetched = await fetchServerInfo();
    if (!serverInfoFetched) {
        showMessage('Failed to connect to server', 'error');
        return;
    }
    
    // Check for existing session
    const storedToken = localStorage.getItem('life-game-jwt');
    const storedNickname = localStorage.getItem('life-game-nickname');
    
    if (storedToken && storedNickname) {
        // Auto-login with stored session
        authToken = storedToken;
        document.getElementById('nickname').value = storedNickname;
        
        document.getElementById('auth-section').style.display = 'none';
        document.getElementById('game-area').style.display = 'flex';
        
        // Reset and fetch fresh trainer data
        resetTrainerState();
        fetchInitialTrainerData().then(() => {
            startPositionSync();
            connectSSE(); // Connect to real-time updates
            showMessage(`Welcome back ${storedNickname}!`, 'success');
        }).catch(error => {
            console.error('Failed to fetch trainer data:', error);
            // Force logout on error
            logout();
        });
    }
});

// Logout function
function logout() {
    // Clear stored data
    localStorage.removeItem('life-game-jwt');
    localStorage.removeItem('life-game-user-id');
    localStorage.removeItem('life-game-nickname');
    
    // Clear auth token
    authToken = '';
    
    // Stop position sync and SSE connection
    stopPositionSync();
    disconnectSSE();
    
    // Clear other trainers data
    otherTrainers.clear();
    
    // Reset direction tracking
    currentDirection = { x: 0, y: 0 };
    isCurrentlyMoving = false;
    
    // Clear pressed keys
    pressedKeys.clear();
    
    // Reset UI
    document.getElementById('auth-section').style.display = 'block';
    document.getElementById('game-area').style.display = 'none';
    document.getElementById('nickname').value = '';
    
    // Reset trainer state completely
    resetTrainerState();
    updateUI();
    showMessage('Logged out successfully', 'success');
}

// Reset trainer state to default values
function resetTrainerState() {
    trainerState = {
        position: { x: 15.0, y: 10.0 },
        movement: {
            direction: { x: 0, y: 0 },
            speed: 5.0, // Will be updated from server
            start_time: null,
            start_pos: { x: 15.0, y: 10.0 },
            is_moving: false
        }
    };
}

// Fetch initial trainer data from server
async function fetchInitialTrainerData() {
    if (!authToken) {
        showMessage('Please authenticate first', 'error');
        return;
    }
    
    try {
        const response = await fetch(buildApiUrl('/api/v1/trainer.Get'), {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'trainer.Get',
                params: {},
                id: Date.now()
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            showMessage(`Error fetching trainer: ${result.error.message}`, 'error');
            return;
        }
        
        // Update trainer state with server data
        const trainer = result.result;
        console.log('Raw trainer data from server:', trainer);
        
        // Validate and set position
        if (trainer.position && 
            typeof trainer.position.x === 'number' && !isNaN(trainer.position.x) &&
            typeof trainer.position.y === 'number' && !isNaN(trainer.position.y)) {
            trainerState.position = trainer.position;
        } else {
            console.warn('Invalid position from server, using default:', trainer.position);
            trainerState.position = { x: 15.0, y: 10.0 };
        }
        
        // Validate and set movement
        if (trainer.movement) {
            trainerState.movement = trainer.movement;
            // Ensure direction is valid, handling both lowercase and uppercase field names
            if (!trainerState.movement.direction) {
                trainerState.movement.direction = { x: 0, y: 0 };
            } else {
                // Handle both {x,y} and {X,Y} field names for backward compatibility
                const dirX = trainerState.movement.direction.x !== undefined ? trainerState.movement.direction.x : trainerState.movement.direction.X;
                const dirY = trainerState.movement.direction.y !== undefined ? trainerState.movement.direction.y : trainerState.movement.direction.Y;
                
                if (isNaN(dirX) || isNaN(dirY)) {
                    trainerState.movement.direction = { x: 0, y: 0 };
                } else {
                    // Normalize to lowercase field names
                    trainerState.movement.direction = { x: dirX, y: dirY };
                }
            }
        } else {
            trainerState.movement = {
                direction: { x: 0, y: 0 },
                speed: 5.0, // Default, will be updated from server
                start_time: null,
                start_pos: trainerState.position,
                is_moving: false
            };
        }
        
        updateUI();
        console.log('Processed trainer state:', trainerState);
        
    } catch (error) {
        showMessage(`Network error: ${error.message}`, 'error');
        throw error;
    }
}

// SSE Connection Management
function connectSSE() {
    if (!authToken) {
        console.warn('No auth token available for SSE connection');
        return;
    }
    
    if (eventSource) {
        eventSource.close();
    }
    
    // EventSource doesn't support custom headers, so we pass the token as URL parameter
    const sseUrl = `${serverInfo.url}/api/v1/stream/positions?token=${encodeURIComponent(authToken)}`;
    eventSource = new EventSource(sseUrl);
    
    eventSource.onopen = function(event) {
        console.log('SSE connection opened');
        showMessage('Real-time sync connected', 'success');
        // Reset reconnection attempts on successful connection
        window.sseReconnectAttempts = 0;
    };
    
    eventSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleSSEMessage(data);
        } catch (error) {
            console.error('Failed to parse SSE message:', error, event.data);
        }
    };
    
    eventSource.onerror = function(event) {
        console.error('SSE connection error:', event);
        
        if (eventSource.readyState === EventSource.CLOSED || eventSource.readyState === EventSource.CONNECTING) {
            showMessage('Real-time sync disconnected', 'error');
            
            // Close existing connection if still open
            if (eventSource.readyState !== EventSource.CLOSED) {
                eventSource.close();
            }
            
            // Try to reconnect with exponential backoff
            const reconnectDelay = Math.min(30000, 1000 * Math.pow(2, (window.sseReconnectAttempts || 0)));
            window.sseReconnectAttempts = (window.sseReconnectAttempts || 0) + 1;
            
            setTimeout(() => {
                if (authToken) { // Only reconnect if still authenticated
                    console.log(`Attempting SSE reconnection (attempt ${window.sseReconnectAttempts})`);
                    connectSSE();
                }
            }, reconnectDelay);
        }
    };
}

function disconnectSSE() {
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
}

// Handle incoming SSE messages (JSON-RPC 2.0 notifications and system messages)
function handleSSEMessage(notification) {
    
    // Handle system messages (non-JSON-RPC)
    if (notification.type) {
        switch (notification.type) {
            case 'connected':
                console.log('SSE connected with client ID:', notification.client_id);
                return;
            case 'heartbeat':
                console.debug('SSE heartbeat received at:', notification.timestamp);
                return;
            default:
                console.debug('Unknown SSE system message:', notification);
                return;
        }
    }
    
    // Handle JSON-RPC 2.0 notifications
    if (!notification.jsonrpc || notification.jsonrpc !== '2.0') {
        console.warn('Invalid JSON-RPC notification:', notification);
        return;
    }
    
    switch (notification.method) {
        case 'trainer.position.updated':
            handleTrainerPositionUpdate(notification.params, true); // isOwnUpdate = true
            break;
            
        case 'trainer.position.broadcast':
            handleTrainerPositionUpdate(notification.params, false); // isOwnUpdate = false
            break;
            
        case 'trainer.movement.stopped':
            handleTrainerMovementStopped(notification.params, true); // isOwnUpdate = true
            break;
            
        case 'trainer.movement.broadcast':
            handleTrainerMovementStopped(notification.params, false); // isOwnUpdate = false
            break;
            
        case 'trainer.created':
            handleTrainerCreated(notification.params);
            break;
            
        case 'connected':
            break;
            
        case 'heartbeat':
            break;
            
        default:
    }
}

// Handle trainer position updates from SSE
function handleTrainerPositionUpdate(params, isOwnUpdate) {
    if (!params.user_id) return;
    
    if (isOwnUpdate) {
        // Update our own trainer with changes from server
        if (params.changes) {
            applyChanges(params.changes);
        }
    } else {
        // Update other trainer's position for real-time sync
        const trainerId = params.user_id;
        if (trainerId !== getCurrentUserId()) { // Don't update our own position from broadcasts
            
            // Store or update other trainer info
            let otherTrainer = otherTrainers.get(trainerId) || {};
            if (params.position) otherTrainer.position = params.position;
            if (params.movement) otherTrainer.movement = params.movement;
            otherTrainers.set(trainerId, otherTrainer);
            
            // Update other trainers display
            updateOtherTrainersDisplay();
        }
    }
}

// Handle trainer movement stopped from SSE  
function handleTrainerMovementStopped(params, isOwnUpdate) {
    if (!params.user_id) return;
    
    if (isOwnUpdate) {
        // Update our own movement state
        if (params.changes) {
            applyChanges(params.changes);
        }
    } else {
        // Update other trainer's movement state
        const trainerId = params.user_id;
        if (trainerId !== getCurrentUserId()) {
            
            let otherTrainer = otherTrainers.get(trainerId) || {};
            if (params.position) otherTrainer.position = params.position;
            if (params.movement) otherTrainer.movement = params.movement;
            otherTrainers.set(trainerId, otherTrainer);
            
            updateOtherTrainersDisplay();
        }
    }
}

// Handle new trainer created
function handleTrainerCreated(params) {
    if (!params.user_id || params.user_id === getCurrentUserId()) return;
    
    showMessage(`${params.nickname || 'New trainer'} joined the game!`, 'info');
    
    // Add to other trainers
    otherTrainers.set(params.user_id, {
        position: params.position || { x: 15.0, y: 10.0 },
        movement: { is_moving: false, direction: { x: 0, y: 0 } },
        nickname: params.nickname || 'Unknown',
        level: params.level || 1
    });
    
    updateOtherTrainersDisplay();
}

// Update display of other trainers on the map with camera following
function updateOtherTrainersDisplay() {
    // Remove existing other trainer elements
    document.querySelectorAll('.other-trainer').forEach(el => el.remove());
    
    // Add other trainers to the map
    const mapContainer = document.querySelector('.map-container');
    const mapWidth = 600;
    const mapHeight = 400;
    const centerX = mapWidth / 2;
    const centerY = mapHeight / 2;
    
    otherTrainers.forEach((trainer, userId) => {
        if (trainer.position) {
            const trainerElement = document.createElement('div');
            trainerElement.className = 'other-trainer';
            trainerElement.title = `${trainer.nickname || userId} (Level ${trainer.level || 1})`;
            
            // Calculate relative position to our trainer
            const relativeX = trainer.position.x - trainerState.position.x;
            const relativeY = trainer.position.y - trainerState.position.y;
            
            // Position relative to camera center with scale factor
            const screenX = centerX + (relativeX * 20); // Scale factor for visual movement
            const screenY = centerY + (relativeY * 20);
            
            trainerElement.style.position = 'absolute';
            trainerElement.style.left = screenX + 'px';
            trainerElement.style.top = screenY + 'px';
            trainerElement.style.width = '18px';
            trainerElement.style.height = '18px';
            trainerElement.style.background = '#4444ff'; // Blue for other trainers
            trainerElement.style.border = '2px solid #fff';
            trainerElement.style.borderRadius = '50%';
            trainerElement.style.transform = 'translate(-50%, -50%)';
            trainerElement.style.zIndex = '5'; // Below own trainer
            trainerElement.style.transition = 'all 0.3s ease';
            
            if (trainer.movement && trainer.movement.is_moving) {
                trainerElement.style.background = '#6666ff';
                trainerElement.style.boxShadow = '0 0 8px rgba(68, 68, 255, 0.5)';
            }
            
            mapContainer.appendChild(trainerElement);
        }
    });
}

// Get current user ID from stored data
function getCurrentUserId() {
    return localStorage.getItem('life-game-user-id') || '';
}