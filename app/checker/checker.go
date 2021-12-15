package checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

type Case struct {
	Url  string
	Code int
}
type Service struct {
	Title  string
	Domain string
	Cases  []Case
}
type Checker struct {
	path     string
	Services []Service
}

func New(path string, messages chan<- string) (*Checker, error) {
	c := &Checker{
		path: path,
	}
	err := c.uploadConfigs()
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (c *Checker) uploadConfigs() error {
	files, err := os.ReadDir(c.path)
	if err != nil {
		return err
	}
	for _, file := range files {
		content, _ := os.ReadFile(path.Join(c.path, file.Name()))
		service := &Service{}
		err := json.Unmarshal([]byte(content), &service)
		if err != nil {
			log.Println(err)
		}
		c.Services = append(c.Services, *service)
	}
	return nil
}

func (c *Checker) Check() (string, bool) {
	var buffer bytes.Buffer
	errors := make(chan string)
	for _, service := range c.Services {
		go checkService(service, errors)
	}
	for err := range errors {
		buffer.WriteString(err)
	}
	result := buffer.String()
	ok := true
	if len(result) > 0 {
		ok = false
	}
	return result, ok
}

func checkService(s Service, errors chan<- string) {
	for _, testCase := range s.Cases {
		url := s.Domain + testCase.Url
		res, err := http.Get(url)
		if err != nil {
			log.Println(err)
			errors <- err.Error()
			continue
		}
		if res.StatusCode != testCase.Code {
			errors <- fmt.Sprintf("%s response with status: %d", url, testCase.Code)
			continue
		}
	}
	close(errors)
}
