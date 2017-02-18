// bitcoinsourceinterface.go
package main

import (
	"regexp"
	"strconv"
	"time"
)

type RateSource struct {
	Title         string
	Scraper       Scraper
	Expires       int
	LastScrapTime int
	LastResult    float64
}

type RateInfo struct {
	Rate        float64
	Expires     int
	SourceTitle string
}

func (this *RateSource) init(config RateFeed) error {
	this.Expires = config.Expires

	if this.Expires < 1 {
		this.Expires = 30 // 30 seconds by default
	}

	this.LastScrapTime = 0

	this.Title = config.Title

	var err error

	this.Scraper, err = getScraper(config.ScrapperClass, config.Url, config.ScraperOptions)

	if err != nil {
		return err
	}

	return nil
}

func (this *RateSource) getTitle() string {
	return this.Title
}

func (this *RateSource) getRate() (RateInfo, error) {
	now := time.Now()
	unixsecs := int(now.Unix())

	Trace.Println("get Rate")

	if this.LastScrapTime > 0 &&
		this.LastScrapTime+this.Expires < unixsecs {
		// not expired yet
		Trace.Println("Rate was remembered and is not yet expired")

		result := RateInfo{Rate: this.LastResult, SourceTitle: this.Title}

		result.Expires = this.LastScrapTime + this.Expires
		return result, nil
	}

	scraperdata, err := this.Scraper.getData()

	if err != nil {
		this.LastScrapTime = 0
		Warning.Println(err)
		return RateInfo{}, err
	}

	// convert string to number
	var rate float64

	rate, err = this.getFloat(scraperdata)

	if err != nil {
		this.LastScrapTime = 0
		Warning.Println(err.Error())
		return RateInfo{}, err
	}

	this.LastResult = rate
	this.LastScrapTime = unixsecs

	Trace.Printf("Return rate %0.7f", rate)

	result := RateInfo{Rate: rate}

	result.Expires = unixsecs + this.Expires

	result.SourceTitle = this.Title

	return result, nil
}

func (this RateSource) getFloat(str string) (float64, error) {

	re := regexp.MustCompile("[^0-9.]")

	str = re.ReplaceAllString(str, "")

	f, err := strconv.ParseFloat(str, 64)

	if err != nil {
		return 0, err
	}

	return f, nil
}
