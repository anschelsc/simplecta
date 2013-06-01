package app

import (
	"time"
)

type Atom struct {
	Title   string `xml:"title"`
	Link    string `xml:"-"`
	XMLLink struct {
		Href string `xml:"href,attr"`
	} `xml:"link" datastore:"-"`
	Entries []Entry `xml:"entry" datastore:"-"`
	IsAtom  bool    `xml:"-"`
}

type Entry struct {
	Title   string `xml:"title"`
	Link    string `xml:"-"`
	XMLLink struct {
		Href string `xml:"href,attr"`
	} `xml:"link" datastore:"-"`
	GUID    string    `xml:"id" datastore:"-"`
	PubDate time.Time `xml:"-"`
	RawPD   string    `xml:"updated" datastore:"-"`
}
