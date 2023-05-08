package client

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"pizza/data"
	"pizza/server"
	"strconv"
)

func checkOrderStatusClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter order id :")
	scanner.Scan()
	idS := scanner.Text()
	checkOrderStatusRequest(idS)
}

func cancelOrderClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter order id :")
	scanner.Scan()
	idS := scanner.Text()
	id, _ := strconv.Atoi(idS)
	cancelOrderRequest(id)
}

func createOrderClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter pizza id :")
	scanner.Scan()
	id := scanner.Text()
	createOrderRequest(id)
}

func deletePizzaClient() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter pizza id :")
	scanner.Scan()
	id := scanner.Text()
	idInt, _ := strconv.Atoi(id)
	deletePizzaRequest(idInt)
}

func signOut() {
	SessionUser = data.User{ID: 0, Name: "", Email: "", Password: "", Address: "", Token: ""}
	server.MySessionName = data.User{0, "", "", "", "", ""}
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
