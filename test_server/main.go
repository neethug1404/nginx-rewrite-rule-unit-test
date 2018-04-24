package main

import (
	"context"
	"encoding/json"
	"fmt"
	htemplate "html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"crypto/tls"

	"gopkg.in/yaml.v2"
)

type TestCases []struct {
	Name     string `yaml:"name"`
	Test     string `yaml:"test"`
	Expected string `yaml:"expected"`
	Redirect string
	Result   bool
}

type Response struct {
	Host    string
	UrlPath string
	Scheme	string
}

func welcome(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Nginx Rewrite Rule Test Server") // send data to client side
	fmt.Fprintf(w, "<h1>Nginx Rewrite Rule Test Server</h1>"+
		"<div>"+
		"<form action='/run-test'>"+
		"<input type='submit' value='Run Tests' />"+
		"</form>"+
		"</div>")
}

func callServer(w http.ResponseWriter, r *http.Request) {

	filename, _ := filepath.Abs("./test-urls.yml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var testcases TestCases
	err = yaml.Unmarshal(yamlFile, &testcases)
	if err != nil {
		panic(err)
	}

	failCount := 0
	for index, each := range testcases {

		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				_, port, _ := net.SplitHostPort(addr)
				if true {
					addr = "172.20.128.20:" + port
				}
				return dialer.DialContext(ctx, network, addr)
			},
		}

		client := &http.Client{Transport: tr}

		req, err := http.NewRequest("GET", each.Test, nil)
		if err != nil {
			panic(err)
		}

		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			testcases[index].Redirect = req.URL.Scheme + "://" + req.URL.Host + req.URL.RequestURI()
			if strings.Compare(each.Expected, testcases[index].Redirect) == 0 {
				testcases[index].Result = true
			} else {
				failCount++
			}
			return http.ErrUseLastResponse
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			if resp.StatusCode == http.StatusFound { //status code 302
				fmt.Println(resp.Location())
			} else {
				panic(err)
			}
		}

		contentType := resp.Header.Get("Content-type")
		if contentType != "application/json" {
			//testcases[index].Result = false
		} else {
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				panic(err.Error())
			}

			var data = new(Response)
			err = json.Unmarshal(body, &data)
			if err != nil {
				panic(err.Error())
			}
			testcases[index].Redirect = data.Scheme + "://" + data.Host + data.UrlPath

			if strings.Compare(each.Expected, testcases[index].Redirect) == 0 {
				testcases[index].Result = true
			} else {
				failCount++
			}
		}
	}

	testCount := len(testcases)
	successRate := (testCount - failCount) * 100 / testCount

	var userAgent = r.Header.Get("User-Agent")
	data := struct {
		TestCount   int
		FailCount   int
		SuccessRate int
		TestCases   TestCases
	}{
		testCount,
		failCount,
		successRate,
		testcases,
	}

	cwd, _ := os.Getwd()

	var reportTemplate string
	if strings.Contains(userAgent, "curl") {
		//var tmpl *template.Template
		reportTemplate = "./junit-report.xml"
		tmpl := template.Must(template.ParseFiles(filepath.Join(cwd, reportTemplate)))
		tmpl.Execute(w, data)
	} else {
		reportTemplate = "./report.html"
		htmpl := htemplate.Must(htemplate.ParseFiles(filepath.Join(cwd, reportTemplate)))
		htmpl.Execute(w, data)
	}
}

func main() {
	http.HandleFunc("/", welcome)            // set router
	http.HandleFunc("/run-test", callServer) // set router
	err := http.ListenAndServe(":80", nil)   // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
