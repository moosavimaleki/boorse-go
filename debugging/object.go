package debugging

import (
	"sync"
	"time"
)

var singletonDebugging *debugging

func init() {
	createMarketWatch()
}

type debugging struct {
	startStep     int
	startUnix     int64
	startStepLock sync.RWMutex
	startUnixLock sync.RWMutex
}

func (d *debugging) getStartStep() int {
	d.startStepLock.RLock()
	defer d.startStepLock.RUnlock()
	return d.startStep
}

func (d *debugging) getStartUnix() int64 {
	d.startUnixLock.RLock()
	defer d.startUnixLock.RUnlock()
	return d.startUnix
}

func createMarketWatch() {

	singletonDebugging = &debugging{
		startStep:     GetMinStep(),
		startStepLock: sync.RWMutex{},
		startUnix:     time.Now().Unix(),
		startUnixLock: sync.RWMutex{},
	}
}

func GetDebugging() *debugging {
	return singletonDebugging
}
