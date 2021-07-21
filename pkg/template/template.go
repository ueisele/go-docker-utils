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

type DubTemplate interface {

	// Uses Go template and environment variables to create configuration files.
	//
	// Args:
	//   templateFileName: template file path.
	//   outputFileName: output file path.
	//
	// Returns:
	//   Returns error if an Exception occurs.
	//
	// See:
	//   Original dub: https://github.com/confluentinc/confluent-docker-utils/blob/master/confluent/docker_utils/dub.py
	TemplateFileToFile(templateFileName string, outputFileName string) error
	MustTemplateFileToFile(templateFileName string, outputFileName string)

	TemplateOsFileToOsFile(templateFile *os.File, outputFile *os.File) error
	MustTemplateOsFileToOsFile(templateFile *os.File, outputFile *os.File)

	TemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) error
	MustTemplateInputToWriter(tplReader io.Reader, outWriter io.Writer)

	TemplateText(tplString string) (string, error)
	MustTemplateText(tplString string) string

	TemplateTextToWriter(tplString string, outWriter io.Writer) error
	MustTemplateTextToWriter(tplString string, outWriter io.Writer)
}

// Uses Go template to create configuration files with environment variables as context.
//
// Args:
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
func NewDubTemplateWithDefaults(options ...string) DubTemplate {
	return NewDubTemplate(readEnv(), options...)
}

// Uses Go template to create configuration files.
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
func NewDubTemplate(context interface{}, options ...string) DubTemplate {
	return &dubtemplate{
		context: context,
		options: options,
	}
}

type dubtemplate struct {
	context interface{}
	options []string
}

func (t *dubtemplate) MustTemplateFileToFile(templateFileName string, outputFileName string) {
	err := t.TemplateFileToFile(templateFileName, outputFileName)
	if err != nil {
		panic(err)
	}
}

func (t *dubtemplate) TemplateFileToFile(templateFileName string, outputFileName string) error {
	tplFile, err := os.Open(templateFileName)
	if err != nil {
		return fmt.Errorf("could not open template file %s: %v", templateFileName, err)
	}
	defer tplFile.Close()

	outFile, err := os.Create(outputFileName);
	if err != nil {
		return fmt.Errorf("could not create output file %s: %v", outputFileName, err)
	}
	defer outFile.Close()

	return t.TemplateOsFileToOsFile(tplFile, outFile)
}

func (t *dubtemplate) MustTemplateOsFileToOsFile(templateFile *os.File, outputFile *os.File) {
	err := t.TemplateOsFileToOsFile(templateFile, outputFile)
	if err != nil {
		panic(err)
	}
}

func (t *dubtemplate) TemplateOsFileToOsFile(templateFile *os.File, outputFile *os.File) error {
	tplReader := bufio.NewReader(templateFile)
	outWriter := bufio.NewWriter(outputFile)
	defer outWriter.Flush()
	err := t.TemplateInputToWriter(tplReader, outWriter)
	if err != nil {
		return err
	}
	err = outWriter.Flush()
	if err != nil {
		return fmt.Errorf("could not write to output: %v", err)
	}
	return nil
}

func (t *dubtemplate) MustTemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) {
	err := t.TemplateInputToWriter(tplReader, outWriter)
	if err != nil {
		panic(err)
	}
}

func (t *dubtemplate) TemplateInputToWriter(tplReader io.Reader, outWriter io.Writer) error {
	tplBytes, err := ioutil.ReadAll(tplReader)
	if err != nil {
		return fmt.Errorf("could not read template: %v", err)
	}
	return t.TemplateTextToWriter(string(tplBytes), outWriter)
}

func (t *dubtemplate) MustTemplateText(tplString string) string {
	output, err := t.TemplateText(tplString)
	if err != nil {
		panic(err)
	}
	return output
}

func (t *dubtemplate) TemplateText(tplString string) (string, error) {
	outBuf := new(bytes.Buffer)
	err := t.TemplateTextToWriter(tplString, outBuf)
	if err != nil {
		return "", err
	}
	return outBuf.String(), nil
}

func (t *dubtemplate) MustTemplateTextToWriter(tplString string, outWriter io.Writer) {
	err := t.TemplateTextToWriter(tplString, outWriter)
	if err != nil {
		panic(err)
	}
}

func (t *dubtemplate) TemplateTextToWriter(tplString string, outWriter io.Writer) error {
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

func (t *dubtemplate) createTemplate(tplString string) (*template.Template, error) {
	return template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(DubTxtFuncMap()).
		Funcs(t.tplFuncMap()).
		Option(t.options...).
		Parse(tplString)
}

func (t *dubtemplate) tplFuncMap() map[string]interface{} {
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