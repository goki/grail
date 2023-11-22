// Code generated by "goki generate"; DO NOT EDIT.

package grail

import (
	"github.com/emersion/go-sasl"
	"goki.dev/gi/v2/gi"
	"goki.dev/goosi/events"
	"goki.dev/gti"
	"goki.dev/ki/v2"
	"goki.dev/ordmap"
	"golang.org/x/oauth2"
)

// AppType is the [gti.Type] for [App]
var AppType = gti.AddType(&gti.Type{
	Name:       "goki.dev/grail/grail.App",
	ShortName:  "grail.App",
	IDName:     "app",
	Doc:        "App is an email client app.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"AuthToken", &gti.Field{Name: "AuthToken", Type: "map[string]*golang.org/x/oauth2.Token", LocalType: "map[string]*oauth2.Token", Doc: "AuthToken contains the [oauth2.Token] for each account.", Directives: gti.Directives{}, Tag: ""}},
		{"AuthClient", &gti.Field{Name: "AuthClient", Type: "map[string]github.com/emersion/go-sasl.Client", LocalType: "map[string]sasl.Client", Doc: "AuthClient contains the [sasl.Client] authentication for sending messages for each account.", Directives: gti.Directives{}, Tag: ""}},
		{"ComposeMessage", &gti.Field{Name: "ComposeMessage", Type: "*goki.dev/grail/grail.Message", LocalType: "*Message", Doc: "ComposeMessage is the current message we are editing", Directives: gti.Directives{}, Tag: ""}},
		{"ReadMessage", &gti.Field{Name: "ReadMessage", Type: "*goki.dev/grail/grail.Message", LocalType: "*Message", Doc: "ReadMessage is the current message we are reading", Directives: gti.Directives{}, Tag: ""}},
		{"Messages", &gti.Field{Name: "Messages", Type: "[]*goki.dev/grail/grail.Message", LocalType: "[]*Message", Doc: "Messages are the current messages we are viewing", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Frame", &gti.Field{Name: "Frame", Type: "goki.dev/gi/v2/gi.Frame", LocalType: "gi.Frame", Doc: "", Directives: gti.Directives{}, Tag: ""}},
	}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{
		{"Compose", &gti.Method{Name: "Compose", Doc: "Compose pulls up a dialog to send a new message", Directives: gti.Directives{
			&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
		}, Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}), Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{})}},
		{"SendMessage", &gti.Method{Name: "SendMessage", Doc: "SendMessage sends the current message", Directives: gti.Directives{
			&gti.Directive{Tool: "gti", Directive: "add", Args: []string{}},
		}, Args: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}), Returns: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
			{"error", &gti.Field{Name: "error", Type: "error", LocalType: "error", Doc: "", Directives: gti.Directives{}, Tag: ""}},
		})}},
	}),
	Instance: &App{},
})

// NewApp adds a new [App] with the given name
// to the given parent. If the name is unspecified, it defaults
// to the ID (kebab-case) name of the type, plus the
// [ki.Ki.NumLifetimeChildren] of the given parent.
func NewApp(par ki.Ki, name ...string) *App {
	return par.NewChild(AppType, name...).(*App)
}

// KiType returns the [*gti.Type] of [App]
func (t *App) KiType() *gti.Type {
	return AppType
}

// New returns a new [*App] value
func (t *App) New() ki.Ki {
	return &App{}
}

// SetAuthToken sets the [App.AuthToken]:
// AuthToken contains the [oauth2.Token] for each account.
func (t *App) SetAuthToken(v map[string]*oauth2.Token) *App {
	t.AuthToken = v
	return t
}

// SetAuthClient sets the [App.AuthClient]:
// AuthClient contains the [sasl.Client] authentication for sending messages for each account.
func (t *App) SetAuthClient(v map[string]sasl.Client) *App {
	t.AuthClient = v
	return t
}

// SetComposeMessage sets the [App.ComposeMessage]:
// ComposeMessage is the current message we are editing
func (t *App) SetComposeMessage(v *Message) *App {
	t.ComposeMessage = v
	return t
}

// SetReadMessage sets the [App.ReadMessage]:
// ReadMessage is the current message we are reading
func (t *App) SetReadMessage(v *Message) *App {
	t.ReadMessage = v
	return t
}

// SetMessages sets the [App.Messages]:
// Messages are the current messages we are viewing
func (t *App) SetMessages(v []*Message) *App {
	t.Messages = v
	return t
}

// SetTooltip sets the [App.Tooltip]
func (t *App) SetTooltip(v string) *App {
	t.Tooltip = v
	return t
}

// SetClass sets the [App.Class]
func (t *App) SetClass(v string) *App {
	t.Class = v
	return t
}

// SetPriorityEvents sets the [App.PriorityEvents]
func (t *App) SetPriorityEvents(v []events.Types) *App {
	t.PriorityEvents = v
	return t
}

// SetCustomContextMenu sets the [App.CustomContextMenu]
func (t *App) SetCustomContextMenu(v func(m *gi.Scene)) *App {
	t.CustomContextMenu = v
	return t
}

// SetStackTop sets the [App.StackTop]
func (t *App) SetStackTop(v int) *App {
	t.StackTop = v
	return t
}

// SetStripes sets the [App.Stripes]
func (t *App) SetStripes(v gi.Stripes) *App {
	t.Stripes = v
	return t
}
