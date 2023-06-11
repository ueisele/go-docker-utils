package template

import (
	"path"
)

type Renderer struct {
	config      Config
	fromFuncs   []Source
	toFuncs     []Transform
	refFuncs    []Source
	valuesFuncs []Source
	filesFuncs  []Source
}

func NewRenderer() *Renderer {
	return &Renderer{
		config:      Config{Strict: true},
		fromFuncs:   make([]Source, 0),
		toFuncs:     make([]Transform, 0),
		refFuncs:    make([]Source, 0),
		valuesFuncs: make([]Source, 0),
		filesFuncs:  make([]Source, 0),
	}
}

func (r *Renderer) Render() (err error) {
	engine := NewEngine(r.config)

	err = waitUntilDone(transformerAnyOrder(contentConsumer(engine.AddReferenceTemplate))(mergeSourcesAnyOrder(r.refFuncs...)()))
	if err != nil {
		return
	}

	contextBuilder := NewContextBuilder()
	err = waitUntilDone(transformerInOrder(contentConsumer(contextAdder(contextBuilder, "Values")))(mergeSourcesInOrder(r.valuesFuncs...)()))
	if err != nil {
		return
	}

	contextFiles := newFiles()
	err = waitUntilDone(transformerInOrder(contentConsumer(contextFiles.put))(mergeSourcesInOrder(r.filesFuncs...)()))
	if err != nil {
		return
	}
	contextBuilder.WithAnyInScope(contextFiles, "Files")

	context, err := contextBuilder.WithEnv("Env").Build()
	if err != nil {
		return
	}

	err = waitUntilDone(fanOutSink(r.toFuncs...)(transformerAnyOrder(renderer(engine, context))(mergeSourcesAnyOrder(r.fromFuncs...)())))

	return
}

func (r *Renderer) WithConfig(config Config) *Renderer {
	r.config = config
	return r
}

func (r *Renderer) From(fromFunc Source) *Renderer {
	r.fromFuncs = append(r.fromFuncs, fromFunc)
	return r
}

func (r *Renderer) To(toFunc Transform) *Renderer {
	r.toFuncs = append(r.toFuncs, toFunc)
	return r
}

func (r *Renderer) WithReferenceTemplates(refFunc Source) *Renderer {
	r.refFuncs = append(r.refFuncs, refFunc)
	return r
}

func (r *Renderer) WithValues(valuesFunc Source) *Renderer {
	r.valuesFuncs = append(r.valuesFuncs, valuesFunc)
	return r
}

func (r *Renderer) WithFiles(filesFunc Source) *Renderer {
	r.filesFuncs = append(r.filesFuncs, filesFunc)
	return r
}

func (r *Renderer) Clone() *Renderer {
	c := NewRenderer()
	c.config = r.config
	copy(c.refFuncs, r.refFuncs)
	copy(c.valuesFuncs, r.valuesFuncs)
	copy(c.fromFuncs, r.fromFuncs)
	copy(c.toFuncs, r.toFuncs)
	return c
}

func contentConsumer(consumer func(name, content string) error) func(*Data) *Data {
	return func(data *Data) *Data {
		err := consumer(data.Name, data.Content)
		return &Data{Name: data.Name, Content: data.Content, Error: err}
	}
}

func contextAdder(contextBuilder *ContextBuilder, scope string) func(name, content string) error {
	return func(name, content string) error {
		return contextBuilder.WithByTypeInScope(path.Ext(name), content, scope)
	}
}

func renderer(engine Engine, context map[string]interface{}) func(*Data) *Data {
	return func(data *Data) *Data {
		rendered, err := engine.Render(data.Content, context)
		return &Data{Name: data.Name, Content: rendered, Error: err}
	}
}
