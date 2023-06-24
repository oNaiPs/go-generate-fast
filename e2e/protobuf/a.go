//go:generate protoc -I . --go_out=. --go_opt=paths=import person.proto  folder/gender.proto
//

package main

import (
	"fmt"
)

func main() {
	// Create a new Person message
	person := &example.Person{
		Id:        1,
		Name:      "John Doe",
		Age:       30,
		Interests: []string{"Programming", "Hiking"},
	}

	// Print the person object
	fmt.Println(person)
}
