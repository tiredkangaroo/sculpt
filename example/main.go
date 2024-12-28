package main

import (
	"fmt"
	"log"
	"sculpt"
)

type Task struct {
	ID    int `pk:"true" autoincrement:"true"`
	Title sculpt.Optional[string]
}

func main() {
	// connect
	if err := sculpt.Connect("postgres://postgres:@localhost:5432/sculpt_example"); err != nil {
		log.Fatal(err.Error())
	}
	// new
	task, err := sculpt.New[Task]()
	if err != nil {
		log.Fatal(err.Error())
	}
	// create
	err = task.Create()
	if err != nil {
		log.Fatal(err.Error())
	}

	// save
	t := Task{
		Title: sculpt.OptionalValue("arrested development"),
	}
	if err := task.Save(t); err != nil {
		log.Fatal(err.Error())
	}

	// query
	tasks, err := task.Query().Conditions(
		sculpt.EqualsTo("id", "2"),
	).Do()
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, t := range tasks {
		fmt.Println(t)
	}
}
