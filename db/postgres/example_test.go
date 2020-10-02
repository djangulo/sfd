package postgres

import "fmt"

func ExampleBuildValueStrings() {
	type Obj struct {
		A string `db:"A"`
		B int    `db:"B"`
		C int    `db:"C"`
	}
	objs := []*Obj{
		{A: "a", B: 1},
		{A: "b", B: 2},
		{A: "c", B: 3},
		{A: "d", B: 4},
	}
	// 2 fields, we're not inserting C
	valueStrings := BuildValueStrings(len(objs), 2)
	args := make([]interface{}, 0) // fed into sql.DB.Query(stmt, args...)
	for _, obj := range objs {
		args = append(args, obj.A)
		args = append(args, obj.B)
	}

	stmt := fmt.Sprintf(`INSERT INTO objects_table (a, b)
VALUES
%s;`, valueStrings)
	fmt.Println(stmt)
	// Output: INSERT INTO objects_table (a, b)
	//VALUES
	//($1,$2),
	//($3,$4),
	//($5,$6),
	//($7,$8);
}
