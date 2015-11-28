package app

import (
	"encoding/xml"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

// Valid only if k has no descendents of its own kind.
func exists(c appengine.Context, k *datastore.Key) (bool, error) {
	count, err := datastore.NewQuery(k.Kind()).KeysOnly().Ancestor(k).Count(c)
	if err != nil {
		return false, err
	}
	return (count != 0), err
}

func fetchRSS(c appengine.Context, url string) (*RSS, error) {
	ret := new(RSS)
	cl := urlfetch.Client(c)
	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(ret)
	if err == nil && ret.Title == "" {
		return nil, errors.New("Not an RSS feed.")
	}
	return ret, err
}

func fetchAtom(c appengine.Context, url string) (*Atom, error) {
	ret := new(Atom)
	cl := urlfetch.Client(c)
	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(ret)
	ret.IsAtom = true
	ret.Link = ret.XMLLink.Href
	if err == nil && ret.Title == "" {
		return nil, errors.New("Not an Atom feed.")
	}
	return ret, err
}

type Recent struct {
	Item *datastore.Key
	PubDate time.Time
}

func (f *RSS) update(c appengine.Context, fk *datastore.Key) error {
	var recentDate time.Time // Zero value is very long ago
	var recentItem *datastore.Key
	for _, it := range f.Items {
		if it.GUID == "" {
			it.GUID = it.Link
		}
		var err error
		it.PubDate, err = time.Parse(time.RFC1123Z, it.RawPD)
		if err != nil {
			it.PubDate, err = time.Parse(time.RFC822Z, it.RawPD)
			if err != nil {
				it.PubDate = time.Now()
			}
		}
		if it.PubDate.Year() < 1990 { // Stupid Internet
			it.PubDate = time.Now()
		}
		ik := datastore.NewKey(c, "item", it.GUID, 0, fk)
		done, err := exists(c, ik)
		if err != nil {
			return err
		}
		if !done {
			_, err := datastore.Put(c, ik, &it)
			if err != nil {
				return err
			}
			propagate.Call(c, ik)
		}
		if it.PubDate.After(recentDate) {
			recentItem = ik
			recentDate = it.PubDate
		}
	}
	if recentItem != nil {
		feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
		recentKey := datastore.NewKey(c, "recent", fk.StringID(), 0, feedRoot)
		_, err := datastore.Put(c, recentKey, &Recent{recentItem, recentDate})
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Atom) update(c appengine.Context, fk *datastore.Key) error {
	var recentDate time.Time // Zero value is very long ago
	var recentItem *datastore.Key
	for _, it := range f.Entries {
		it.Link = it.XMLLink.Href
		var err error
		it.PubDate, err = time.Parse(time.RFC3339, it.RawPD)
		if err != nil || it.PubDate.Year() < 1990 {
			it.PubDate = time.Now()
		}
		ik := datastore.NewKey(c, "item", it.GUID, 0, fk)
		done, err := exists(c, ik)
		if err != nil {
			return err
		}
		if !done {
			_, err := datastore.Put(c, ik, &it)
			if err != nil {
				return err
			}
			propagate.Call(c, ik)
		}
		if it.PubDate.After(recentDate) {
			recentItem = ik
			recentDate = it.PubDate
		}
	}
	if recentItem != nil {
		feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
		recentKey := datastore.NewKey(c, "recent", fk.StringID(), 0, feedRoot)
		_, err := datastore.Put(c, recentKey, &Recent{recentItem, recentDate})
		if err != nil {
			return err
		}
	}
	return nil
}

func addRSS(c appengine.Context, url string) error {
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	fk := datastore.NewKey(c, "feed", url, 0, feedRoot)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		done, err := exists(c, fk)
		if err != nil {
			return err
		}
		if !done {
			f, err := fetchRSS(c, url)
			if err != nil {
				return err
			}
			_, err = datastore.Put(c, fk, f)
			if err != nil {
				return err
			}
			err = f.update(c, fk)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	return subscribe(c, fk, true)
}

func addAtom(c appengine.Context, url string) error {
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	fk := datastore.NewKey(c, "feed", url, 0, feedRoot)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		done, err := exists(c, fk)
		if err != nil {
			return err
		}
		if !done {
			f, err := fetchAtom(c, url)
			if err != nil {
				return err
			}
			_, err = datastore.Put(c, fk, f)
			if err != nil {
				return err
			}
			err = f.update(c, fk)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	return subscribe(c, fk, true)
}

func atomAdder(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	url := r.URL.Query()["url"][0]
	err := addAtom(c, url)
	if err != nil {
		handleError(w, err)
		return
	}
	http.Redirect(w, r, "/feeds/", http.StatusFound)
}

func rssAdder(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	url := r.URL.Query()["url"][0]
	err := addRSS(c, url)
	if err != nil {
		handleError(w, err)
		return
	}
	http.Redirect(w, r, "/feeds/", http.StatusFound)
}
