package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func FileGlobsToFileNames(globs ...string) ([]string, error) {
	filenames := make([]string, 0)
	for _, glob := range globs {
		filenamesOfGlob, err := filepath.Glob(glob)
		if err!= nil {
			return nil, err
		}
		for _, filename := range filenamesOfGlob {
			if !contains(filenames, filename) {
				filenames = append(filenames, filename)
			}
		}
	}
	return filenames, nil
}

func FileInputSource(filenames ...string) Source {
	return func () <-chan *Data {
		return FileInputProvider(filenames...)
	}
}

func FileInputProvider(filenames ...string) <-chan *Data {
	out := make(chan *Data)
	go func() {
		defer close(out)
		for _, filename := range filenames {
			b, err := os.ReadFile(filename)
			if err == nil {
				out <- &Data{Name: filename, Content: string(b)}
			} else {
				out <- &Data{Name: filename, Content: "", Error: err}
			}
		}
	}()
	return out	
}

func ReaderSource(name string, reader io.Reader) Source {
	return func () <-chan *Data {
		return ReaderProvider(name, reader)
	}
}

func ReaderProvider(name string, reader io.Reader) <-chan *Data {
	out := make(chan *Data)
	go func() {
		defer close(out)
		b, err := ioutil.ReadAll(reader)
		if err == nil {
			out <- &Data{Name: name, Content: string(b)}
		} else {
			out <- &Data{Name: name, Content: "", Error: err}
		}
	}()
	return out	
}

func FileOutputSink(filename string) (Transform, error) {
	info, err := os.Stat(filename)
	if err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a file", filename) 	
	} else if err == nil && info.Mode().IsRegular() {
		if err := unix.Access(filename, unix.W_OK); err != nil {
			return nil, fmt.Errorf("%s could not be accessed, %v", filename, err)
		}
	} else if err := unix.Access(filepath.Dir(filename), unix.R_OK | unix.W_OK | unix.X_OK); err != nil {
		return nil, fmt.Errorf("%s could not be accessed, %v", filepath.Dir(filename), err)
	}
	return func (input <-chan *Data) <-chan *Data {
		out := make(chan *Data)
		go func() {
			defer close(out)
			file, fileerr := os.Create(filename)
			if err == nil {
				defer file.Close()
			}
			for data := range input {
				if fileerr == nil && data.Error == nil {
					_, err := io.WriteString(file, data.Content)
					out <- &Data{Name: data.Name, Content: data.Content, Error: err}
				} else if fileerr != nil {
					out <- &Data{Name: data.Name, Content: data.Content, Error: fileerr}
				} else {
					out <- data
				}
			}
		}()
		return out
	}, nil
}

func DirOutputSink(dirpath string, removeexts ...string) (Transform, error) {
	info, err := os.Stat(dirpath)
	if err != nil {
		err := os.MkdirAll(dirpath, 0777)
		if err != nil {
			return nil, err
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dirpath)
	}
	if err := unix.Access(dirpath, unix.R_OK | unix.W_OK | unix.X_OK); err != nil {
		return nil, fmt.Errorf("%s could not be accessed, %v", dirpath, err)
	}
 	return transformerInOrder(func (data *Data) *Data {
		basename := filepath.Base(data.Name)
		for _, removeext := range removeexts {
			if filepath.Ext(basename) == removeext {
				basename = strings.TrimSuffix(basename, removeext)
				break
			}
		} 
		targetfilepath := filepath.Join(dirpath, basename)
		targetfile, err := os.Create(targetfilepath)
		if err == nil {
			defer targetfile.Close()
			_, err = io.WriteString(targetfile, data.Content)
		}
		return &Data{Name: data.Name, Content: data.Content, Error: err} 
	}), nil
}

func WriterSink(writer io.Writer) Transform {
	return transformerInOrder(func (data *Data) *Data {
		_, err := io.WriteString(writer, data.Content)
		return &Data{Name: data.Name, Content: data.Content, Error: err}
	})
}
