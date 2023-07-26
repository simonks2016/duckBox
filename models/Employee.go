package models

type Employee struct {
	Id       string `orm:"pk"`
	Name     string
	Icon     string
	Account  string
	Password string
	Power    string
	State    int
}
