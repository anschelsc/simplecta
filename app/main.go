package app

import (
	"net/http"
	"fmt"
)

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("internal server error: %s", err),
		http.StatusInternalServerError)
}

func init() {
	//http.HandleFunc("/", sender)
	http.HandleFunc("/feed/", showFeed)
	http.HandleFunc("/list/", lister)
	http.HandleFunc("/all/", showAll)
}
