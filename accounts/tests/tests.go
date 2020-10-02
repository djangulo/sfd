package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi"

	pkg "github.com/djangulo/sfd/accounts"
	"github.com/djangulo/sfd/db/mock"
	"github.com/djangulo/sfd/db/models"
	testutils "github.com/djangulo/sfd/testing"
)

var (
	users []*models.User
)

func init() {
	users = mock.Users()
}

func Test(t *testing.T, s *pkg.Server) {
	t.Run("TestGetUser", func(t *testing.T) { TestGetUser(t, s) })
	t.Run("TestPasswordResetInit", func(t *testing.T) { TestPasswordResetInit(t, s) })
	t.Run("TestCheckPassResetToken", func(t *testing.T) { TestCheckPassResetToken(t, s) })
	t.Run("TestPasswordResetConfirm", func(t *testing.T) { TestPasswordResetConfirm(t, s) })
}

func TestGetUser(t *testing.T, s *pkg.Server) {
	chain := s.UserContext(http.HandlerFunc(pkg.GetUser))
	// need the router for the chi.URLParam extraction
	r := chi.NewRouter()
	r.Get(`/{userID:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}`, chain.ServeHTTP)
	r.Get(`/{username:[a-z0-9-._]{1,255}}`, chain.ServeHTTP)

	getByPublic := func(profilePublic bool) *models.User {
		for _, u := range users {
			if u.ProfilePublic == profilePublic {
				return u
			}
		}
		return nil
	}

	ts := httptest.NewTLSServer(r)
	defer ts.Close()
	client := ts.Client()
	t.Run("can get user", func(t *testing.T) {
		t.Run("by ID", func(t *testing.T) {
			want := getByPublic(true)
			res, err := client.Get(ts.URL + "/" + want.ID.String())
			if err != nil {
				t.Fatal(err)
			}
			var user = new(models.User)
			testutils.AssertJSON(t, res)
			testutils.DecodeJSON(t, res, user)
			testutils.AssertUsersEqual(t, user, want)
		})
		t.Run("by username", func(t *testing.T) {
			want := getByPublic(true)
			res, err := client.Get(ts.URL + "/" + want.Username)
			if err != nil {
				t.Fatal(err)
			}
			var user = new(models.User)
			testutils.AssertJSON(t, res)
			testutils.DecodeJSON(t, res, user)
			testutils.AssertUsersEqual(t, user, want)
		})

	})
}

func TestPasswordResetInit(t *testing.T, s *pkg.Server) {
	chain := http.HandlerFunc(s.PasswordResetInit)

	ts := httptest.NewTLSServer(chain)
	defer ts.Close()
	client := ts.Client()
	t.Run("success", func(t *testing.T) {
		var want = pkg.CallBackResponse{Status: "OK", Message: "reset email sent"}
		for _, tt := range []struct {
			name, usernameOrEmail string
		}{
			{"success with existing user username", users[0].Username},
			{"success with existing user email", users[0].Email},
			{"success with non-existing user username", "non-existing-user"},
			{"success with non-existing user email", "non-existing-email@emailland.com"},
		} {
			t.Run(tt.name, func(t *testing.T) {
				req := &pkg.PassResetInitRequest{
					UsernameOrEmail: tt.usernameOrEmail,
				}
				b, err := json.Marshal(req)
				if err != nil {
					t.Fatal(err)
				}
				res, err := client.Post(
					ts.URL,
					"application/json; charset=utf-8",
					bytes.NewReader(b),
				)
				if err != nil {
					t.Fatal(err)
				}

				testutils.AssertJSON(t, res)
				var got = new(pkg.CallBackResponse)
				testutils.DecodeJSON(t, res, got)
				if got.Status != want.Status {
					t.Errorf("expected status %q got %q", want.Status, got.Status)
				}
			})
		}
	})
}

func TestCheckPassResetToken(t *testing.T, s *pkg.Server) {
	r := chi.NewRouter()
	r.Post("/init", http.HandlerFunc(s.PasswordResetInit))
	r.Get("/check", http.HandlerFunc(s.CheckPassResetToken))

	ts := httptest.NewTLSServer(r)
	defer ts.Close()
	client := ts.Client()
	// generate an email and extract the token from said email
	// in order to use it in the test server
	extractToken := func(t *testing.T, ts *httptest.Server, username string) string {
		req := &pkg.PassResetInitRequest{UsernameOrEmail: username}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Post(
			ts.URL+"/init",
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		b, err = ioutil.ReadFile(filepath.Join(os.TempDir(), "sfd_password_reset.txt"))
		if err != nil {
			t.Fatal(err)
		}
		re := regexp.MustCompile(`\?token=([\w\d-]+)`)
		match := re.FindAllStringSubmatch(string(b), -1)
		return match[0][1]
	}

	verifyToken := extractToken(t, ts, users[0].Username)

	res, err := client.Get(ts.URL + "/check?token=" + verifyToken)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertJSON(t, res)
	var got = new(pkg.ValidationTokenResponse)
	testutils.DecodeJSON(t, res, got)
	// ensure tokens are different
	if strings.Contains(verifyToken, got.Token) || strings.Contains(got.Token, verifyToken) {
		t.Errorf("tokens are equal: %q %q", verifyToken, got.Token)
	}
	// ensure original token is no longer valid
	res, err = client.Get(ts.URL + "/check?token=" + verifyToken)
	if err != nil {
		t.Fatal(err)
	}
	bb, _ := ioutil.ReadAll(res.Body)
	if !strings.Contains(string(bb), "token is invalid") {
		t.Errorf("unexpected error in %q", string(bb))
	}
}

func TestPasswordResetConfirm(t *testing.T, s *pkg.Server) {
	r := chi.NewRouter()
	r.With(s.UserContext).Get("/users/{username}", pkg.GetUser)
	r.Post("/init", s.PasswordResetInit)
	r.Get("/check", s.CheckPassResetToken)
	r.Post("/confirm", s.PasswordResetConfirm)

	ts := httptest.NewTLSServer(r)
	defer ts.Close()
	client := ts.Client()
	// generate an email and extract the token from said email
	// in order to use it in the test server
	extractToken := func(t *testing.T, ts *httptest.Server, username string) string {
		req := &pkg.PassResetInitRequest{UsernameOrEmail: username}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Post(
			ts.URL+"/init",
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		b, err = ioutil.ReadFile(filepath.Join(os.TempDir(), "sfd_password_reset.txt"))
		if err != nil {
			t.Fatal(err)
		}
		re := regexp.MustCompile(`\?token=([\w\d-]+)`)
		match := re.FindAllStringSubmatch(string(b), -1)

		res, err := client.Get(ts.URL + "/check?token=" + match[0][1])
		if err != nil {
			t.Fatal(err)
		}
		var got = new(pkg.ValidationTokenResponse)
		testutils.DecodeJSON(t, res, got)
		return got.Token
	}
	t.Run("success", func(t *testing.T) {

		req := &pkg.PassResetRequest{
			Token:          extractToken(t, ts, users[0].Username),
			Password:       "newPassword12345",
			RepeatPassword: "newPassword12345",
		}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Post(
			ts.URL+"/confirm",
			"application/json; charset=utf-8",
			bytes.NewReader(b),
		)
		if err != nil {
			t.Fatal(err)
		}
		match, err := s.ComparePassword(&users[0].ID, req.Password)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !match {
			t.Error("passwords do not match")
		}

	})

}
