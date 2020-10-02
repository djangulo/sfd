//Package console is a simple os.Stdout implementation of the mail.Mailer interface.
// Console additionally writes all emails to `os.Tempdir()/sfd_email.txt` so as to be captured
// by tests if needed.
package console

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/djangulo/sfd/mail"
)

type Console struct {
	out   io.Writer
	quiet bool
}

func init() {
	c := &Console{out: os.Stdout}
	mail.Register("console", c)
}

const format = "%v:\t%v\n"

func (c *Console) Open(urlString string) (mail.Driver, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if quiet := q.Get("quiet"); quiet != "" {
		c.quiet = true
	}
	return c, nil
}

func (c *Console) Close() error {
	return nil
}

func (c *Console) SendMail(contentType, subject, body, from string, recipients ...mail.Recipient) error {
	switch contentType {
	case "text/html":
		c.write(subject, "", body, from, recipients...)
	default:
		c.write(subject, body, "", from, recipients...)
	}
	return nil
}

// SendMail sends an email with the parameters passed to the recipients passed.
func (c *Console) SendMultipart(subject, txtBody, htmlBody, from string, recipients ...mail.Recipient) error {
	c.write(subject, txtBody, htmlBody, from, recipients...)
	return nil
}

// SendMultiPart email
func (c *Console) SendTemplate(contentType, tplName string, data interface{}, subject, from string, recipients ...mail.Recipient) error {
	switch contentType {
	case "text/html":
		c.writeTpl("", nil, tplName, data, subject, from, recipients...)
	default:
		c.writeTpl(tplName, data, "", nil, subject, from, recipients...)
	}

	return nil
}

func (c *Console) SendMultipartTemplate(txtTpl string,
	txtData interface{},
	htmlTpl string,
	htmlData interface{},
	subject, from string, recipients ...mail.Recipient) error {
	c.writeTpl(txtTpl, txtData, htmlTpl, htmlData, subject, from, recipients...)

	return nil
}

func (c *Console) writeTmp(subject string, data []byte) {
	subject = strings.ToLower(strings.ReplaceAll(subject, " ", "_"))
	path := filepath.Join(os.TempDir(), "sfd_"+subject+".txt")
	fh, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	_, err = fh.Write(data)
	if err != nil {
		panic(err)
	}
}

func (c *Console) write(subject, txtBody, htmlBody, from string, recipients ...mail.Recipient) {
	var b strings.Builder
	tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "From", from)
	var to = make([]string, 0)
	for _, r := range recipients {
		to = append(to, fmt.Sprintf("%s <%s>", r.Name, r.Address))
	}

	fmt.Fprintf(tw, format, "To", strings.Join(to, "; "))
	fmt.Fprintf(tw, format, "Subject", subject)
	if txtBody == "" {
		fmt.Fprintf(tw, format, "Body", htmlBody)
	} else if htmlBody == "" {
		fmt.Fprintf(tw, format, "Body", txtBody)
	} else {
		fmt.Fprintf(tw, format, "Text body", txtBody)
		fmt.Fprintf(tw, format, "HTML body", htmlBody)
	}
	tw.Flush()
	if !c.quiet {
		fmt.Fprint(c.out, b.String())
	}
	c.writeTmp(subject, []byte(b.String()))

}

func (c *Console) writeTpl(
	txtTpl string,
	txtData interface{},
	htmlTpl string,
	htmlData interface{},
	subject, from string, recipients ...mail.Recipient) {
	var b strings.Builder

	tw := new(tabwriter.Writer).Init(&b, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "From", from)

	var to = make([]string, 0)
	for _, r := range recipients {
		to = append(to, fmt.Sprintf("%s <%s>", r.Name, r.Address))
	}

	fmt.Fprintf(tw, format, "To", strings.Join(to, "; "))
	fmt.Fprintf(tw, format, "subject", subject)

	if txtTpl == "" {
		fmt.Fprint(tw, "Body:\t")
		if err := mail.Execute(htmlTpl, tw, htmlData); err != nil {
			panic(err)
		}
		fmt.Fprint(tw, "\n")
	} else if htmlTpl == "" {
		fmt.Fprint(tw, "Body:\t")
		if err := mail.Execute(txtTpl, tw, txtData); err != nil {
			panic(err)
		}
		fmt.Fprint(tw, "\n")
	} else {
		fmt.Fprint(tw, "Text body:\n")
		if err := mail.Execute(txtTpl, tw, txtData); err != nil {
			panic(err)
		}
		fmt.Fprint(tw, "\n")
		fmt.Fprint(tw, "HTML body:\n")
		if err := mail.Execute(htmlTpl, tw, htmlData); err != nil {
			panic(err)
		}
		fmt.Fprint(tw, "\n")
	}

	tw.Flush()
	if !c.quiet {
		fmt.Fprint(c.out, b.String())
	}
	c.writeTmp(subject, []byte(b.String()))
}
