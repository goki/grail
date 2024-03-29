// Copyright (c) 2023, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-maildir"
	"goki.dev/gi/v2/giv"
	"goki.dev/goosi/events"
	"goki.dev/grows/jsons"
)

// CacheData contains the data stored for a cached message in the cached messages file.
// It contains basic information about the message so that it can be displayed in the
// mail list in the GUI.
type CacheData struct {
	imap.Envelope
	UID      uint32
	Filename string
}

// CacheMessages caches all of the messages from the server that
// have not already been cached. It caches them using maildir in
// the app's prefs directory.
func (a *App) CacheMessages() error {
	if a.Cache == nil {
		a.Cache = map[string]map[string][]*CacheData{}
	}
	if a.IMAPClient == nil {
		a.IMAPClient = map[string]*imapclient.Client{}
	}
	mbox := a.FindPath("splits/mbox").(*giv.TreeView)
	updt := mbox.UpdateStart()
	mbox.DeleteChildren(true)
	for _, account := range Prefs.Accounts {
		err := a.CacheMessagesForAccount(account)
		if err != nil {
			mbox.Update()
			mbox.UpdateEndAsync(updt)
			return fmt.Errorf("caching messages for account %q: %w", account, err)
		}
	}
	// can't defer update end because it screws up the stack trace
	mbox.Update()
	mbox.UpdateEndAsync(updt)
	return nil
}

// CacheMessages caches all of the messages from the server that
// have not already been cached for the given email account. It
// caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForAccount(email string) error {
	if a.Cache[email] == nil {
		a.Cache[email] = map[string][]*CacheData{}
	}

	c, err := imapclient.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return fmt.Errorf("TLS dialing: %w", err)
	}
	defer c.Logout()

	a.IMAPClient[email] = c

	err = c.Authenticate(a.AuthClient[email])
	if err != nil {
		return fmt.Errorf("authenticating: %w", err)
	}

	mailboxes, err := c.List("", "*", nil).Collect()
	if err != nil {
		return fmt.Errorf("getting mailboxes: %w", err)
	}

	for _, mailbox := range mailboxes {
		err := a.CacheMessagesForMailbox(c, email, mailbox.Mailbox)
		if err != nil {
			return fmt.Errorf("caching messages for mailbox %q: %w", mailbox.Mailbox, err)
		}
	}
	return nil
}

// CacheMessagesForMailbox caches all of the messages from the server
// that have not already been cached for the given email account and mailbox.
// It caches them using maildir in the app's prefs directory.
func (a *App) CacheMessagesForMailbox(c *imapclient.Client, email string, mailbox string) error {
	if a.CurMailbox == "" {
		a.CurMailbox = mailbox
	}

	bemail := FilenameBase32(email)
	bmbox := FilenameBase32(mailbox)

	mbox := a.FindPath("splits/mbox").(*giv.TreeView)
	embox := mbox.ChildByName(bemail)
	if embox == nil {
		embox = giv.NewTreeView(mbox, bemail).SetText(email)
	}
	giv.NewTreeView(embox, bmbox).SetText(mailbox).OnClick(func(e events.Event) {
		a.CurMailbox = mailbox
		a.UpdateMessageList()
	})

	dir := maildir.Dir(filepath.Join(a.App().DataDir(), "mail", bemail, bmbox))
	err := os.MkdirAll(string(dir), 0700)
	if err != nil {
		return err
	}
	err = dir.Init()
	if err != nil {
		return fmt.Errorf("initializing maildir: %w", err)
	}

	cachedFile := filepath.Join(a.App().DataDir(), "caching", bemail, bmbox, "cached-messages.json")
	err = os.MkdirAll(filepath.Dir(cachedFile), 0700)
	if err != nil {
		return err
	}

	var cached []*CacheData
	err = jsons.Open(&cached, cachedFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, io.EOF) {
		return fmt.Errorf("opening cache list: %w", err)
	}
	a.Cache[email][mailbox] = cached

	_, err = c.Select(mailbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("opening mailbox: %w", err)
	}

	// we want messages with UIDs not in the list we already cached
	criteria := &imap.SearchCriteria{}
	if len(cached) > 0 {
		seqset := imap.SeqSet{}
		for _, c := range cached {
			seqset.AddNum(c.UID)
		}

		nc := imap.SearchCriteria{}
		nc.UID = []imap.SeqSet{seqset}
		criteria.Not = append(criteria.Not, nc)
	}

	// these are the UIDs of the new messages
	uidsData, err := c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("searching for uids: %w", err)
	}

	uids := uidsData.AllNums()
	if len(uids) == 0 {
		a.UpdateMessageList()
		return nil
	}

	// we only fetch the new messages
	fseqset := imap.SeqSet{}
	fseqset.AddNum(uids...)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		UID:      true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
	}

	mcmd := c.Fetch(fseqset, fetchOptions)

	for {
		msg := mcmd.Next()
		if msg == nil {
			break
		}

		mdata, err := msg.Collect()
		if err != nil {
			return err
		}

		key, w, err := dir.Create([]maildir.Flag{})
		if err != nil {
			return fmt.Errorf("making maildir file: %w", err)
		}

		var header, text []byte

		for k, v := range mdata.BodySection {
			if k.Specifier == imap.PartSpecifierHeader {
				header = v
			} else if k.Specifier == imap.PartSpecifierText {
				text = v
			}
		}

		_, err = w.Write(append(header, text...))
		if err != nil {
			return fmt.Errorf("writing message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("closing message: %w", err)
		}

		cd := &CacheData{
			Envelope: *mdata.Envelope,
			UID:      mdata.UID,
			Filename: key,
		}

		// we need to save the list of cached messages every time in case
		// we get interrupted or have an error
		cached = append(cached, cd)
		err = jsons.Save(&cached, cachedFile)
		if err != nil {
			return fmt.Errorf("saving cache list: %w", err)
		}

		a.Cache[email][mailbox] = cached
		a.UpdateMessageList()
	}

	err = mcmd.Close()
	if err != nil {
		return fmt.Errorf("fetching messages: %w", err)
	}

	return nil
}
