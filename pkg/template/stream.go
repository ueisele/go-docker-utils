package template

import (
	"fmt"
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