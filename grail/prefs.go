// Copyright (c) 2023, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grail

import (
	"errors"
	"io/fs"
	"path/filepath"

	"goki.dev/goosi"
	"goki.dev/grows/jsons"
	"goki.dev/grr"
)

// Prefs is the global instance of [Preferences], loaded on startup.
var Prefs = &Preferences{}

func init() {
	file := filepath.Join(goosi.TheApp.AppPrefsDir(), "prefs.json")
	err := jsons.Open(Prefs, file)
	if errors.Is(err, fs.ErrNotExist) {
		grr.Log0(jsons.Save(Prefs, file))
	} else {
		grr.Log0(err)
	}
}

// Preferences are the preferences that control grail.
type Preferences struct {

	// Accounts are the email accounts the user is signed into.
	Accounts []string
}
