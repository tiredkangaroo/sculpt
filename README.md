# sculpt
type-safe abstraction for defining models and querying data
```golang
package main

import (
	"log"
	"sculpt"
)

type User struct {
	ID          int `pk:"true" autoincrement:"true"`
	Name        string
	PhoneNumber sculpt.Optional[string]
}

func main() {
	err := sculpt.Connect("<postgres_connection_uri>")

	// create a new model
	userModel, err := sculpt.New[User]()

	// query the model for every user but John Doe
	users, err := userModel.Query().Conditions(
		sculpt.NotEqualsTo("Name", "John Doe"),
	).Do()

	// loop through the users, welcoming them
	for _, user := range users {
		log.Printf("Welcome, %s (ID: %d)!", user.Name, user.ID)
	}

	// save a new user
	err = userModel.Save(User{
		Name:        "Ajitesh Kumar",
		PhoneNumber: sculpt.OptionalValue("123-456-7890"),
	})
}
```

Sculpt provides a high-level, type-safe abstraction for defining models and querying data, as well as reducing query and modeling boilerplate.
