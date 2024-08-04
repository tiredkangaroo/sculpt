package sculpt

import (
	"testing"
)

type ExampleModel1 struct {
	ID           string `kind:"IDField"`
	Name         string `kind:"TextField"`
	AddressLine1 string `kind:"TextField"`
	AddressLine2 string `kind:"TextField"`
	City         string `kind:"TextField"`
	PostalCode   int    `kind:"IntegerField"`
}

var exm1 *Model = NewModel(new(ExampleModel1))

func TestIn(t *testing.T) {
	got := In("user1", "user1", "user2")
	t.Log(got)
}

func handle_err(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Errorf(err.Error())
	t.FailNow()
}

func TestSeed(t *testing.T) {
	err := Connect("postgres", "", "example")
	handle_err(t, err)

	return
	err = exm1.Save()
	handle_err(t, err)
	m := exm1

	row1, err := m.NewRow(&ExampleModel1{
		ID:           "AAA-BBB-CCC-DDD",
		Name:         "Ajitesh Kumar",
		AddressLine1: "One Venus Way",
		AddressLine2: "",
		City:         "Dallas",
		PostalCode:   01234,
	})

	row2, err := m.NewRow(&ExampleModel1{
		ID:           "EEE-FFF-GGG-HHH",
		Name:         "School Bus",
		AddressLine1: "3 Jupiter Way",
		AddressLine2: "",
		City:         "Brentwood",
		PostalCode:   43321,
	})

	row3, err := m.NewRow(&ExampleModel1{
		ID:           "BAB-ABA-BAB-ABA",
		Name:         "Wheels On The",
		AddressLine1: "17-43 Earth Street",
		AddressLine2: "",
		City:         "New York",
		PostalCode:   12321,
	})

	Seed(row1, row2, row3)
}

func TestRunQuery(t *testing.T) {
	query := Query{
		Conditions: []Condition{
			EqualTo("Name", "Ajitesh Kumar"),
		},
	}
	got, err := RunQuery(exm1, new(ExampleModel1), query)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
		return
	}
	t.Log(*got[0])
}
