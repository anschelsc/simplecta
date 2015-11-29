package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type uLogData struct {
	Time time.Time
}

func logUser(c appengine.Context) error {
	u := user.Current(c)
	if u == nil {
		return errors.New("User should be logged in before calling logUser.")
	}
	key := datastore.NewKey(c, "userLog", u.ID, 0, nil)
	now := uLogData{time.Now()}
	_, err := datastore.Put(c, key, &now)
	return err
}

func vanity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	c := appengine.NewContext(r)
	since := time.Now().Add(-30 * 24 * time.Hour)
	q := datastore.NewQuery("userLog").Filter("Time>", since).Order("-Time")
	count, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "%d unique visitors in the last 30 days. First 100:\n", count)
	q = q.Limit(100)
	times := make([]uLogData, 0, 100)
	ks, err := q.GetAll(c, &times)
	if err != nil {
		handleError(w, err)
		return
	}
	for i, k := range ks {
		fmt.Fprintf(w, "%v: %s\n", times[i].Time, k.StringID())
	}
	fmt.Fprintf(w, "I am %s\n", user.Current(c).ID)
}
