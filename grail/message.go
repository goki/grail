// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"bytes"
	"log/slog"
	"slices"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/yuin/goldmark"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
)

// Message contains the relevant information for an email message.
type Message struct {
	From    []*mail.Address `view:"inline"`
	To      []*mail.Address `view:"inline"`
	Subject string
	// only for sending
	Body string `view:"text-editor" viewif:"BodyReader==nil"`
	// only for receiving
	BodyReader imap.Literal `view:"-"`
}

// Compose pulls up a dialog to send a new message
func (a *App) Compose() { //gti:add
	a.ComposeMessage = &Message{}
	a.ComposeMessage.From = []*mail.Address{{Address: gi.Prefs.User.Email}}
	b := gi.NewBody().AddTitle("Send message")
	giv.NewStructView(b).SetStruct(a.ComposeMessage)
	b.AddBottomBar(func(pw gi.Widget) {
		b.AddCancel(pw)
		b.AddOk(pw).SetText("Send").OnClick(func(e events.Event) {
			a.SendMessage()
		})
	})
	b.NewFullDialog(a).Run()
}

// SendMessage sends the current message
func (a *App) SendMessage() error { //gti:add
	var b bytes.Buffer

	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", a.ComposeMessage.From)
	h.SetAddressList("To", a.ComposeMessage.To)
	h.SetSubject(a.ComposeMessage.Subject)

	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		return err
	}

	tw, err := mw.CreateInline()
	if err != nil {
		return err
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		return err
	}
	err = goldmark.Convert([]byte(a.ComposeMessage.Body), w)
	if err != nil {
		return err
	}
	w.Close()
	tw.Close()

	to := make([]string, len(a.ComposeMessage.To))
	for i, t := range a.ComposeMessage.To {
		to[i] = t.Address
	}

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		a.AuthClient,
		gi.Prefs.User.Email,
		to,
		&b,
	)
	if err != nil {
		se := err.(*smtp.SMTPError)
		slog.Error("error sending message: SMTP error:", "code", se.Code, "enhancedCode", se.EnhancedCode, "message", se.Message)
	}
	return err
}

// GetMessages fetches the messages from the server
func (a *App) GetMessages() error { //gti:add
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	err = c.Authenticate(a.AuthClient)
	if err != nil {
		return err
	}

	ibox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	// Get the last 40 messages
	from := uint32(1)
	to := ibox.Messages
	if ibox.Messages > 39 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = ibox.Messages - 39
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	var sect imap.BodySectionName

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, sect.FetchItem()}, messages)
	}()

	a.Messages = make([]*Message, 0)
	for msg := range messages {

		from := make([]*mail.Address, len(msg.Envelope.From))
		for i, fr := range msg.Envelope.From {
			from[i] = &mail.Address{Name: fr.PersonalName, Address: fr.Address()}
		}
		to := make([]*mail.Address, len(msg.Envelope.To))
		for i, fr := range msg.Envelope.To {
			to[i] = &mail.Address{Name: fr.PersonalName, Address: fr.Address()}
		}

		m := &Message{
			From:       from,
			To:         to,
			Subject:    msg.Envelope.Subject,
			BodyReader: msg.GetBody(&sect),
		}
		a.Messages = append(a.Messages, m)
	}
	slices.Reverse(a.Messages)

	if err := <-done; err != nil {
		return err
	}
	return nil
}
