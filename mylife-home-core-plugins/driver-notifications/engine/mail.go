package engine

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

func NewMailChannel(smtpServer string, user string, pass string, to []string) Channel {
	return &mailChannel{smtpServer, user, pass, to}
}

type mailChannel struct {
	host string
	user string
	pass string
	to   []string
}

func (channel *mailChannel) Send(title string, text string) {
	go channel.sendSync(title, text)
}

func (channel *mailChannel) sendSync(title string, text string) {
	auth := smtp.PlainAuth("", channel.user, channel.pass, channel.host)
	msg := createMessage(title, text)

	err := smtp.SendMail(channel.host+":587", auth, channel.user, channel.to, msg)

	if err != nil {
		logger.Errorf("Error sending mail: %s", err)
	}
}

var model = template.Must(template.New("mail").Parse(`
  <!DOCTYPE html>
  <html>
    <body>
      {{.Content}}
    </body>
  </html>
`))

var mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

func createMessage(title string, text string) []byte {
	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", title, mimeHeaders)))

	model.Execute(&body, struct {
		Content string
	}{
		Content: template.HTMLEscapeString(text),
	})

	return body.Bytes()
}
