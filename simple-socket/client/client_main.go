package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	wait     sync.WaitGroup = sync.WaitGroup{}
	running  bool           = true
	username string
)

const CLR_RESET = "\033[0m"

func AnsiColor(clr string) string {
	return "\033[38;5;" + clr + "m"
}

func AnsiBackground(clr string) string {
	return "\033[48;5;" + clr + "m"
}

func ClearRecentlyWritten() {
	fmt.Print("\033[1A\033[K")
}

func write(connection net.Conn) {
	fmt.Println("You are connected, type anything you want to send to the server or 'exit' to exit.")
	connection.Write([]byte("$connection-handshake:" + username))
	reader := bufio.NewReader(os.Stdin)

	for {
		if !running {
			break
		}

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		ClearRecentlyWritten()

		if text == "exit" {
			running = false
			break
		}

		connection.Write([]byte(text))
	}

	connection.Close()
	wait.Done()

	fmt.Println("[write] connection closed")

	running = false
}

func read(connection net.Conn) {
	// connection.SetDeadline(time.Now().Add(200 * time.Millisecond))

	for {
		if !running {
			break
		}

		var read []byte = make([]byte, 1024)
		_, err := connection.Read(read)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}

			fmt.Println("[read] connection closed")
			running = false
			break
		}

		fmt.Printf("%v\n", string(read))
	}

	connection.Close()
	wait.Done()
}

func main() {
	var address string
	fmt.Print("Please enter a username: ")
	fmt.Scanf("%v", &username)
	fmt.Print("Please enter an address: ")
	fmt.Scanf("%v", &address)
	fmt.Println()

	connection, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("error occurred while connecting to %v: %v\n", address, err.Error())
		return
	}

	wait.Add(1)
	go write(connection)

	wait.Add(1)
	go read(connection)

	wait.Wait()
}
