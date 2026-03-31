package examples

type Person struct {
	Name    string  `structtomap:"name"`
	Age     int     `structtomap:"age"`
	Address Address `structtomap:"address"`
}

type Address struct {
	City    string `structtomap:"city"`
	Country string `structtomap:"country"`
}
