package client

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func InitInterface() {
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
			if SessionUser.Name == "" {
				fmt.Println("Bad login. Try again!")
				break
			}
			if SessionUser.Token == "usertoken" {
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

func userInterface() {
	for {
		fmt.Printf("\nWelcome to GoPizza, %v!\n\n", SessionUser.Name)
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
		fmt.Printf("\nWelcome to GoPizza, %v!\n\n", SessionUser.Name)
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
