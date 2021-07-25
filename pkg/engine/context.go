package engine

import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"github.com/BurntSushi/toml"
	"github.com/magiconair/properties"
	"github.com/imdario/mergo"
)

type ContextBuilder struct {
	contexts     []map[string]interface{}
	typeRegistry map[string]func (string) error
}

func NewContextBuilder() *ContextBuilder {
	cb := &ContextBuilder{contexts: make([]map[string]interface{}, 0)}
	cb.initTypeRegistry()
	return cb
}

func (cb *ContextBuilder) initTypeRegistry() {
	cb.typeRegistry = make(map[string]func (string) error)
	cb.typeRegistry["json"] = cb.WithJson
	cb.typeRegistry["yml"] = cb.WithYaml
	cb.typeRegistry["yaml"] = cb.WithYaml
	cb.typeRegistry["toml"] = cb.WithToml
	cb.typeRegistry["properties"] = cb.WithProperties
}

func (cb *ContextBuilder) Build() (context map[string]interface{}, err error) {
	context = make(map[string]interface{})
	for _, src := range cb.contexts {
		err = mergo.Merge(context, src, mergo.WithOverride)
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

func (cb *ContextBuilder) WithEnv() *ContextBuilder {
	env := make(map[string]interface{})
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	cb.contexts = append(cb.contexts, env)
	return cb
}

func (cb *ContextBuilder) WithYaml(text string) error {
	context := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(text), &context)
	if err != nil {
		return err
	}
	cb.contexts = append(cb.contexts, context)
	return nil
}

func (cb *ContextBuilder) WithJson(text string) error {
	context := make(map[string]interface{})
	err := json.Unmarshal([]byte(text), &context)
	if err != nil {
		return err
	}
	cb.contexts = append(cb.contexts, context)
	return nil
}

func (cb *ContextBuilder) WithToml(text string) error {
	context := make(map[string]interface{})
	err := toml.Unmarshal([]byte(text), &context)
	if err != nil {
		return err
	}
	cb.contexts = append(cb.contexts, context)
	return nil
}

func (cb *ContextBuilder) WithProperties(text string) error {
	context := make(map[string]interface{})
	props, err := properties.Load([]byte(text), properties.ISO_8859_1)
	for key, value := range props.Map() {
		context[key] = value
	}
	if err != nil {
		return err
	}
	cb.contexts = append(cb.contexts, context)
	return nil
}

func (cb *ContextBuilder) WithByType(ext, text string) error {
	adder, ok := cb.typeRegistry[ext]
	if !ok {
		return fmt.Errorf("file type unkown, it was %s but requires one of %v", ext, cb.SupportedTypes())
	}
	return adder(text)
}

func (cb *ContextBuilder) SupportedTypes() []string {
	return toKeySet(cb.typeRegistry)
}

func toKeySet(v map[string]func (string) error) []string {
	keys := make([]string, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}