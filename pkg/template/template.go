package template

import (
	"bufio"
	"bytes"
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

	godubtpl := NewGoDubTpl(context, options...)
	outString, err := godubtpl.TemplateText(string(tplBytes))
	if err != nil {
		return err
	}

	io.WriteString(outWriter, outString)

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

type GoDubTemplate interface {

	// Uses Go template and environment variables to create configuration files.
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
	TemplateFileToFile(template_file string, output_file string) error
	MustTemplateFileToFile(template_file string, output_file string)

	TemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) error
	MustTemplateInputToWriter(tplReader io.Reader, outWriter io.Writer)

	TemplateText(tplString string) (string, error)
	MustTemplateText(tplString string) string

	TemplateTextToWriter(tplString string, outWriter io.Writer) error
	MustTemplateTextToWriter(tplString string, outWriter io.Writer)
}

// Uses Go template and environment variables to create configuration files with environment variables as context.
func NewGoDubTplWithDefaults(options ...string) GoDubTemplate {
	return NewGoDubTpl(readEnv(), options...)
}

// Uses Go template and environment variables to create configuration files.
//
// Args:
//   context: The data for the filling in the template.
//	 options: Option sets options for the template
//     Known options:
//
//     missingkey: Control the behavior during execution if a map is
//     indexed with a key that is not present in the map.
//      "missingkey=default" or "missingkey=invalid"
//	 	  The default behavior: Do nothing and continue execution.
//		  If printed, the result of the index operation is the string
//		  "<no value>".
//	    "missingkey=zero"
//		  The operation returns the zero value for the map type's element.
//	    "missingkey=error"
//		  Execution stops immediately with an error.
//
// Returns:
//   Returns a new templating instance.
func NewGoDubTpl(context interface{}, options ...string) GoDubTemplate {
	return &godubtpl{
		context: context,
		options: options,
	}
}

type godubtpl struct {
	context interface{}
	options []string
}

func (t *godubtpl) MustTemplateFileToFile(template_file string, output_file string) {
	err := t.TemplateFileToFile(template_file, output_file)
	if err != nil {
		panic(err)
	}
}

func (t *godubtpl) TemplateFileToFile(template_file string, output_file string) error {
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

	return t.TemplateInputToWriter(tplReader, outWriter)
}

func (t *godubtpl) MustTemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) {
	err := t.TemplateInputToWriter(tplReader, outWriter)
	if err != nil {
		panic(err)
	}
}

func (t *godubtpl) TemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) error {
	tplBytes, err := ioutil.ReadAll(tplReader)
	if err != nil {
		return fmt.Errorf("could not read template: %v", err)
	}
	return t.TemplateTextToWriter(string(tplBytes), outWriter)
}

func (t *godubtpl) MustTemplateText(tplString string) string {
	output, err := t.TemplateText(tplString)
	if err != nil {
		panic(err)
	}
	return output
}

func (t *godubtpl) TemplateText(tplString string) (string, error) {
	outBuf := new(bytes.Buffer)
	err := t.TemplateTextToWriter(tplString, outBuf)
	if err != nil {
		return "", err
	}
	return outBuf.String(), nil
}

func (t *godubtpl) MustTemplateTextToWriter(tplString string, outWriter io.Writer) {
	err := t.TemplateTextToWriter(tplString, outWriter)
	if err != nil {
		panic(err)
	}
}

func (t *godubtpl) TemplateTextToWriter(tplString string, outWriter io.Writer) error {
	tpl, err := t.createTemplate(tplString)
	if err != nil {
		return fmt.Errorf("could not parse template: %v", err)
	}

	tpl.Execute(outWriter, t.context)
	if err != nil {
		return fmt.Errorf("could not render template: %v", err)
	}

	return nil
}

func (t *godubtpl) createTemplate(tplString string) (*template.Template, error) {
	return template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(DubTxtFuncMap()).
		Funcs(t.tplFuncMap()).
		Option(t.options...).
		Parse(tplString)
}

func (t *godubtpl) tplFuncMap() map[string]interface{} {
	gfm := make(map[string]interface{}, len(genericMap))
	gfm["tpl"] = t.MustTemplateText
	return gfm
}

// returns map of environment variables
func readEnv() (env map[string]string) {
	env = make(map[string]string)
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	return
}