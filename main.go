package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

type Post struct {
	Title string
	Body  string
	Slug  string
}

func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", HomeHandler).Methods("GET")
	router.HandleFunc("/articles/new", NewArticleHandler).Methods("GET")
	router.HandleFunc("/articles", CreateArticleHandler).Methods("POST")
	router.HandleFunc("/articles/{title}", ShowArticleHandler).Methods("GET")

	log.Println("Listening on port 3000...")
	http.ListenAndServe(":3000", router)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := ioutil.ReadDir("posts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	html := "<!doctype html><html><body>Hello world! <a href='/articles/new'>Want to create an article?</a><ul>"
	sort.Sort(byLatestDate(posts))
	for _, file := range posts {
		post, err := LoadPost(file.Name()[:len(file.Name())-5])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		postListing := fmt.Sprintf("<li><a href='articles/%s'>%s</a></li>", post.Slug, post.Title)
		html += postListing
	}
	html += "</ul></body></html>"
	fmt.Fprintf(w, html)
}

func NewArticleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<!doctype html><html><body><form action='/articles' method='post'><input type='text' name='title' placeholder='enter your title&hellip;'/><br/><textarea name='body' placeholder='your thoughts...'></textarea><br/><input type='submit' value='Submit'/></form></body></html>")
}

func CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	post := NewPost(r.FormValue("title"), r.FormValue("body"))
	err := post.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/articles/"+post.Slug, http.StatusFound)
}

func ShowArticleHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	post, err := LoadPost(params["title"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "<!doctype html><html><head><title>%s</title></head><body><h1>%s</h1><p>%s</p></body></html>", post.Title, post.Title, post.Body)
}

func NewPost(title string, body string) *Post {
	return &Post{Title: title, Body: body, Slug: slugify(title)}
}

func LoadPost(slug string) (post *Post, err error) {
	b, err := ioutil.ReadFile("posts/" + slug + ".json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &post)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (post *Post) save() error {
	b, err := json.Marshal(post)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("posts/"+post.Slug+".json", b, 0600)
}

func slugify(title string) (slug string) {
	slug = strings.ToLower(title)
	slug = regexp.MustCompile("[^a-z0-9 -]").ReplaceAllString(slug, "")
	slug = regexp.MustCompile(" +").ReplaceAllString(slug, "-")
	slug = regexp.MustCompile("-+").ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, " -")
	return
}

type byLatestDate []os.FileInfo

func (f byLatestDate) Len() int           { return len(f) }
func (f byLatestDate) Less(i, j int) bool { return f[i].ModTime().After(f[j].ModTime()) }
func (f byLatestDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
