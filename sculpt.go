package sculpt

import (
	// "fmt"
	"sculpt/Log"
	"sculpt/Manager"
	"sculpt/Models"
)

func main() {
	err := Manager.Connect("postgres", "password", "example")
	if err != nil {
		log.Error("Line 10: " + err.Error())
		return
	}
	test_model := models.Model{
		Name: "test_model",
		Fields: []interface{}{
			models.IDField{
				PRIMARY_KEY: true,
				UNIQUE:      true,
				Name:        "id",
				Auto:        true,
			},
			models.TextField{
				Name:           "name",
				Minimum_Length: 2,
				Maximum_Length: 30,
			},
			models.IntegerField{
				Name: "cool_number",
			},
		},
	}
	err = test_model.Create(true)
	if err != nil {
		log.Error("Line 34: " + err.Error())
		return
	}
	new_model, err := test_model.New(0, "jimmy", 14)
	if err != nil {
		log.Error("Line 41: " + err.Error())
		return
	}
	err = new_model.Save()
	if err != nil {
		log.Error("Line 46: " + err.Error())
		return
	}
	_, err = test_model.Get(&models.Query{DISTINCT: true, Columns: []string{"name"}})
	if err != nil {
		log.Error("Line 53: " + err.Error())
		return
	}
	err = test_model.Delete(map[string]interface{}{"id": "KLXNDEYFBE12"})
	if err != nil {
		log.Error("Line 57: " + err.Error())
		return
	}
	// fmt.Println(result)
}
