package saveArticleOnSlack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"cloud.google.com/go/datastore"
)

type Parameter struct {
	SubCommand string
	Tag        string
	Url      string
}

type Article struct {
	ID        int64     `datastore:"-"`
	Tag       string    `datastore:"tag"`
	Url       string    `datastore:"url"`
	CreatedAt time.Time `datastore:"createdAt"`
}

func (p *Parameter) parse(text string) {
	t := strings.TrimSpace(text)
	if len(t) < 1 {
		return
	}
	s := strings.SplitN(t, " ", 3)
	p.SubCommand = s[0]
	if len(s) == 1 {
		return
	}
	p.Tag = s[1]
	if len(s) == 2 {
		return
	}
	p.Url = s[2]
}

func add(tag string, url string) error {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, os.Getenv("PROJECT_NAME"))
	if err != nil {
		return err
	}
	newKey := datastore.IncompleteKey("Article", nil)
	r := Article{
		Tag:       tag,
		Url:       url,
		CreatedAt: time.Now(),
	}
	if _, err := client.Put(ctx, newKey, &r); err != nil {
		return err
	}
	return nil
}

func list(tag ...string) ([]Article, error) {
	// log.Printf("DatastoreDebug: %v\n", tag[0])
	// if len(tag[0]) < 1 {
	// 	log.Printf("DatastoreDebug: test")
	// }

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, os.Getenv("PROJECT_NAME"))
	if err != nil {
		return nil, err
	}
	var r []Article
	q := datastore.NewQuery("Post")
	if len(tag[0]) < 1 {
		q = datastore.NewQuery("Article").Order("-createdAt")
	}else{
		q = datastore.NewQuery("Article").Filter("tag =", tag[0]).Order("-createdAt")
	}

	keys, err := client.GetAll(ctx, q, &r)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(r); i++ {
		r[i].ID = keys[i].ID
	}
	return r, nil
}

func sprint(list []Article) (s string){
	for _, r := range list {
		s = s + fmt.Sprintf("[%v] %v\n", r.Tag, r.Url)
	}
	return s
}

func SaveArticleOnSlack(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		e := "Method Not Allowed."
		log.Println(e)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(e))
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ReadAllError: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	parsed, err := url.ParseQuery(string(b))
	if err != nil {
		log.Printf("ParceQueryError: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if parsed.Get("token") != os.Getenv("SLACK_TOKEN") {
		e := "Unauthorized Token."
		log.Printf(e)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(e))
		return
	}

	p := new(Parameter)
	p.parse(parsed.Get("text"))
	switch p.SubCommand {
	case "add":
		if err := add(p.Tag, p.Url); err != nil {
			log.Printf("DatastorePutError: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(p.Tag))
		w.Write([]byte(p.Url))
	case "list":
		list, err := list(p.Tag)
		if err != nil {
			log.Printf("DatastoreGetAllError: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sprint(list)))
	default:
		e := "Invalid SubCommand."
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(parsed.Get("text")))
}