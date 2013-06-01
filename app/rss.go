package app

import (
	"time"
)

type RSS struct {
	Title string `xml:"channel>title"`
	Link  string `xml:"channel>link"`
	//Description string `xml:"channel>description"`
	Items []Item `xml:"channel>item" datastore:"-"`
	IsAtom bool `xml:"-"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	GUID  string `xml:"guid" datastore:"-"`
	PubDate time.Time `xml:"-"`
	RawPD string `xml:"pubDate" datastore:"-"`
}
