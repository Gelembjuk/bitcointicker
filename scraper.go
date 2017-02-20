/*
* by Roman Gelembjuk <roman@gelembjuk.com>
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Scraper interface {
	init(Url string, config ScraperOptions) error
	getData() (string, error)
}

type ScraperOptions struct {
	Opt1 string
	Opt2 string
	Opt3 string
}

type ScraperError struct {
	code    string
	message string
}

func (e *ScraperError) Error() string {
	return e.message
}

func (e *ScraperError) Code() string {
	return e.code
}

func getScraper(class string, Url string, opts ScraperOptions) (Scraper, error) {
	if class == "json" {
		js := JSONScraper{}
		err := js.init(Url, opts)

		if err != nil {
			return nil, err
		}

		return &js, nil
	}

	return nil, errors.New("Scraper class not found")
}

// JSON scraper, implements Scraper interface

type JSONScraper struct {
	Url      string
	DataPath string
}

func (this *JSONScraper) init(Url string, config ScraperOptions) error {

	if Url == "" {
		return &ScraperError{"init", "Url not provided"}
	}
	this.Url = Url

	// check Url is really http url. TODO

	this.DataPath = config.Opt1

	if this.DataPath == "" {
		fmt.Println(config)
		return &ScraperError{"init", "Data Path not provided"}
	}

	return nil
}

func (this *JSONScraper) getData() (string, error) {
	// call the url.
	// check response is  correct JSON
	// parse JSON and get data by path

	if this.Url == "" {
		return "", &ScraperError{"url", "Url not provided"}
	}

	if this.DataPath == "" {
		return "", &ScraperError{"json", "Data Path not provided"}
	}
	// wait max 10 sec
	timeout := time.Duration(10 * time.Second)

	client := http.Client{
		Timeout: timeout,
	}

	response, err := client.Get(this.Url)

	if err != nil {

		return "", &ScraperError{"http", err.Error()}
	} else {
		defer response.Body.Close()

		jsondata, err := ioutil.ReadAll(response.Body)

		if err != nil {

			return "", &ScraperError{"http", err.Error()}
		}
		m := map[string]interface{}{}

		errj := json.Unmarshal([]byte(jsondata), &m)

		if errj != nil {

			return "", &ScraperError{"json", errj.Error()}
		}

		path := strings.Split(this.DataPath, "/")

		for i, key := range path {
			if val, ok := m[key]; ok {

				if reflect.TypeOf(val).Kind() == reflect.String ||
					reflect.TypeOf(val).Kind() == reflect.Float64 ||
					reflect.TypeOf(val).Kind() == reflect.Float32 ||
					reflect.TypeOf(val).Kind() == reflect.Int {

					if i == len(path)-1 {
						var retval string

						if reflect.TypeOf(val).Kind() == reflect.Float64 {
							retval = strconv.FormatFloat(val.(float64), 'f', -1, 64)
						} else if reflect.TypeOf(val).Kind() == reflect.Float32 {
							retval = strconv.FormatFloat(val.(float64), 'f', -1, 32)
						} else if reflect.TypeOf(val).Kind() == reflect.Int {
							retval = strconv.FormatInt(val.(int64), 10)
						} else {
							retval = val.(string)
						}
						// this is our value
						return retval, nil
					}

					return "", &ScraperError{"data", "Not found by Path. No enought deep"}
				}
				if i == len(path)-1 {

					return "", &ScraperError{"data", "Not found by Path. Given path is array, not a single value"}
				}
				// go deeper
				m = val.(map[string]interface{})
			} else {
				return "", &ScraperError{"data", "Can not get data by path"}
			}
		}

		return "", &ScraperError{"data", "Not found by Path. Given path is array, not a single value"}
	}
}
