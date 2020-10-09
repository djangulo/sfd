package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/djangulo/sfd/config"
	testutils "github.com/djangulo/sfd/testing"
)

func TestNew(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	values := Values{"testkey": "testval"}
	ses, err := man.New(values)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	ss, ok := ms[ses.ID()]
	if !ok {
		t.Errorf("expected mockstore to contain session, it did not")
	}
	if ss.ID() != ses.ID() {
		t.Errorf("expected id %q got %q", ses.ID(), ss.ID())
	}
	t.Run("with nil values", func(t *testing.T) {
		ses, err := man.New(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		ss, ok := ms[ses.ID()]
		if !ok {
			t.Errorf("expected mockstore to contain session, it did not")
		}
		if ss.ID() != ses.ID() {
			t.Errorf("expected id %q got %q", ses.ID(), ss.ID())
		}
	})
}

func TestDelete(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	values := Values{"testkey": "testval"}
	ses, err := man.New(values)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := man.Delete(ses.ID()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_, ok := ms[ses.ID()]
	if ok {
		t.Errorf("expected mockstore not to contain session, it did")
	}
}

func TestGet(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	values := Values{"testkey": "testval"}
	ses, err := man.New(values)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if v := ses.Get("testkey").(string); v != "testval" {
		t.Errorf("expected 'testval' got %q", v)
	}
	ss, ok := ms[ses.ID()]
	if !ok {
		t.Errorf("expected mockstore to contain session, it did not")
	}
	if ss.ID() != ses.ID() {
		t.Errorf("expected id %q got %q", ses.ID(), ss.ID())
	}
}

func TestSave(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	values := Values{"testkey": "testval"}
	ses, err := man.New(values)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err = ses.Set("testkey2", "testval2"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := man.Save(ses); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	ss, ok := ms[ses.ID()]
	if !ok {
		t.Errorf("expected mockstore to contain session, it did not")
	}
	if ss.ID() != ses.ID() {
		t.Errorf("expected id %q got %q", ses.ID(), ss.ID())
	}
	if v := ss.Get("testkey2").(string); v != "testval2" {
		t.Errorf("expected 'testval2' got %q", v)
	}
}

func TestContext(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	chain := man.Context(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"response": "OK"})
	}))
	t.Run("no cookie", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		chain.ServeHTTP(rec, req)

		res := rec.Result()
		testutils.AssertJSON(t, res)
		var buf []byte
		var wantErr = bytes.NewBuffer(buf)
		json.NewEncoder(wantErr).Encode(`{"status": "session not available, redirect to login"}`)
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected %d got %d", http.StatusUnauthorized, res.StatusCode)
		}
	})
	t.Run("has cookie", func(t *testing.T) {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{
			Jar: jar,
		}
		ts := httptest.NewServer(chain)
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		ses, err := man.New(Values{"testkey": "testvalue"})
		if err != nil {
			t.Fatal(err)
		}

		cookie, err := man.NewAuthCookie(ses, "")
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(cookie)

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		var want = &struct {
			Response string `json:"response"`
		}{}

		if res.StatusCode != 200 {
			t.Errorf("expected %d got %d", 200, res.StatusCode)
		}
		testutils.AssertJSON(t, res)

		testutils.DecodeJSON(t, res, want)
		if want.Response != "OK" {
			t.Errorf("expected %q got %q", "OK", want.Response)
		}
	})
}

func TestNoErrContext(t *testing.T) {
	ms := make(mockStore)
	cnf := config.NewConfig()
	cnf.Defaults()
	man, err := NewManager(ms, "test-session", 60, cnf)
	if err != nil {
		t.Fatal(err)
	}

	chain := man.NoErrContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"response": "OK"})
	}))
	t.Run("no error without cookie", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		chain.ServeHTTP(rec, req)

		res := rec.Result()
		var want = &struct {
			Response string `json:"response"`
		}{}

		if res.StatusCode != 200 {
			t.Errorf("expected %d got %d", 200, res.StatusCode)
		}
		testutils.AssertJSON(t, res)

		testutils.DecodeJSON(t, res, want)
		if want.Response != "OK" {
			t.Errorf("expected %q got %q", "OK", want.Response)
		}
	})
	t.Run("no error with cookie", func(t *testing.T) {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{
			Jar: jar,
		}
		ts := httptest.NewServer(chain)
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		ses, err := man.New(Values{"testkey": "testvalue"})
		if err != nil {
			t.Fatal(err)
		}

		cookie, err := man.NewAuthCookie(ses, "")
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(cookie)

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		var want = &struct {
			Response string `json:"response"`
		}{}

		if res.StatusCode != 200 {
			t.Errorf("expected %d got %d", 200, res.StatusCode)
		}
		testutils.AssertJSON(t, res)

		testutils.DecodeJSON(t, res, want)
		if want.Response != "OK" {
			t.Errorf("expected %q got %q", "OK", want.Response)
		}
	})
}

type mockStore map[string]Session

func (ms mockStore) NewSession(ses Session) error {
	id, _, _ := ses.Metadata()
	ms[id] = ses
	return nil
}
func (ms mockStore) ReadSession(id string) ([]byte, error) {
	ses, ok := ms[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return ses.Bytes()
}

// DeleteToken explicitly deletes a token from the store.
func (ms mockStore) DeleteSession(id string) error {
	delete(ms, id)
	return nil
}

func (ms mockStore) UpdateSession(ses Session) error {
	id, _, _ := ses.Metadata()
	ms[id] = ses
	return nil
}
func (ms mockStore) SessionGC() error {
	for k, v := range ms {
		if v.Expiry().Before(time.Now()) {
			delete(ms, k)
		}
	}
	return nil
}
