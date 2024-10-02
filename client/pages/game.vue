<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { TheChessboard, type BoardApi, type MoveableColor } from 'vue3-chessboard';
import 'vue3-chessboard/style.css';

let boardAPI: BoardApi | null = null;
let socket: WebSocket | null = null;
const playerColor = ref<MoveableColor | ''>('');
const readyToStart = ref(false);
const gameTime = ref(0); // seconds

onMounted(() => {
    console.log("WebSocket connection initializing...");
    socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
        console.log("WebSocket connection established.");
    };

    socket.onerror = (error) => {
        console.error("WebSocket error:", error);
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        console.log("Received WebSocket data:", data);

        switch (data.state) {
            case 1:
                playerColor.value = data.data.color as MoveableColor;
                readyToStart.value = true;
                gameTime.value = data.data.gameTime;
                break;
            case 78:
                boardAPI?.move(data.data.move);
            default:
                console.log("Unexpected game state:", data.state);
        }

        if (!boardAPI) {
            console.error("Board API is not initialized.");
            return;
        }
    };
});

// Handle board creation
const handleBoardCreated = (boardApi: BoardApi) => {
    boardAPI = boardApi;
    console.log("Board API initialized.");
};

// Game event handlers
const handleCheckmate = () => {
    console.log('Checkmate');
    handleEndGame();
};

const handleDraw = () => {
    console.log('Draw');
    handleEndGame();
};

const handleStalemate = () => {
    console.log('Stalemate');
    handleEndGame();
};

const handleEndGame = () => {
    console.log('Game Over');
};

// Handle player move
const handleMove = () => {
    const history = boardAPI?.getHistory(true);

    const moves = history?.map((move) => {
        if (typeof move === 'object') {
            return move; // e.g., "e2e4"
        } else {
            return move;
        }
    });

    if (moves) {
        const lastMove = moves[moves.length - 1];
        let isFirstMove = false;
        // check if it's first player's move
        if (moves.length === 2 && playerColor.value === 'black') {
            isFirstMove = true;
        } else if (moves.length === 1 && playerColor.value === 'white') {
            isFirstMove = true;
        }

        console.log(lastMove)
        socket?.send(JSON.stringify({ move: lastMove.lan, isFirstMove: isFirstMove }));
    }
}
</script>

<template>
    <div class="min-h-screen bg-black flex flex-col items-center justify-center text-center">
        <div className="w-1/3 flex flex-row justify-between">
            <div v-show="playerColor !== ''" class="text-white">
                You are playing as: {{ playerColor }}
            </div>
            <div v-show="gameTime !== 0" class="text-white">
                Game time: {{ gameTime }} seconds
            </div>
        </div>
        <TheChessboard v-if="readyToStart" 
                       @board-created="handleBoardCreated" 
                       @checkmate="handleCheckmate" 
                       @draw="handleDraw" 
                       @stalemate="handleStalemate" 
                       @move="handleMove" 
                       :player-color="(playerColor as MoveableColor)" />
        <div v-else class="text-white">
            Waiting for opponent...
        </div>
    </div>
</template>
