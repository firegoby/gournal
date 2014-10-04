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
	"log"
	"net/http"
	"text/template"

	"github.com/firegoby/gournal/article"
	"github.com/firegoby/mux"
)

// Main creates a gorilla/mux router & dispatches requests on port :3000
func main() {
	r := mux.NewRouter().StrictSlash(true).HTTPMethodOverride(true)

	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/articles/new", NewArticleHandler).Methods("GET")
	r.HandleFunc("/articles", CreateArticleHandler).Methods("POST")
	r.HandleFunc("/articles/{title}", ShowArticleHandler).Methods("GET")
	r.HandleFunc("/articles/{title}/edit", EditArticleHandler).Methods("GET")
	r.HandleFunc("/articles/{title}", UpdateArticleHandler).Methods("PUT")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Listening on 3000...")
	http.ListenAndServe(":3000", r)
}

// HomeHandler provides a welcome/index page with a listing of recents posts,
// and a link to create a new post.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := article.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderTemplate(w, "home", articles)
}

// Article REST Functions - implements RESTfulResource interface ==============

// IndexArticleHandler is a RESTful function for GET /articles
func IndexArticleHandler(w http.ResponseWriter, r *http.Request) {
	// no need for a dedicated Articles index, gournal Home serves that purpose
	http.Redirect(w, r, "/", http.StatusFound)
}

// NewArticleHandler is a RESTful function for GET /articles/new
func NewArticleHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new_article", nil)
}

// CreateArticleHandler is a RESTful function for POST /articles/new
func CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	a := article.New(r.FormValue("title"), r.FormValue("body"))
	err := a.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/articles/"+a.Slug, http.StatusFound)
}

// ShowArticleHandler is a RESTful function for GET /articles/:id
func ShowArticleHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	a, err := article.Load(params["title"])
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "show_article", a)
}

// EditArticleHandler is a RESTful function for GET /articles/:id/edit
func EditArticleHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	a, err := article.Load(params["title"])
	if err != nil {
		log.Println(err.Error())
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "edit_article", a)
}

// UpdateArticleHandler is a RESTful function for PUT /articles/:id
func UpdateArticleHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	a, err := article.Load(params["title"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Title = r.FormValue("title")
	a.Body = r.FormValue("body")

	err = a.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/articles/"+a.Slug, http.StatusFound)
}

// DestroyArticleHandler is a RESTful function for DELETE /articles/:id
func DestroyArticleHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Utilities ==================================================================

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
