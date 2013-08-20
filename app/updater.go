package app

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/mail"
	"appengine/urlfetch"
	"appengine/user"
)

func updateFeed(c appengine.Context, cl *http.Client, fk *datastore.Key) error {
	resp, err := cl.Get(fk.StringID())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	var rfeed RSS
	err = datastore.Get(c, fk, &rfeed)
	if err != nil {
		return err
	}
	if rfeed.IsAtom {
		var afeed Atom
		err = decoder.Decode(&afeed)
		if err != nil {
			return err
		}
		return afeed.update(c, fk)
	} else {
		err = decoder.Decode(&rfeed)
		if err != nil {
			return err
		}
		return rfeed.update(c, fk)
	}
	panic("unreachable")
}

func updater(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	cl := urlfetch.Client(c)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	q := datastore.NewQuery("feed").Ancestor(feedRoot).KeysOnly()
	iter := q.Run(c)
	ch := make(chan error)
	count := 0
	for {
		fk, err := iter.Next(c)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		go func(fk *datastore.Key) {
			err = updateFeed(c, cl, fk)
			ch <- err
		}(fk)
		count++
	}
	buf := new(bytes.Buffer)
	for count != 0 {
		err := <-ch
		if err != nil {
			fmt.Fprintln(buf, <-ch)
		}
		count--
	}
	fmt.Fprintf(buf, "User: %s\n", user.Current(c))
	err := mail.Send(c, &mail.Message{
		Sender:  "updates@simplecta.appspotmail.com",
		To:      []string{"anschelsc@gmail.com"},
		Subject: "Errors from simplecta update",
		Body:    buf.String(),
	})
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintln(w, "Done.")
}
