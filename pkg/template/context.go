package template

import (
	"fmt"
	"github.com/imdario/mergo"
)

type ContextBuilder struct {
	contexts            []map[string]interface{}
	typeDecoderRegistry map[string]func(string) (interface{}, error)
}

func NewContextBuilder() *ContextBuilder {
	cb := &ContextBuilder{contexts: make([]map[string]interface{}, 0)}
	cb.initTypeDecoderRegistry()
	return cb
}

func (cb *ContextBuilder) initTypeDecoderRegistry() {
	cb.typeDecoderRegistry = make(map[string]func(string) (interface{}, error))
	cb.typeDecoderRegistry[".json"] = fromJSON
	cb.typeDecoderRegistry[".yml"] = fromYAML
	cb.typeDecoderRegistry[".yaml"] = fromYAML
	cb.typeDecoderRegistry[".toml"] = fromTOML
	cb.typeDecoderRegistry[".properties"] = func(text string) (interface{}, error) { return fromProperties(text) }
}

func (cb *ContextBuilder) Build() (context map[string]interface{}, err error) {
	context = make(map[string]interface{})
	for _, src := range cb.contexts {
		err = mergo.Merge(&context, src, mergo.WithOverride)
		if err != nil {
			err = fmt.Errorf("could not create context, merge was not possible: %v", err)
			return
		}
	}
	return
}

func (cb *ContextBuilder) WithMap(context map[string]interface{}) *ContextBuilder {
	cb.contexts = append(cb.contexts, context)
	return cb
}

func (cb *ContextBuilder) WithAnyInScope(context interface{}, scope string) *ContextBuilder {
	return cb.WithMap(map[string]interface{}{scope: context})
}

func (cb *ContextBuilder) WithEnv(scope string) *ContextBuilder {
	return cb.WithAnyInScope(fromEnv(), scope)
}

func (cb *ContextBuilder) WithYamlInScope(text, scope string) error {
	out, err := fromYAML(text)
	if err != nil {
		return err
	}
	cb.WithAnyInScope(out, scope)
	return nil
}

func (cb *ContextBuilder) WithJsonInScope(text, scope string) error {
	out, err := fromJSON(text)
	if err != nil {
		return err
	}
	cb.WithAnyInScope(out, scope)
	return nil
}

func (cb *ContextBuilder) WithToml(text, scope string) error {
	out, err := fromTOML(text)
	if err != nil {
		return err
	}
	cb.WithAnyInScope(out, scope)
	return nil
}

func (cb *ContextBuilder) WithProperties(text, scope string) error {
	out, err := fromProperties(text)
	if err != nil {
		return err
	}
	cb.WithAnyInScope(out, scope)
	return nil
}

func (cb *ContextBuilder) WithByTypeInScope(ext, text, scope string) error {
	decoder, ok := cb.typeDecoderRegistry[ext]
	if !ok {
		return fmt.Errorf("file type unkown, it was %s but requires one of %v", ext, cb.SupportedTypes())
	}
	out, err := decoder(text)
	if err != nil {
		return err
	}
	cb.WithAnyInScope(out, scope)
	return nil
}

func (cb *ContextBuilder) SupportedTypes() []string {
	return toKeySet(cb.typeDecoderRegistry)
}

func toKeySet[V any](v map[string]V) []string {
	keys := make([]string, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}
