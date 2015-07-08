package lycos

import (
	"appengine/datastore"
	"fmt"
	"time"
)

type PostType int

const (
	SystemMessage PostType = iota + 1
	Public
	Personal
	Whisper
	SystemSecret
	Graveyard
)

func (t PostType) String() string {
	switch t {
	case SystemMessage:
		return "SystemMessage"
	case Public:
		return "Public"
	case Personal:
		return "Personal"
	case Whisper:
		return "Whisper"
	case SystemSecret:
		return "SystemSecret"
	case Graveyard:
		return "Graveyard"
	default:
		return "None"
	}
}

type Post struct {
	ID        int64          `datastore:"-" goon:"id"`
	NumberTag string         `datastore:"NumberTag,noindex"`
	ParentKey *datastore.Key `datastore:"-" goon:"parent"`
	Day       int            `datastore:"Day"`
	Text      string         `datastore:"Text,noindex"`
	Author    string         `datastore:"Author,noindex"`
	AuthorID  string         `datastore:"AuthorID,noindex"`
	Type      PostType       `datastore:"Type,noindex"`
	Face      string         `datastore:"Face,noindex"`
	Time      time.Time      `datastore:"Time"`
}

func (p Post) HasFace() bool {
	return p.Face != ""
}

func (p Post) IsPersonal() bool {
	return p.Type == Personal
}

func (p Post) JstAboutTime() string {
	t := p.Time.In(jst)
	m := (t.Minute() / 5) * 5
	str := fmt.Sprintf("%d/%d/%d %d時%02d分 頃", t.Year(), t.Month(), t.Day(), t.Hour(), m)
	return str
}
