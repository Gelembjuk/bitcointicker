package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var config Configuration

func main() {

	var err error
	config, err = getConfig()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if config.LogFile == "stdout" {
		logToStd()
	} else if config.LogFile != "" {
		logToFile(config.LogFile)
	}

	Info.Println("The tool started")

	// create rates pulls
	bpull, initerr := getBitcoinRateCheckerPull(config.BitcoinFeeds)

	if initerr != nil {
		fmt.Println(initerr.Error())
		os.Exit(1)
	}

	epull, initeuroerr := getEuroRateCheckerPull(config.EuroFeeds)

	if initeuroerr != nil {
		fmt.Println(initeuroerr.Error())
		os.Exit(1)
	}

	stopmainchan := make(chan struct{})
	stopmainconfirmchan := make(chan struct{})
	theendchan := make(chan struct{})
	// prepare to catch a signal of
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func(bp *RatesPull, ep *RatesPull) {
		sig := <-sigs

		Info.Printf("Interupted %d . Stop workers", sig)

		// stop checking rates
		close(stopmainchan)

		// wait main thread confirms it stopped to read pulls
		<-stopmainconfirmchan

		// stop all goroutines

		ep.Destroy()

		bp.Destroy()
		// final exit from the program
		close(theendchan)

		return
	}(&bpull, &epull)

	// read sources and show state
	for {
		r1 := bpull.GetRate()

		r2 := epull.GetRate()

		res := ""

		if r1.CountSources >= config.MinBitcoin {
			res = fmt.Sprintf("BTC/USD: %0.2f", r1.Rate)
		} else {
			res = "BTC/USD: Undefined"
		}

		res += "\t"

		if r2.CountSources >= config.MinEuro {
			res += fmt.Sprintf("EUR/USD: %0.5f", r2.Rate)
		} else {
			res += "EUR/USD: Undefined"
		}

		res += "\t"

		if r1.CountSources >= config.MinBitcoin && r2.CountSources >= config.MinEuro {
			be := float64(0)

			if r2.Rate > 0 {
				be = r1.Rate / r2.Rate
			}
			res += fmt.Sprintf("BTC/EUR: %0.2f", be)
		} else {
			res += "BTC/EUR: Undefined"
		}

		res += "\t"

		res += fmt.Sprintf("Active sources: BTC/USD (%d of %d)  EUR/USD (%d of %d)", r1.CountSources, r1.TotalSources, r2.CountSources, r2.TotalSources)

		fmt.Println(res)

		stop := false

		select {
		case _, ok := <-stopmainchan:
			if !ok {
				stop = true
			}
		default:
			time.Sleep(1 * time.Second)
		}

		if stop {
			close(stopmainconfirmchan)
			break
		}
	}
	<-theendchan

	Info.Println("The tool completed")
}
