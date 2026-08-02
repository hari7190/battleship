// Preload game data if we already have a game
async function getGameData() {
    if (getCookie('token')) {
        const response = await fetch('http://localhost:8080/api/playerInfo', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'token': getCookie('token')
            },
        }).catch((error) => {
            console.error('Failed to get game:', error);
        });

        const data = await response.json();
        gameState = normalizeBoardState(data);
        console.log('2. Data received:', gameState);
        rebuildBoardFromState(gameState);
    }
}

function normalizeBoardState(data) {
    if (Array.isArray(data)) {
        return { ships: data, misses: [] };
    }
    return {
        ships: Array.isArray(data?.ships) ? data.ships : [],
        misses: Array.isArray(data?.misses) ? data.misses : [],
    };
}

function rebuildBoardFromState(state) {
    fleetCells.forEach((cell) => {
        cell.classList.remove('placed');
        cell.classList.remove('hit-marker');
        cell.classList.remove('miss-marker');
        cell.classList.remove('ship-blue');
        cell.classList.remove('ship-green');
        cell.classList.remove('ship-orange');
    });

    shipOptions.forEach((option) => option.classList.remove('is-placed'));

    const board = normalizeBoardState(state);

    board.ships.forEach((shipPlacement) => {
        const shipColor = shipPlacement.ship;
        const positions = Array.isArray(shipPlacement.positions) ? shipPlacement.positions : [];

        const matchingOption = Array.from(shipOptions).find((option) => option.dataset.color === shipColor);
        if (matchingOption) {
            matchingOption.remove();
        }

        positions.forEach((position) => {
            const x = position && typeof position === 'object' ? position.x : undefined;
            const y = position && typeof position === 'object' ? position.y : undefined;
            const isHit = position && typeof position === 'object' && position.hit === true;
            const cell = typeof x === 'number' && typeof y === 'number'
                ? getBoardCell('fleet', x, y)
                : null;

            if (cell) {
                cell.classList.add('placed');
                cell.classList.add(`ship-${shipColor}`);

                if (isHit) {
                    cell.classList.add('hit-marker');
                }
            }
        });
    });

    board.misses.forEach((position) => {
        const x = position && typeof position === 'object' ? position.x : undefined;
        const y = position && typeof position === 'object' ? position.y : undefined;
        const cell = typeof x === 'number' && typeof y === 'number'
            ? getBoardCell('fleet', x, y)
            : null;

        if (cell) {
            cell.classList.add('miss-marker');
        }
    });
}

// Utility methods
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
    setOpponentStatus(false);
}

function setOpponentStatus(joined) {
    if (!opponentStatus) {
        return;
    }
    if (joined) {
        opponentStatus.textContent = 'Opponent joined';
        opponentStatus.classList.add('joined');
        opponentStatus.classList.remove('ready');
    } else {
        opponentStatus.textContent = 'Waiting for opponent…';
        opponentStatus.classList.remove('joined', 'ready');
    }
}

function setGameReadyStatus() {
    if (!opponentStatus) {
        return;
    }
    opponentStatus.textContent = 'Game ready — both fleets placed';
    opponentStatus.classList.add('joined', 'ready');
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
        setOpponentStatus(!data.waiting_for_opponent);
    } catch (error) {
        console.error('Failed to join game:', error);
    }

    establishEventSource();
}

const shipOptions = Array.from(document.querySelectorAll('.ship-option'));
const rotateShipBtn = document.getElementById('rotateShipBtn');
const clearSessionBtn = document.getElementById('clearSessionBtn');
const playerIdField = document.getElementById('playerIdField');
const gameIdField = document.getElementById('gameIdField');
const opponentStatus = document.getElementById('opponentStatus');
const viewTabs = Array.from(document.querySelectorAll('.view-tab'));
const fleetBoard = document.getElementById('fleetBoard');
const fireBoard = document.getElementById('fireBoard');
const fleetCells = Array.from(document.querySelectorAll('#fleetBoard .cell'));
const fireCells = Array.from(document.querySelectorAll('#fireBoard .cell'));
const cells = [...fleetCells, ...fireCells];
let selectedShip = null;
let shipOrientation = 'horizontal';
let gameState = { ships: [], misses: [] };
let activeView = 'fleet';
let fireDisabled = true;

function getCellCoordinates(cell) {
    const match = cell.id.match(/(\d+)-(\d+)$/);
    if (!match) {
        return null;
    }

    return {
        col: parseInt(match[1], 10),
        row: parseInt(match[2], 10)
    };
}

function getBoardCell(boardName, col, row) {
    return document.getElementById(`${boardName}-cell-${col}-${row}`) || document.getElementById(`${boardName}-cell-${row}-${col}`);
}

updateSessionDisplay();
clearSessionBtn.addEventListener('click', clearSession);

function setActiveView(view) {
    activeView = view;
    viewTabs.forEach((tab) => {
        tab.classList.toggle('active', tab.dataset.view === view);
    });

    fleetBoard.classList.toggle('board-hidden', view !== 'fleet');
    fleetBoard.classList.toggle('board-active', view === 'fleet');
    fireBoard.classList.toggle('board-hidden', view !== 'fire');
    fireBoard.classList.toggle('board-active', view === 'fire');
    fireBoard.classList.toggle('fire-control', view === 'fire');
}

async function fireAtCell(cell) {
    if (cell.classList.contains('fired')) {
        return;
    }

    fireDisabled = true;

    const coords = getCellCoordinates(cell);
    const payload = coords ? { x: coords.col, y: coords.row } : null;

    try {
        const response = await fetch('http://localhost:8080/api/fire', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'token': getCookie('token')
            },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            throw new Error(`Fire failed: ${response.status}`);
        }

        const isHit = (await response.text()) === 'true';
        cell.classList.add('fired');
        cell.classList.add(isHit ? 'hit' : 'miss');

        setTimeout(() => setActiveView('fleet'), 2000);

    } catch (error) {
        console.error('Failed to fire:', error);
    }
}

viewTabs.forEach((tab) => {
    tab.addEventListener('click', () => setActiveView(tab.dataset.view));
});

setActiveView(activeView);
joinGame().then(getGameData());

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

[fleetCells, fireCells].forEach((boardCells) => {
    boardCells.forEach((cell) => {
        cell.addEventListener('dragover', (event) => event.preventDefault());

        cell.addEventListener('drop', (event) => {
            event.preventDefault();
            if (activeView === 'fleet' && selectedShip) {
                placeShip(cell);
            }
        });

        cell.addEventListener('click', () => {
            if (activeView === 'fire' && !fireDisabled) {
                fireAtCell(cell);
                return;
            }

            if (selectedShip) {
                placeShip(cell);
            }
        });
    });
});

function highlightReadyCells() {
    clearReadyCells();
    if (!selectedShip) {
        return;
    }

    const length = parseInt(selectedShip.dataset.length, 10);
    fleetCells.forEach((cell) => {
        const coords = getCellCoordinates(cell);
        if (!coords) {
            return;
        }

        const { col, row } = coords;
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
    fleetCells.forEach((cell) => cell.classList.remove('ready'));
    shipOptions.forEach((item) => item.classList.remove('selected'));
}

function placeShip(startCell) {
    if (!selectedShip) {
        return;
    }

    const length = parseInt(selectedShip.dataset.length, 10);
    const coords = getCellCoordinates(startCell);
    if (!coords) {
        return;
    }

    const startCol = coords.col;
    const startRow = coords.row;

    const positions = [];
    for (let offset = 0; offset < length; offset += 1) {
        const col = shipOrientation === 'horizontal' ? startCol + offset : startCol;
        const row = shipOrientation === 'horizontal' ? startRow : startRow + offset;
        const targetCell = getBoardCell('fleet', col, row);

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
        const coords = getCellCoordinates(cell);
        return coords ? { x: coords.col, y: coords.row } : null;
    }).filter(Boolean);
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

function establishEventSource() {
    const eventSource = new EventSource("http://localhost:8080/api/events/" + getCookie('token'));

    // Listen for custom named events ("event: ping")
    eventSource.addEventListener("ping", (event) => {
        console.log("Ping received:", JSON.parse(event.data));
    });

    // Or listen for generic messages (when no 'event:' tag is set in Go)
    eventSource.onmessage = (event) => {
        console.log("Generic message:", event.data);
    };

    eventSource.onerror = (err) => {
        console.error("EventSource failed:", err);
    };

    eventSource.addEventListener("update", (event) => {
        try {
            const data = JSON.parse(event.data);
            gameState = normalizeBoardState(data);
            rebuildBoardFromState(gameState);

            console.log("Board rebuilt with new state:", gameState);
            setTimeout(() => setActiveView('fire'), 2000);
        } catch (err) {
            console.error("Failed to parse SSE JSON:", err);
        }
    });

    eventSource.addEventListener("fire-control", (event) => {
        fireDisabled = event.data === 'true';
    });

    eventSource.addEventListener("player-joined", () => {
        setOpponentStatus(true);
    });

    eventSource.addEventListener("game-ready", () => {
        setGameReadyStatus();
    });
}
