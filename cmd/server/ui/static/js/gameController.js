
function setCookie(name, value, days = 7) {
    const expires = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString();
    document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; path=/`;
}

function getCookie(name) {
    const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
    return match ? decodeURIComponent(match[1]) : null;
}

function eraseCookie(name) {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
}

function updateSessionDisplay() {
    const playerId = getCookie('player_id') || '';
    const gameId = getCookie('game_id') || '';
    playerIdField.value = playerId;
    gameIdField.value = gameId;
}

function clearSession() {
    ['game_id', 'token', 'player_id'].forEach(eraseCookie);
    updateSessionDisplay();
}

async function joinGame() {
    try {
        const response = await fetch('http://localhost:8080/api/join', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'token': getCookie("token")
            },
            body: JSON.stringify({})
        });

        if (!response.ok) {
            throw new Error(`Join failed: ${response.status}`);
        }

        const data = await response.json();
        if (data.game_id) {
            setCookie('game_id', data.game_id);
        }
        if (data.token) {
            setCookie('token', data.token);
        }
        if (data.player_id) {
            setCookie('player_id', data.player_id);
        }

        updateSessionDisplay();
    } catch (error) {
        console.error('Failed to join game:', error);
    }
}

const shipOptions = Array.from(document.querySelectorAll('.ship-option'));
const rotateShipBtn = document.getElementById('rotateShipBtn');
const clearSessionBtn = document.getElementById('clearSessionBtn');
const playerIdField = document.getElementById('playerIdField');
const gameIdField = document.getElementById('gameIdField');
const cells = Array.from(document.querySelectorAll('.grid .cell'));
let selectedShip = null;
let shipOrientation = 'horizontal';

updateSessionDisplay();
clearSessionBtn.addEventListener('click', clearSession);
joinGame();

shipOptions.forEach((option) => {
    option.addEventListener('dragstart', (event) => {
        event.dataTransfer.setData('text/plain', option.dataset.length);
        option.classList.add('dragging');
        selectedShip = option;
        highlightReadyCells();
    });

    option.addEventListener('dragend', () => {
        option.classList.remove('dragging');
        clearReadyCells();
    });

    option.addEventListener('click', () => {
        selectedShip = option;
        highlightReadyCells();
        shipOptions.forEach((item) => item.classList.remove('selected'));
        option.classList.add('selected');
    });
});

rotateShipBtn.addEventListener('click', () => {
    shipOrientation = shipOrientation === 'horizontal' ? 'vertical' : 'horizontal';
    rotateShipBtn.textContent = `Rotate Ship (${shipOrientation.charAt(0).toUpperCase() + shipOrientation.slice(1)})`;
    if (selectedShip) {
        highlightReadyCells();
    }
});

cells.forEach((cell) => {
    cell.addEventListener('dragover', (event) => event.preventDefault());

    cell.addEventListener('drop', (event) => {
        event.preventDefault();
        if (selectedShip) {
            placeShip(cell);
        }
    });

    cell.addEventListener('click', () => {
        if (selectedShip) {
            placeShip(cell);
        }
    });
});

function highlightReadyCells() {
    clearReadyCells();
    if (!selectedShip) {
        return;
    }

    const length = parseInt(selectedShip.dataset.length, 10);
    cells.forEach((cell) => {
        const [_, colStr, rowStr] = cell.id.split('-');
        const col = parseInt(colStr, 10);
        const row = parseInt(rowStr, 10);
        const canPlace = shipOrientation === 'horizontal'
            ? col <= 10 - length
            : row <= 10 - length;

        if (canPlace) {
            cell.classList.add('ready');
        }
    });
    selectedShip.classList.add('selected');
}

function clearReadyCells() {
    cells.forEach((cell) => cell.classList.remove('ready'));
    shipOptions.forEach((item) => item.classList.remove('selected'));
}

function placeShip(startCell) {
    if (!selectedShip) {
        return;
    }

    const length = parseInt(selectedShip.dataset.length, 10);
    const [_, colStr, rowStr] = startCell.id.split('-');
    const startCol = parseInt(colStr, 10);
    const startRow = parseInt(rowStr, 10);

    const positions = [];
    for (let offset = 0; offset < length; offset += 1) {
        const col = shipOrientation === 'horizontal' ? startCol + offset : startCol;
        const row = shipOrientation === 'horizontal' ? startRow : startRow + offset;
        const cellId = `cell-${col}-${row}`;
        const targetCell = document.getElementById(cellId);

        if (!targetCell || targetCell.classList.contains('placed')) {
            return;
        }

        positions.push(targetCell);
    }

    positions.forEach((cell) => {
        cell.classList.add('placed');
        cell.classList.add(`ship-${selectedShip.dataset.color}`);
    });

    const shipType = selectedShip.dataset.color;
    const coordinates = positions.map((cell) => {
        const parts = cell.id.split('-');
        return [parseInt(parts[2], 10), parseInt(parts[1], 10)];
    });
    const payload = {
        ship: shipType,
        positions: coordinates
    };

    fetch('http://localhost:8080/api/placement', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'token': getCookie('token')
        },
        body: JSON.stringify(payload)
    }).catch((error) => {
        console.error('Failed to submit placement:', error);
    });

    selectedShip.remove();
    selectedShip = null;
    clearReadyCells();
}
