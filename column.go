package sculpt

import (
	"fmt"
	"strconv"
)

// IntegerField
type IntegerField struct{}

func (i IntegerField) String() string {
	return "IntegerField"
}

// TextField
type TextField struct {
	MaximumLength uint
}

func (t TextField) String() string {
	return "TextField"
}

// BooleanField
type BooleanField struct{}

func (b BooleanField) String() string {
	return "BooleanField"
}

// ReferenceField
type ReferenceField struct {
	References *Model
	OnDelete   string
}

func (r ReferenceField) String() string {
	return "ReferenceField"
}

type Field interface {
	String() string
}

type Column struct {
	model       *Model
	PRIMARY_KEY bool
	UNIQUE      bool
	NULLABLE    bool
	Name        string
	Kind        Field
	Validations []string
}

func compareColumns(oldC []*Column, newC []*Column) (additions []Column, alterations []string, deletions []Column) {
	additionsMap := make(map[string]Column)
	deletionsMap := make(map[string]Column)
	for _, c := range oldC {
		deletionsMap[c.Name] = *c
	}

	for _, c := range newC {
		tc, found := deletionsMap[c.Name]
		if found { // it means its present in both old and new
			delete(deletionsMap, c.Name)
			if tc.NULLABLE != c.NULLABLE {
				sd := ""
				if c.NULLABLE {
					sd = "DROP"
				} else {
					sd = "SET"
					var input string
					fmt.Printf("Migrator: %s is becoming a NOT NULL column. Please provide a default value for rows that have this column set to NULL.", c.Name)
					_, err := fmt.Scan(&input)
					if err != nil {
						panic(err)
					}
					ActiveDB.Execute("UPDATE $1 SET $2 = '$3' WHERE $4 IS NULL;", c.model.Name, c.Name, input, c.Name)
				}
				alterations = append(alterations, fmt.Sprintf(`ALTER COLUMN "%s" %s NOT NULL`, c.Name, sd))
			}
			if tc.PRIMARY_KEY != c.PRIMARY_KEY {
				if c.PRIMARY_KEY {
					alterations = append(alterations, fmt.Sprintf(`DROP CONSTRAINT %s_pkey`, c.model.Name))
				} else {
					alterations = append(alterations, fmt.Sprintf(`ADD PRIMARY KEY ("%s")`, c.Name))
				}
			}
			if tc.Kind != c.Kind {
				defaultCase := ""
				if c.Kind.String() == "IntegerField" { //new kind is integer field
					var input string
					fmt.Printf("Migrator: %s is changing types from %s to integerfield. Default integer value: ", c.Name, tc.Kind.String())
					_, err := fmt.Scan(&input)
					if err != nil {
						panic(err)
					}
					i, err := strconv.Atoi(input)
					if err != nil {
						panic("the default integer value you provided was not an integer.")
					}
					defaultCase = fmt.Sprintf(" USING CASE WHEN %s ~ '^[0-9]+$' THEN %s::integer ELSE %d END", c.Name, c.Name, i)
				}
				newType, _ := kindToSQL(c.Kind)
				alterations = append(alterations, fmt.Sprintf(`ALTER COLUMN "%s" TYPE %s%s`, c.Name, newType, defaultCase))
			}
			if tc.UNIQUE != c.UNIQUE {
				if c.UNIQUE {
					if c.NULLABLE == false {
						s := "Migrator: Column %s is becoming unique however the column contains a not null constraint. By default, Sculpt sets duplicate values to null then applies the unique constraint to avoid errors. Please set nullable to true (if there are no duplicates, nothing will be set to null!)."
						LogError(s)
						continue
					}

					statement := fmt.Sprintf(`WITH cte AS (SELECT ctid, "%s", ROW_NUMBER() OVER (PARTITION BY "%s" ORDER BY ctid) AS rn FROM "%s")
							UPDATE "%s"
							SET "%s" = NULL
							FROM cte
							WHERE "%s".ctid = cte.ctid AND cte.rn > 1;`, c.Name, c.Name, c.model.Name, c.model.Name, c.Name, c.Name)
					_, err := ActiveDB.Execute(statement)
					if err != nil {
						panic(err)
					}

					alterations = append(alterations, fmt.Sprintf(`ADD UNIQUE("%s")`, c.model.Name))
				} else {
					alterations = append(alterations, fmt.Sprintf("DROP CONSTRAINT %s_%s_key", c.model.Name, c.Name))
				}
			}
		} else { // only present in new
			additionsMap[c.Name] = *c
		}
	}

	for _, v := range additionsMap {
		additions = append(additions, v)
	}
	for _, v := range deletionsMap {
		deletions = append(deletions, v)
	}
	return
}
