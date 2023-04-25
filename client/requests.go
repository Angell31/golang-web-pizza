package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pizza/data"
	"pizza/helper"
	"strconv"
)

func getMenuClient() {
	url := "http://localhost:8080/menu"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var menu []data.Pizza
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

	var orders []data.Order
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

	var customerOrders []data.Order
	json.NewDecoder(resp.Body).Decode(&customerOrders)

	for _, order := range customerOrders {
		fmt.Printf("ID: %v | UserID: %v | PizzaID: %v | Status: %s\n", order.ID, order.UserID, order.PizzaID, order.Status)
	}

	helper.ReadJSONFile("data/order.json", &data.Orders)
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

	var menu []data.Pizza
	if err := json.NewDecoder(resp.Body).Decode(&menu); err != nil {
		return err
	}

	fmt.Printf("Pizza with ID %d successfully deleted\n", id)
	return nil
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

	var ret []data.Order
	json.NewDecoder(resp.Body).Decode(&ret)

	fmt.Println(ret)

	return nil
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
