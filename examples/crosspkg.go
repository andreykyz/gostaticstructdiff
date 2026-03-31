package examples

import "github.com/andreykyz/gostaticstructdiff/examples/models"

type Account struct {
	ID      int         `structtomap:"id"`
	Owner   models.User `structtomap:"owner"`
	Balance float64     `structtomap:"balance"`
}
