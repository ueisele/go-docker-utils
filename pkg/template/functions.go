package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig"
	"github.com/magiconair/properties"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra functionality
	extra := template.FuncMap{
		"heygodub": func() string { return "Hello :)" },

		// Env functions
		"hasEnv":    hasEnv,
		"fromEnv":   fromEnv,
		"envToMap":  envToMap,
		"envToProp": envToProp,

		// Map functions
		"excludeKeys":      excludeKeys,
		"replaceKeyPrefix": replaceKeyPrefix,
		"toPropertiesKey":  toPropertiesKey,

		// String functions
		"kvCsvToMap": kvCsvToMap,

		// List functions
		"filterHasPrefix": filterHasPrefix,

		// Verify functions
		"required": required,

		// Network functions
		"ipAddresses":  ipAddresses,
		"ipAddress":    ipAddress,
		"anyIpAddress": anyIpAddress,

		// Format functions
		"toYAML":         toYAML,
		"fromYAML":       fromYAML,
		"toJSON":         toJSON,
		"toJSONPretty":   toJSONPretty,
		"fromJSON":       fromJSON,
		"toTOML":         toTOML,
		"fromTOML":       fromTOML,
		"toProperties":   toProperties,
		"fromProperties": fromProperties,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// checks if an environment variable with the given key exists
func hasEnv(key string) bool {
	_, has := os.LookupEnv(key)
	return has
}

func fromEnv() map[string]interface{} {
	env := make(map[string]interface{})
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	return env
}

// custom function that returns key/value for all environment variable keys matching prefix
func envToMap(prefix string) map[string]interface{} {
	env := make(map[string]interface{})
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
//
//	CONTROL_CENTER_STREAMS_NUM_STREAM_THREADS=4
//	CONTROL_CENTER_STREAMS_SECURITY_PROTOCOL=SASL_SSL
//	CONTROL_CENTER_STREAMS_WITH__UNDERSCORE=foo
//	CONTROL_CENTER_STREAMS_WITH___DASH=bar
//	CONTROL_CENTER_STREAMS_SASL_KERBEROS_SERVICE_NAME=kafka
//
// then
//
//	env_to_props('CONTROL_CENTER_STREAMS_', 'confluent.controlcenter.streams.', exclude=['CONTROL_CENTER_STREAMS_NUM_STREAM_THREADS'])
//
// will produce
//
//	{
//		'confluent.controlcenter.streams.security.protocol': 'SASL_SSL',
//		'confluent.controlcenter.streams.with_underscore': 'foo',
//		'confluent.controlcenter.streams.with_dash': 'bar',
//		'confluent.controlcenter.streams.sasl.kerberos.service.name': 'kafka'
//	}
//
// Args:
//
//	env_prefix: prefix of environment variables to include. (e.g. CONTROL_CENTER_STREAMS_)
//	prop_prefix: prefix of the resulting properties (e.g. confluent.controlcenter.streams.)
//	exclude: list of environment variables to exclude
//
// Returns:
//
//	Map of matching properties.
//
// See:
//
//	Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func envToProp(env_prefix string, prop_prefix string, exclude ...interface{}) map[string]interface{} {
	return toPropertiesKey(replaceKeyPrefix(env_prefix, prop_prefix, excludeKeys(exclude, envToMap(env_prefix))))
}

func excludeKeys(exclude interface{}, sourceMap interface{}) map[string]interface{} {
	sourceMapVal := reflect.ValueOf(sourceMap)
	switch sourceMapVal.Kind() {
	case reflect.Map:
		resultMap := make(map[string]interface{})
		excludeStrings := toFlatListOfStrings(exclude)
		iter := sourceMapVal.MapRange()
		for iter.Next() {
			key := strval(iter.Key().Interface())
			value := iter.Value().Interface()
			if !contains(excludeStrings, key) {
				resultMap[key] = value
			}
		}
		return resultMap
	default:
		panic(fmt.Errorf("must be a map but was %T", sourceMap))
	}
}

func strval(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case error:
		return t.Error()
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

func replaceKeyPrefix(prefix string, replacement string, sourceMap interface{}) map[string]interface{} {
	sourceMapVal := reflect.ValueOf(sourceMap)
	switch sourceMapVal.Kind() {
	case reflect.Map:
		resultMap := make(map[string]interface{})
		iter := sourceMapVal.MapRange()
		for iter.Next() {
			key := strval(iter.Key().Interface())
			value := iter.Value().Interface()
			resultMap[replacement+strings.TrimPrefix(key, prefix)] = value
		}
		return resultMap
	default:
		panic(fmt.Errorf("must be a map but was %T", sourceMap))
	}
}

func toPropertiesKey(sourceMap interface{}) map[string]interface{} {
	sourceMapVal := reflect.ValueOf(sourceMap)
	switch sourceMapVal.Kind() {
	case reflect.Map:
		props := make(map[string]interface{})
		to_dot_pattern := regexp.MustCompile("[^_](_)[^_]")
		iter := sourceMapVal.MapRange()
		for iter.Next() {
			key := strval(iter.Key().Interface())
			value := iter.Value().Interface()
			raw_name := strings.ToLower(key)
			var prop_dot string = raw_name
			for matches := to_dot_pattern.FindAllString(prop_dot, -1); len(matches) > 0; matches = to_dot_pattern.FindAllString(prop_dot, -1) {
				for _, frac := range matches {
					prop_dot = strings.Replace(prop_dot, frac, strings.ReplaceAll(frac, "_", "."), 1)
				}
			}
			prop_dash := strings.Join(strings.Split(prop_dot, "___"), "-")
			prop_underscore := strings.Join(strings.Split(prop_dash, "__"), "_")
			props[prop_underscore] = value
		}
		return props
	default:
		panic(fmt.Errorf("must be a map but was %T", sourceMap))
	}
}

// Parses a list of key/value pairs separated by commas.
//
// For example for "foo.bar=DEBUG,baz.bam=TRACE"
//
//	the this function will return {"foo.bar: "DEBUG", "baz.bam": "TRACE"}
//
// Args:
//
//	kvList: String containing the comma separated list of key/value pairs.
//
// Returns:
//
//	Map of key/value pairs.
//
// See:
//
//	Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func kvCsvToMap(kvList string) map[string]interface{} {
	props := make(map[string]interface{})
	for _, override := range strings.Split(kvList, ",") {
		tokens := strings.SplitN(override, "=", 2)
		if len(tokens) == 2 {
			props[tokens[0]] = tokens[1]
		}
	}
	return props
}

func filterHasPrefix(prefix string, list interface{}) (interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		filteredList := make([]interface{}, 0)

		l2 := reflect.ValueOf(list)

		if l2.Len() == 0 {
			return filteredList, nil
		}

		for i := 0; i < l2.Len(); i++ {
			result, err := filterHasPrefix(prefix, l2.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			if result != nil {
				filteredList = append(filteredList, result)
			}
		}

		return filteredList, nil
	default:
		switch t := list.(type) {
		case string:
			if strings.HasPrefix(t, prefix) {
				return t, nil
			}
			return nil, nil
		case fmt.Stringer:
			if strings.HasPrefix(t.String(), prefix) {
				return t, nil
			}
			return nil, nil
		case []byte:
			if strings.HasPrefix(string(t), prefix) {
				return t, nil
			}
			return nil, nil
		case error:
			if strings.HasPrefix(t.Error(), prefix) {
				return t, nil
			}
			return nil, nil
		case nil:
			return nil, nil
		default:
			return nil, fmt.Errorf("should be type of slice, array, string, fmt.Stringer, []byte or error, but %s", tp)
		}
	}
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

func toFlatListOfStrings(args ...interface{}) []string {
	stringList := make([]string, 0)
	for _, arg := range args {
		switch reflect.TypeOf(arg).Kind() {
		case reflect.Slice, reflect.Array:
			value := reflect.ValueOf(arg)
			for i := 0; i < value.Len(); i++ {
				stringList = append(stringList, toFlatListOfStrings(value.Index(i).Interface())...)
			}
		default:
			switch t := arg.(type) {
			case nil:
				break
			case string:
				stringList = append(stringList, t)
			case []byte:
				stringList = append(stringList, string(t))
			case error:
				stringList = append(stringList, t.Error())
			case fmt.Stringer:
				stringList = append(stringList, t.String())
			default:
				stringList = append(stringList, fmt.Sprintf("%v", t))
			}
		}
	}
	return stringList
}

func required(warn string, val interface{}) (interface{}, error) {
	if val == nil || reflect.ValueOf(val).IsNil() {
		return nil, fmt.Errorf(warn)
	} else if _, ok := val.(string); ok {
		if val == "" || val == "<nil>" {
			return nil, fmt.Errorf(warn)
		}
	}
	return val, nil
}

const (
	require string = "require"
	prefer  string = "prefer"
)

const (
	ipv4 string = "ipv4"
	ipv6 string = "ipv6"
)

func anyIpAddress() (*net.IP, error) {
	return ipAddress(prefer, ipv4, 0)
}

func ipAddress(preference string, ipVersion string, iface int) (*net.IP, error) {
	addresses, err := ipAddresses(preference, ipVersion)
	if err != nil {
		return nil, err
	}
	if len(addresses) <= iface {
		return nil, fmt.Errorf("less than %d interfaces are available with a global unicast address matching '%s %s'", iface+1, preference, ipVersion)
	}
	return addresses[iface], nil
}

func ipAddresses(preference string, ipVersion string) ([]*net.IP, error) {
	if preference != require && preference != prefer {
		return nil, fmt.Errorf("preference argument must be one of [%s, %s], but was: %s", require, prefer, preference)
	}
	if ipVersion != ipv4 && ipVersion != ipv6 {
		return nil, fmt.Errorf("IP version argument must be one of [%s, %s], but was: %s", ipv4, ipv6, ipVersion)
	}

	addresses := make([]*net.IP, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		var foundIfaceAddr *net.IP
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsGlobalUnicast() {
				if foundIfaceAddr == nil || ipAddrVersion(*foundIfaceAddr) != ipVersion {
					foundIfaceAddr = &ip
				}
				if ipAddrVersion(*foundIfaceAddr) == ipVersion {
					break
				}
			}
		}
		if foundIfaceAddr != nil && (preference == prefer || ipAddrVersion(*foundIfaceAddr) == ipVersion) {
			addresses = append(addresses, foundIfaceAddr)
		}
	}
	return addresses, nil
}

func ipAddrVersion(ip net.IP) string {
	if ip.To4() != nil {
		return ipv4
	}
	return ipv6
}

// toYAML takes an interface, marshals it to yaml, and returns a string.
func toYAML(v interface{}) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}

func fromYAML(text string) (interface{}, error) {
	var out interface{}
	err := yaml.Unmarshal([]byte(text), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// toJSON takes an interface, marshals it to json, and returns a string.
func toJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toJSONPretty(indent string, v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fromJSON(text string) (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(text), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// toTOML takes an interface, marshals it to toml, and returns a string.
func toTOML(v interface{}) (string, error) {
	b := bytes.NewBuffer(nil)
	e := toml.NewEncoder(b)
	err := e.Encode(v)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func fromTOML(text string) (interface{}, error) {
	var out interface{}
	_, err := toml.Decode(text, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// toProperties takes an interface, marshals it to properties, and returns a string.
func toProperties(v interface{}) (string, error) {
	ofStrings, err := toPropertiesStringMap(v)
	if err != nil {
		return "", err
	}
	props := properties.LoadMap(ofStrings)
	return props.String(), nil
}

func fromProperties(text string) (map[string]interface{}, error) {
	propsMap := make(map[string]interface{})
	props, err := properties.LoadString(text)
	for key, value := range props.Map() {
		propsMap[key] = value
	}
	if err != nil {
		return nil, err
	}
	return propsMap, nil
}

func toPropertiesStringMap(v interface{}) (map[string]string, error) {
	switch t := v.(type) {
	case nil:
		return map[string]string{}, nil
	case map[string]string:
		return t, nil
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Map:
			stringMap := make(map[string]string)
			iter := val.MapRange()
			for iter.Next() {
				key := iter.Key().Interface()
				value := iter.Value().Interface()
				stringKey, err := toPropertiesString(key)
				if err != nil {
					return nil, err
				}
				stringValue, err := toPropertiesString(value)
				if err != nil {
					return nil, err
				}
				stringMap[stringKey] = stringValue
			}
			return stringMap, nil
		case reflect.Array, reflect.Slice:
			stringMap := make(map[string]string)
			l := val.Len()
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				subMap, err := toPropertiesStringMap(value)
				if err != nil {
					return nil, err
				}
				for subKey, subValue := range subMap {
					stringMap[subKey] = subValue
				}
			}
			return stringMap, nil
		}
		if isStruct(t) {
			m, err := structToMap(t)
			if err != nil {
				return nil, err
			}
			return toPropertiesStringMap(m)
		}
		stringKey, err := toPropertiesString(t)
		if err != nil {
			return nil, err
		}
		return map[string]string{stringKey: ""}, nil
	}
}

func toPropertiesString(v interface{}) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case []byte:
		return string(t), nil
	case error:
		return t.Error(), nil
	case fmt.Stringer:
		return t.String(), nil
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Map:
			stringList := make([]string, 0)
			iter := val.MapRange()
			for iter.Next() {
				key := iter.Key().Interface()
				value := iter.Value().Interface()
				stringKey, err := toPropertiesString(key)
				if err != nil {
					return "", err
				}
				stringValue, err := toPropertiesString(value)
				if err != nil {
					return "", err
				}
				stringList = append(stringList, fmt.Sprintf("%s=%s", stringKey, stringValue))
			}
			return strings.Join(stringList, ","), nil
		case reflect.Array, reflect.Slice:
			l := val.Len()
			stringList := make([]string, 0)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				subString, err := toPropertiesString(value)
				if err != nil {
					return "", err
				}
				stringList = append(stringList, subString)
			}
			return strings.Join(stringList, ","), nil
		}
		if isStruct(t) {
			m, err := structToMap(t)
			if err != nil {
				return "", err
			}
			return toPropertiesString(m)
		}
		return fmt.Sprint(t), nil
	}
}

func isStruct(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr &&
		reflect.ValueOf(v).Elem().Kind() == reflect.Struct
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	if !isStruct(v) {
		return nil, fmt.Errorf("%T is not a ptr to a struct", v)
	}
	out := make(map[string]interface{})
	err := mapstructure.Decode(v, &out)
	return out, err
}
