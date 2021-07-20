package template

import (
	"text/template"
	"os"
	"strings"
)

// TxtFuncMap returns a 'text/template'.FuncMap
func DubTxtFuncMap() template.FuncMap {
	return template.FuncMap(DubFuncMap())
}

// DubFuncMap returns a copy of the basic function map as a map[string]interface{}.
func DubFuncMap() map[string]interface{} {
	gfm := make(map[string]interface{}, len(genericMap))
	for k, v := range genericMap {
		gfm[k] = v
	}
	return gfm
}

var genericMap = map[string]interface{}{
	"heygodub": func() string { return "Hello :)" },

	// Env functions
	"environment":  		environment,
	"env_to_prop":			env_to_prop,
	"parse_log4j_loggers":	parse_log4j_loggers,
}

// custom function that returns key, value for all environment variable keys matching prefix
// (see original envtpl: https://pypi.org/project/envtpl/)
func environment(prefix string) map[string]string {
	env := make(map[string]string)
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		if strings.HasPrefix(pair[0], prefix) {
			env[pair[0]] = pair[1]
		}
	}
	return env
}

// Converts environment variables with a prefix into key/value properties
// in order to support wildcard handling of properties.  Naming convention
// is to convert env vars to lower case and replace '_' with '.'.
// Additionally, two underscores '__' are replaced with a single underscore '_'
// and three underscores '___' are replaced with a dash '-'.
//
// For example: if these are set in the environment
//   CONTROL_CENTER_STREAMS_NUM_STREAM_THREADS=4
//   CONTROL_CENTER_STREAMS_SECURITY_PROTOCOL=SASL_SSL
//   CONTROL_CENTER_STREAMS_WITH__UNDERSCORE=foo
//   CONTROL_CENTER_STREAMS_WITH___DASH=bar
//   CONTROL_CENTER_STREAMS_SASL_KERBEROS_SERVICE_NAME=kafka
//
// then
//   env_to_props('CONTROL_CENTER_STREAMS_', 'confluent.controlcenter.streams.', exclude=['CONTROL_CENTER_STREAMS_NUM_STREAM_THREADS'])
// will produce
// 	{
// 		'confluent.controlcenter.streams.security.protocol': 'SASL_SSL',
// 		'confluent.controlcenter.streams.with_underscore': 'foo',
// 		'confluent.controlcenter.streams.with_dash': 'bar',
// 		'confluent.controlcenter.streams.sasl.kerberos.service.name': 'kafka'
// 	}
//
// Args:
//   env_prefix: prefix of environment variables to include. (e.g. CONTROL_CENTER_STREAMS_)
//   prop_prefix: prefix of the resulting properties (e.g. confluent.controlcenter.streams.)
//   exclude: list of environment variables to exclude
//
// Returns:
//   Map of matching properties.
//
// See:
//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func env_to_prop(env_prefix string, prop_prefix string, exclude ...string) map[string]string {
	props := make(map[string]string)
	for key, value := range environment(env_prefix) {
		if !contains(exclude, key) {
			raw_name := strings.ToLower(strings.TrimPrefix(key, env_prefix))
			prop_dash := strings.Join(strings.Split(raw_name, "___"), "-")
			prop_underscore := strings.Join(strings.Split(prop_dash, "__"), "_")
			prop_dot := strings.Join(strings.Split(prop_underscore, "_"), ".")
			prop_name := prop_prefix + prop_dot
			props[prop_name] = value
		}
	}
	return props
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Parses *_LOG4J_PROPERTIES string and returns a list of log4j properties.
//
// For example: if LOG4J_PROPERTIES = "foo.bar=DEBUG,baz.bam=TRACE"
//   and
//   defaults={"foo.bar: "INFO"}
//   the this function will return {"foo.bar: "DEBUG", "baz.bam": "TRACE"}
//
// Args:
//   overrides_str: String containing the overrides for the default properties.
//   defaults: Map of default log4j properties.
//
// Returns:
//   Map of log4j properties.
//
// See:
//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func parse_log4j_loggers(overrides_str string, defaults ...interface{}) map[string]string {
	props := to_map_of_strings(defaults)
	for _, override := range strings.Split(overrides_str, ",") {
		tokens := strings.SplitN(override, "=", 2)
		if len(tokens) == 2 {
			props[tokens[0]] = tokens[1]
		}
	}
	return props
}

func to_map_of_strings(args []interface{}) map[string]string {
	stringMap := make(map[string]string)
	for _, arg := range args {
		switch t := arg.(type) {
		  case map[string]string:
			for k, v := range t {
				stringMap[k] = v
			}
		  default:
			panic("Unknown argument")
		}
	}
	return stringMap
}