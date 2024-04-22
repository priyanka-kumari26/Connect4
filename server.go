package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Define constants for the game
const (
	BoardWidth  = 7
	BoardHeight = 6
)

type Game struct {
	Board       [BoardHeight][BoardWidth]int
	Players     [2]string
	CurrentTurn int
	Mutex       sync.Mutex
}

// Function to initialize the game
func (game *Game) InitGame() {
	game.CurrentTurn = 0
}

// Function to handle the game logic
func (game *Game) PlayMove(column, playerIdx int) bool {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	player := playerIdx + 1
	for row := BoardHeight - 1; row >= 0; row-- {
		if game.Board[row][column] == 0 {
			game.Board[row][column] = player
			return true
		}
	}
	return false
}

// Function to check if a player has won
func (game *Game) CheckWin(player int) bool {
	// Check horizontal
	for row := 0; row < BoardHeight; row++ {
		for col := 0; col <= BoardWidth-4; col++ {
			if game.Board[row][col] == player &&
				game.Board[row][col+1] == player &&
				game.Board[row][col+2] == player &&
				game.Board[row][col+3] == player {
				return true
			}
		}
	}

	// Check vertical
	for col := 0; col < BoardWidth; col++ {
		for row := 0; row <= BoardHeight-4; row++ {
			if game.Board[row][col] == player &&
				game.Board[row+1][col] == player &&
				game.Board[row+2][col] == player &&
				game.Board[row+3][col] == player {
				return true
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for row := 0; row <= BoardHeight-4; row++ {
		for col := 0; col <= BoardWidth-4; col++ {
			if game.Board[row][col] == player &&
				game.Board[row+1][col+1] == player &&
				game.Board[row+2][col+2] == player &&
				game.Board[row+3][col+3] == player {
				return true
			}
		}
	}

	// Check diagonal (bottom-left to top-right)
	for row := 3; row < BoardHeight; row++ {
		for col := 0; col <= BoardWidth-4; col++ {
			if game.Board[row][col] == player &&
				game.Board[row-1][col+1] == player &&
				game.Board[row-2][col+2] == player &&
				game.Board[row-3][col+3] == player {
				return true
			}
		}
	}

	return false
}

func handleConnection(conn net.Conn, playerIdx int, game *Game, playerCh chan int) {
	defer conn.Close()

	// Prompt player for their name
	fmt.Fprintf(conn, "Enter your name:\n")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	playerName := scanner.Text()
	game.Players[playerIdx] = playerName

	fmt.Printf("Player %d (%s) connected\n", playerIdx+1, playerName)

	// Notify both players about the game start
	fmt.Fprintf(conn, "Game starts!\n")

	waitingMsgPrinted := false

	// Loop to handle moves from the client
	for {

		if playerIdx != game.CurrentTurn {

			if !waitingMsgPrinted {
				sendBoardState(conn, game)
				fmt.Fprintf(conn, "Waiting for opponent's move...\n")
				waitingMsgPrinted = true
			}
			continue
		} else {
			waitingMsgPrinted = false
		}

		// Send the current game board to the player
		sendBoardState(conn, game)

		fmt.Fprintf(conn, "Your turn!\n")

		// Read the player's move
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		input := scanner.Text()
		column, err := strconv.Atoi(input)
		if err != nil || column < 0 || column >= BoardWidth {
			fmt.Fprintf(conn, "Invalid input. Please enter a number between 0 and 6.\n")
			continue
		}

		// Attempt to make the move
		if !game.PlayMove(column, playerIdx) {
			fmt.Fprintf(conn, "Column is full. Please choose another column.\n")
			continue
		}
		if game.CheckWin(playerIdx + 1) {
			for i := 0; i < 2; i++ {
				if i == playerIdx {
					fmt.Fprintf(conn, "Congratulations! You win!\n")
				} else {
					fmt.Fprintf(conn, "Player %d (%s) wins!\n", playerIdx+1, game.Players[playerIdx])
				}
			}
			playerCh <- playerIdx + 1 // Notify main goroutine about the winner
			return
		}

		game.CurrentTurn = (game.CurrentTurn + 1) % 2
	}
}

// Function to send the current game board to a player
func sendBoardState(conn net.Conn, game *Game) {
	for _, row := range game.Board {
		fmt.Fprintf(conn, "%s\n", strings.Join(strings.Fields(fmt.Sprint(row)), " "))
	}
}

func main() {
	fmt.Println("Connect Four Server")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		//fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	game := &Game{}
	game.InitGame()

	playerCh := make(chan int)

	// Accept connections and handle them in separate goroutines
	for i := 0; i < 2; i++ {
		conn, err := listener.Accept()
		if err != nil {
			//fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, i, game, playerCh)
	}

	winner := <-playerCh
	fmt.Printf("Player %d wins the game!\n", winner)
}
