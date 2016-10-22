package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-commonmark/markdown"
)

func main() {
	//prepare output directory
	err := os.MkdirAll("output/posts", 0755)
	FailOnErr(err)

	//list posts
	posts, err := allPosts()
	FailOnErr(err)
	sort.Sort(Posts(posts))

	//deduplicate slugs
	slugSeen := make(map[string]bool)
	for _, post := range posts {
		if slugSeen[post.Slug] {
			//deduplicate "$slug" to "$slug-1", "$slug-2" etc.
			i := 0
			for {
				i++
				altSlug := fmt.Sprintf("%s-%d", post.Slug, i)
				if !slugSeen[altSlug] {
					post.Slug = altSlug
					break
				}
			}
		}
		slugSeen[post.Slug] = true
		continue
	}

	//render posts
	for _, post := range posts {
		writeFile(post.OutputFileName(), post.HTML)
	}

	//index.html and all.html show posts in reverse order
	reverse(posts)
	RenderIndex(posts)
	RenderAll(posts)

	//TODO: generate index.html, all.html
}

////////////////////////////////////////////////////////////////////////////////
// Post

//Post is a blog post.
type Post struct {
	Timestamp uint64
	Slug      string
	Markdown  []byte
	HTML      string
}

//Posts is a list of Post (only required for sorting).
type Posts []*Post

func (p Posts) Len() int           { return len(p) }
func (p Posts) Less(i, j int) bool { return p[i].Timestamp < p[j].Timestamp }
func (p Posts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var postFilenameRx = regexp.MustCompile(`^(\d{10})-([^/]+)\.md$`)

func allPosts() ([]*Post, error) {
	dir, err := os.Open("posts")
	if err != nil {
		return nil, err
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var posts Posts
	for _, fi := range fis {
		if fi.Mode().IsRegular() && postFilenameRx.MatchString(fi.Name()) {
			post, err := NewPost(fi.Name())
			FailOnErr(err) //should be unreachable
			posts = append(posts, post)
		}
	}

	return posts, nil
}

//NewPost creates a new Post instance.
func NewPost(fileName string) (*Post, error) {
	//parse filename
	match := postFilenameRx.FindStringSubmatch(fileName)
	if match == nil {
		return nil, fmt.Errorf("%s is not a valid post filename", fileName)
	}
	timestamp, _ := strconv.ParseUint(match[1], 10, 64)
	slug := match[2]

	//read contents
	markdownBytes, err := ioutil.ReadFile("posts/" + fileName)
	FailOnErr(err)

	//generate HTML

	return &Post{
		Timestamp: timestamp,
		Slug:      slug,
		Markdown:  markdownBytes,
		HTML:      markdown.New().RenderToString(markdownBytes),
	}, nil
}

//OutputFileName returns the output filename below output/ for this Post.
func (p *Post) OutputFileName() string {
	return "posts/" + p.Slug + ".html"
}

////////////////////////////////////////////////////////////////////////////////
// output formatting

var innerHeadingsRx = regexp.MustCompile(`(?s)^(.+?)<h[1-6]>`)

//RenderIndex generates the index.html page.
func RenderIndex(posts []*Post) {
	//not more than 10 posts
	if len(posts) > 10 {
		posts = posts[:10]
	}

	//accumulate posts
	articlesStr := ""
	if len(posts) > 0 {
		articles := make([]string, 0, len(posts))
		for _, post := range posts {
			//shorten post.HTML if it contains multiple headings
			htmlStr := post.HTML
			match := innerHeadingsRx.FindStringSubmatch(htmlStr)
			if match != nil {
				htmlStr = match[1]
				htmlStr += fmt.Sprintf(
					"<p class=\"more\"><a href=\"%s\">Read more...</a></p>",
					post.OutputFileName(),
				)
			}
			articles = append(articles, htmlStr)
		}
		articlesStr = "<article>" + strings.Join(articles, "</article><article>") + "</article>"
	}

	writeFile("index.html", articlesStr)
}

var initialHeadingRx = regexp.MustCompile(`^<h1>(.+?)</h1>`)

//RenderAll generates the all.html page.
func RenderAll(posts []*Post) {
	items := ""
	for _, post := range posts {
		//show either the initial <h1> or fall back to the slug
		var itemText string
		match := initialHeadingRx.FindStringSubmatch(post.HTML)
		if match == nil {
			itemText = post.Slug
		} else {
			itemText = match[1]
		}
		items += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", post.OutputFileName(), itemText)
	}

	writeFile("all.html", "<section class=\"all\"><ul>"+items+"</ul></section>")
}

////////////////////////////////////////////////////////////////////////////////
// utilities

//FailOnErr complains and aborts if `err != nil`.
func FailOnErr(err error) {
	if err != nil {
		os.Stderr.Write([]byte(err.Error() + "\n"))
		os.Exit(1)
	}
}

func reverse(list []*Post) {
	max := len(list) - 1
	cnt := len(list) / 2
	for idx := 0; idx < cnt; idx++ {
		list[idx], list[max-idx] = list[max-idx], list[idx]
	}
}

func writeFile(path string, contents string) {
	FailOnErr(ioutil.WriteFile("output/"+path, []byte(contents), 0644))
}
