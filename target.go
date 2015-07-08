package lycos

import (
	gae "appengine"
	"appengine/user"
	"errors"
	"github.com/mjibson/goon"
	"net/http"
	"strconv"
)

func changeVoteTargetHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil || user.Current(c) == nil {
		bad(w)
		return
	}
	village := Village{No: no}
	if err := g.Get(&village); err != nil {
		bad(w)
		return
	}
	vKey := g.Key(village)
	err = g.RunInTransaction(func(g *goon.Goon) error {
		person := Person{CharacterID: r.FormValue("characterID"), ParentKey: vKey}
		err := g.Get(&person)
		if err != nil {
			return errors.New("Can't Get User Data")
		}
		if person.Dead {
			return errors.New("User is Dead Status")
		}
		if person.UserID != u.ID {
			return errors.New("UserID is wrong!")
		}
		person.VoteTarget = r.FormValue("VoteTarget")
		_, err = g.Put(&person)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		bad(w)
		return
	}
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=recent", http.StatusFound)
}

func changeAbilityTargetHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	no, err := strconv.ParseInt(r.FormValue("vno"), 10, 64)
	if err != nil || user.Current(c) == nil {
		bad(w)
		return
	}
	village := Village{No: no}
	if err := g.Get(&village); err != nil {
		c.Debugf("74 %v", err)
		bad(w)
		return
	}
	vKey := g.Key(village)
	err = g.RunInTransaction(func(g *goon.Goon) error {
		person := Person{CharacterID: r.FormValue("characterID"), ParentKey: vKey}
		err := g.Get(&person)
		if err != nil {
			return errors.New("Can't Get User Data")
		}
		if !person.Job.HasAbility() {
			return errors.New("User hasn't Ability")
		} else if person.Dead {
			return errors.New("User is Dead Status")
		} else if person.UserID != u.ID {
			return errors.New("UserID is wrong!")
		}
		person.AbilityTarget = r.FormValue("AbilityTarget")
		_, err = g.Put(&person)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		bad(w)
		return
	}
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=recent", http.StatusFound)
}
