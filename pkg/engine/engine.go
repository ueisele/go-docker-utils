package engine

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

type Engine struct {
	referenceTpls map[string]string
	config Config
}

type Config struct {
	Strict bool
}

func NewEngine(referenceTpls map[string]string, config Config) *Engine {
	return &Engine{
		referenceTpls: referenceTpls,
		config: config,
	}
}

func (e *Engine) RenderOne(tpl string, context interface{}) (string, error) {
	tplUuid := uuid.New().String()
	renderedMap, err := e.RenderAll(map[string]string{tplUuid: tpl}, context)
	if err != nil {
		return "", err
	}
	return renderedMap[tplUuid], nil
}

func (e *Engine) RenderAll(tpls map[string]string, context interface{}) (rendered map[string]string, err error) {
	t, err := e.createTemplate(tpls)
	if err != nil {
		return map[string]string{}, fmt.Errorf("could not parse %v", err)
	}

	rendered, err = e.renderTemplate(t, e.keySet(tpls), context)
	if err != nil {
		return map[string]string{}, fmt.Errorf("could not render %v", err)
	}

	return
}

func (e *Engine) createTemplate(tpls map[string]string) (t *template.Template, err error) {
	t = template.New("gotpl")

	if e.config.Strict {
		t.Option("missingkey=error")
	} else {
		// Not that zero will attempt to add default values for types it knows,
		// but will still emit <no value> for others. We mitigate that later.
		t.Option("missingkey=zero")
	}

	t.Funcs(funcMap()).Funcs(template.FuncMap{"tp2": e.RenderOne})

	for name, tpl := range tpls {
		if _, err := t.New(name).Parse(tpl); err != nil {
			return nil, err
		}
	}

	for name, refTpl := range e.referenceTpls {
		if t.Lookup(name) == nil {
			if _, err := t.New(name).Parse(refTpl); err != nil {
				return nil, err
			}
		}
	}

	return
}

func (e *Engine) renderTemplate(t *template.Template, tplNames []string, context interface{}) (rendered map[string]string, err error) {
	rendered = make(map[string]string, len(tplNames))
	for _, name := range tplNames {
		var buf strings.Builder
		if err := t.ExecuteTemplate(&buf, name, context); err != nil {
			return map[string]string{}, err
		}

		// Work around the issue where Go will emit "<no value>" even if Options(missing=zero)
		// is set. Since missing=error will never get here, we do not need to handle
		// the Strict case.
		rendered[name] = strings.ReplaceAll(buf.String(), "<no value>", "")
	}
	return
}

func (e *Engine) keySet(m map[string]string) (keys []string) {
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return
}
