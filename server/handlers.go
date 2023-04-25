package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"pizza/data"
	"pizza/helper"
	"strconv"
	"time"
)

func getCustomerOrders(c *gin.Context) {
	idCustomer := c.Param("id")
	id, _ := strconv.Atoi(idCustomer)

	helper.ReadJSONFile("data/order.json", &data.Orders)

	var customerOrders []data.Order

	for _, v := range data.Orders {
		if v.UserID == id {
			customerOrders = append(customerOrders, v)
		}
	}
	c.JSON(http.StatusOK, customerOrders)
}

func getMenu(c *gin.Context) {
	helper.ReadJSONFile("data/menu.json", &data.Menu)
	c.IndentedJSON(http.StatusOK, &data.Menu)
}

func getOrders(c *gin.Context) {
	helper.ReadJSONFile("data/menu.json", &data.Orders)
	c.IndentedJSON(http.StatusOK, &data.Orders)
}

func getUsers(c *gin.Context) {
	helper.ReadJSONFile("data/menu.json", &data.Users)
	c.IndentedJSON(http.StatusOK, &data.Users)
}

func createPizza(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price parameter"})
		return
	}

	file, err := os.Open("data/menu.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open menu file"})
		return
	}
	defer file.Close()

	var menu []data.Pizza
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read menu file"})
		return
	}

	pizzaCounter := len(menu) + 1
	newPizza := data.Pizza{ID: pizzaCounter, Name: name, Description: description, Price: price}
	menu = append(menu, newPizza)

	file, err = os.Create("data/menu.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu file"})
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to menu file"})
		return
	}

	c.JSON(http.StatusOK, newPizza)
}

func createOrder(c *gin.Context) {
	idPizzaInt := c.PostForm("pizza")
	idUserInt := c.PostForm("IdUser")
	idPizza, _ := strconv.Atoi(idPizzaInt)
	idUser, _ := strconv.Atoi(idUserInt)

	file, err := os.Open("data/order.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open order file"})
		return
	}
	defer file.Close()

	var orders []data.Order
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read order file"})
		return
	}

	orderCounter := len(orders) + 1
	newOrder := data.Order{ID: orderCounter, UserID: idUser, PizzaID: idPizza, Status: "preparing"}

	orders = append(orders, newOrder)

	file, err = os.Create("data/order.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order file"})
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to orders file"})
		return
	}

	StartOrderTicker(newOrder.ID)

}

func deletePizza(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	_, i := helper.GetPizzaById(id)
	data.Menu = append(data.Menu[:i], data.Menu[i+1:]...)
	c.JSON(http.StatusOK, data.Menu)

	file, err := os.Create("data/menu.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu file"})
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data.Menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to menu file"})
		return
	}
}

func cancelOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var targetOrder *data.Order
	for i, order := range data.Orders {
		if order.ID == orderID {
			if MySessionName.Token == "usertoken" && order.Status == "ready_to_be_delivered" {
				c.AbortWithStatusJSON(http.StatusFailedDependency, nil)
				return
			}
			targetOrder = &data.Orders[i]
			//orders[i].Status = "cancelled"
			break
		}
	}

	if targetOrder == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	StopOrderTicker(orderID)

	targetOrder.Status = "cancelled"

	file, err := os.Create("data/order.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order file"})
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data.Orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to orders file"})
		return
	}

	c.JSON(http.StatusOK, targetOrder)
}

func getOrderStatus(c *gin.Context) {

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	for _, order := range data.Orders {
		if order.ID == orderID && order.UserID == MySessionName.ID {
			c.JSON(http.StatusOK, order.Status)
			return
		}
	}

}

func registerUser(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	address := c.PostForm("address")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	password = string(hashedPassword)

	file, err := os.Open("data/users.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open users file"})
		return
	}
	defer file.Close()

	var users []data.User
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read users file"})
		return
	}

	userCounter := len(users) + 1
	newUser := data.User{ID: userCounter, Name: name, Email: email, Password: password, Address: address, Token: "usertoken"}
	users = append(users, newUser)

	file, err = os.Create("data/users.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu file"})
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to menu file"})
		return
	}

	c.JSON(http.StatusOK, newUser)
}

func loginUser(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	for _, v := range data.Users {
		if v.Name == name {

			pass := bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(password))
			if pass == nil {
				MySessionName = &v
				c.JSON(http.StatusOK, v)
			} else {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
			}

		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
	return
}

func StartOrderTicker(orderID int) {
	ticker := time.NewTicker(10 * time.Second)
	OrderTickers[orderID] = ticker

	go func() {
		<-ticker.C
		for i := range data.Orders {
			if data.Orders[i].ID == orderID {
				data.Orders[i].Status = "ready_to_be_delivered"
				file, err := os.Create("data/order.json")
				if err != nil {
					log.Println("Failed to create order file:", err)
					return
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				if err := encoder.Encode(data.Orders); err != nil {
					log.Println("Failed to write to orders file:", err)
					return
				}
				break
			}
		}
		ticker.Stop()

		// Update order status to "delivered" after 2 minutes
		ticker = time.NewTicker(10 * time.Second)
		OrderTickers[orderID] = ticker
		<-ticker.C
		for i := range data.Orders {
			if data.Orders[i].ID == orderID {
				data.Orders[i].Status = "delivered"
				// Write updated orders slice to orders file
				file, err := os.Create("data/order.json")
				if err != nil {
					log.Println("Failed to create order file:", err)
					return
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				if err := encoder.Encode(data.Orders); err != nil {
					log.Println("Failed to write to orders file:", err)
					return
				}
				break
			}
		}
		ticker.Stop()
		delete(OrderTickers, orderID)
	}()
}

func StopOrderTicker(orderID int) {
	ticker, ok := OrderTickers[orderID]
	if !ok {
		return
	}
	ticker.Stop()
	delete(OrderTickers, orderID)
}
