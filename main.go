package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var sessionUser User
var mySessionName *User
var OrderTickers = make(map[int]*time.Ticker)

type Pizza struct {
	ID          int
	Name        string
	Description string
	Price       float64
}

type Order struct {
	ID      int
	UserID  int
	PizzaID int
	Status  string // "in_progress", "ready_to_be_delivered", "delivered", "cancelled"
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	Address  string
	Token    string
}

var menu []Pizza
var users []User
var orders []Order

func getCustomerOrders(c *gin.Context) {
	idCustomer := c.Param("id")
	id, _ := strconv.Atoi(idCustomer)

	readJSONFile("data/order.json", &orders)
	var customerOrders []Order

	for _, v := range orders {
		if v.UserID == id {
			customerOrders = append(customerOrders, v)
		}
	}
	c.JSON(http.StatusOK, customerOrders)
}

func getMenu(c *gin.Context) {
	readJSONFile("data/menu.json", &menu)
	c.IndentedJSON(http.StatusOK, menu)
}

func getOrders(c *gin.Context) {
	readJSONFile("data/order.json", &orders)
	c.IndentedJSON(http.StatusOK, orders)
}

func getUsers(c *gin.Context) {
	readJSONFile("data/users.json", &users)
	c.IndentedJSON(http.StatusOK, users)
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

	var menu []Pizza
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read menu file"})
		return
	}

	pizzaCounter := len(menu) + 1
	newPizza := Pizza{ID: pizzaCounter, Name: name, Description: description, Price: price}
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

	var orders []Order
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read order file"})
		return
	}

	orderCounter := len(orders) + 1
	newOrder := Order{ID: orderCounter, UserID: idUser, PizzaID: idPizza, Status: "preparing"}

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

	startOrderTicker(newOrder.ID)

}

func startOrderTicker(orderID int) {
	ticker := time.NewTicker(10 * time.Second)
	OrderTickers[orderID] = ticker

	go func() {
		<-ticker.C
		for i := range orders {
			if orders[i].ID == orderID {
				orders[i].Status = "ready_to_be_delivered"
				file, err := os.Create("data/order.json")
				if err != nil {
					log.Println("Failed to create order file:", err)
					return
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				if err := encoder.Encode(orders); err != nil {
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
		for i := range orders {
			if orders[i].ID == orderID {
				orders[i].Status = "delivered"
				// Write updated orders slice to orders file
				file, err := os.Create("data/order.json")
				if err != nil {
					log.Println("Failed to create order file:", err)
					return
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				if err := encoder.Encode(orders); err != nil {
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

func stopOrderTicker(orderID int) {
	ticker, ok := OrderTickers[orderID]
	if !ok {
		return
	}
	ticker.Stop()
	delete(OrderTickers, orderID)
}

func deletePizza(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	_, i := getPizzaById(id)
	menu = append(menu[:i], menu[i+1:]...)
	c.JSON(http.StatusOK, menu)

	file, err := os.Create("data/menu.json")
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
}

func cancelOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var targetOrder *Order
	for i, order := range orders {
		if order.ID == orderID {
			if mySessionName.Token == "usertoken" && order.Status == "ready_to_be_delivered" {
				c.AbortWithStatusJSON(http.StatusFailedDependency, nil)
				return
			}
			targetOrder = &orders[i]
			//orders[i].Status = "cancelled"
			break
		}
	}

	if targetOrder == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	stopOrderTicker(orderID)

	targetOrder.Status = "cancelled"

	file, err := os.Create("data/order.json")
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

	c.JSON(http.StatusOK, targetOrder)
}

func getOrderStatus(c *gin.Context) {

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	for _, order := range orders {
		if order.ID == orderID && order.UserID == mySessionName.ID {
			c.JSON(http.StatusOK, order.Status)
			return
		}
	}

}

func getPizzaById(id int) (*Pizza, int) {
	for i, o := range orders {
		if o.ID == id {
			return &menu[i], i
		}
	}
	return nil, 0
}

func readJSONFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open menu file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&v); err != nil {
		log.Fatalf("Failed to parse menu file: %v", err)
	}
	return nil
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

	var users []User
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read users file"})
		return
	}

	userCounter := len(users) + 1
	newUser := User{ID: userCounter, Name: name, Email: email, Password: password, Address: address, Token: "usertoken"}
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

	for _, v := range users {
		if v.Name == name {

			pass := bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(password))
			if pass == nil {
				mySessionName = &v
				c.JSON(http.StatusOK, v)
			} else {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
			}

		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
	return
}

func initServer() {

	readJSONFile("data/menu.json", &menu)
	readJSONFile("data/order.json", &orders)
	readJSONFile("data/users.json", &users)

	router := gin.Default()

	router.GET("/menu", getMenu)
	router.GET("/order", getOrders)
	router.GET("/users", getUsers)
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

func getMenuClient() {
	url := "http://localhost:8080/menu"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var menu []Pizza
	if err := json.NewDecoder(resp.Body).Decode(&menu); err != nil {
		panic(err)
	}

	for _, pizza := range menu {
		fmt.Printf("ID: %d | Name: %s | Description: %s | Price: $%.2f\n", pizza.ID, pizza.Name, pizza.Description, pizza.Price)
	}
}

func getAllOrders() {
	url := "http://localhost:8080/order"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		panic(err)
	}

	for _, order := range orders {
		fmt.Printf("ID: %v | UserID: %v | PizzaID: %v | Status: %s\n", order.ID, order.UserID, order.PizzaID, order.Status)
	}
}

func getClientOrders() error {

	url := fmt.Sprintf("http://localhost:8080/orders/%v", sessionUser.ID)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to get customers orders. Status code: %d", resp.StatusCode)
	}

	var customerOrders []Order
	json.NewDecoder(resp.Body).Decode(&customerOrders)

	for _, order := range customerOrders {
		fmt.Printf("ID: %v | UserID: %v | PizzaID: %v | Status: %s\n", order.ID, order.UserID, order.PizzaID, order.Status)
	}

	readJSONFile("data/order.json", &orders)
	return nil

}

func createPizzaRequest(name string, description string, price float64) error {

	formData := url.Values{}
	formData.Set("name", name)
	formData.Set("description", description)
	formData.Set("price", strconv.FormatFloat(price, 'f', 2, 64))

	resp, err := http.PostForm("http://localhost:8080/menu", formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create pizza: %v", resp.Status)
	}
	return nil
}

func createPizzaClient() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter pizza name:")
	scanner.Scan()
	name := scanner.Text()

	fmt.Println("Enter pizza description:")
	scanner.Scan()
	desc := scanner.Text()

	fmt.Println("Enter pizza price:")
	scanner.Scan()
	price := scanner.Text()

	// convert the price to a float
	fPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		fmt.Println("Invalid price entered")
		os.Exit(1)
	}
	err = createPizzaRequest(name, desc, fPrice)
	if err != nil {
		return
	}
}

func registerRequest(name string, email string, password string, address string) error {
	formData := url.Values{}
	formData.Set("name", name)
	formData.Set("email", email)
	formData.Set("password", password)
	formData.Set("address", address)

	resp, err := http.PostForm("http://localhost:8080/register", formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create user: %v", resp.Status)
	}
	return nil
}

func registerClient() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter your name:")
	scanner.Scan()
	name := scanner.Text()

	fmt.Println("Enter your email:")
	scanner.Scan()
	email := scanner.Text()

	fmt.Println("Enter your password:")
	scanner.Scan()
	password := scanner.Text()

	fmt.Println("Enter your address:")
	scanner.Scan()
	address := scanner.Text()

	err := registerRequest(name, email, password, address)
	if err != nil {
		return
	}

}

func loginRequest(name string, password string) {
	formData := url.Values{}
	formData.Set("name", name)
	formData.Set("password", password)
	resp, _ := http.PostForm("http://localhost:8080/login", formData)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&sessionUser)
	return
}

func loginClient() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter your name:")
	scanner.Scan()
	name := scanner.Text()

	fmt.Println("Enter your password:")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	password := string(bytePassword)
	loginRequest(name, password)

}

func signOut() {
	sessionUser = User{ID: 0, Name: "", Email: "", Password: "", Address: "", Token: ""}
	mySessionName = nil
	return
}

func deletePizzaRequest(id int) error {
	url := fmt.Sprintf("http://localhost:8080/menu/%d", id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete pizza with ID %d", id)
	}

	var menu []Pizza
	if err := json.NewDecoder(resp.Body).Decode(&menu); err != nil {
		return err
	}

	fmt.Printf("Pizza with ID %d successfully deleted\n", id)
	return nil
}

func deletePizzaClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter pizza id :")
	scanner.Scan()
	id := scanner.Text()
	idInt, _ := strconv.Atoi(id)
	deletePizzaRequest(idInt)
}

func createOrderRequest(pizza string) error {
	formData := url.Values{}
	formData.Set("pizza", pizza)
	formData.Set("IdUser", strconv.Itoa(sessionUser.ID))

	fmt.Println(sessionUser.ID)
	resp, _ := http.PostForm("http://localhost:8080/order", formData)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create order: %v", resp.Status)
	}

	return nil
}

func createOrderClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter pizza id :")
	scanner.Scan()
	id := scanner.Text()
	createOrderRequest(id)
}

func cancelOrderRequest(id int) error {

	url := fmt.Sprintf("http://localhost:8080/order/%d", id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFailedDependency {
		fmt.Printf("Your order is ready to be delivered, you cannot cancel it.")
	}

	var ret []Order
	json.NewDecoder(resp.Body).Decode(&ret)

	fmt.Println(ret)

	return nil
}

func cancelOrderClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter order id :")
	scanner.Scan()
	idS := scanner.Text()
	id, _ := strconv.Atoi(idS)
	cancelOrderRequest(id)
}

func checkOrderStatusRequest(id string) error {

	url := fmt.Sprintf("http://localhost:8080/order/%v/status", id)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to get order status. Status code: %d", resp.StatusCode)
	}

	var status string
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	fmt.Printf("Status of order %v : %v \n", id, status)
	return nil
}

func checkOrderStatusClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter order id :")
	scanner.Scan()
	idS := scanner.Text()
	checkOrderStatusRequest(idS)
}

func checkServer() {
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

func userInterface() {
	for {
		fmt.Printf("\nWelcome to GoPizza, %v!\n\n", sessionUser.Name)
		fmt.Println("1. List menu")
		fmt.Println("2. Order pizza")
		fmt.Println("3. List my orders")
		fmt.Println("4. Status of order")
		fmt.Println("5. Cancel order")
		fmt.Printf("6. Sign out\n\n")

		fmt.Println("Enter option:")
		byteOpt, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		opt := string(byteOpt)

		switch opt {
		case "1":
			getMenuClient()
		case "2":
			createOrderClient()
		case "3":
			getClientOrders()
		case "4":
			checkOrderStatusClient()
		case "5":
			cancelOrderClient()
		case "6":
			signOut()
			return
		}
	}
}

func adminInterface() {
	for {
		fmt.Printf("\nWelcome to GoPizza, %v!\n\n", sessionUser.Name)
		fmt.Println("1. List menu")
		fmt.Println("2. Add pizza")
		fmt.Println("3. DeletePizza")
		fmt.Println("4. List all orders")
		fmt.Println("5. Cancel order")
		fmt.Printf("6. Sign out\n\n")

		fmt.Println("Enter option:")
		byteOpt, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		opt := string(byteOpt)

		switch opt {
		case "1":
			getMenuClient()
		case "2":
			createPizzaClient()
		case "3":
			deletePizzaClient()
		case "4":
			getAllOrders()
		case "5":
			cancelOrderClient()
		case "6":
			signOut()
			return
		}
	}
}

func main() {

	checkServer()

	for {
		fmt.Println("\nWelcome to GoPizza!\n")
		fmt.Println("1. Login")
		fmt.Println("2. Register")
		fmt.Printf("3. Exit\n\n")

		fmt.Println("Enter option:")
		byteOpt, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
		opt := string(byteOpt)

		switch opt {
		case "1":
			loginClient()
			if sessionUser.Name == "" {
				fmt.Println("Bad login. Try again!")
				break
			}
			if sessionUser.Token == "usertoken" {
				userInterface()
			} else {
				adminInterface()
			}
		case "2":
			registerClient()
		case "3":
			return
		}
	}

}
