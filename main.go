package main

import (
	"pizza/client"
	"pizza/server"
)

func main() {

	server.CheckServer()
	client.InitInterface()

}
