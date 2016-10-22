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
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-commonmark/markdown"
)

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

var initialHeadingRx = regexp.MustCompile(`^<h1>(.+?)</h1>`)

//Title returns the contents of the first <h1>, or the slug as a fallback.
func (p *Post) Title() string {
	match := initialHeadingRx.FindStringSubmatch(p.HTML)
	if match != nil {
		return match[1]
	}
	return p.Slug
}

//Time returns the post timestamp as time.Time object in UTC.
func (p *Post) Time() time.Time {
	return time.Unix(int64(p.Timestamp), 0).UTC()
}

//Render writes the post to its output file.
func (p *Post) Render() {
	timeStr := p.Time().Format(time.RFC1123)
	str := p.HTML + fmt.Sprintf("<p><i>Written: %s</i></p>", timeStr)

	writeFile(p.OutputFileName(), p.Title(), str)
}
