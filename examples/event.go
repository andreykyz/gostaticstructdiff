package examples

import "time"

type Event struct {
	ID        int       `structtomap:"id"`
	Name      string    `structtomap:"name"`
	CreatedAt time.Time `structtomap:"created_at"`
	UpdatedAt time.Time `structtomap:"updated_at"`
}
