package template

import (
	"strings"

	"github.com/gobwas/glob"
)

// https://github.com/helm/helm/blob/main/pkg/engine/files.go

type files map[string]string

func newFiles() files {
	return make(map[string]string)
}

func (f files) put(name, content string) error {
	f[name] = content
	return nil
}

// Get returns a string representation of the given file.
//
// Fetch the contents of a file as a string. It is designed to be called in a
// template.
//
// {{.Files.Get "foo"}}
//
// In a template context, this is identical to calling {{index .Files $path}}.
//
// This is intended to be accessed from within a template, so a missed key returns
// an empty string.
func (f files) Get(name string) string {
	if v, ok := f[name]; ok {
		return v
	}
	return ""
}

// GetBytes gets a file by path.
//
// The returned data is raw. In a template context, this is identical to calling
// {{index .Files $path}}.
//
// This is intended to be accessed from within a template, so a missed key returns
// an empty []byte.
func (f files) GetBytes(name string) []byte {
	return []byte(f.Get(name))
}

// Glob takes a glob pattern and returns another files object only containing
// matched  files.
//
// This is designed to be called from a template.
//
// {{ range $name, $content := .Files.Glob("foo/**") }}
// {{ $name }}: |
// {{ .Files.Get($name) | indent 4 }}{{ end }}
func (f files) Glob(pattern string) files {
	g, err := glob.Compile(pattern, '/')
	if err != nil {
		g, _ = glob.Compile("**")
	}

	nf := newFiles()
	for name, contents := range f {
		if g.Match(name) {
			nf[name] = contents
		}
	}

	return nf
}

// Lines returns each line of a named file (split by "\n") as a slice, so it can
// be ranged over in your templates.
//
// This is designed to be called from a template.
//
// {{ range .Files.Lines "foo/bar.html" }}
// {{ . }}{{ end }}
func (f files) Lines(path string) []string {
	if f == nil || f[path] == "" {
		return []string{}
	}
	return strings.Split(f[path], "\n")
}
