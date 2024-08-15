package main

import (
	"fmt"
	"time"

	"github.com/tiredkangaroo/sculpt"
)

type User struct {
	PublicKeyHex   string
	ID             int `kind:"IntegerField"`
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
	one := 1
	two := 2
	three := 3
	row1 := user.NewNE(&User{
		PublicKeyHex:   "040cb72f6902e9f29773c594a63ad4ef690aefde9d53c0764e6c532622fd58566fe62609c699d2b58d41d0e5ca2e425577dd69b824b5e3e3891bb5caa818fd75c5",
		ID:             one,
		Username:       "ajiteshkumar",
		Email:          "ajiteshkumar@smtp.local",
		HashedPassword: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
	})
	row2 := user.NewNE(&User{
		PublicKeyHex:   "0447471e608fadc63e9667d8e90b44432df6330deb168e371f7f48e599eba5c0222ce06ab5411d6383097badb006ccf5473e972c674dbd0bc66fe0ab278e4fb7cb5498af91f3664366a5f37ba56114bfe039fb1a896a1e47ab990a7f9726edf7f1",
		ID:             two,
		Username:       "testuser1",
		Email:          "testuser1@smtp.local",
		HashedPassword: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
	})
	row3 := user.NewNE(&User{
		PublicKeyHex:   "049a62d65eb6111ced9af5a3ab4c82c918e1ef444b9cedbc8621fab246d8d4b7ceb19a979a70a0404101a70e508762b02c4d08ffdfc66dc057e08f46d1e2dd285f8ada6bada4630c793561b900f4f78c1e4a915ecd919c28f2765154df1b40462c",
		ID:             three,
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
