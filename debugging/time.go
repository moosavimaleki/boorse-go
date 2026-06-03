package debugging

import (
	"time"
	"tsetmc/utils/mytime"
)

var todayStartUnix int64

func init() {
	loc := mytime.GetIran()
	now := time.Now().In(loc)
	todayStartUnix = time.Date(
		now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, loc).Unix()
}

func GetMinStep() int {
	return int((time.Now().Unix() - todayStartUnix) / 60)
}

func GetTodayStartUnix() int64 {
	return todayStartUnix
}

func GetNowUnix() int64 {
	return time.Now().Unix()
}
