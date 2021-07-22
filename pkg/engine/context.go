package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type ContextBuilder struct {
	contexts []interface{}
	errors	 []string
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		contexts: make([]interface{}, 0),
		errors:	make([]string, 0),
	}
}

func (cb *ContextBuilder) Build() (context interface{}, err error) {
	if len(cb.errors) > 0 {
		err = fmt.Errorf("could not create context:\n%v", strings.Join(cb.errors, "\n"))
		return
	}

	context = make(map[string]interface{})
	for _, src := range cb.contexts {
		mergo.Merge(context, src, mergo.WithOverride)
	}
	return
}

func (cb *ContextBuilder) WithMap(context map[string]interface{}) *ContextBuilder {
	cb.contexts = append(cb.contexts, context)
	return cb
}

func (cb *ContextBuilder) WithKeyValues(keyValues []string) *ContextBuilder {
	env := make(map[string]interface{})
	for _, setting := range keyValues {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	cb.contexts = append(cb.contexts, env)
	return cb
}

func (cb *ContextBuilder) WithEnv() *ContextBuilder {
	return cb.WithKeyValues(os.Environ())
}

func (cb *ContextBuilder) WithYaml(text string) *ContextBuilder {
	context := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(text), &context)
	if err != nil {
		cb.errors = append(cb.errors, err.Error())
	} else {
		cb.contexts = append(cb.contexts, context)
	}
	return cb
}

func (cb *ContextBuilder) WithJson(text string) *ContextBuilder {
	context := make(map[string]interface{})
	err := json.Unmarshal([]byte(text), &context)
	if err != nil {
		cb.errors = append(cb.errors, err.Error())
	} else {
		cb.contexts = append(cb.contexts, context)
	}
	return cb
}