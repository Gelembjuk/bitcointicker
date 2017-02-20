/*
* by Roman Gelembjuk <roman@gelembjuk.com>
 */
package main

import (
	"encoding/json"
	//	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type RateFeed struct {
	Title          string
	Url            string
	ScrapperClass  string
	ScraperOptions ScraperOptions
	Expires        int
}

type Configuration struct {
	BitcoinFeeds []RateFeed
	EuroFeeds    []RateFeed
	MinBitcoin   int
	MinEuro      int
	LogFile      string
}

func init() {
	// set logging to dev/null
	Trace = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(ioutil.Discard,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(ioutil.Discard,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(ioutil.Discard,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

}

// Parse config file
func getConfig() (Configuration, error) {
	file, errf := os.Open("config.json")

	if errf != nil {
		return Configuration{}, errf
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)

	if err != nil {
		return Configuration{}, err
	}
	return configuration, nil
}

// set logging to stdout
func logToStd() {
	Trace.SetOutput(os.Stdout)
	Info.SetOutput(os.Stdout)
	Warning.SetOutput(os.Stdout)
	Error.SetOutput(os.Stdout)
}

// set logging to given file
func logToFile(filepath string) {

	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("error opening log file: %v", err)
		os.Exit(1)
	}
	Trace.SetOutput(f)
	Info.SetOutput(f)
	Warning.SetOutput(f)
	Error.SetOutput(f)

}
