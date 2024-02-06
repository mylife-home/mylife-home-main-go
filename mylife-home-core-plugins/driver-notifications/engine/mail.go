package engine

import (
	"bytes"
	"text/template"

	gomail "gopkg.in/mail.v2"
)

func NewMailChannel(smtpServer string, smtpPort int64, user string, pass string, from string, to []string) Channel {
	return &mailChannel{
		dialer: gomail.NewDialer(smtpServer, int(smtpPort), user, pass),
		from:   from,
		to:     to,
	}
}

type mailChannel struct {
	dialer *gomail.Dialer
	from   string
	to     []string
}

func (channel *mailChannel) Send(title string, text string) {
	go channel.sendSync(title, text)
}

func (channel *mailChannel) sendSync(title string, text string) {
	msg := gomail.NewMessage()

	msg.SetHeader("From", channel.from)
	msg.SetHeader("To", channel.to...)
	msg.SetHeader("Subject", title)

	var body bytes.Buffer
	model.Execute(&body, struct{ Content string }{Content: template.HTMLEscapeString(text)})
	msg.SetBody("text/html", body.String())

	err := channel.dialer.DialAndSend(msg)

	if err == nil {
		logger.Infof("Mail sent from '%s'", channel.dialer.Username)
	} else {
		logger.Errorf("Error sending mail: %s", err)
	}
}

var model = template.Must(template.New("mail").Parse(`
  <!DOCTYPE html>
  <html style="color:#4da6ff; font-family: -apple-system,BlinkMacSystemFont, Segoe UI, Roboto, Helvetica Neue, Arial, Noto Sans, Liberation Sans, sans-serif, Apple Color Emoji, Segoe UI Emoji, Segoe UI Symbol, Noto Color Emoji ;">
    <body>
			<p style="border: solid; border-color:#4da6ff; border-width: 1px; padding: 12px; margin: 12px;">
      	{{.Content}}
			</p>
    </body>
  </html>
`))
