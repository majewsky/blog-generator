# blog-generator

This is the program that renders https://blog.bethselamin.de.

## Installation

Just `go get` this repo and find the `blog-generator` binary in `$GOBIN`. If
you're not familiar with Go, use the following commands to install to
`$HOME/bin` without leaving any garbage lying around:

```
TEMPDIR="$(mktemp -d)"
GOPATH="$TEMPDIR" GOBIN="$HOME/bin" go get https://github.com/majewsky/blog-generator
rm -r -- "$TEMPDIR"
```

## Usage

Call with one argument, the path to the configuration file. The configuration looks like so:

```
source-url https://github.com/majewsky/blog-data
source-dir /home/stefan/git/blog-data

target-url https://blog.bethselamin.de
target-dir ./output

page-name  Stefan's Blog
page-desc  Personal blog of Stefan Majewsky
```

Empty lines, leading and trailing whitespace, and lines starting with `#` are
ignored. The following directives are known:

* `source-url <url>` sets the URL to the Git repository on Github (or any other
  service that lists commits to a file at `<url>/commits/<branch>/<file>`).
* `source-dir <path>` sets the path to the source directory, where the repo from
  above is checked out.
* `target-url <url>` sets the URL where the blog is published. This is only
  needed for rendering the RSS feed; the HTML templates use relative links.
* `target-dir <path>` set the path to the target directory, the directory that
  is equivalent to the `target-url`.
* `page-name <string>` is the name of the website, as used in the HTML page titles.
* `page-desc <string>` is a slightly longer description of the website,
  currently used for the RSS feed only.

### The source directory

The source repository must contain the following files:

* `posts/*.md`: the actual blog posts, in Markdown format -- The generator will
  look at the `git log` of each file and use the first commit timestamp as
  "Created" and the last one as "Last Modified". The creation timestamp
  determines sorting of posts.
* `template.html`: the template for all generated HTML pages

You can also put any further assets (images, CSS, JS) into the repo, but the
generator won't touch them. But since you're invoking this program, you can
probably invoke other programs, such as `cp` or `rsync`, to copy these to the
output directory easily. ;)

### The page template

The `template.html` is just HTML, except for the following placeholders, which
will be replaced by the generator:

* `%TITLE%` will be replaced by the value of the `page-name` configuration
  variable (on single post pages: preceded by the post title).
* `%PATH_TO_ROOT%` can be used in place of the `target-url` to generate relative
  instead of absolute URLs.
* `%CONTENT%` will be replaced by the generated page contents.
* `%META%` will be replaced by a bunch of `<meta>` tags that describe the
  rendered page in terms of the [Open Graph Protocol](http://ogp.me/). This
  must go into the `<head>` of the template.

If these descriptions confuse you, have a look at [my own template.html](https://github.com/majewsky/blog-data/blob/master/template.html).

### A note about Markdown

The markdown dialect recognized by the generator is CommonMark. There is one
caveat: The generator sort of expects every post to start with a top-level
heading, either

    # The head line

or

    The head line
    =============

I don't say that it's required, but I don't guarantee stuff to look good
otherwise.
