package app

import (
	"fmt"
	"net/http"
)

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("internal server error: %s", err),
		http.StatusInternalServerError)
}

func init() {
	http.HandleFunc("/", showAll)
	http.HandleFunc("/feed/", showFeed)
	http.HandleFunc("/all/", showAll)
	http.HandleFunc("/feeds/", lister)
	http.HandleFunc("/addAtom/", atomAdder)
	http.HandleFunc("/addRSS/", rssAdder)
	http.HandleFunc("/read/", reader)
	http.HandleFunc("/markRead/", readMarker)
	http.HandleFunc("/markUnread/", unreadMarker)
	http.HandleFunc("/update/", updater)

	http.HandleFunc("/convertSubs/", convertSubs)
}
