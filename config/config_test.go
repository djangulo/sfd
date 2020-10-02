package config

import "testing"

func TestFromFile(t *testing.T) {
	c := &Config{}

	err := c.FromFile("config.conf")
	if err != nil {
		t.Errorf("didn't expect an error but got one %v", err)
	}
}

func TestSiteHost(t *testing.T) {
	c := NewConfig()
	c.Defaults()

	want := "localhost"
	got := c.SiteHost()
	if got != want {
		t.Errorf("expected %q got %q", want, got)
	}
}
