package main

import (
	"github.com/rs/zerolog"
	"time"
	"tsetmc/debugging"
	"tsetmc/tse_crawler"
	"tsetmc/ui"
	"tsetmc/utils/logger"
)

func init() {
	zerolog.TimeFieldFormat = "2000-01-01 00:00:00"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger.GetFlog().Info().Msg("This is an entry from my log")
	logger.GetFlog().Info().
		Str("Start Program", "start").
		Time("Now is", time.Now()).
		Send()
}

func _main() {
	//todo: dont save where market closed
	if debugging.GetNowUnix() > 1638259381 {
		return
	}

	stockChannel := make(chan map[uint64]tse_crawler.Stock, 5)
	go tse_crawler.StartCrawl(stockChannel)
	go func() {
		time.Sleep(4 * time.Second)
		tse_crawler.CheckFilters(stockChannel)
	}()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			tse_crawler.StartCodal()
		}
	}()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			tse_crawler.StartCodalSoorat()
		}
	}()

	lorcaUI, listener := ui.UiInit()

	defer lorcaUI.Close()
	defer listener.Close()
	ui.WaitUntilUiClosed()

}
