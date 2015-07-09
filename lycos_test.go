package lycos

import (
	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"appengine/user"
	"github.com/mjibson/goon"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildPageHandler(t *testing.T) {
	opt := &aetest.Options{AppID: "unittest", StronglyConsistentDatastore: true}
	user := &user.User{Email: "test@example.com", Admin: true, ID: "1234567890"}
	inst, err := aetest.NewInstance(opt)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer inst.Close()
	request, err := inst.NewRequest("POST", "/build", strings.NewReader("Name=ABC&chip=true&fox=true&hour=3&minute=0"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	aetest.Login(user, request)
	recoder := httptest.NewRecorder()
	buildPageHandler(recoder, request)
	if recoder.Code != 302 {
		b, _ := ioutil.ReadAll(recoder.Body)
		t.Fatalf("unexpected %d, expected 302, body=%s", recoder.Code, string(b))
	}
	context := appengine.NewContext(request)
	g := goon.FromContext(context)
	v := Village{No: 1}
	if err := g.Get(&v); err != nil {
		t.Fatalf("unexpected err, expected Get Village{No: 1}")
	}
	if v.Name != "ABC" || v.Chip == false || v.IncludeFox == false || v.UpdatetimeHour != 3 || v.UpdatetimeMinute != 0 {
		t.Fatalf("Failed: Cant't get correct Data from datastore")
	}
}

func TestPostWriteHandler(t *testing.T) {
	opt := &aetest.Options{AppID: "unittest", StronglyConsistentDatastore: true}
	user := &user.User{Email: "test@example.com", Admin: true, ID: "1234567890"}
	inst, err := aetest.NewInstance(opt)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer inst.Close()
	args := "comment=TestMessage&characterID=CharacterID&vno=1&commentType=" + Public.String()
	request, err := inst.NewRequest("POST", "/write", strings.NewReader(args))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	aetest.Login(user, request)
	recoder := httptest.NewRecorder()
	context := appengine.NewContext(request)
	g := goon.FromContext(context)

	// Pre
	v := Village{No: 1}
	vKey, err := g.Put(&v)
	if err != nil {
		t.Fatalf("Failed to Put Village: %v", err)
	}
	p := Person{ParentKey: vKey, CharacterID: "CharacterID", Name: "Player1", UserID: user.ID}
	if _, err := g.Put(&p); err != nil {
		t.Fatalf("Failed to Put Person: %v", err)
	}

	villagePostWriteHandler(recoder, request)
	if recoder.Code != 302 {
		b, _ := ioutil.ReadAll(recoder.Body)
		t.Fatalf("unexpected %d, expected 302, body=%s", recoder.Code, string(b))
	}
	query := datastore.NewQuery("Post").Ancestor(vKey).Filter("Day =", 0).Order("-Time").Limit(1)
	var posts []Post
	if _, err := g.GetAll(query, &posts); err != nil || len(posts) < 1 {
		t.Fatalf("Failed to Get Post: %v", err)
	}
	po := posts[0]
	if po.Text != "TestMessage" || po.AuthorID != user.ID || po.Type != Public || po.Author != p.Name {
		t.Log(po)
		t.Fatalf("Failed: Cant't get correct Data from datastore")
	}
}
