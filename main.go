package main

import (
	"pizza/client"
	"pizza/server"
)

func main() {

	server.CheckServer()
	//data.ConnectDatabase
	client.InitInterface()
	//data.DisconnectDatabase()

}
