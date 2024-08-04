package sculpt

import (
	"testing"
)

type ExampleModel1 struct {
	ID   int    `kind:"IntegerField"`
	Name string `kind:"TextField"`
}

func TestNewModel(t *testing.T) {
	err := Connect("postgres", "", "example")
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
		return
	}
	got := NewModel(new(ExampleModel1))
	err = got.Save()
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
		return
	}
	got2, err := got.NewRow(&ExampleModel1{
		Name: "hello world",
	})
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
		return
	}
	got2.Save()
	t.Log(got.Name)
}
