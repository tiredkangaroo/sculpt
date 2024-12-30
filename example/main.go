package main

import (
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/tiredkangaroo/sculpt"
)

type ExampleUser struct {
	CreatedAt time.Time

	ID    int `pk:"true" autoincrement:"true"`
	Name  string
	Email sculpt.Optional[string] `validators:"email"`
}

func handleError(err error, m string, a ...any) {
	if err != nil {
		log.Fatalf(fmt.Sprintf("error: %s. "+m, append([]any{err.Error()}, a...)))
	}
}

func main() {
	err := sculpt.Connect("postgres://postgres:@localhost:5432/sculpt_example?sslmode=disable")
	defer sculpt.Close()
	handleError(err, "connect error")

	// register validators
	sculpt.RegisterValidator("email", func(e string) (err error) {
		fmt.Println("Validating email:", e)
		_, err = mail.ParseAddress(e)
		return
	})

	// create a new model
	userModel, err := sculpt.New[ExampleUser]()
	handleError(err, "new model error")

	handleError(userModel.Create(), "create table error")

	// query the model for every user but John Doe
	users, err := userModel.Query().Conditions(
		sculpt.NotEqualsTo("Name", "John Doe"),
	).Do()
	handleError(err, "query error")

	// loop through the users, welcoming them
	for _, user := range users {
		log.Printf("Welcome, %s (ID: %d)! You've been here for: %v.", user.Name, user.ID, time.Since(user.CreatedAt))
	}

	// save a new user
	err = userModel.Save(ExampleUser{
		CreatedAt: time.Now(),
		Name:      "Ajitesh Kumar",
		Email:     sculpt.OptionalValue("ajinest6@gmail.com"),
	})
	handleError(err, "save error")

}
