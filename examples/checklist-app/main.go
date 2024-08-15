package main

import (
	"fmt"
	"time"

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
	user.Delete()
	user.Save()

	start := time.Now()
	row1 := user.NewNE(&User{
		ID:             1,
		Username:       "ajiteshkumar",
		Email:          "ajiteshkumar@smtp.local",
		HashedPassword: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
	})
	row2 := user.NewNE(&User{
		ID:             2,
		Username:       "testuser1",
		Email:          "testuser1@smtp.local",
		HashedPassword: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
	})
	row3 := user.NewNE(&User{
		ID:             3,
		Username:       "testuser2",
		Email:          "helloworld@smtp.local",
		HashedPassword: "0000",
	})
	err = row1.Save()
	if err != nil {
		sculpt.LogError(err.Error())
		return
	}
	err2 := row2.Save()
	if err2 != nil {
		sculpt.LogError(err2.Error())
		return
	}
	err3 := row3.Save()
	if err3 != nil {
		sculpt.LogError(err3.Error())
		return
	}

	fmt.Println(time.Since(start).Microseconds())
}
