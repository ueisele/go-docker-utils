package template

import (
	"fmt"
	"strings"
	"text/template"
)

type Engine interface {
	AddReferenceTemplate(name string, renderable string) error
	Render(renderable string, context map[string]interface{}) (string, error)
}

type Config struct {
	Strict bool
}

type engine struct {
	config Config
	tpl *template.Template
} 

func NewEngine(config Config) Engine {
	e := &engine{config: config}
	e.initTemplate()
	return e
}

func (e *engine) initTemplate() {
	e.tpl = template.New("gotpl")
	e.tpl.Funcs(funcMap()).Funcs(template.FuncMap{"tpl": e.Render})
}

func (e *engine) AddReferenceTemplate(name string, renderable string) error {
	if e.tpl.Lookup(name) == nil {
		if _, err := e.tpl.New(name).Parse(renderable); err != nil {
			return fmt.Errorf("could not parse %s, %v", name, err)
		}
	}
	return nil
}

func (e *engine) Render(renderable string, context map[string]interface{}) (rendered string, err error) {
	t, err := e.createTemplate(renderable)
	if err != nil {
		return "", fmt.Errorf("could not parse %v", err)
	}

	rendered, err = e.renderTemplate(t, context)
	if err != nil {
		return "", fmt.Errorf("could not render %v", err)
	}

	return
}

func (e *engine) createTemplate(renderable string) (t *template.Template, err error) {
	t = template.Must(e.tpl.Clone())
	
	if e.config.Strict {
		t.Option("missingkey=error")
	} else {
		// Not that zero will attempt to add default values for types it knows,
		// but will still emit <no value> for others. We mitigate that later.
		t.Option("missingkey=zero")
	}

	_, err = t.Parse(renderable)
	return 
}

func (e *engine) renderTemplate(t *template.Template, context map[string]interface{}) (string, error) {
	var buf strings.Builder
	if err := t.Execute(&buf, context); err != nil {
		return "", err
	}

	if e.config.Strict {
		return buf.String(), nil
	} else {
		// Work around the issue where Go will emit "<no value>" even if Options(missing=zero)
		// is set. Since missing=error will never get here, we do not need to handle
		// the Strict case.
		return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
	}
}
