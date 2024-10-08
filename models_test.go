package sculpt

import (
	"testing"
)

type ExampleModel struct {
	Name int
}

var examplemodel *Model

func handle_err(t *testing.T, err error) {
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
}

func RequireConnection() {
	if Connected() == false {
		Connect("postgres://postgres:@localhost:5432/checklistapp_test?sslmode=disable")
	}
}
func RequireModels() {
	RequireConnection()
	if examplemodel == nil {
		examplemodel = Register(new(ExampleModel))
		examplemodel.Save()
	}
}

// func TestMigrate(t *testing.T) {
// 	RequireModels()
//
// 	handle_err(t, err)
// }
