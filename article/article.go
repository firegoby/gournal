// Article represents a single post to gournal
package article

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

// An Article contains a title, body and slug (used as a permalink).
type Article struct {
	Title string
	Body  string
	Slug  string
}

// the location on disk to store Articles in JSON representation
const Dir = "./articles/"

// byLatestDate implements the sort.Interface
type byLatestDate []os.FileInfo

func (f byLatestDate) Len() int           { return len(f) }
func (f byLatestDate) Less(i, j int) bool { return f[i].ModTime().After(f[j].ModTime()) }
func (f byLatestDate) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Article Creation/Aquisition Functions ======================================

// New returns a new Article with Title and Body
func New(title string, body string) *Article {
	return &Article{Title: title, Body: body, Slug: Slugify(title)}
}

// Load attempts to load an Article from Dir identified by slug, returning the
// error if one occurs
func Load(slug string) (a *Article, err error) {
	b, err := ioutil.ReadFile(Dir + slug + ".json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// All returns a slice of all Articles located in Dir, sorted by latest date,
// returning the error if one occurs
func All() (res []*Article, err error) {
	files, err := ioutil.ReadDir(Dir)
	if err != nil {
		return
	}
	sort.Sort(byLatestDate(files))
	for _, f := range files {
		a, err := Load(f.Name()[:len(f.Name())-len(".json")])
		if err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return
}

// Article Methods ============================================================

// Save stores a JSON representation of an Article in the Dir directory
func (a *Article) Save() error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(Dir+a.Slug+".json", b, 0600)
}

// String returns a simple single line representation of an Article,
// implementing the fmt.Stringer interface
func (a *Article) String() string {
	return fmt.Sprintf("%s (%s)", a.Title, a.Slug)
}

// Utilities ==================================================================

// Slugify converts a title string into a url-friendly slug string
func Slugify(title string) (slug string) {
	slug = strings.ToLower(title)
	slug = regexp.MustCompile("[^a-z0-9 -]").ReplaceAllString(slug, "")
	slug = regexp.MustCompile(" +").ReplaceAllString(slug, "-")
	slug = regexp.MustCompile("-+").ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, " -")
	return
}
