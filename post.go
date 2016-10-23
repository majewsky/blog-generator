/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
*
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* This program is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* this program. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-commonmark/markdown"
)

//Post is a blog post.
type Post struct {
	CreationTimestamp   uint64
	LastEditedTimestamp uint64
	Slug                string
	Markdown            []byte
	HTML                string
}

//Posts is a list of Post (only required for sorting).
type Posts []*Post

func (p Posts) Len() int           { return len(p) }
func (p Posts) Less(i, j int) bool { return p[i].CreationTimestamp < p[j].CreationTimestamp }
func (p Posts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func allPosts() []*Post {
	dir, err := os.Open("data/posts")
	FailOnErr(err)
	fis, err := dir.Readdir(-1)
	FailOnErr(err)

	var posts Posts
	for _, fi := range fis {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".md") {
			posts = append(posts, NewPost(fi.Name()))
		}
	}

	return posts
}

//NewPost creates a new Post instance.
func NewPost(fileName string) *Post {
	//check `git log` for creation and last modification timestamp
	cmd := exec.Command(
		"git", "-C", "data",
		"log", "--pretty=%at", "-M", "--follow",
		"--", "posts/"+fileName,
	)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	FailOnErr(cmd.Run())

	var creationTimestamp uint64 = 0
	var lastEditedTimestamp uint64 = 0
	for _, str := range strings.Fields(string(buf.Bytes())) {
		timestamp, _ := strconv.ParseUint(str, 10, 64)
		//"last edited" is the chronologically largest timestamp
		if lastEditedTimestamp < timestamp {
			lastEditedTimestamp = timestamp
		}
		//"creation" is the timestamp on the hierarchically lowest timestamp
		creationTimestamp = timestamp
	}

	//read contents
	markdownBytes, err := ioutil.ReadFile("data/posts/" + fileName)
	FailOnErr(err)

	//generate HTML
	return &Post{
		CreationTimestamp:   creationTimestamp,
		LastEditedTimestamp: lastEditedTimestamp,
		Slug:                strings.TrimSuffix(fileName, ".md"),
		Markdown:            markdownBytes,
		HTML:                markdown.New(markdown.HTML(true)).RenderToString(markdownBytes),
	}
}

//OutputFileName returns the output filename below output/ for this Post.
func (p *Post) OutputFileName() string {
	return "posts/" + p.Slug + ".html"
}

var initialHeadingRx = regexp.MustCompile(`^<h1>(.+?)</h1>`)

//Title returns the contents of the first <h1>, or the slug as a fallback.
func (p *Post) Title() string {
	match := initialHeadingRx.FindStringSubmatch(p.HTML)
	if match != nil {
		return match[1]
	}
	return p.Slug
}

var innerHeadingsRx = regexp.MustCompile(`(?s)^(.+?)<h[1-6]>`)

//ShortenedHTML is like HTML, but cut off before the second heading.
func (p *Post) ShortenedHTML() string {
	match := innerHeadingsRx.FindStringSubmatch(p.HTML)
	if match == nil {
		return p.HTML
	}
	return match[1] + fmt.Sprintf(
		"<p class=\"more\"><a href=\"%s\">Read more...</a></p>",
		p.OutputFileName(),
	)
}

//CreationTime returns the creation timestamp as time.Time object in UTC.
func (p *Post) CreationTime() time.Time {
	return time.Unix(int64(p.CreationTimestamp), 0).UTC()
}

//LastEditedTime returns the timestamp of the last edit as time.Time object in UTC.
func (p *Post) LastEditedTime() time.Time {
	return time.Unix(int64(p.LastEditedTimestamp), 0).UTC()
}

//Render writes the post to its output file.
func (p *Post) Render() {
	str := p.HTML
	ctime := p.CreationTime().Format(time.RFC1123)
	mtime := p.LastEditedTime().Format(time.RFC1123)

	if ctime == mtime {
		str += fmt.Sprintf("<p><i>Created: %s</i></p>", ctime)
	} else {
		historyUrl := fmt.Sprintf("https://github.com/majewsky/blog-data/commits/master/posts/%s.md", p.Slug)
		str += fmt.Sprintf(
			"<p><i>Created: %s</i><br><i>Last edited: <a href=\"%s\" title=\"Commits on GitHub\">%s</a></i></p>",
			ctime, historyUrl, mtime)
	}

	writeFile(p.OutputFileName(), p.Title(), str)
}
