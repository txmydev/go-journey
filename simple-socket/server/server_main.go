package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const CLR_RESET = "\033[0m"

func AnsiColor(clr string) string {
	return "\033[38;5;" + clr + "m"
}

func AnsiBackground(clr string) string {
	return "\033[48;5;" + clr + "m"
}

func ClearRecentlyWritten() {
	fmt.Print("\033[1A\033[K\r")
}

var wait = sync.WaitGroup{}

type config struct {
	maxClients uint32
	port       uint16
	address    string
}

func loadConfiguration(fileName string) config {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var config config = config{2000, 5001, "127.0.0.1"}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		portRaw, foundPort := strings.CutPrefix(text, "port=")
		if foundPort {
			port, err := strconv.ParseUint(portRaw, 10, 16)
			if err != nil {
				fmt.Printf("config.properties: expected number < 63739 while setting port but found %v, using default value.\n", portRaw)
				continue
			}
			config.port = uint16(port)
		}

		maxClientsRaw, foundMaxClients := strings.CutPrefix(text, "max-clients-allowed=")
		if foundMaxClients {
			maxClients, err := strconv.ParseUint(maxClientsRaw, 10, 32)
			if err != nil {
				fmt.Printf("config.properties: expected number but found %v while setting max-clients-allowed, using default value.\n", maxClientsRaw)
				continue
			}

			config.maxClients = uint32(maxClients)
		}

		address, foundAddress := strings.CutPrefix(text, "address=")
		if foundAddress {
			config.address = address
		}
	}

	return config
}

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[3J")
	fmt.Print("\033[H")
}

func main() {
	// Initialize the configuration
	var config config = loadConfiguration("config/config.properties")
	server := server{
		nil,
		config.address,
		make([]*client, config.maxClients),
		uint32(config.port),
		0,
		false,
		make([]Message, 0),
	}

	ClearScreen()

	wait.Add(1)
	go server.Listen()
	wait.Add(1)
	go AcceptInput(&server)

	wait.Wait()

	fmt.Println("goodbye!")
}

func AcceptInput(server *server) {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error occurred while reading input: %v\n", err.Error())
			break
		}

		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")

		if len(args) == 0 {
			continue
		}

		exited := false
		switch cmd := args[0]; cmd {
		case "broadcast":
			ClearRecentlyWritten()
			server.Broadcast(strings.Join(args[1:], " "))
		case "setcolor":
			ClearRecentlyWritten()

			if len(args) < 3 {
				fmt.Println("usage: setcolor <user> <color>")
				continue
			}
			userName := args[1]
			user, err := server.GetClient(userName)
			if err != nil {
				fmt.Printf("error: user %v is not connected.\n", userName)
				continue
			}

			color := args[2]
			num, err := strconv.ParseInt(color, 10, 8)
			if num < 0 || err != nil {
				fmt.Println("error: colors must go from 0 to 255")
				continue
			}

			user.color = color
			fmt.Printf("success: you've set the color of %v to %v\n", userName, AnsiColor(user.color)+user.name+CLR_RESET)
		case "debug":
			ClearRecentlyWritten()

			fmt.Println("debug:")

			if server.connectedClients == 0 {
				fmt.Println(" no connections found")
				continue
			}

			for _, user := range server.users {
				if user != nil && user.connection != nil {
					fmt.Printf(" %v: { id: %v, color: %v, addr: %v }\n", user.name, user.id, user.color, user.addr)
				}
			}
		case "exit":
			ClearRecentlyWritten()

			exited = true
			server.Shutdown()
		case "clear":
			ClearScreen()

			server.PrintState()
		default:
			ClearRecentlyWritten()

			fmt.Println(AnsiColor("1") + "error: unknown command")
			fmt.Println(AnsiColor("10") + "available:" + CLR_RESET)
			fmt.Println(AnsiColor("10") + " debug " + CLR_RESET + "- shows the current connected users and their info")
			fmt.Println(AnsiColor("10") + " broadcast <message> " + CLR_RESET + " sends a message to every connected user")
			fmt.Println(AnsiColor("10") + " setcolor <user> <color> " + CLR_RESET + " update's a user color")
			fmt.Println(AnsiColor("10") + " exit " + CLR_RESET + " shutdowns the server and exists the program")
			fmt.Println(AnsiColor("10") + " clear " + CLR_RESET + " clears the screen")
		}

		if exited {
			break
		}
	}

	wait.Done()
}
