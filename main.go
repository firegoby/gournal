// Simplebog implements the most minimal blog imaginable.
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
	"text/template"

	"github.com/gorilla/mux"
)

// A Post contains a title, body and slug (used as a permalink).
type Post struct {
	Title string
	Body  string
	Slug  string
}

// String returns a simple single line representation of a Post, implementing the Stringer interface
func (p *Post) String() string {
	return fmt.Sprintf("%s (%s)", p.Title, p.Slug)
}

// Main creates a gorilla/mux router and dispatches incoming requests on port :3000.
func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", HomeHandler).Methods("GET")
	router.HandleFunc("/articles/new", NewArticleHandler).Methods("GET")
	router.HandleFunc("/articles", CreateArticleHandler).Methods("POST")
	router.HandleFunc("/articles/{title}", ShowArticleHandler).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Listening on port 3000...")
	http.ListenAndServe(":3000", router)
}

// HomeHandler provides a welcome/index page with a listing of recents posts, and a link to create a new post.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := ioutil.ReadDir("posts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sort.Sort(byLatestDate(posts))
	var postsData = make([]*Post, 0)
	for _, file := range posts {
		post, err := LoadPost(file.Name()[:len(file.Name())-5])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		postsData = append(postsData, post)
	}
	renderTemplate(w, "home", postsData)
}

// NewArticleHandler is a RESTful function for GET /photos/new
func NewArticleHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new_article", nil)
}

// CreateArticleHandler is a RESTful function for POST /photos/new
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
		log.Println(err.Error())
		//http.Redirect(w, r, "/404", http.StatusMovedPermanently)
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "article", post)
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

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/"+tmpl+".html", "templates/layout.html"))
	/*
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/
	err := t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type byLatestDate []os.FileInfo

func (f byLatestDate) Len() int           { return len(f) }
func (f byLatestDate) Less(i, j int) bool { return f[i].ModTime().After(f[j].ModTime()) }
func (f byLatestDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
