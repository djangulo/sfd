package validators

import (
	"testing"
)

func TestCleanPhoneNumber(t *testing.T) {
	for _, test := range []struct {
		in, want string
	}{
		{"+1 (555)-555-5555", "+15555555555"},
		{"+1 (555)-555-5555 x 99123", "+15555555555x99123"},
		{"+1 (555)-555-5555 ext 12321", "+15555555555x12321"},
		{"+1 (555)-555-5555ext12301", "+15555555555x12301"},
		{"+1 (555)-555-5555x13132", "+15555555555x13132"},
		{"+1      (555)-555-5555", "+15555555555"},
		{"+1           (555)   --  555      5555              x 2132", "+15555555555x2132"},
	} {
		t.Run(test.in, func(t *testing.T) {
			got := cleanPhone(test.in)
			if got != test.want {
				t.Errorf("expected %q got %q", test.want, got)
			}
		})
	}
}
