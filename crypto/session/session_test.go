package session

import (
	"context"
	"errors"
	"testing"
)

func TestFromContext(t *testing.T) {
	for _, tt := range []struct {
		name string
		want Session
		err  error
	}{
		{"found", &sessionObject{id: "abcd1234", values: Values{"testkey": "testval"}}, nil},
		{"not found", nil, ErrNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != nil {
				ctx := context.WithValue(context.Background(), CtxKey, tt.want)

				ses, err := FromContext(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if ses.ID() != tt.want.ID() {
					t.Errorf("expected %v got %v", tt.want.ID(), ses.ID())
				}
			} else { // error check
				ctx := context.Background()

				_, err := FromContext(ctx)
				if err == nil {
					t.Errorf("expected an error but didn't get one")
				}
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("expected %q but got %q", ErrNotFound.Error(), err.Error())
				}
			}
		})
	}
}

// func TestValues(t *testing.T) {
// 	t.Run("encode/decode", func(t *testing.T) {
// 		// for _, tt := range []struct {
// 		// 	in interface{}
// 		// 	out interface{}
// 		// } {
// 		// 	{&models.User{Username:"testuser"}, &models.User{Username:"testuser"}}
// 		// }
// 		user := &models.User{Username: "testuser"}
// 		v1 := Values{}
// 		v1.Set("user", user)
// 		var b bytes.Buffer
// 		enc := gob.NewEncoder(&b)
// 		if err := enc.Encode(v1); err != nil {
// 			t.Errorf("unexpected error: %v", err)
// 		}

// 		v2 := Values{}
// 		bv := bytes.NewBuffer(b.Bytes())
// 		dec := gob.NewDecoder(bv)
// 		if err := dec.Decode(&v2); err != nil {
// 			t.Errorf("unexpected error: %v", err)
// 		}
// 		u1 := v1.Get("user").(*models.User)
// 		u2 := v2.Get("user").(*models.User)
// 		if u1.Username != u2.Username {
// 			t.Errorf("expected %q got %q", u2.Username, u1.Username)
// 		}
// 	})
// }
