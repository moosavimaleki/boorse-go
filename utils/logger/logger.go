package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"time"
)

var Flog zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = "2000-01-01 00:00:00"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if err := os.MkdirAll("./log", 0755); err != nil {
		log.Error().Err(err).Msg("there was an error creating log directory")
		Flog = zerolog.New(os.Stderr).With().Logger()
		return
	}
	tempFile, err := ioutil.TempFile("./log", "deleteme")
	if err != nil {
		// Can we log an error before we have our logger? :)
		log.Error().Err(err).Msg("there was an error creating a temporary file four our log")
		Flog = zerolog.New(os.Stderr).With().Logger()
		return
	}
	Flog = zerolog.New(tempFile).With().Logger()
	log.Info().Msg("This is an entry from my log")
	log.Info().
		Str("Start Program", "start").
		Time("Now is", time.Now()).
		Send()
}
func GetFlog() *zerolog.Logger {
	return &Flog
}
