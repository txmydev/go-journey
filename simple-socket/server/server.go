package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type client struct {
	name       string
	id         int
	addr       string
	connection net.Conn
	color      string
}

type server struct {
	address          string
	port             uint32
	connectedClients uint8
	users            []*client
	running          bool
	connection       net.Listener
	messageCount     uint64
}

const MAX_MESSAGE_SIZE uint16 = 1024

func (server *server) CreateClient(addr string) *client {
	var cl client = client{"", int(server.connectedClients), addr, nil, "0"}

	server.users[server.connectedClients] = &cl
	server.connectedClients++

	return &cl
}

func (server *server) PrintState() {
	fmt.Printf(AnsiBackground("6")+"started listener at port %v%v\n", server.port, CLR_RESET)
	for _, user := range server.users {
		if user != nil {
			fmt.Printf("%v connected.\n", user.name)
		}
	}
}

func (server *server) Listen() {
	var addr string = server.address + ":" + fmt.Sprintf("%d", server.port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal(err)
	}

	server.PrintState()

	server.connection = listener
	server.running = true

	// listener.(*net.TCPListener).SetDeadline(time.Now().Add(100 * time.Millisecond))

	for {
		if !server.running {
			fmt.Println("server shutdown")
			break
		}

		var connection, err = listener.Accept()
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			} else if errors.Is(err, net.ErrClosed) {
				continue
			}

			fmt.Printf("error occurred attempting to accept a connection: %v\n", err.Error())
			continue
		}

		go AcceptConnection(server, connection)
	}
	wait.Done()
}

func (server *server) GetClient(nameOrAddress string) (*client, error) {
	var found *client = nil
	for _, client := range server.users {
		if client == nil {
			continue
		}

		if client.name != nameOrAddress && client.addr != nameOrAddress {
			continue
		}

		found = client
	}

	if found == nil {
		return nil, errors.New("client with address " + nameOrAddress + " not found")
	}

	return found, nil
}

func (server *server) Broadcast(message string) {
	fmt.Println(message)

	for _, user := range server.users {
		if user != nil && user.connection != nil {
			user.connection.Write([]byte(message))
		}
	}

	server.messageCount++
}

func (server *server) ListUserNames() ([]string, error) {
	if server.connectedClients == 0 {
		return []string{""}, errors.New("there aren't any clients connected to the server")
	}

	var names []string = make([]string, server.connectedClients)
	for index, user := range server.users {
		if user.name == "" {
			continue
		}

		names[index] = user.name
	}
	return names, nil
}

func AcceptConnection(server *server, connection net.Conn) {
	var user client = client{}

	user.connection = connection
	user.addr = connection.LocalAddr().String()
	user.id = int(server.connectedClients)

	prefix := "[" + user.addr + "]"

	server.users[server.connectedClients] = &user
	server.connectedClients++

	for {
		if !server.running {
			connection.Close()
			break
		}

		read := make([]byte, MAX_MESSAGE_SIZE)
		len, err := connection.Read(read)
		if err != nil {
			fmt.Printf("[%v] disconnected\n", user.name)
			break
		}

		var data string = string(read[:len])
		if strings.HasPrefix(data, "$connection-handshake") {
			if user.name != "" {
				fmt.Printf("%v received handshake when already connected (weird)\n", "["+user.name+"]")
				continue
			}

			name, _ := strings.CutPrefix(data, "$connection-handshake:")
			user.name = name

			fmt.Printf("%v connected.\n", name)
			continue
		}

		if user.name == "" {
			fmt.Printf("%v first message wasn't a handshake, can't broadcast message\n", prefix)
			continue
		}

		server.Broadcast(fmt.Sprintf("%v: %v", AnsiColor(user.color)+user.name+CLR_RESET, data))
	}

	server.users[user.id] = nil
	server.connectedClients--
}

func (server *server) Shutdown() {
	server.running = false
	server.connection.Close()
	server.connectedClients = 0

	for _, user := range server.users {
		if user != nil && user.connection != nil {
			user.connection.Close()
		}
	}
}
