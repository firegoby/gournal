// Copyright (c) 2014 Chris Batchelor.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// gournal implements the most minimal go blog (go-journal) imaginable.
//
// It's just a project for learning about building web apps in Go and isn't
// meant for any real-world usage.
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

// String returns a simple single line representation of a Post, implementing
// the Stringer interface
func (p *Post) String() string {
	return fmt.Sprintf("%s (%s)", p.Title, p.Slug)
}

// Main creates a gorilla/mux router & dispatches requests on port :3000
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

// HomeHandler provides a welcome/index page with a listing of recents posts,
// and a link to create a new post.
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

// ShowArticle tries to load a Post identified by query param 'title'
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

// NewPost creates a new Post stuct with title and body
func NewPost(title string, body string) *Post {
	return &Post{Title: title, Body: body, Slug: slugify(title)}
}

// LoadPost attempts to load a Post from posts/ dir identified by slug
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

// save method for Posts
func (post *Post) save() error {
	b, err := json.Marshal(post)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("posts/"+post.Slug+".json", b, 0600)
}

// slugify converts a title string into a url-friendly slug string
func slugify(title string) (slug string) {
	slug = strings.ToLower(title)
	slug = regexp.MustCompile("[^a-z0-9 -]").ReplaceAllString(slug, "")
	slug = regexp.MustCompile(" +").ReplaceAllString(slug, "-")
	slug = regexp.MustCompile("-+").ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, " -")
	return
}

// renderTemplate is a utility function to simplify rendering a nested template
// tmpl with data
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

// byLatestDate implements the sort.Interface
type byLatestDate []os.FileInfo

func (f byLatestDate) Len() int           { return len(f) }
func (f byLatestDate) Less(i, j int) bool { return f[i].ModTime().After(f[j].ModTime()) }
func (f byLatestDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
