package gomail

import (
	"fmt"
	"io"
	"log"
	nurl "net/url"
	"regexp"
	"strconv"

	"gopkg.in/gomail.v2"
	gomailPkg "gopkg.in/gomail.v2"

	"github.com/djangulo/sfd/mail"
)

type Gomail struct {
	host     string
	port     int
	username string
	password string
	dialer   *gomailPkg.Dialer
	tlsCert  string
	tlsKey   string
}

func init() {
	g := &Gomail{}
	mail.Register("gomail", g)
}

var parseRe = regexp.MustCompile(`^gomail://([\w\.\-\_\@]+):(.*)@([\w\-\.\_]+):(\d{1,5}).*`)

// Open parses a url in the format
// gomail://user:password@host:port/?tls-cert=/path/to/cert&tls-key=path/to/key
func (g *Gomail) Open(urlString string) (mail.Driver, error) {
	url, err := nurl.Parse(urlString)
	if err != nil {
		return nil, err
	}
	if url.Scheme != "gomail" {
		return nil, fmt.Errorf("mail::gomail - wrong scheme, should be \"gomail\", got %s", url.Scheme)
	}
	match := parseRe.FindStringSubmatch(urlString)
	g.username = match[1]
	g.password = match[2]
	g.host = match[3]
	n, err := strconv.Atoi(match[4])
	if err != nil {
		return nil, err
	}
	g.port = n

	return g, nil
}

// Close noop, as the connection is not persisted.
func (g *Gomail) Close() error {
	return nil
}

func (g *Gomail) SendMail(contentType, subject, body, from string, recipients ...mail.Recipient) error {

	d := gomailPkg.NewDialer(g.host, g.port, g.username, g.password)
	s, err := d.Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	for _, recipient := range recipients {
		m.SetHeader("From", from)
		m.SetAddressHeader("To", recipient.Address, recipient.Name)
		m.SetHeader("Subject", subject)

		m.SetBody(contentType, body)

		if err := gomailPkg.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", recipient.Address, err)
		}
		m.Reset()
	}

	return nil
}

func (g *Gomail) SendMultipart(subject, txtBody, htmlBody, from string, recipients ...mail.Recipient) error {

	d := gomailPkg.NewDialer(g.host, g.port, g.username, g.password)
	s, err := d.Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	for _, recipient := range recipients {
		m.SetHeader("From", from)
		m.SetAddressHeader("To", recipient.Address, recipient.Name)
		m.SetHeader("Subject", subject)

		m.SetBody("text/plain", txtBody)
		m.AddAlternative("text/html", htmlBody)

		if err := gomailPkg.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", recipient.Address, err)
		}
		m.Reset()
	}

	return nil
}

func (g *Gomail) SendTemplate(
	contentType, tplName string,
	data interface{},
	subject, from string,
	recipients ...mail.Recipient) error {

	d := gomailPkg.NewDialer(g.host, g.port, g.username, g.password)
	s, err := d.Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	for _, recipient := range recipients {
		m.SetHeader("From", from)
		m.SetAddressHeader("To", recipient.Address, recipient.Name)
		m.SetHeader("Subject", subject)

		m.AddAlternativeWriter(contentType, func(w io.Writer) error {
			return mail.Execute(tplName, w, data)
		})

		if err := gomailPkg.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", recipient.Address, err)
		}
		m.Reset()
	}

	return nil
}

func (g *Gomail) SendMultipartTemplate(
	txtTpl string,
	txtData interface{},
	htmlTpl string,
	htmlData interface{},
	subject, from string,
	recipients ...mail.Recipient) error {

	d := gomailPkg.NewDialer(g.host, g.port, g.username, g.password)
	s, err := d.Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	for _, recipient := range recipients {
		m.SetHeader("From", from)
		m.SetAddressHeader("To", recipient.Address, recipient.Name)
		m.SetHeader("Subject", subject)

		m.AddAlternativeWriter("text/plain", func(w io.Writer) error {
			return mail.Execute(txtTpl, w, txtData)
		})

		m.AddAlternativeWriter("text/html", func(w io.Writer) error {
			return mail.Execute(htmlTpl, w, htmlData)
		})

		if err := gomailPkg.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", recipient.Address, err)
		}
		m.Reset()
	}

	return nil
}
