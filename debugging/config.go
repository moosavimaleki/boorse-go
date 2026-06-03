package debugging

func IsDebug() bool {
	return false
}

func GetMaxTimeForFillHistory() int {
	return 2
}

func IfDebugStock(inscode uint64) bool {
	if inscode == 4942127026063388 {
		return true
	}
	return false
}
