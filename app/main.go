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
		if u.Email == "Anschelsc@gmail.com" {
			h(w, r)
			return
		}
		logoutURL, err := user.LogoutURL(c, "/")
		if err != nil {
			handleError(w, err)
			return
		}
		http.Redirect(w, r, logoutURL, http.StatusFound)
	}
}

func init() {
	http.HandleFunc("/", ensureAnschel(showAll))
	http.HandleFunc("/feed/", ensureAnschel(showFeed))
	http.HandleFunc("/list/", ensureAnschel(lister))
	http.HandleFunc("/all/", ensureAnschel(showAll))
	http.HandleFunc("/addAtom/", ensureAnschel(atomAdder))
	http.HandleFunc("/addRSS/", ensureAnschel(rssAdder))
	http.HandleFunc("/read/", ensureAnschel(reader))
	http.HandleFunc("/markRead/", ensureAnschel(readMarker))
	http.HandleFunc("/markUnread/", ensureAnschel(unreadMarker))
	http.HandleFunc("/rehash/", ensureAnschel(rehasher))
	http.HandleFunc("/update/", updater)
}
