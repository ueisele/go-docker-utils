package template

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"github.com/Masterminds/sprig"
)

// Uses Go template and environment variables to create configuration files with environment variables as context.
//
// Args:
//   template_file: template file path.
//   output_file: output file path.
//
// Returns:
//   Returns error if an Exception occurs.
//
// See:
//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func FillAndWriteTemplateFileWithEnvContext(template_file string, output_file string, options ...string) error {
	return FillAndWriteTemplateFile(template_file, output_file, ReadEnv(), options...)
}

// Uses Go template and environment variables to create configuration files.
//
// Args:
//   template_file: template file path.
//   output_file: output file path.
//   context: the data for the filling in the template.
//
// Returns:
//   Returns error if an Exception occurs.
//
// See:
//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
func FillAndWriteTemplateFile(template_file string, output_file string, context interface{}, options ...string) error {
	var tplReader *bufio.Reader
	tplFile, err := os.Open(template_file)
	if err != nil {
		return fmt.Errorf("could not open template file %s: %v", template_file, err)
	}
	defer tplFile.Close()
	tplReader = bufio.NewReader(tplFile)

	var outWriter *bufio.Writer
	outFile, err := os.Create(output_file);
	if err != nil {
		return fmt.Errorf("could not create output file %s: %v", output_file, err)
	}
	defer outFile.Close()
	outWriter = bufio.NewWriter(outFile)

	return FillAndWriteTemplateIo(tplReader, outWriter, context, options...)
}

func FillAndWriteTemplateIoWithEnvContext(tplReader io.Reader, outWriter io.Writer, options ...string) error {
	return FillAndWriteTemplateIo(tplReader, outWriter, ReadEnv(), options...)
}

func FillAndWriteTemplateIo(tplReader io.Reader, outWriter io.Writer, context interface{}, options ...string) error {
	tplBytes, err := ioutil.ReadAll(tplReader)
	if err != nil {
		return fmt.Errorf("could not read template: %v", err)
	}

	tpl, err := template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(DubTxtFuncMap()).
		Option(options...).
		Parse(string(tplBytes))
	if err != nil {
		return fmt.Errorf("could not parse template: %v", err)
	}

	err = tpl.Execute(outWriter, context)
	if err != nil {
		return fmt.Errorf("could not render template: %v", err)
	}

	return nil
}

// returns map of environment variables
func ReadEnv() (env map[string]string) {
	env = make(map[string]string)
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	return
}
