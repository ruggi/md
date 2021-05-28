# .md

- [.md](#md)
  - [Yet-another-one :)](#yet-another-one-)
  - [Usage](#usage)
    - [Building](#building)
    - [Live server](#live-server)
  - [Templating](#templating)
    - [Dates](#dates)
    - [Pages](#pages)
    - [Functions](#functions)
  - [Configuration](#configuration)
    - [SyntaxHighlighting](#syntaxhighlighting)
    - [Custom page attributes](#custom-page-attributes)

This is a very straightforward static-site generator that uses Markdown files and a single HTML layout template.

It can be used for any kind of purpose where simple HTML and CSS are required, and it is pretty fast at generating the output files.

## Yet-another-one :)

There's definitely plenty of tools like this, in my case I wanted to have something that is as close as possible to "zero config" and just writing plain Markdown for my blog, website, and project pages.

I really don't like the complexity of having comment headers in the source files, as I don't want to spend time writing things that are _not_ the contents of my pages.

I know this may be too limiting for a lot of people, but for me this is pretty much all I need.

## Usage

First, install via `go get`:

```sh
go get github.com/ruggi/md
```

Then, run `md init` to set up a target directory:

```sh
md -d blog init
```

If the `blog` directory does not exist, it will be created. Its contents will look like this:

```sh
$ tree -a blog
blog
├── .md
│   ├── config.json
│   ├── layout.html
│   └── out
│       └── index.html
└── index.md
```

* `.md` is the directory containing the configuration files:
  * `.md/config.json` contains the configuration variables
  * `.md/layout.html` contains the HTML template used when generating the pages
* `index.md` is a sample source file that is only created if the whole directory was created as well.

### Building

Pages are generated with the `build` command:

```sh
md -d blog build
```

All `.md` and `.markdown` files will be converted to HTML files according to their relative path and put into the `.md/out` directory.

All other files (e.g. `.css`) will be simply copied over the output directory.

For example, see this example:

```plain
$ tree -a blog
blog
├── .md
│   ├── config.json
│   └── layout.html
├── about.md
├── index.md
├── posts
│   ├── 2020
│   │   ├── 1-hello.md
│   │   └── 2-syntax.md
│   └── 2021
│       └── 1-happy-new-year.md
└── projects
    ├── bar.md
    └── foo.md

$ md -d blog build
2021/05/26 22:58:20 Building blog
2021/05/26 22:58:20 /about.md
2021/05/26 22:58:20 /index.md
2021/05/26 22:58:20 /posts/2020/1-hello.md
2021/05/26 22:58:20 /posts/2020/2-syntax.md
2021/05/26 22:58:20 /posts/2021/1-happy-new-year.md
2021/05/26 22:58:20 /projects/bar.md
2021/05/26 22:58:20 /projects/foo.md
2021/05/26 22:58:20 ✔ Done (3.457977ms)

$ tree -a blog
blog
├── .md
│   ├── config.json
│   ├── layout.html
│   └── out
│       ├── about.html
│       ├── index.html
│       ├── posts
│       │   ├── 2020
│       │   │   ├── 1-hello.html
│       │   │   └── 2-syntax.html
│       │   └── 2021
│       │       └── 1-happy-new-year.html
│       └── projects
│           ├── bar.html
│           └── foo.html
├── about.md
├── index.md
├── posts
│   ├── 2020
│   │   ├── 1-hello.md
│   │   └── 2-syntax.md
│   └── 2021
│       └── 1-happy-new-year.md
└── projects
    ├── bar.md
    └── foo.md
```

### Live server

You can run an HTTP server (with auto-reloading) with the `serve` command:

```sh
md -d blog serve
```

A server will be available at <http://localhost:4000>.

As optional arguments, you can pass:

* `-w` (`--watch`), to trigger a new build when files change
* `-p` (`--port`), to specify a different port for the server (default `4000`)
* `-H` (`--host`), to specify a different listening address (default `127.0.0.1`)

## Templating

There's only one template, `.md/layout.html`.

In your template, you can reference three variables: 

* `.Content` is the HTML contents of the page, converted from its markdown source.
* `.Title` is the title of the page, being either the first line of the file, if it's a H1 heading (starting with a single `#`) or the file name (without extension).
* `.Date` is the date of the page, as parsed from the filename.
* `.Pages` which is a map referencing all the generated pages.

You can also use the `.Title` and `.Pages` variables in your markdown files!

### Dates

Dates are parsed from the file name: you can either use the Unix timestamp in seconds, followed by a dash, or the `YYYY-MM-DD` syntax, followed by a dash. The `{{.Date}}` variable is of `time.Time` type and can be used accordingly.

### Pages

Pages are stored in the `.Pages` variable; it's a map where the keys are the "nesting levels" and the keys are references to the pages contained in them.

For example, the following structure:

```sh
├── about.html
├── index.html
├── posts
│   ├── 2020
│   │   ├── 1-hello.html
│   │   └── 2-syntax.html
│   └── 2021
│       └── 1-happy-new-year.html
└── projects
    ├── bar.html
    └── foo.html
```

will be represented in the `.Pages` variable as:

```json
{
    "_": [                 // <- the root level
        {Title: "About", Path: "about.html"},
        {Title: "Index", Path: "index.html"},
        {Title: "Hello", Path: "posts/2020/1-hello.html"},
        {Title: "Syntax", Path: "posts/2020/2-syntax.html"},
        ...
    ],
    "posts": [
        {Title: "Hello", Path: "posts/2020/1-hello.html"},
        ...
    ],
    "projects": [
        {Title: "Foo", Path: "projects/foo.html"},
        ...
    ]
}
```

This is useful for creating, for example, a blog index page where all posts and projects are listed:

```md
# Welcome

This is my website.

## Blog

<ul>
{{ range $index, $page := .Pages.posts }}
    <li><a href="{{ $page.Path }}">{{ $page.Title }}</a></li>
{{ end }}
</ul>

## Projects

<ul>
{{ range $index, $page := .Pages.projects }}
    <li><a href="{{ $page.Path }}">{{ $page.Title }}</a></li>
{{ end }}
</ul>
```

### Functions

* `reverse` : inverts the order of a slice of pages
* `byDate` : sorts a slice of pages by date

Example:

```md
<ul>
    {{ range $index, $page := reverse ( byDate .Pages.posts ) }}
    <li>
        {{ $page.Date.Format "2006 01 02" }}
        <a href="{{ $page.Path }}">{{ $page.Path }}</a>
    </li>
    {{ end }}
</ul>
```

## Configuration

Configuration variables are set in the `.md/config.json` file.

```ts
{
    SyntaxHighlighting: {
        Enabled: boolean // enable or disable syntax highlighting for fenced code blocks
        Style: string // the visual style of the syntax highlithing blocks.
        LineNumbers: boolean // show line numbers
    }
}
```

### SyntaxHighlighting

You can use any of the [styles available here](https://github.com/alecthomas/chroma/tree/master/styles) in your fenced code blocks (or any other Pygmens-compatible style of your choice, but you'll have to do it manually via CSS).

### Custom page attributes

As written above, page titles are automatically generated from either the first line of the file, if it's a heading, or the file name.

In addition, you can pass custom page attributes by setting the file's first line to a HTML comment like this:

```md
<!-- {"Title": "my-custom-title", "Date": "2006-02-01"} -->
```
