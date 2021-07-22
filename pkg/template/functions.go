package template

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"inet.af/netaddr"
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
	"hasEnv":				hasEnv,
	"envAsMap":		  		envAsMap,
	"envToProp":			envToProp,

	// Map functions
	"filterMapKey":			filterMapKey,
	"replaceMapKeyPrefix":	replaceMapKeyPrefix,
	"mapKeyToProp":			mapKeyToProp,
	"parseKvCsvList":		parseKvCsvList,

	// List functions
	"filterHasPrefix":		filterHasPrefix,

	// Verify functions
	"required":				required,

	// Network functions
	"ipAddresses":			ipAddresses,
	"ipAddress":			ipAddress,
	"anyIpAddress":			anyIpAddress,
}

func hasEnv(key string) bool {
	_, has := os.LookupEnv(key)
	return has
}

// custom function that returns key, value for all envAsMap variable keys matching prefix
// (see original envtpl: https://pypi.org/project/envtpl/)
func envAsMap(prefix string) map[string]string {
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
func envToProp(env_prefix string, prop_prefix string, exclude ...interface{}) map[string]string {
	return mapKeyToProp(replaceMapKeyPrefix(env_prefix, prop_prefix, filterMapKey(exclude, envAsMap(env_prefix))))
}

func filterMapKey(exclude interface{}, sourceMap map[string]string) map[string]string {
	resultMap := make(map[string]string)
	excludeStrings := toFlatListOfStrings(exclude)
	for key, value := range sourceMap {
		if !contains(excludeStrings, key) {
			resultMap[key] = value
		}
	}
	return resultMap
}

func replaceMapKeyPrefix(prefix string, replacement string, sourceMap map[string]string) map[string]string {
	resultMap := make(map[string]string)
	for key, value := range sourceMap {
		resultMap[replacement + strings.TrimPrefix(key, prefix)] = value
	}
	return resultMap
}

func mapKeyToProp(sourceMap map[string]string) map[string]string {
	props := make(map[string]string)
	to_dot_pattern := regexp.MustCompile("[^_](_)[^_]")
	for key, value := range sourceMap {
		raw_name := strings.ToLower(key)
		var prop_dot string = raw_name
		for matches := to_dot_pattern.FindAllString(prop_dot, -1); 
			len(matches) > 0; 
			matches = to_dot_pattern.FindAllString(prop_dot, -1) {
			for _, frac := range matches {
				prop_dot = strings.Replace(prop_dot, frac, strings.ReplaceAll(frac, "_", "."), 1)
			}
		}
		prop_dash := strings.Join(strings.Split(prop_dot, "___"), "-")
		prop_underscore := strings.Join(strings.Split(prop_dash, "__"), "_")
		props[prop_underscore] = value
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

func toFlatListOfStrings(args ...interface{}) []string {
	stringList := make([]string, 00)
	for _, arg := range args {
		switch reflect.TypeOf(arg).Kind() {
		case reflect.Slice, reflect.Array:
			value := reflect.ValueOf(arg)
			for i := 0; i < value.Len(); i++ {
				stringList = append(stringList, toFlatListOfStrings(value.Index(i).Interface())...)
			}
		default:
			switch t := arg.(type) {
			case string:
				stringList = append(stringList, t)
			case fmt.Stringer:
				stringList = append(stringList, t.String())
			default:
				stringList = append(stringList, fmt.Sprintf("%v", t))
		  }
		}
	}
	return stringList
}

// Parses a list of key/value pairs separated by commas.
//
// For example for "foo.bar=DEBUG,baz.bam=TRACE"
//   the this function will return {"foo.bar: "DEBUG", "baz.bam": "TRACE"}
//
// Args:
//   kvList: String containing the comma separated list of key/value pairs.
//
// Returns:
//   Map of key/value pairs.
//
// See:
//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func parseKvCsvList(kvList string) map[string]interface{} {
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
		case nil:
			return nil, nil
		default:
			return nil, fmt.Errorf("should be type of slice, array, string or fmt.Stringer but %s", tp)
		}
	}
}

func required(warn string, val interface{}) interface{} {
	if val == nil || reflect.ValueOf(val).IsNil() {
		panic(warn)
	} else if _, ok := val.(string); ok {
		if val == "" || val == "<nil>" {
			panic(warn)
		}
	}
	return val
}

const (
	require string = "require"
	prefer  string = "prefer"
)

const (
	ipv4 string = "ipv4"
	ipv6 string = "ipv6"
)

func anyIpAddress() (*netaddr.IP, error) {
	return ipAddress(prefer, ipv4, 0)
}

func ipAddress(preference string, ipVersion string, iface int) (*netaddr.IP, error) {
	addresses, err := ipAddresses(preference, ipVersion)
	if err != nil {
		return nil, err
	}
	if (len(addresses) <= iface) {
		return nil, fmt.Errorf("less than %d interfaces are available with a global unicast address matching '%s %s'", iface + 1, preference, ipVersion)
	}
	return addresses[iface], nil
}

func ipAddresses(preference string, ipVersion string) ([]*netaddr.IP, error) {
	if preference != require && preference != prefer {
		return nil, fmt.Errorf("preference argument must be one of [%s, %s], but was: %s", require, prefer, preference)
	}
	if ipVersion != ipv4 && ipVersion != ipv6 {
		return nil, fmt.Errorf("IP version argument must be one of [%s, %s], but was: %s", ipv4, ipv6, ipVersion)
	}

	addresses := make([]*netaddr.IP, 0)

	ifaces, err := net.Interfaces()
	if err != nil { return nil, err }

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil { return nil, err }
		
		var foundIfaceAddr *netaddr.IP
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
					ip = v.IP
			case *net.IPAddr:
					ip = v.IP
			}
			if ip.IsGlobalUnicast() {
				nIp, _ := netaddr.FromStdIP(ip)
				if foundIfaceAddr == nil || ipAddrVersion(*foundIfaceAddr) != ipVersion {
					foundIfaceAddr = &nIp
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

func ipAddrVersion(ip netaddr.IP) string {
	if ip.Is4() {
		return ipv4
	}
	return ipv6
}