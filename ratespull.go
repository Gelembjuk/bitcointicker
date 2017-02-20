// fabric.go
package main

import "time"

// struct to keep info about running routines to get info from them
type RateRoutine struct {
	StopChan     chan struct{}
	StopBackChan chan struct{}
	RateChan     chan RateInfo
	LastRate     RateInfo
	Title        string
}

// pull of routines for one type of scrappers
type RatesPull struct {
	Title   string
	Threads []RateRoutine
}

// rate result for one pull
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

	for ind, routine := range this.Threads {
		// read latest from this buffered channel
		for {
			// we repeat reading till nothing in the buffer
			wasread := false
			select {
			case rate, ok := <-routine.RateChan:
				if ok {
					// check if not expired. It can be that a result waited for a read for some time
					// TODO
					Trace.Printf("Success received rate %0.5f from %s channel", rate.Rate, rate.SourceTitle)
					this.Threads[ind].LastRate = rate
					routine.LastRate = rate
					wasread = true

				} else {
					Warning.Println("Channel closed!")
				}
			default:

			}

			if !wasread {
				break
			}
		}

		now := time.Now()
		unixsecs := int(now.Unix())

		if routine.LastRate.Expires > unixsecs {
			// it can be just received or earlier but not yet expired
			sum += routine.LastRate.Rate
			result.CountSources++

			Trace.Printf("Summ a rate from %s", routine.LastRate.SourceTitle)
		}
	}

	if result.CountSources > 0 {
		result.Rate = sum / float64(result.CountSources)
	}

	result.TotalSources = len(this.Threads)

	return result
}

func (this *RatesPull) AddSource(source RateSource) {
	// the source is posted as argument to a goroutine and is processed there.
	// we create 2 channels. One to get data from sources, other to stop a goroutine
	routine := RateRoutine{}

	routine.StopChan = make(chan struct{})
	routine.StopBackChan = make(chan struct{})
	routine.RateChan = make(chan RateInfo, 2) // buffered!
	routine.Title = source.Title

	go func(source RateSource, r RateRoutine) {
		// read source, put result to channel, sleep, check if to stop , repeat

		for {

			rate, err := source.getRate()

			select {
			default:

			case <-r.StopChan:
				// stop
				Trace.Printf("Stop signal received for %s", source.Title)
				close(r.StopBackChan)
				return
			}

			if err != nil {
				Trace.Printf("Source %s fails to return rate. Error: %s", source.getTitle(), err.Error())
			} else {
				r.RateChan <- rate

				Trace.Printf("Source %s returns rate %0.6f\n", source.getTitle(), rate.Rate)
			}

			time.Sleep(1 * time.Second)
		}

	}(source, routine)

	this.Threads = append(this.Threads, routine)
}

func (this *RatesPull) Destroy() {
	Trace.Printf("Destroy %s pull, has %d sources\n", this.Title, len(this.Threads))

	for _, routine := range this.Threads {
		Trace.Printf("Stopping channel %s", routine.Title)
		// send stop signal to a routine
		close(routine.StopChan)

		// wait it is closed
		Trace.Printf("Waiting channel %s to stop", routine.Title)
		<-routine.StopBackChan
		// close corresponding rates channel

		close(routine.RateChan)
		Trace.Printf("Stopped channel %s", routine.Title)
	}

	this.Threads = []RateRoutine{}
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
