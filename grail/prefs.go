// Copyright (c) 2023, The Goki Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"errors"
	"io/fs"
	"path/filepath"

	"goki.dev/grows/jsons"
	"goki.dev/grr"
)

// Prefs is the global instance of [Preferences], loaded on startup.
var Prefs = &Preferences{}

// Init initializes the grail preferences.
func (a *App) Init() {
	grr.Log(a.OpenPrefs())
}

// OpenPrefs opens [Prefs] from the default location.
func (a *App) OpenPrefs() error {
	file := filepath.Join(a.App().DataDir(), "prefs.json")
	err := jsons.Open(Prefs, file)
	if errors.Is(err, fs.ErrNotExist) {
		return jsons.Save(Prefs, file)
	}
	return err
}

// SavePrefs saves [Prefs] to the default location.
func (a *App) SavePrefs() error {
	file := filepath.Join(a.App().DataDir(), "prefs.json")
	return jsons.Save(Prefs, file)
}

// Preferences are the preferences that control grail.
type Preferences struct {

	// Accounts are the email accounts the user is signed into.
	Accounts []string
}
