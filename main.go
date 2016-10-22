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
	"sort"
	"strings"
)

func main() {
	//prepare output directory
	err := os.MkdirAll("output/posts", 0755)
	FailOnErr(err)

	//list posts
	posts := allPosts()
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
		post.Render()
	}

	//index.html and sitemap.html show posts in reverse order
	reverse(posts)
	RenderIndex(posts)
	RenderAll(posts)

	//write additional assets
	FailOnErr(ioutil.WriteFile("output/style.css", []byte(AssetStyleCss), 0644))
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
			//include permalink in initial heading
			htmlStr = initialHeadingRx.ReplaceAllStringFunc(htmlStr, func(h1str string) string {
				match := initialHeadingRx.FindStringSubmatch(h1str)
				return fmt.Sprintf("<h1><a href=\"%s\" title=\"Permalink\">[l]</a> %s</h1>",
					post.OutputFileName(), match[1],
				)
			})
			articles = append(articles, htmlStr)
		}
		articlesStr = "<article>" + strings.Join(articles, "</article><article>") + "</article>"
	}

	writeFile("index.html", "", articlesStr)
}

//RenderAll generates the sitemap.html page.
func RenderAll(posts []*Post) {
	items := ""
	currentMonth := ""

	for _, post := range posts {
		//add a month header when this post is from a different month than the previous one
		month := post.CreationTime().Format("Jan 2006")
		if month != currentMonth {
			items += fmt.Sprintf("</ul><h2>%s</h2><ul>", month)
			currentMonth = month
		}
		//show either the initial <h1> or fall back to the slug
		items += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", post.OutputFileName(), post.Title())
	}

	items = strings.TrimPrefix(items, "</ul>")
	writeFile("sitemap.html", "Article list",
		"<section class=\"sitemap\">"+items+"</ul></section>",
	)
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

func writeFile(path, title, contents string) {
	str := AssetTemplateHtml

	slashCount := strings.Count(path, "/")
	dotdots := make([]string, 0, slashCount)
	for idx := 0; idx < slashCount; idx++ {
		dotdots = append(dotdots, "..")
	}
	if len(dotdots) == 0 {
		dotdots = []string{"."}
	}
	str = strings.Replace(str, "%PATH_TO_ROOT%", strings.Join(dotdots, "/"), -1)

	if title == "" {
		str = strings.Replace(str, "%TITLE%", "Stefan's Blog", -1)
	} else {
		str = strings.Replace(str, "%TITLE%", title+" &ndash; Stefan's Blog", -1)
	}
	str = strings.Replace(str, "%CONTENT%", contents, -1)

	FailOnErr(ioutil.WriteFile("output/"+path, []byte(str), 0644))
}
