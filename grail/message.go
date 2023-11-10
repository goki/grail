// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
)

// Message contains the relevant information for an email message.
type Message struct {
	To      []string
	Subject string
	Body    string
}

// Compose pulls up a dialog to send a new message
func (a *App) Compose() { //gti:add
	d := gi.NewDialog(a).Title("Send message")
	giv.NewStructView(d).SetStruct(&a.Message)
	d.OnAccept(func(e events.Event) {
		a.SendMessage()
	}).Cancel().Ok().Run()
}

// SendMessage sends the current message
func (a *App) SendMessage() error { //gti:add
	var b bytes.Buffer

	from := []*mail.Address{{Address: gi.Prefs.User.Email}}
	to := make([]*mail.Address, len(a.Message.To))
	for i, t := range a.Message.To {
		to[i] = &mail.Address{Address: t}
	}

	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)
	h.SetSubject(a.Message.Subject)

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
	io.WriteString(w, a.Message.Body)
	w.Close()
	tw.Close()

	fmt.Println(b.String())

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		a.Auth,
		gi.Prefs.User.Email,
		a.Message.To,
		&b,
	)
	if err != nil {
		se := err.(*smtp.SMTPError)
		slog.Error("error sending message: SMTP error:", "code", se.Code, "enhancedCode", se.EnhancedCode, "message", se.Message)
	}
	return err
}
