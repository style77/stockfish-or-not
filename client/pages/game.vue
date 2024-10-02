<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { formatTime } from '../lib/utils';
import { TheChessboard, type BoardApi, type MoveableColor } from 'vue3-chessboard';
import 'vue3-chessboard/style.css';

let boardAPI: BoardApi | null = null;
let socket: WebSocket | null = null;
const playerColor = ref<MoveableColor | ''>('');
const readyToStart = ref(false);

const playerTimeLeft = ref(0);
const opponentTimeLeft = ref(0);

// Modal state
const showModal = ref(false);
const gameResultText = ref('');
const isAI = ref(false);
const aiEngine = ref('');
const aiRank = ref(0);

const guessedCorrecly = ref();

// Session score
const sessionCorrect = ref(0);
const sessionTotal = ref(0);

// Persistent score
const persistentCorrect = ref(0);
const persistentTotal = ref(0);

const revealExplanation = ref(false);
const revealScore = ref(false);

const updatePersistentScore = () => {
    const score = {
        correct: persistentCorrect.value,
        total: persistentTotal.value,
    };
    localStorage.setItem('persistentScore', JSON.stringify(score));
};

const loadPersistentScore = () => {
    const score = localStorage.getItem('persistentScore');
    if (score !== null) {
        const parsedScore = JSON.parse(score);
        persistentCorrect.value = parsedScore.correct;
        persistentTotal.value = parsedScore.total;
    }
};

const guess = (option: "AI" | "Human") => {
    sessionTotal.value++;
    persistentTotal.value++;

    if (isAI.value && option === "AI") {
        sessionCorrect.value++;
        persistentCorrect.value++;
        guessedCorrecly.value = true;
    } else if (!isAI.value && option === "Human") {
        sessionCorrect.value++;
        persistentCorrect.value++;
        guessedCorrecly.value = true;
    } else {
        guessedCorrecly.value = false;
    }

    updatePersistentScore();

    setTimeout(() => {
        revealExplanation.value = true;
    }, 500);

    setTimeout(() => {
        revealScore.value = true;
    }, 500);
};

const tryAgain = () => {
    sessionCorrect.value = 0;
    sessionTotal.value = 0;
    guessedCorrecly.value = undefined;
    showModal.value = false;

    revealExplanation.value = false;
    revealScore.value = false;

    startGame();
};

const startGame = () => {
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

        switch (data.state) {
            case 1:
                playerColor.value = data.data.color as MoveableColor;
                readyToStart.value = true;

                playerTimeLeft.value = data.data.gameTime;
                opponentTimeLeft.value = data.data.gameTime;
                break;
            case 78:
                boardAPI?.move(data.data.move);
                break;
            case 80:
                if (data.data.color === playerColor.value) {
                    playerTimeLeft.value = data.data.time;
                } else {
                    opponentTimeLeft.value = data.data.time;
                }
                break;
            case 99:
                handleEndGame(data.data);
                break;
            default:
                console.log("Unexpected game state:", data.state);
                break;
        }

        if (!boardAPI) {
            console.error("Board API is not initialized.");
            return;
        }
    };
}

onMounted(() => {
    loadPersistentScore();

    startGame();
});

// Handle board creation
const handleBoardCreated = (boardApi: BoardApi) => {
    boardAPI = boardApi;
    console.log("Board API initialized.");
};

const handleEndGame = (data: any) => {
    console.log('Game ended');
    socket?.close();
    boardAPI = null;
    socket = null;

    playerColor.value = '';
    readyToStart.value = false;

    isAI.value = data.isAI;
    if (data.isAI) {
        aiEngine.value = data.AIMeta.engine;
        aiRank.value = data.AIMeta.rank;
    }

    switch (data.result) {
        case "1/2-1/2":
            gameResultText.value = "It's a draw!";
            break;
        case "1-0":
            gameResultText.value = "White won!";
            break;
        case "0-1":
            gameResultText.value = "Black won!";
            break;
    }

    showModal.value = true;
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
        const isFirstMove = (moves.length === 1 && playerColor.value === 'white') || (moves.length === 2 && playerColor.value === 'black');

        // send socket data only if it's player's turn

        if ((lastMove.color as string) !== (playerColor.value[0] as string)) {
            return;
        }

        socket?.send(JSON.stringify({ move: lastMove.lan, isFirstMove: isFirstMove }));
    }
}
</script>

<template>
    <div class="min-h-screen bg-black flex flex-col items-center justify-center text-center">
        <div class="w-1/3 flex flex-row justify-between">
            <div v-show="playerColor !== ''" class="text-white">
                You are playing as: {{ playerColor }}
            </div>
        </div>

        <div className="flex flex-row gap-4 h-full" v-if="readyToStart || showModal">
            <TheChessboard @board-created="handleBoardCreated" @move="handleMove"
                :player-color="(playerColor as MoveableColor)" :board-config="{
                    'orientation': playerColor === 'white' ? 'white' : 'black',
                }" />

            <div className="flex flex-col min-h-[80vh]">
                <h1 class="text-white text-2xl">Stockfish or not</h1>
                <div className="h-full flex flex-col justify-between">
                    <div className="bg-gray-300 py-4 w-full">
                        <h2 class="text-black text-2xl">{{ formatTime(opponentTimeLeft) }}</h2>
                    </div>
                    <div className="bg-gray-300 py-4 w-full">
                        <h2 class="text-black text-2xl">{{ formatTime(playerTimeLeft) }}</h2>
                    </div>
                </div>
            </div>
        </div>

        <div v-else class="text-white">
            Waiting for opponent...
        </div>

        <!-- Modal -->
        <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-80 z-10 flex items-center justify-center">
            <div class="bg-white p-6 rounded-lg text-center">
                <h2 class="text-2xl mb-4">{{ gameResultText }}</h2>
                <span class="block mb-2 font-semibold text-gray-800" v-show="guessedCorrecly === undefined">
                    Who do you think you played against?
                </span>
                <div class="flex flex-row justify-around mb-4 gap-8" v-if="guessedCorrecly === undefined">
                    <button @click="() => guess('AI')"
                        class="bg-gray-800 text-white py-2 px-4 rounded w-full transition duration-300 hover:bg-green-800">Stockfish</button>
                    <button @click="() => guess('Human')"
                        class="bg-gray-800 text-white py-2 px-4 rounded w-full transition duration-300 hover:bg-yellow-800">Human</button>
                </div>
                <div class="mb-4" v-else>
                    <span class="block mb-4 font-semibold text-gray-800 text-xl">
                        You guessed
                        <span v-if="guessedCorrecly" class="text-green-500 transition duration-300"
                            :class="revealScore ? 'blur-0' : 'blur-xl'">correctly!</span>
                        <span v-else class="text-red-500 transition duration-300"
                            :class="revealScore ? 'blur-0' : 'blur-xl'">incorrectly!</span>
                    </span>
                    <span class="block mb-2 font-light text-gray-700/50 text-xs" v-if="isAI">
                        You played versus <span class="transition duration-300"
                            :class="revealExplanation ? 'blur-0' : 'blur-xl'"><span class="font-bold">{{ aiEngine }}</span>
                            with ELO of <span class="font-bold">{{ aiRank }}</span></span>
                    </span>
                    <span class="block mb-2 font-light text-gray-700/50 text-xs transition duration-300"
                        :class="revealExplanation ? 'blur-0' : 'blur-xl'" v-else>
                        You played versus <span class="transition duration-300"
                            :class="revealExplanation ? 'blur-0' : 'blur-xl'">another human being.</span>
                    </span>
                        <div class="flex flex-col">
                            <button @click="tryAgain" class="bg-green-700 text-white py-2 px-6 rounded">Try Again</button>
                            <span class="block mt-2 font-light text-gray-700/50 text-xs transition duration-300"
                                :class="revealScore ? 'blur-0' : 'blur-xl'">
                                You have already guessed correctly <span class="font-bold">{{ persistentCorrect }}/{{
                                    persistentTotal }}</span> times.
                            </span>
                        </div>
                </div>
            </div>
        </div>

    <div class="blur-xl"></div>
</div></template>
