package mytime

import (
	"embed"
	"io/ioutil"
	"time"
)

//go:embed zoneinfo/Asia/Tehran
var fs embed.FS

func GetIran() *time.Location {

	rc, err := fs.Open("zoneinfo/Asia/Tehran")
	if err != nil {
		loc, _ := time.LoadLocation("Asia/Tehran")
		return loc
	}

	fbts, err := ioutil.ReadAll(rc)
	if err != nil {
		rc.Close()
		loc, _ := time.LoadLocation("Asia/Tehran")
		return loc
	}

	tz, err := time.LoadLocationFromTZData("Asia/Tehran", fbts)
	if err != nil {
		rc.Close()
		loc, _ := time.LoadLocation("Asia/Tehran")
		return loc
	}
	rc.Close()

	return tz
}
