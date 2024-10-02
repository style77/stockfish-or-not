package internal

import (
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/notnil/chess"
	"github.com/style77/stockfish-or-not/internal/constants"
	"github.com/style77/stockfish-or-not/internal/engine"
	"github.com/style77/stockfish-or-not/internal/game"
	"github.com/style77/stockfish-or-not/internal/models"
	"github.com/style77/stockfish-or-not/internal/timer"
	"github.com/style77/stockfish-or-not/internal/utils"
)

type App struct {
	WaitingPlayers []*models.Player
	Rooms          map[string]*models.Room
	mux            sync.Mutex
}

func CreateApp() *App {
	return &App{
		WaitingPlayers: make([]*models.Player, 0),
		Rooms:          make(map[string]*models.Room),
	}
}

func setRoomTurn(room *models.Room, player1Color string, player1, player2 *models.Player) {
	if player1Color == "black" {
		room.Turn = player2
	} else {
		room.Turn = player1
	}
}

func (app *App) createRoom(player1, player2 *models.Player, isAI bool) *models.Room {
	roomID := uuid.New().String()

	room := &models.Room{
		ID:        roomID,
		Player1:   player1,
		Player2:   player2,
		IsAI:      isAI,
		Moves:     make([]string, 0),
		GameEnded: false,
	}

	player1.Room = room
	if player2 != nil {
		player2.Room = room
	}

	app.mux.Lock()
	app.Rooms[roomID] = room
	app.mux.Unlock()

	return room
}

func getPlayerColor() string {
	if rand.Float64() < 0.5 {
		return "black"
	}
	return "white"
}

func getOpponentColor(playerColor string) string {
	if playerColor == "white" {
		return "black"
	}
	return "white"
}

func notifyPlayersAboutTime(room *models.Room, color string, remainingTime int) {
	err := utils.NotifyBothPlayers(room, map[string]interface{}{
		"message": "Time left for " + color,
		"roomID":  room.ID,
		"state":   80,
		"data": map[string]interface{}{
			"time":  remainingTime,
			"color": color,
		},
	})

	if err != nil {
		log.Println("Error notifying players about time:", err)
	}
}

func (app *App) HandleAIOpponent(player *models.Player) {
	selectedEngine := "stockfish"

	manager, elo := engine.DeterminateAI()

	aiOpponent := &models.Player{IsAI: true, Rank: &elo, Engine: &selectedEngine, AI: manager}
	room := app.createRoom(player, aiOpponent, true)

	playerColor := getPlayerColor()
	opponentColor := getOpponentColor(playerColor)

	player.Color = &playerColor
	aiOpponent.Color = &opponentColor

	aiOpponent.Timer = timer.NewTimer(constants.GameTime, func(remainingTime int) {
		notifyPlayersAboutTime(room, opponentColor, remainingTime)

		if remainingTime == 0 {
			var outcome chess.Outcome

			// aiOpponent timer is up, so player wins
			if playerColor == "white" {
				outcome = chess.WhiteWon
			} else {
				outcome = chess.BlackWon
			}

			game.HandleGameEnd(aiOpponent, room, "Time is up", &utils.GameResult{
				Outcome:       outcome,
				OutcomeReason: "Time is up",
			})
		}
	})
	player.Timer = timer.NewTimer(constants.GameTime, func(remainingTime int) {
		notifyPlayersAboutTime(room, playerColor, remainingTime)

		if remainingTime == 0 {
			var outcome chess.Outcome

			// player timer is up, so aiOpponent wins
			if playerColor == "white" {
				outcome = chess.BlackWon
			} else {
				outcome = chess.WhiteWon
			}

			game.HandleGameEnd(player, room, "Time is up", &utils.GameResult{
				Outcome:       outcome,
				OutcomeReason: "Time is up",
			})
		}
	})

	setRoomTurn(room, playerColor, player, aiOpponent)

	log.Println("Player", player.Conn.RemoteAddr(), "has been matched with an AI opponent with Elo", elo)
	player.Conn.WriteJSON(map[string]interface{}{
		"message": "You have been matched with an opponent! You are playing as " + playerColor,
		"roomID":  room.ID,
		"state":   1,
		"data": map[string]interface{}{
			"color":    playerColor,
			"gameTime": constants.GameTime, // seconds
		},
	})

	// if player is black, AI makes the first move
	if playerColor == "black" {
		move, err := manager.ProcessMove("", constants.MaxDepth)
		if err != nil {
			log.Println("Error processing move for AI opponent:", err)
			return
		}

		app.ProcessMove(aiOpponent, move, true)
	}
}

func (app *App) FindOpponent(player *models.Player) {
	isOpponentAi := rand.Float64() < constants.AIPlayerPosibility

	if isOpponentAi {
		waitingTime := rand.IntN(constants.PlayerLookingIntervalRangeTo-constants.PlayerLookingIntervalRangeFrom) + constants.PlayerLookingIntervalRangeFrom

		time.AfterFunc(time.Duration(waitingTime)*time.Second, func() {
			app.HandleAIOpponent(player)
		})
	} else {
		go waitForRealPlayer(player, app)
	}
}

func getRandomPlayerThatsNot(player *models.Player, app *App) *models.Player {
	app.mux.Lock()
	defer app.mux.Unlock()

	if len(app.WaitingPlayers) == 0 {
		return nil
	}

	for _, p := range app.WaitingPlayers {
		if p != player {
			return p
		}
	}

	return nil
}

func removePlayerFromWaitingList(player *models.Player, app *App) {
	app.mux.Lock()
	defer app.mux.Unlock()

	for i, p := range app.WaitingPlayers {
		if p == player {
			app.WaitingPlayers = append(app.WaitingPlayers[:i], app.WaitingPlayers[i+1:]...)
			break
		}
	}
}

func isPlayerInWaitingList(player *models.Player, app *App) bool {
	app.mux.Lock()
	defer app.mux.Unlock()

	for _, p := range app.WaitingPlayers {
		if p == player {
			return true
		}
	}
	return false
}

func isPlayerInGame(player *models.Player, app *App) bool {
	app.mux.Lock()
	defer app.mux.Unlock()

	for _, room := range app.Rooms {
		if room.Player1 == player || room.Player2 == player {
			return true
		}
	}
	return false
}

func waitForRealPlayer(player *models.Player, app *App) {
	timeout := time.After(constants.TimeoutForOpponent * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if isPlayerInGame(player, app) {
				log.Println("Player is already in a game:", player.Conn.RemoteAddr())
				return
			}

			log.Println("Looking for an opponent for player:", player.Conn.RemoteAddr())
			log.Println("Players waiting:", len(app.WaitingPlayers))

			opponent := getRandomPlayerThatsNot(player, app)
			if opponent != nil {
				removePlayerFromWaitingList(opponent, app)
				removePlayerFromWaitingList(player, app)
			}

			if opponent != nil {
				log.Println("Found an opponent for player:", player.Conn.RemoteAddr())
				log.Println("Player 1:", player.Conn.RemoteAddr())
				log.Println("Player 2:", opponent.Conn.RemoteAddr())

				player1Color := getPlayerColor()
				player2Color := getOpponentColor(player1Color)

				player.Color = &player1Color
				opponent.Color = &player2Color

				room := app.createRoom(player, opponent, false)
				setRoomTurn(room, player1Color, player, opponent)

				player.Timer = timer.NewTimer(constants.GameTime, func(remainingTime int) {
					notifyPlayersAboutTime(room, player1Color, remainingTime)

					if remainingTime == 0 {
						var outcome chess.Outcome

						// player timer is up, so opponent wins
						if player1Color == "white" {
							outcome = chess.BlackWon
						} else {
							outcome = chess.WhiteWon
						}

						game.HandleGameEnd(player, room, "Time is up", &utils.GameResult{
							Outcome:       outcome,
							OutcomeReason: "Time is up",
						})
					}
				})
				opponent.Timer = timer.NewTimer(constants.GameTime, func(remainingTime int) {
					notifyPlayersAboutTime(room, player2Color, remainingTime)

					if remainingTime == 0 {
						var outcome chess.Outcome

						// opponent timer is up, so player wins
						if player1Color == "white" {
							outcome = chess.WhiteWon
						} else {
							outcome = chess.BlackWon
						}

						game.HandleGameEnd(opponent, room, "Time is up", &utils.GameResult{
							Outcome:       outcome,
							OutcomeReason: "Time is up",
						})
					}
				})

				player.Conn.WriteJSON(map[string]interface{}{
					"message": "You have been matched with a player! You are playing as " + player1Color,
					"roomID":  room.ID,
					"state":   1,
					"data": map[string]interface{}{
						"color":    player1Color,
						"gameTime": constants.GameTime, // seconds
					},
				})
				opponent.Conn.WriteJSON(map[string]interface{}{
					"message": "You have been matched with a player! You are playing as " + player2Color,
					"roomID":  room.ID,
					"state":   1,
					"data": map[string]interface{}{
						"color":    player2Color,
						"gameTime": constants.GameTime,
					},
				})

				log.Println("Players matched:", player.Conn.RemoteAddr(), opponent.Conn.RemoteAddr())
				return
			}

			log.Println("No opponent found for player:", player.Conn.RemoteAddr())
			log.Println("Adding player to waiting list:", player.Conn.RemoteAddr())

			if !isPlayerInWaitingList(player, app) {
				app.mux.Lock()
				app.WaitingPlayers = append(app.WaitingPlayers, player)
				app.mux.Unlock()
			}
		case <-timeout:
			// Timeout: if no real player is found, match with AI
			app.HandleAIOpponent(player)

			return
		}
	}
}

func (app *App) ProcessMove(player *models.Player, move string, isFirstMove bool) {
	room := player.Room
	if room == nil {
		log.Println("Player is not in a room.")
		return
	}

	if room.Turn != player {
		log.Println("Not player's turn.")
		return
	}

	opponent := room.Player1
	if player == room.Player1 {
		opponent = room.Player2
	}

	err := utils.SafelyNotifyPlayer(opponent, map[string]interface{}{
		"message": "Opponent made move",
		"roomID":  room.ID,
		"state":   78,
		"data": map[string]interface{}{
			"move": move,
		},
	})

	if err != nil {
		log.Println("Error notifying opponent about move:", err)
	}

	room.Mux.Lock()
	room.Moves = append(room.Moves, move)
	room.Mux.Unlock()

	game.ChangeTurn(room)

	if room.IsAI && opponent.AI != nil {
		go func() {
			processAIMove(room, opponent)
		}()
	}
}

func processAIMove(room *models.Room, aiPlayer *models.Player) {
	waitTime := rand.IntN(constants.AIMoveWaitTimeFrom-constants.AIMoveWaitTimeFrom+1) + constants.AIMoveWaitTimeFrom
	log.Printf("AI will take %d seconds to make its move...\n", waitTime)
	time.Sleep(time.Duration(waitTime) * time.Second)

	randomDepth := rand.IntN(constants.MaxDepth) + 1
	aiMove, err := aiPlayer.AI.ProcessMove(utils.GetPosition(room.Moves), randomDepth)
	if err != nil {
		log.Println("Error getting AI move:", err)
		return
	}

	log.Println("Processing AI move with depth", randomDepth, ":", aiMove)

	humanPlayer := room.Player1
	if aiPlayer == room.Player1 {
		humanPlayer = room.Player2
	}

	err = utils.SafelyNotifyPlayer(humanPlayer, map[string]interface{}{
		"message": "Opponent made move",
		"roomID":  room.ID,
		"state":   78,
		"data": map[string]interface{}{
			"move": aiMove,
		},
	})

	if err != nil {
		log.Println("Error notifying human player about AI move:", err)
		return
	}

	room.Mux.Lock()
	room.Moves = append(room.Moves, aiMove)
	room.Mux.Unlock()

	game.ChangeTurn(room)
}
