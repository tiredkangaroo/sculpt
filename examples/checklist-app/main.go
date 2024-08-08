package main

import (
	"fmt"

	"github.com/tiredkangaroo/sculpt"
)

type User struct {
	ID             int
	Username       string
	Email          string `validators:"email"`
	HashedPassword string
}

func main() {
	sculpt.SetLogLevel(sculpt.DEBUG)

	err := sculpt.Connect("postgres", "", "checklistapp")
	if err != nil {
		sculpt.LogError(err.Error())
		return
	}
	sculpt.RegisterValidator("email", sculpt.EmailValidator)
	user := sculpt.Register(new(User))
	err = user.Save()
	if err != nil {
		sculpt.LogError(err.Error())
		return
	}

	user.Migrate()
	seed := []*sculpt.Row{
		user.NewNE(&User{
			ID:             0,
			Username:       "ajiteshkumar",
			Email:          "ajiteshkumar@smtp.local",
			HashedPassword: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		}),
		user.NewNE(&User{
			ID:             1,
			Username:       "testuser1",
			Email:          "testuser1@smtp.local",
			HashedPassword: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
		}),
		user.NewNE(&User{
			ID:             2,
			Username:       "testuser2",
			Email:          "helloworld@monster",
			HashedPassword: "0000",
		}),
	}
	sculpt.Seed(seed...)

	users, err := sculpt.RunQuery[*User](user, sculpt.Query{
		Columns:  []string{},
		Distinct: true,
		Conditions: []sculpt.Condition{
			sculpt.EqualTo("ID", 0),
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(users[0])
}
