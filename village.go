package lycos

import (
	gae "appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/user"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/mjibson/goon"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func enterToVillageHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	person := Person{}
	person.Name = r.FormValue("Name")
	person.WantJob = r.FormValue("wantjob")
	person.Face = r.FormValue("charaSet")
	text := r.FormValue("comment")
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil || u == nil || len(person.Name) <= 2 || len(text) <= 6 || len(person.Name) > 36 {
		bad(w)
		return
	}
	person.UserID = u.ID
	person.CreatedTime = time.Now()
	data := []byte(person.UserID + person.Name + person.CreatedTime.String())
	person.CharacterID = fmt.Sprintf("%x", md5.Sum(data))

	village := Village{No: no}
	err = g.Get(&village)
	if err != nil || village.Day != 0 {
		bad(w)
		return
	}
	vKey := g.Key(village)
	person.ParentKey = vKey
	village.NumberOfPeople++
	village.PublicPostNo++
	user := User{ID: u.ID}
	err = g.Get(&user)
	if err != nil {
		s := strings.Split(u.Email, "@")
		handle := "Noname"
		if len(s) > 0 {
			handle = s[0]
		}
		user.Handle = handle
		user.Email = u.Email
		_, err = g.Put(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = g.RunInTransaction(func(g *goon.Goon) error {
		_, err := g.Put(&person)
		if err != nil {
			return err
		}
		_, err = g.Put(&village)
		if err != nil {
			return err
		}
		post := Post{NumberTag: fmt.Sprintf("%d", village.PublicPostNo),
			Text: text, AuthorID: u.ID, Type: Public, Time: time.Now(), Face: person.Face, Author: person.Name, ParentKey: vKey}
		_, err = g.Put(&post)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	memcache.Delete(c, memcacheKey("Post", no, village.Day))
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=recent&page=recent", http.StatusFound)
}

func villageHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	village := Village{No: no}
	err = g.Get(&village)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	vKey := g.Key(village)
	schedule := UpdateSchedule{VillageNo: no}
	err = g.Get(&schedule)
	updateNoticeText := ""
	if err != nil {
		updateNoticeText = fmt.Sprintf("更新設定(%d:%02d)", village.UpdatetimeHour, village.UpdatetimeMinute)
	} else {
		t := schedule.UpdateTime.In(jst)
		updateNoticeText = fmt.Sprintf("%d/%d/%d %d時%02d分 頃", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	}
	day, err := strconv.Atoi(r.URL.Query().Get("day"))
	if err != nil {
		if r.URL.Query().Get("day") == "recent" {
			day = village.Day
			if day <= -1 {
				day = -1
			}
		} else {
			day = 0
		}
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		if r.URL.Query().Get("day") == "recent" {
			page = -1
		} else {
			page = 0
		}
	}
	// Illegal Access
	if (day == -1 && village.Day >= 0) || (day > village.Day && village.Day >= 0) || day < -1 {
		http.NotFound(w, r)
		return
	}
	villageView := VillageView{
		No:               no,
		CharacterSet:     characterSet,
		Village:          village,
		Day:              day,
		UpdatetimeNotice: updateNoticeText,
		NpcName:          setting.NpcName,
	}
	u := user.Current(c)
	if u != nil {
		villageView.Login = true
		villageView.LogoutURL, _ = user.LogoutURL(c, r.URL.String())
	} else {
		villageView.Login = false
		villageView.LoginURL, _ = user.LoginURL(c, r.URL.String())
	}
	q1 := datastore.NewQuery("Person").Ancestor(vKey).Order("CreatedTime")
	people := make([]Person, 0, 10)
	if _, err := g.GetAll(q1, &people); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	villageView.People = people
	var reader Person
	for _, person := range people {
		if u != nil && person.UserID == u.ID {
			villageView.Enter = true
			villageView.UserFace = person.Face
			if village.Day > 0 {
				villageView.ShowAbility = true
				villageView.AbilityDescription = person.Job.Description()
				if person.Job.CanUseAbility(village.Day) && !person.Dead {
					villageView.ShowAbilitySelect = true
				}
			}
			reader = person
			villageView.Reader = person
			break
		}
	}
	if u != nil && village.Builder == u.ID && village.NumberOfPeople >= 8 && village.Day == 0 {
		villageView.ShowStartButton = true
	}
	posts := make([]Post, 0, 30)
	memPostKey := memcacheKey("Post", no, day)
	if cache, err := memcache.Get(c, memPostKey); err == memcache.ErrCacheMiss {
		q2 := datastore.NewQuery("Post").Ancestor(vKey).Filter("Day =", day).Order("Time")
		if _, err := g.GetAll(q2, &posts); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if val, err := serialize(&posts); err == nil {
			item := memcache.Item{Key: memPostKey, Value: val, Expiration: time.Hour * 12}
			memcache.Add(c, &item)
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		deserialize(cache.Value, &posts)
	}

	viewPosts := make([]Post, 0, 30)
	if villageView.Enter {
		for i := range posts {
			addOK := false
			pType := posts[i].Type
			if village.Day <= -1 {
				addOK = true
			} else if (pType == Personal || pType == SystemSecret) && posts[i].AuthorID == reader.UserID {
				addOK = true
			} else if (pType == Whisper) && reader.Job.CanSpeakWhisper() {
				addOK = true
			} else if pType == Public || pType == SystemMessage {
				addOK = true
			} else if pType == Graveyard && reader.Dead {
				addOK = true
			}
			if addOK {
				viewPosts = append(viewPosts, posts[i])
			}
		}
	} else {
		for i := range posts {
			if village.Day <= -1 || posts[i].Type == Public || posts[i].Type == SystemMessage {
				viewPosts = append(viewPosts, posts[i])
			}
		}
	}
	var maxPage int
	if len(viewPosts) > 19 {
		maxPage = (len(viewPosts) / 15)
		if page == -1 || page >= maxPage {
			viewPosts = viewPosts[len(viewPosts)-15:]
			if day == -1 || day == village.Day {
				villageView.Recent = true
			}
		} else {
			viewPosts = viewPosts[15*page : 15*(page+1)]
		}
	} else {
		maxPage = 0
		if day == -1 || day == village.Day {
			villageView.Recent = true
		}
	}
	villageView.Posts = viewPosts
	villageView.Indexes = make([]Page, maxPage+1)
	for i := 0; i <= maxPage; i++ {
		p := Page{Number: i}
		if page == i {
			p.Invalid = true
		}
		villageView.Indexes[i] = p
	}

	for i, po := range villageView.Posts {
		buf := new(bytes.Buffer)
		template.HTMLEscape(buf, []byte(po.Text))
		t := buf.String()
		t = strings.Replace(t, "\n", "<br />", -1)
		villageView.Posts[i].Text = t
	}

	chap := []Chapter{Chapter{Day: 0, Name: "プロローグ", Invalid: day == 0}}
	if d := village.Day; d > 0 {
		for i := 1; i <= village.Day; i++ {
			chap = append(chap, Chapter{Day: i, Name: strconv.Itoa(i) + "日目", Invalid: day == i})
		}
	} else if d < 0 {
		d *= -1
		for i := 1; i < d; i++ {
			chap = append(chap, Chapter{Day: i, Name: strconv.Itoa(i) + "日目", Invalid: day == i})
		}
		chap = append(chap, Chapter{Day: -1, Name: "エピローグ", Invalid: day == -1})
	}
	villageView.Chapters = chap
	if day == -1 && village.Day <= -1 {
		villageView.ShowResult = true
		rCols := make([]ResultCol, 0, 10)
		j := Judge(people)
		for i := range people {
			rc := ResultCol{Name: people[i].Name, Dead: people[i].Dead, Job: people[i].Job, Victory: people[i].Job.GotVictory(j)}
			if people[i].WantJob == "1" {
				rc.WantJob = "おまかせ"
			} else if people[i].WantJob == "2" {
				rc.WantJob = "村陣営"
			} else if people[i].WantJob == "3" {
				rc.WantJob = "村陣営(役職)"
			} else if people[i].WantJob == "4" {
				rc.WantJob = "人外陣営"
			}
			user := User{ID: people[i].UserID}
			if err := g.Get(&user); err != nil {
				rc.Handle = "Unknown"
			} else {
				rc.Handle = user.Handle
			}
			rCols = append(rCols, rc)
		}
		villageView.Result = rCols
	}
	if err := villagePageTmpl.ExecuteTemplate(w, "base", villageView); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renameHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	handle := r.FormValue("InputName")
	if u != nil && handle != "" {
		user := User{ID: u.ID}
		if err := g.Get(&user); err != nil {
			reader := User{Handle: handle, Email: u.Email, ID: u.ID}
			_, err = g.Put(&reader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			user.Handle = handle
			_, err = g.Put(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func villagePreWriteHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	preWriteView := PreWriteView{}
	buf := new(bytes.Buffer)
	template.HTMLEscape(buf, []byte(r.FormValue("comment")))
	t := buf.String()
	preWriteView.Text = strings.Replace(t, "\n", "<br>", -1)
	preWriteView.HiddenText = r.FormValue("comment")
	commentType := r.FormValue("commentType")
	characterID := r.FormValue("characterID")
	preWriteView.CharacterID = characterID
	if commentType == "personal" {
		preWriteView.IsPersonal = true
	} else if commentType == "whisper" {
		preWriteView.IsWhisper = true
	} else if commentType == "graveyard" {
		preWriteView.IsGraveyard = true
	} else {
		preWriteView.IsPublic = true
	}
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil || len(preWriteView.Text) <= 5 || user.Current(c) == nil || len(preWriteView.Text) > 1000 {
		bad(w)
		return
	}
	preWriteView.VillageNo = no
	village := Village{No: no}
	if err := g.Get(&village); err != nil {
		bad(w)
		return
	}
	vKey := g.Key(village)
	person := Person{UserID: u.ID, ParentKey: vKey, CharacterID: characterID}
	if err := g.Get(&person); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	preWriteView.Face = person.Face
	preWriteView.Author = person.Name
	if err = prewriteTmpl.ExecuteTemplate(w, "base", preWriteView); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func villagePostWriteHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	text := r.FormValue("comment")
	commentType := r.FormValue("commentType")
	characterID := r.FormValue("characterID")
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil || len(text) <= 5 || u == nil || len(text) > 1000 {
		bad(w)
		return
	}
	if r.FormValue("button") == "やめる" {
		http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=recent&page=recent", http.StatusFound)
		return
	}
	village := Village{No: no}
	if err := g.Get(&village); err != nil {
		c.Debugf("%v", err)
		bad(w)
		return
	}
	vKey := g.Key(village)
	err = g.RunInTransaction(func(g *goon.Goon) error {
		nTag := ""
		var pType PostType
		switch commentType {
		case "Public":
			pType = Public
			village.PublicPostNo += 1
			nTag = fmt.Sprintf("%d", village.PublicPostNo)
		case "Personal":
			pType = Personal
			village.PersonalPostNo += 1
			nTag = fmt.Sprintf("*%d", village.PersonalPostNo)
		case "Whisper":
			pType = Whisper
			village.WhisperNo += 1
			nTag = fmt.Sprintf("@%d", village.WhisperNo)
		case "Graveyard":
			pType = Graveyard
			village.GraveyardPostNo += 1
			nTag = fmt.Sprintf("$%d", village.GraveyardPostNo)
		}
		post := Post{Text: text, AuthorID: u.ID, Type: pType,
			Time: time.Now(), NumberTag: nTag, ParentKey: vKey}
		if village.Day <= -1 {
			post.Day = -1
		} else {
			post.Day = village.Day
		}
		person := Person{ParentKey: vKey, CharacterID: characterID}
		if err := g.Get(&person); err != nil {
			return err
		}
		if !person.HasWriteAuth(pType) && !(village.Day <= -1 && pType == Public) {
			return errors.New("Reader has no Write Auth.")
		}
		if person.UserID != u.ID {
			return errors.New("UserID is wrong!")
		}
		post.Face = person.Face
		post.Author = person.Name
		if _, err := g.Put(&post); err != nil {
			return err
		}
		if _, err := g.Put(&village); err != nil {
			return err
		}
		memcache.Delete(c, memcacheKey("Post", no, village.Day))
		return nil
	}, nil)
	if err != nil {
		c.Debugf("%v", err)
		bad(w)
	}
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=recent&page=recent", http.StatusFound)
}
