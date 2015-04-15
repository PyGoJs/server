package handlers

import (
	"fmt"
	"net/http"

	"github.com/pygojs/server/util"
)

// Home shows information about the server configuration.
// (if it's in dev/debug mode, and if it is what time it is stuck at)
func Home(w http.ResponseWriter, r *http.Request) {
	text := "'Production'"
	if util.Cfg().Debug.Enabled {
		text = "Dev/debug mode\nTime: " + util.Cfg().Debug.TimeStr
	}
	fmt.Fprintf(w, text)
}
