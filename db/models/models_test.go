package models

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	unique := "the-very-unique-name"
	for _, test := range []struct {
		name string
		in   string
		want string
	}{
		{"spaces", "The Very Unique Name", unique},
		{"multiple spaces", "  tHe    VeRy     UnIqUe     nAmE   ", unique},
		{"dots", "...tHe..VeRy.UnIqUe....nAmE...", unique},
		{"commas", "tHe,VeRy,,,UnIqUe,,,,nAmE,", unique},
		{"slashes", "tHe\\VeRy\\\\UnIqUe/\\/\\/\\nAmE\\\\//\\\\//\\\\//", unique},
		{"percent", `tHe%VeRy%%UnIqUe%nAmE%%`, unique},
		{"octochorpe", "###tHe##VeRy#UnIqUe###nAmE####", unique},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := Slugify(test.in)
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
		})
	}
}
