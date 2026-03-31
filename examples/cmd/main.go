package main

import (
	"fmt"
	"log"
	"time"

	"github.com/andreykyz/gostaticstructdiff/examples"
	"github.com/andreykyz/gostaticstructdiff/examples/models"
)

func main() {
	fmt.Println("=== Go Struct to HashMap Generator Demo ===")

	// 1. Person
	fmt.Println("\n--- Person ---")
	person := examples.Person{
		Name: "Alice",
		Age:  30,
		Address: examples.Address{
			City:    "Moscow",
			Country: "Russia",
		},
	}
	personMap := examples.PersonToMap(person)
	fmt.Printf("PersonToMap: %#v\n", personMap)
	personBack, err := examples.MapToPerson(personMap)
	if err != nil {
		log.Fatalf("MapToPerson error: %v", err)
	}
	fmt.Printf("MapToPerson result: %+v\n", personBack)

	// 2. Event
	fmt.Println("\n--- Event ---")
	event := examples.Event{
		ID:        42,
		Name:      "Conference",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now().Add(2 * time.Hour),
	}
	eventMap := examples.EventToMap(event)
	fmt.Printf("EventToMap: %#v\n", eventMap)
	eventBack, err := examples.MapToEvent(eventMap)
	if err != nil {
		log.Fatalf("MapToEvent error: %v", err)
	}
	fmt.Printf("MapToEvent result: %+v\n", eventBack)

	// 3. ComplexStruct
	fmt.Println("\n--- ComplexStruct ---")
	complex := examples.ComplexStruct{
		Name:   "Test",
		Count:  5,
		Active: true,
		Tags:   []string{"go", "generator", "demo"},
		Users: []models.User{
			{ID: 1, Username: "user1", Email: "user1@example.com", Active: true},
			{ID: 2, Username: "user2", Email: "user2@example.com", Active: false},
		},
		Metadata: map[string]models.Metadata{
			"meta1": {
				Label: "First",
				Values: map[string]string{
					"key1": "value1",
				},
				Score: 9.5,
				Extra: struct {
					Note string  `structtomap:"note"`
					Cost float64 `structtomap:"cost"`
				}{
					Note: "note1",
					Cost: 100.0,
				},
			},
		},
		Inner: struct {
			Title string `structtomap:"title"`
			Value int    `structtomap:"value"`
		}{
			Title: "Inner Title",
			Value: 777,
		},
		Ref: &models.User{
			ID:       99,
			Username: "admin",
			Email:    "admin@example.com",
			Active:   true,
		},
		Categories: map[string][]string{
			"cat1": {"a", "b", "c"},
			"cat2": {"x", "y"},
		},
	}
	complexMap := examples.ComplexStructToMap(complex)
	fmt.Printf("ComplexStructToMap (partial): %#v\n", limitMapDepth(complexMap, 2))
	complexBack, err := examples.MapToComplexStruct(complexMap)
	if err != nil {
		log.Fatalf("MapToComplexStruct error: %v", err)
	}
	fmt.Printf("MapToComplexStruct result: Name=%s, Count=%d, Active=%v\n",
		complexBack.Name, complexBack.Count, complexBack.Active)

	// 4. Account (cross‑package)
	fmt.Println("\n--- Account (cross‑package) ---")
	account := examples.Account{
		ID: 1001,
		Owner: models.User{
			ID:       1,
			Username: "john_doe",
			Email:    "john@example.com",
			Active:   true,
		},
		Balance: 1234.56,
	}
	accountMap := examples.AccountToMap(account)
	fmt.Printf("AccountToMap: %#v\n", accountMap)
	accountBack, err := examples.MapToAccount(accountMap)
	if err != nil {
		log.Fatalf("MapToAccount error: %v", err)
	}
	fmt.Printf("MapToAccount result: ID=%d, Balance=%.2f, Owner.Username=%s\n",
		accountBack.ID, accountBack.Balance, accountBack.Owner.Username)

	fmt.Println("\n=== All conversions completed successfully ===")
}

// limitMapDepth recursively limits the depth of map printing for readability.
func limitMapDepth(m map[string]any, depth int) map[string]any {
	if depth <= 0 {
		return map[string]any{"...": "depth exceeded"}
	}
	result := make(map[string]any)
	for k, v := range m {
		if nested, ok := v.(map[string]any); ok {
			result[k] = limitMapDepth(nested, depth-1)
		} else if nestedSlice, ok := v.([]any); ok {
			// handle slices of maps
			var limitedSlice []any
			for _, item := range nestedSlice {
				if nestedMap, ok := item.(map[string]any); ok {
					limitedSlice = append(limitedSlice, limitMapDepth(nestedMap, depth-1))
				} else {
					limitedSlice = append(limitedSlice, item)
				}
			}
			result[k] = limitedSlice
		} else {
			result[k] = v
		}
	}
	return result
}
