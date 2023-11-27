// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"errors"
	"io/fs"
	"path/filepath"

	"goki.dev/gi/v2/gi"
	"goki.dev/grows/jsons"
	"goki.dev/grr"
)

// Prefs is the global instance of [Preferences], loaded on startup.
var Prefs = &Preferences{}

// Init initializes the grail preferences. It needs to be called inside of
// a [gimain.Run] app function.
func Init() {
	grr.Log(OpenPrefs())
}

// OpenPrefs opens [Prefs] from the default location.
func OpenPrefs() error {
	file := filepath.Join(gi.AppPrefsDir(), "prefs.json")
	err := jsons.Open(Prefs, file)
	if errors.Is(err, fs.ErrNotExist) {
		return jsons.Save(Prefs, file)
	}
	return err
}

// SavePrefs saves [Prefs] to the default location.
func SavePrefs() error {
	file := filepath.Join(gi.AppPrefsDir(), "prefs.json")
	return jsons.Save(Prefs, file)
}

// Preferences are the preferences that control grail.
type Preferences struct {

	// Accounts are the email accounts the user is signed into.
	Accounts []string
}
