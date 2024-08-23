package main

import (
	"fmt"
	"time"

	"github.com/tiredkangaroo/sculpt"
)

type User struct {
	PrimaryKeyID   int `kind:"IntegerField"`
	Username       string
	Email          string `validators:"email"`
	HashedPassword string
}
type Task struct {
	User  *User `on_delete:"cascade"`
	Title string
}

func main() {
	sculpt.SetLogLevel(sculpt.DEBUG)

	err := sculpt.Connect("postgres://postgres:@localhost:5432/checklistapp_test?sslmode=disable")
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

	task := sculpt.Register(new(Task))
	err = task.Save()
	if err != nil {
		sculpt.LogError(err.Error())
		return
	}

	row1 := user.NewNE(&User{
		PrimaryKeyID:   1,
		Username:       "ajiteshkumar",
		Email:          "ajiteshkumar@smtp.local",
		HashedPassword: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
	})
	row2 := user.NewNE(&User{
		PrimaryKeyID:   2,
		Username:       "testuser1",
		Email:          "testuser1@smtp.local",
		HashedPassword: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
	})
	row3 := user.NewNE(&User{
		PrimaryKeyID:   3,
		Username:       "testuser2",
		Email:          "helloworld@smtp.local",
		HashedPassword: "0000",
	})
	row4 := task.NewNE(&Task{
		User: &User{
			PrimaryKeyID: 1,
		},
		Title: "laundry",
	})
	sculpt.Seed(row4, row1, row2, row3)

	// sculpt.RunQuery[*Task](task, sculpt.Query{})

	start := time.Now()
	tasks, err := sculpt.RunQuery[*Task](task, sculpt.Query{
		Conditions: []sculpt.Condition{
			sculpt.EqualTo(`User`, &User{
				PrimaryKeyID: 1,
			}),
		},
	})
	fmt.Println(time.Since(start).Microseconds())
	if err != nil {
		sculpt.LogError(err.Error())
		return
	}

	fmt.Println(tasks[0].User)
}
