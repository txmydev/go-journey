package main

type client struct {
	serverAddress string
}

func CreateClientAndConnect(address string) *client {
	var client client = client{address}
	return &client
}
