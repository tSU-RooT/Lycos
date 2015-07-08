package lycos

import (
	"time"
)

type Village struct {
	No               int64     `datastore:"No" goon:"id"`
	Name             string    `datastore:"Name,noindex"`
	CreatedTime      time.Time `datastore:"CreatedTime"`
	NumberOfPeople   int       `datastore:"NumberOfPeople,noindex"`
	IncludeFreemason bool      `datastore:"IncludeFreemason,noindex"`
	IncludeFox       bool      `datastore:"IncludeFox,noindex"`
	Chip             bool      `datastore:"Chip,noindex"`
	UpdatetimeHour   int       `datastore:"UpdatetimeHour,noindex"`
	UpdatetimeMinute int       `datastore:"UpdatetimeMinute,noindex"`
	Day              int       `datastore:"Day,noindex"` // 0:プロローグ 1〜N:N日目 -N:N日目で終了したエピローグ
	PublicPostNo     int       `datastore:"PublicPostNo,noindex"`
	PersonalPostNo   int       `datastore:"PersonalPostNo,noindex"`
	WhisperNo        int       `datastore:"WhisperNo,noindex"`
	GraveyardPostNo  int       `datastore:"GraveyardPostNo,noindex"`
	Builder          string    `datastore:"Builder, noindex"`
	Close            bool      `datastore:"Close"`
}

type Management struct {
	Key       string `datastore:"-" goon:"id"`
	VillageNo int64  `datastore:"VillageNo,noindex"`
}

type User struct {
	Handle string `datastore:"Handle,noindex"`
	ID     string `datastore:"ID" goon:"id"`
	Email  string `datastore:"Email,noindex"`
}

type Chara struct {
	File string `yaml:"File"`
	Name string `yaml:"Name"`
}
