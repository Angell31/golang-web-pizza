package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
)

func CheckServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Printf("Error resolving TCP address: %v", err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err == nil {
		listener.Close()
		initServer()
	}
}

func initServer() {

	router := gin.Default()

	router.GET("/menu", getMenu)
	router.GET("/order", getOrders)
	router.GET("/order/:id/status", getOrderStatus)
	router.GET("/orders/:id", getCustomerOrders)

	router.POST("/login", loginUser)

	router.POST("/menu", createPizza)
	router.POST("/order", createOrder)

	router.POST("/register", registerUser)

	router.DELETE("/menu/:id", deletePizza)
	router.DELETE("/order/:id", cancelOrder)

	router.Run("localhost:8080")
}
