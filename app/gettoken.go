package app

import (
	"encoding/json"
	"net/http"

	"appengine"
	"appengine/channel"
)

func getToken(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	client := r.FormValue("client")
	token, err := channel.Create(c, client)
	if err != nil {
		handleError(w, err)
		return
	}
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		handleError(w, err)
		return
	}
}
