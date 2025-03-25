package checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
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
	var wg sync.WaitGroup
	errors := make(chan string)
	for _, service := range c.Services {
		wg.Add(1)
		go checkService(service, errors, &wg)
	}
	go func() { //https://stackoverflow.com/questions/21819622/let-golang-close-used-channel-after-all-goroutines-finished
		wg.Wait()
		close(errors)
	}()
	for err := range errors {
		buffer.WriteString(err)
	}
	result := buffer.String()
	log.Println(result)
	ok := true
	if len(result) > 0 {
		ok = false
	}
	return result, ok
}

func checkService(s Service, errors chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	for _, testCase := range s.Cases {
		log.Println("Checking", s.Title, "for", testCase.Url)
		url := s.Domain + testCase.Url
		resp, err := client.Get(url)
		if err != nil {
			log.Println("Error: for url", url, err)
			errors <- err.Error()
			continue
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != testCase.Code {
			errors <- fmt.Sprintf("%s response status: %d != %d", url, resp.StatusCode, testCase.Code)
			continue
		}
	}
}
