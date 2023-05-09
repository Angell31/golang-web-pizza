package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"pizza/data"
	"strconv"
	"time"
)

var MySessionName data.User

func getCustomerOrders(c *gin.Context) {
	idCustomer := c.Param("id")
	id, _ := strconv.Atoi(idCustomer)

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("mojabaza").Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var orders []data.Order
	for cursor.Next(ctx) {
		var menuItem data.Order
		err := cursor.Decode(&menuItem)
		if err != nil {
			log.Fatal(err)
		}
		if menuItem.UserID == id {
			orders = append(orders, menuItem)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	data.Orders = orders
	c.IndentedJSON(http.StatusOK, orders)
}

func getMenu(c *gin.Context) {
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("mojabaza").Collection("menu")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var menu []data.Pizza
	for cursor.Next(ctx) {
		var menuItem data.Pizza
		err := cursor.Decode(&menuItem)
		if err != nil {
			log.Fatal(err)
		}
		menu = append(menu, menuItem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
	data.Menu = menu
	c.IndentedJSON(http.StatusOK, menu)
}

func getOrders(c *gin.Context) {
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("mojabaza").Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var orders []data.Order
	for cursor.Next(ctx) {
		var menuItem data.Order
		err := cursor.Decode(&menuItem)
		if err != nil {
			log.Fatal(err)
		}
		orders = append(orders, menuItem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	data.Orders = orders
	c.IndentedJSON(http.StatusOK, orders)
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

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("menu")
	count, _ := collection.CountDocuments(ctx, bson.M{})
	pizzaCounter := int(count) + 1

	newPizza := data.Pizza{ID: pizzaCounter, Name: name, Description: description, Price: price}
	_, err = collection.InsertOne(ctx, newPizza)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, newPizza)
}

func createOrder(c *gin.Context) {
	idPizzaInt := c.PostForm("pizza")
	idUserInt := c.PostForm("IdUser")
	idPizza, _ := strconv.Atoi(idPizzaInt)
	idUser, _ := strconv.Atoi(idUserInt)

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("orders")
	count, _ := collection.CountDocuments(ctx, bson.M{})
	orderCounter := int(count) + 1

	newOrder := data.Order{ID: orderCounter, UserID: idUser, PizzaID: idPizza, Status: "preparing"}
	_, err = collection.InsertOne(ctx, newOrder)
	if err != nil {
		log.Fatal(err)
	}
	StartOrderTicker(newOrder.ID)
	c.JSON(http.StatusOK, newOrder)

}

func deletePizza(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("menu")
	_, err = collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var menu []data.Pizza
	if err = cursor.All(ctx, &menu); err != nil {
		log.Fatal(err)
	}

	data.Menu = menu

	c.JSON(http.StatusOK, data.Menu)
}

func cancelOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("orders")

	var order data.Order
	err = collection.FindOne(ctx, bson.M{"id": orderID}).Decode(&order)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
		return
	}

	if MySessionName.ID != order.UserID && MySessionName.Token == "usertoken" {
		c.AbortWithStatusJSON(http.StatusFailedDependency, nil)
		return
	}

	if MySessionName.Token == "usertoken" && order.Status == "ready_to_be_delivered" {
		c.AbortWithStatusJSON(http.StatusFailedDependency, nil)
		return
	}

	StopOrderTicker(orderID)

	filter := bson.M{"id": orderID}
	update := bson.M{"$set": bson.M{"status": "cancelled"}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order in database"})
		return
	}

	c.JSON(http.StatusOK, order)

}

func getOrderStatus(c *gin.Context) {

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("orders")

	var order data.Order
	err = collection.FindOne(ctx, bson.M{"id": orderID}).Decode(&order)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
		return
	}

	c.JSON(http.StatusOK, order.Status)

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

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("users")
	count, _ := collection.CountDocuments(ctx, bson.M{})
	userCounter := int(count) + 1

	newUser := data.User{ID: userCounter, Name: name, Email: email, Password: password, Address: address, Token: "usertoken"}
	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, newUser)
}

func loginUser(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mojabaza").Collection("users")

	var user data.User
	err = collection.FindOne(ctx, bson.M{"name": name}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
		return
	}

	pass := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if pass == nil {
		MySessionName = user
		c.JSON(http.StatusOK, user)
		return
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad login"})
		return
	}

}

func StartOrderTicker(orderID int) {
	ticker := time.NewTicker(15 * time.Second)
	OrderTickers[orderID] = ticker

	go func() {
		<-ticker.C
		clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		collection := client.Database("mojabaza").Collection("orders")

		// Find the order by ID and update its status field
		filter := bson.M{"id": orderID}
		update := bson.M{"$set": bson.M{"status": "ready_to_be_delivered"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Fatal(err)
		}

		ticker.Stop()

		// Update order status to "delivered" after 2 minutes
		ticker = time.NewTicker(15 * time.Second)
		OrderTickers[orderID] = ticker

		<-ticker.C
		clientOptions = options.Client().ApplyURI("mongodb://mongodb:27017")
		client, err = mongo.NewClient(clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		collection = client.Database("mojabaza").Collection("orders")

		// Find the order by ID and update its status field
		filter = bson.M{"id": orderID}
		update = bson.M{"$set": bson.M{"status": "delivered"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Fatal(err)
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
