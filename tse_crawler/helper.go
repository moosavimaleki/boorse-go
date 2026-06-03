package tse_crawler

import (
	"math"
	"strconv"
	"tsetmc/utils/logger"
)

func convertStringToUint(val string, size int) uint64 {
	if val == "" {
		return 0
	}
	res, err := strconv.ParseUint(val, 10, size)
	if err != nil {
		//error for parse 4500.00
		resFloat, _err := strconv.ParseFloat(val, size)
		res = uint64(resFloat)
		if _err != nil {
			logger.GetFlog().Err(err).Str("convertStringToUint", val).Send()
			logger.GetFlog().Err(_err).Str("convertStringToUint", val).Send()
		}
	}
	return res
}

func convertStringToInt(val string, size int) int64 {
	if val == "" {
		return 0
	}
	res, err := strconv.ParseInt(val, 10, size)
	if err != nil {
		logger.GetFlog().Err(err).Str("convertStringToInt", val).Send()
	}
	return res
}

func convertStringToFloat(val string, size int) float64 {
	if val == "" {
		return 0
	}
	res, err := strconv.ParseFloat(val, size)
	if err != nil {
		logger.GetFlog().Err(err).Str("convertStringToFloat", val).Send()
	}
	return res
}

func divide(num float64, denum float64, decimalPlaces int) float64 {
	tmp := math.Pow10(decimalPlaces)
	return math.Floor((num/denum)*tmp) / tmp
}
func divide32(num float32, denum float32, decimalPlaces int) float32 {
	tmp := float32(math.Pow10(decimalPlaces))
	return float32(math.Floor(float64((num/denum)*tmp))) / tmp
}
func divideStrings(num string, denum string, decimalPlaces int) float64 {
	tmp := math.Pow10(decimalPlaces)
	return math.Floor((convertStringToFloat(num, 64)/(convertStringToFloat(denum, 64)))*tmp) / tmp
}
