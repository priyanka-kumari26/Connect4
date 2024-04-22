package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Connect4 Client")

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		//fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Read the player's name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	playerName, _ := reader.ReadString('\n')
	playerName = strings.TrimSpace(playerName)

	fmt.Fprintf(conn, "%s\n", playerName)
	scanner := bufio.NewScanner(os.Stdin)
	serverReader := bufio.NewReader(conn)

	// Start receiving game updates from the server
	go func() {
		for {
			line, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Server disconnected:")
				os.Exit(1)
			}
			fmt.Print(line)

			if strings.Contains(line, "Enter column number") {
				fmt.Print("Enter column number (0-6): ")
			} else if strings.Contains(line, "wins!") {
				fmt.Println(line)
				os.Exit(0)
			}
		}
	}()

	// Loop to send moves to the server
	for scanner.Scan() {
		input := scanner.Text()
		_, err := fmt.Fprintf(conn, "%s\n", input)
		if err != nil {
			//fmt.Println("Error sending move:", err)
			return
		}
	}
}
