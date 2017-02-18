// fabric.go
package main

import "time"

type RatesPull struct {
	Title     string
	StopChans []chan struct{}
	RateChans []chan RateInfo
}

type RateResult struct {
	Rate         float64
	CountSources int
	TotalSources int
	Errors       []string
}

func (this *RatesPull) GetRate() RateResult {
	// get dadta from each source
	result := RateResult{Errors: []string{}}

	result.CountSources = 0
	sum := float64(0)

	for _, rateinfochan := range this.RateChans {
		select {
		case rate, ok := <-rateinfochan:
			if ok {
				// check if not expired. It can be that a result waited for a read for some time
				// TODO
				Trace.Println("Success received rate %0.5f from %s channel", rate.Rate, rate.SourceTitle)
				sum += rate.Rate
				result.CountSources++

			} else {
				Warning.Println("Channel closed!")
			}
		default:
			continue
		}
	}

	if result.CountSources > 0 {
		result.Rate = sum / float64(result.CountSources)
	}

	result.TotalSources = len(this.RateChans)

	return result
}

func (this *RatesPull) AddSource(source RateSource) {
	// the source is posted as argument to a goroutine and is processed there.
	// we create 2 channels. One to get data from sources, other to stop a goroutine

	stopchan := make(chan struct{})
	rateinfochan := make(chan RateInfo)

	go func(source RateSource) {
		// read source, put result to channel, sleep, check if to stop , repeat

		for {

			select {
			default:
				rate, err := source.getRate()

				if err != nil {
					Trace.Printf("Source %s fails to return rate. Error: %s", source.getTitle(), err.Error())
				} else {
					rateinfochan <- rate
					Trace.Printf("Source %s returns rate %0.6f\n", source.getTitle(), rate)
				}

				time.Sleep(1 * time.Second)
			case <-stopchan:
				// stop
				Trace.Println("Stop signal received")
				return
			}
		}

	}(source)

	this.StopChans = append(this.StopChans, stopchan)
	this.RateChans = append(this.RateChans, rateinfochan)
}

func (this *RatesPull) Destroy() {
	Trace.Printf("Destroy %s pull, has %d sources\n", this.Title, len(this.RateChans))

	for _, stopchan := range this.StopChans {
		close(stopchan)
	}
}

func getRatesCheckPull(config []RateFeed, title string) (RatesPull, error) {
	rcp := RatesPull{Title: title}

	for _, bf := range config {
		bs := RateSource{}

		initerr := bs.init(bf)

		if initerr != nil {
			Warning.Printf("Can not init source. Skipping %s", initerr.Error())
			continue
		}
		rcp.AddSource(bs)
	}

	return rcp, nil
}

func getBitcoinRateCheckerPull(config []RateFeed) (RatesPull, error) {
	return getRatesCheckPull(config, "Bitcoin")
}

func getEuroRateCheckerPull(config []RateFeed) (RatesPull, error) {
	return getRatesCheckPull(config, "Euro")
}
