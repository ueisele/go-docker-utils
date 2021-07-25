package engine

import (
	"fmt"
	"path"
	"strings"
	"sync"
)

type Data struct {
	Name    string
	Content string
	Error   error
}

type Source func () <-chan *Data

type Transform func (<-chan *Data) <-chan *Data

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

func mergeSourcesInOrder(sources ...Source) Source {
	return mergeSources(sources, mergeChannelsInOrder)
}

func mergeSourcesAnyOrder(sources ...Source) Source {
	return mergeSources(sources, mergeChannelsAnyOrder)
}

func mergeSources(sources []Source, channelMerger func (cs ...<-chan *Data) <-chan *Data) Source {
	if len(sources) == 0 {
		return emptySource
	}
	if len(sources) == 1 {
		return sources[0]
	}
	return func() <-chan *Data {
		channels := make([]<-chan *Data, 0)
		for _, source := range sources {
			channels = append(channels, source())
		}
		return channelMerger(channels...)
	}
}

func mergeChannelsInOrder(cs ...<-chan *Data) <-chan *Data {
	out := make(chan *Data)
	go func () {
		defer close(out)
		for _, c := range cs {
			for v := range c {
				out <- v
			}
		}
	}()
	return out
}

func mergeChannelsAnyOrder(cs ...<-chan *Data) <-chan *Data {
	out := make(chan *Data)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan *Data) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func emptySource() <-chan *Data {
	out := make(chan *Data)
	close(out)
	return out 
}

func transformerInOrder(function func (*Data) *Data) Transform {
	return func (input <-chan *Data) <-chan *Data {
		out := make(chan *Data)
		go func() {
			defer close(out)
			for data := range input {
				if data.Error == nil {
					out <- function(data)
				} else {
					out <- data
				}
			}
		}()
		return out
	}
}

func transformerAnyOrder(function func (*Data) *Data) Transform {
	return func (input <-chan *Data) <-chan *Data {
		var wg sync.WaitGroup
		out := make(chan *Data)
		
		doRender := func(data *Data) {
			defer wg.Done()
			if data.Error == nil {
				out <- function(data)
			} else {
				out <- data
			}
		}

		go func() {
			defer close(out)
			for data := range input {
				wg.Add(1)
				go doRender(data)
			}
			wg.Wait()
		}()

		return out
	}
}

func fanOutSink(sinks ...Transform) Transform {
	if len(sinks) == 0 {
		return noopSink
	}
	if len(sinks) == 1 {
		return sinks[0]
	}
	return func (input <-chan *Data) <-chan *Data {
		channels := make([]<-chan *Data, 0)
		for _, sink := range sinks {
			channels = append(channels, sink(input))
		}
		out := make(chan *Data)
		go func () {
			defer close(out)
			for _, c := range channels {
				for v := range c {
					out <- v
				}
			}
		}()
		return out
	}
}

func noopSink(input <-chan *Data) <-chan *Data  {
	out := make(chan *Data)
	go func () {
		defer close(out)
		for data := range input {
			out <- data
		}
	}()
	return out
}

func waitUntilDone(channel <-chan *Data) (err error) {
	errors := make([]string, 0)
	for data := range channel {
		if data.Error != nil {
			errors = append(errors, fmt.Errorf("%s was not completed successfully: %v", data.Name, data.Error).Error())
		}
	}
	if len(errors) > 0 {
		err = fmt.Errorf("failed to complete %d inputs:\n\t%s", len(errors), strings.Join(errors, "\n\t"))
	}
	return
}