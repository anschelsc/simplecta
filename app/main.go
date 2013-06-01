package app

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/user"
)

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("internal server error: %s", err),
		http.StatusInternalServerError)
}

func ensureAnschel(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u == nil {
			loginURL, err := user.LoginURL(c, r.RequestURI)
			if err != nil {
				handleError(w, err)
				return
			}
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}
		if u.Email == "anschelsc@gmail.com" {
			h(w, r)
			return
		}
		logoutURL, err := user.LogoutURL(c, r.RequestURI)
		if err != nil {
			handleError(w, err)
			return
		}
		http.Redirect(w, r, logoutURL, http.StatusFound)
	}
}

func init() {
	http.HandleFunc("/", ensureAnschel(lister))
	http.HandleFunc("/feed/", ensureAnschel(showFeed))
	http.HandleFunc("/list/", ensureAnschel(lister))
	http.HandleFunc("/all/", ensureAnschel(showAll))
	http.HandleFunc("/addAtom/", ensureAnschel(atomAdder))
	http.HandleFunc("/addRSS/", ensureAnschel(rssAdder))
}
