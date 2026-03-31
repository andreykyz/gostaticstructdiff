package models

type UserDiff struct {
	ID struct {
		Value int
		Set   bool
	}
	Username struct {
		Value string
		Set   bool
	}
	Email struct {
		string
		Set bool
	}
	Active struct {
		Value bool
		Set   bool
	}
}
