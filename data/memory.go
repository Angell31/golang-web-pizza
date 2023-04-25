package data

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

var Menu []Pizza
var Users []User
var Orders []Order
