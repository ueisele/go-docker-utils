package engine

import (
	"path"
)

type Renderer struct {
	config 		 Config
	refFuncs	 []Source
	contextFuncs []Source
	fromFuncs 	 []Source
	toFuncs		 []Transform
}

func NewRenderer() *Renderer {
	return &Renderer{
		config: Config{Strict: true},
		refFuncs: make([]Source, 0),
		contextFuncs: make([]Source, 0),
		fromFuncs: make([]Source, 0),
		toFuncs: make([]Transform, 0),
	}
}

func (r* Renderer) Render() (err error) {
	engine := NewEngine(r.config)

	err = waitUntilDone(transformerAnyOrder(contentConsumer(engine.AddReferenceTemplate))(mergeSourcesAnyOrder(r.refFuncs...)()))
	if err != nil {	return }

	contextBuilder := NewContextBuilder()
	err = waitUntilDone(transformerInOrder(byTypeContextAdder(contextBuilder))(mergeSourcesInOrder(r.contextFuncs...)()))
	if err != nil {	return }
	context, err := contextBuilder.WithEnv().Build()
	if err != nil {	return }

	err = waitUntilDone(fanOutSink(r.toFuncs...)(transformerAnyOrder(renderer(engine, context))(mergeSourcesAnyOrder(r.fromFuncs...)())))

	return
}

func (r *Renderer) WithConfig(config Config) *Renderer {
	r.config = config
	return r
}

func (r *Renderer) WithReferenceTemplates(fromFunc Source) *Renderer {
	r.refFuncs = append(r.refFuncs, fromFunc)
	return r
}

func (r *Renderer) WithContext(fromFunc Source) *Renderer {
	r.contextFuncs = append(r.contextFuncs, fromFunc)
	return r
}

func (r* Renderer) From(fromFunc Source) *Renderer {
	r.fromFuncs = append(r.fromFuncs, fromFunc)
	return r
}

func (r* Renderer) To(toFunc Transform) *Renderer {
	r.toFuncs = append(r.toFuncs, toFunc)
	return r
}

func (r *Renderer) Clone() *Renderer {
	c := NewRenderer()
	c.config = r.config
	copy(c.refFuncs, r.refFuncs)
	copy(c.contextFuncs, r.contextFuncs)
	copy(c.fromFuncs, r.fromFuncs)
	copy(c.toFuncs, r.toFuncs)
	return c
}

func contentConsumer(consumer func (name, content string) error) func (*Data) *Data {
	return func (data *Data) *Data {
		err := consumer(data.Name, data.Content)
		return &Data{Name: data.Name, Content: data.Content, Error: err}
	}
}

func byTypeContextAdder(contextBuilder *ContextBuilder) func (*Data) *Data {
	return func (data *Data) *Data {
		err := contextBuilder.WithByType(path.Ext(data.Name), data.Content)
		return &Data{Name: data.Name, Content: data.Content, Error: err}
	}
}

func renderer(engine Engine, context map[string]interface{}) func (*Data) *Data {
	return func (data *Data) *Data {
		rendered, err := engine.Render(data.Content, context)
		return &Data{Name: data.Name, Content: rendered, Error: err}
	}
}