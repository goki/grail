// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-maildir"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/yuin/goldmark"
	"goki.dev/gi/v2/gi"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
	"goki.dev/grows/jsons"
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
	a.ComposeMessage.From = []*mail.Address{{Address: Prefs.Accounts[0]}}
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
	if len(a.ComposeMessage.From) != 1 {
		return fmt.Errorf("expected 1 sender, but got %d", len(a.ComposeMessage.From))
	}
	email := a.ComposeMessage.From[0].Address

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
		a.AuthClient[email],
		email,
		to,
		&b,
	)
	if err != nil {
		se := err.(*smtp.SMTPError)
		slog.Error("error sending message: SMTP error:", "code", se.Code, "enhancedCode", se.EnhancedCode, "message", se.Message)
	}
	return err
}

/*
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
*/

// CacheMessages caches all of the messages from the server that
// have not already been cached. It caches them using maildir in
// the app's prefs directory.
func (a *App) CacheMessages() error {
	for _, account := range Prefs.Accounts {
		err := a.CacheMessagesForAccount(account)
		if err != nil {
			return err
		}
	}
	return nil
}

// CacheMessages caches all of the messages from the server that
// have not already been cached for the given email account. It
// caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForAccount(email string) error {
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	err = c.Authenticate(a.AuthClient[email])
	if err != nil {
		return err
	}

	return a.CacheMessagesForMailbox(c, email, "INBOX")
}

// CacheMessagesForMailbox caches all of the messages from the server
// that have not already been cached for the given email account and mailbox.
// It caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForMailbox(c *client.Client, email string, mailbox string) error {
	hemail := hex.EncodeToString([]byte(email))
	dir := maildir.Dir(filepath.Join(gi.AppPrefsDir(), "mail", hemail, mailbox))
	err := os.MkdirAll(string(dir), 0700)
	if err != nil {
		return err
	}
	err = dir.Init()
	if err != nil {
		return err
	}

	cachedFile := filepath.Join(gi.AppPrefsDir(), "caching", hemail, mailbox, "cached-messages.json")
	err = os.MkdirAll(filepath.Dir(cachedFile), 0700)
	if err != nil {
		return err
	}
	var cached []uint32
	err = jsons.Open(&cached, cachedFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	ibox, err := c.Select(mailbox, false)
	if err != nil {
		return err
	}
	_ = ibox

	// we want messages with UIDs not in the list we already cached
	criteria := imap.NewSearchCriteria()
	if len(cached) > 0 {
		seqset := &imap.SeqSet{}
		seqset.AddNum(cached...)

		nc := imap.NewSearchCriteria()
		nc.Uid = seqset
		criteria.Not = append(criteria.Not, nc)
	}

	// these are the UIDs of the new messages
	uids, err := c.UidSearch(criteria)
	if err != nil {
		return err
	}

	if len(uids) == 0 {
		return nil
	}

	// we only fetch the new messages
	fseqset := &imap.SeqSet{}
	fseqset.AddNum(uids...)

	var sect imap.BodySectionName

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(fseqset, []imap.FetchItem{sect.FetchItem()}, messages)
	}()

	for msg := range messages {
		d, err := maildir.NewDelivery(string(dir))
		if err != nil {
			return err
		}
		_, err = io.Copy(d, msg.GetBody(&sect))
		if err != nil {
			return errors.Join(err, d.Close())
		}
		err = d.Close()
		if err != nil {
			return err
		}

		// we need to save the list of cached messages every time in case
		// we get interrupted or have an error
		cached = append(cached, msg.Uid)
		err = jsons.Save(&cached, cachedFile)
		if err != nil {
			return err
		}
	}

	if err := <-done; err != nil {
		return err
	}

	return nil
}
