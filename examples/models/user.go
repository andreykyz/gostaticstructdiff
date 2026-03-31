package models

type User struct {
	ID       int    `structtomap:"id"`
	Username string `structtomap:"username"`
	Email    string `structtomap:"email"`
	Active   bool   `structtomap:"active"`
}
