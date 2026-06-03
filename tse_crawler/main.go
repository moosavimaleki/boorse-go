package tse_crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime/debug"
	"sort"
	"sync"
	"time"
	"tsetmc/debugging"
	"tsetmc/utils/logger"
	"tsetmc/utils/mymath"
	"tsetmc/utils/mytime"
	"tsetmc/utils/settings"
)

var UpdateSpeed int = 2
var CrawlerOn bool = true
var todayStartUnix int64 = 0
var currentStep int = 0
var prvStep int = 0
var mw *MarketWatch
var lockHistoryMap = sync.RWMutex{}

type tmpGroupChart struct {
	Name  string
	Power float32
	Range float32
}

func init() {
	loc := mytime.GetIran()
	now := time.Now().In(loc)
	todayStartUnix = time.Date(
		now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, loc).Unix()
}

func _getMinStep() int {
	return int((time.Now().Unix() - todayStartUnix) / 60)
}

func _getSecStepInMin() int {
	return time.Now().Second()
}

func saveStep() {
	cfg := settings.LoadConfiguration()
	var sumOfBuy uint64 = 0
	var sumOfSell uint64 = 0

	var groupsNowPrice = map[int][]float32{}
	var groupsOldPrice = map[int][]float32{}

	var groupsChild = map[int][]float64{}
	//var groupsSaraneB = map[int][]float64{}
	//var groupsSaraneS = map[int][]float64{}

	var sumOfSellVal map[int]float64 = map[int]float64{}
	var sumOfSellCount map[int]float64 = map[int]float64{}
	var sumOfBuyVal map[int]float64 = map[int]float64{}
	var sumOfBuyCount map[int]float64 = map[int]float64{}

	lockHistoryMap.RLock()
	previousAllRowObj := mw.AllRowsArchiveObj[debugging.GetMinStep()-1]
	for k, stock := range mw.AllRowsObj {
		stockArchive := mw.AllRowsArchiveObj[debugging.GetMinStep()-30][k]
		groupsOldPrice[stock.cs] = append(groupsOldPrice[stock.cs], stockArchive.PLP)
	}
	lockHistoryMap.RUnlock()

	for k, stockInStep := range mw.AllRowsObj {
		stock := StockArchive{
			inscode:     k,
			PC:          stockInStep.PC,
			PCC:         stockInStep.PCC,
			PCP:         stockInStep.PCP,
			PL:          stockInStep.PL,
			PLC:         stockInStep.PLC,
			PLP:         stockInStep.PLP,
			TNO:         stockInStep.TNO,
			TVOL:        stockInStep.TVOL,
			TVAL:        stockInStep.TVAL,
			PMIN:        stockInStep.PMIN,
			PMAX:        stockInStep.PMAX,
			VisitCount:  stockInStep.VisitCount,
			ZO1:         stockInStep.ZO1,
			ZD1:         stockInStep.ZD1,
			PD1:         stockInStep.PD1,
			PO1:         stockInStep.PO1,
			QD1:         stockInStep.QD1,
			QO1:         stockInStep.QO1,
			ZO2:         stockInStep.ZO2,
			ZD2:         stockInStep.ZD2,
			PD2:         stockInStep.PD2,
			PO2:         stockInStep.PO2,
			QD2:         stockInStep.QD2,
			QO2:         stockInStep.QO2,
			ZO3:         stockInStep.ZO3,
			ZD3:         stockInStep.ZD3,
			PD3:         stockInStep.PD3,
			PO3:         stockInStep.PO3,
			QD3:         stockInStep.QD3,
			QO3:         stockInStep.QO3,
			ZO4:         stockInStep.ZO4,
			ZD4:         stockInStep.ZD4,
			PD4:         stockInStep.PD4,
			PO4:         stockInStep.PO4,
			QD4:         stockInStep.QD4,
			QO4:         stockInStep.QO4,
			ZO5:         stockInStep.ZO5,
			ZD5:         stockInStep.ZD5,
			PD5:         stockInStep.PD5,
			PO5:         stockInStep.PO5,
			QD5:         stockInStep.QD5,
			QO5:         stockInStep.QO5,
			BuyCountI:   stockInStep.BuyCountI,
			BuyCountN:   stockInStep.BuyCountN,
			BuyIVolume:  stockInStep.BuyIVolume,
			BuyNVolume:  stockInStep.BuyNVolume,
			SellCountI:  stockInStep.SellCountI,
			SellCountN:  stockInStep.SellCountN,
			SellIVolume: stockInStep.SellIVolume,
			SellNVolume: stockInStep.SellNVolume,
			Power:       stockInStep.Power,
			SaraneB:     stockInStep.SaraneB,
			SaraneS:     stockInStep.SaraneS,
		}
		if k == 43531447361624437 {
			fmt.Println("currentStep", debugging.GetMinStep())
			fmt.Println("archive", stock)
		}
		writeToHistory(debugging.GetMinStep(), k, stock)
	}

	//calc chart
	for k, stockInStep := range mw.AllRowsObj {
		if stockInStep.cs == 68 {
			continue
		}
		if stockInStep.IsBlock {
			continue
		}

		if stockInStep.SaraneB > 0 && stockInStep.SaraneS > 0 {
			groupsChild[stockInStep.cs] = append(groupsChild[stockInStep.cs], stockInStep.SaraneB)
			//groupsSaraneB[stockInStep.cs] = append(groupsSaraneB[stockInStep.cs], stockInStep.SaraneB)
			//groupsSaraneS[stockInStep.cs] = append(groupsSaraneS[stockInStep.cs], stockInStep.SaraneS)
			sumOfBuyVal[stockInStep.cs] += float64(stockInStep.BuyIVolume) * float64(stockInStep.PL)
			sumOfBuyCount[stockInStep.cs] += float64(stockInStep.BuyCountI)
			sumOfSellVal[stockInStep.cs] += float64(stockInStep.SellIVolume) * float64(stockInStep.PL)
			sumOfSellCount[stockInStep.cs] += float64(stockInStep.SellCountI)
		}

		groupsNowPrice[stockInStep.cs] = append(groupsNowPrice[stockInStep.cs], stockInStep.PLP)
		sumOfBuy += (stockInStep.QD1 * uint64(stockInStep.PD1)) + (stockInStep.QD2 * uint64(stockInStep.PD2)) + (stockInStep.QD3 * uint64(stockInStep.PD3)) + (stockInStep.QD4 * uint64(stockInStep.PD4)) + (stockInStep.QD5 * uint64(stockInStep.PD5))
		sumOfSell += (stockInStep.QO1 * uint64(stockInStep.PO1)) + (stockInStep.QO2 * uint64(stockInStep.PO2)) + (stockInStep.QO3 * uint64(stockInStep.PO3)) + (stockInStep.QO4 * uint64(stockInStep.PO4)) + (stockInStep.QO5 * uint64(stockInStep.PO5))

		stockInPreviousStep := previousAllRowObj[k]
		if stockInStep.BuyCountI > stockInPreviousStep.BuyCountI {
			analogSarane := float64(stockInStep.BuyIVolume) - float64(stockInPreviousStep.BuyIVolume)*float64(stockInStep.PL)/float64(stockInStep.BuyCountI) - float64(stockInPreviousStep.BuyCountI)
			if analogSarane > 0 && uint64(analogSarane) > cfg.HotMoneyValue {
				var sig = NewSignal(stockInStep, "hotMoney", 0)
				sig.SignalNameFa = "پول داغ"
				sig.appendInfo("value", fmt.Sprintf("%.2f", analogSarane/10000000000))
				sig.appendInfo("type", "buy")
				SendSignal(sig)
			}
		}
		if stockInStep.SellCountI > stockInPreviousStep.SellCountI {
			analogSarane := float64(stockInStep.SellIVolume) - float64(stockInPreviousStep.SellIVolume)*float64(stockInStep.PL)/float64(stockInStep.SellCountI) - float64(stockInPreviousStep.SellCountI)
			if analogSarane > 0 && uint64(analogSarane) > cfg.HotMoneyValue {
				var sig = NewSignal(stockInStep, "hotMoney", 0)
				sig.SignalNameFa = "پول داغ"
				sig.appendInfo("value", fmt.Sprintf("%.2f", analogSarane/10000000000))
				sig.appendInfo("type", "sell")
				SendSignal(sig)
			}
		}
	}

	if mw.getLoadedGroups() == false {
		return
	}

	//calc range
	_groupsChart := make([]tmpGroupChart, 0, len(groupsNowPrice))
	for cs, newPlp := range groupsNowPrice {

		oldMean := float32(0)
		if old, ok := groupsOldPrice[cs]; ok {
			oldMean = mymath.MeanF32(old)
		}
		newMean := mymath.MeanF32(newPlp)
		gch := tmpGroupChart{}
		gch.Range = newMean - oldMean
		gch.Name = mw.groups[cs]
		gch.Power = 0

		sumOfCSSellVal, ok1 := sumOfSellVal[cs]
		sumOfCSSellCount, ok2 := sumOfSellCount[cs]
		sumOfCSBuyVal, ok3 := sumOfBuyVal[cs]
		sumOfCSBuyCount, ok4 := sumOfBuyCount[cs]

		if ok1 && ok2 && ok3 && ok4 && sumOfCSSellCount > 0 && sumOfCSBuyCount > 0 && sumOfCSSellVal > 0 && sumOfCSBuyVal > 0 {
			meanSaraneS := sumOfCSSellVal / sumOfCSSellCount
			meanSaraneB := sumOfCSBuyVal / sumOfCSBuyCount

			if meanSaraneB >= meanSaraneS {
				gch.Power = float32(meanSaraneB) / float32(meanSaraneS)
				if gch.Power > 5 {
					gch.Power = 5
				}
			} else {
				gch.Power = -1 * float32(meanSaraneS) / float32(meanSaraneB)
				if gch.Power < -5 {
					gch.Power = -5
				}
			}
		}
		//if SaraneB, ok := groupsSaraneB[cs]; ok {
		//	if SaraneS, ok := groupsSaraneS[cs]; ok {
		//		meanSaraneS := mymath.MeanF64(SaraneS)
		//		meanSaraneB := mymath.MeanF64(SaraneB)
		//
		//		if meanSaraneS > float64(0) {
		//			if meanSaraneB >= meanSaraneS {
		//				gch.Power = float32(meanSaraneB) / float32(meanSaraneS)
		//				if gch.Power > 10 {
		//					gch.Power = 10
		//				}
		//			} else {
		//				gch.Power = -1 * float32(meanSaraneS) / float32(meanSaraneB)
		//				if gch.Power < -10 {
		//					gch.Power = -10
		//				}
		//			}
		//		}
		//	}
		//}
		if len(groupsChild[cs]) >= 5 {
			_groupsChart = append(_groupsChart, gch)
		}
	}

	//sort range
	sort.Slice(_groupsChart, func(i, j int) bool {
		return _groupsChart[i].Range > _groupsChart[j].Range
	})

	lenOfGroups := len(_groupsChart)
	listOfGroupName := make([]string, lenOfGroups)
	listOfGroupPower := make([]string, lenOfGroups)
	listOfGroupRange := make([]string, lenOfGroups)
	for index, _groupChart := range _groupsChart {
		listOfGroupName[index] = _groupChart.Name
		listOfGroupPower[index] = fmt.Sprintf("%.2f", _groupChart.Power)
		listOfGroupRange[index] = fmt.Sprintf("%.2f", _groupChart.Range)
	}

	lockHistoryMap.RLock()
	askBidCharts := mw.askBidCharts
	lockHistoryMap.RUnlock()

	askBidCharts.BidValueOfOrders = append(askBidCharts.BidValueOfOrders, sumOfBuy/1e10)
	askBidCharts.AskValueOfOrders = append(askBidCharts.AskValueOfOrders, sumOfSell/1e10)
	askBidCharts.StepOfOrderValue = append(askBidCharts.StepOfOrderValue, debugging.GetMinStep())

	gc := GroupCharts{
		ListOfGroupName:  listOfGroupName,
		ListOfGroupRange: listOfGroupRange,
		ListOfGroupPower: listOfGroupPower,
	}

	func() {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				parseError := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
				logger.GetFlog().Err(parseError).Send()
				return
			}
		}()
		SendGroupCharts(gc)
		SendAskBidCharts(askBidCharts)
	}()

	lockHistoryMap.Lock()
	defer lockHistoryMap.Unlock()
	mw.groupCharts = gc
	mw.askBidCharts = askBidCharts
}

func initStepInArchiveObjIfNeed(step int) {

	needMake := func() bool {
		lockHistoryMap.RLock()
		defer lockHistoryMap.RUnlock()
		if _, ok := mw.AllRowsArchiveObj[step]; !ok {
			return true
		}
		return false
	}()
	if needMake {
		lockHistoryMap.Lock()
		defer lockHistoryMap.Unlock()
		mw.AllRowsArchiveObj[step] = make(map[uint64]StockArchive)
	}
}

func writeToHistory(step int, id uint64, stock StockArchive) {
	if step < 10 || step > 330 {
		return
	}
	initStepInArchiveObjIfNeed(step)
	lockHistoryMap.Lock()

	mw.AllRowsArchiveObj[step][id] = stock
	lockHistoryMap.Unlock()

	if step >= 120 && step <= 330 && step%30 == 0 && !debugging.IsDebug() {
		lockHistoryMap.RLock()
		allArchive := mw.AllRowsArchiveObj
		lockHistoryMap.RUnlock()

		jsonSignals, errr := json.Marshal(Signals)
		if errr != nil {
			fmt.Print(errr)
		} else {
			errr = ioutil.WriteFile(fmt.Sprintf("./archive/%d_Signals.txt", step), jsonSignals, 0644)
			if errr != nil {
				fmt.Print(errr)
			}
		}

		allArchive2 := map[uint64]map[int]StockArchive{}
		for _step, _idArchive := range allArchive {
			for _id, _archive := range _idArchive {
				if _, ok := allArchive2[_id]; !ok {
					allArchive2[_id] = map[int]StockArchive{}
				}
				allArchive2[_id][_step] = _archive
			}
		}
		for _id, _stepArchive := range allArchive2 {
			jsonStepArchive, err := json.Marshal(_stepArchive)
			if err != nil {
				fmt.Println(err)
			} else {
				err := ioutil.WriteFile(fmt.Sprintf("./archive/%d_%d.txt", step, _id), jsonStepArchive, 0644)
				if err != nil {
					fmt.Print(err)
				}
			}
		}
	}

}

func triggerChangeStep() {
	saveStep()
	//if debugging.GetMinStep()%5 == 0 {
	//
	//}
}

var stockChannel chan map[uint64]Stock

func StartCrawl(_stockChannel chan map[uint64]Stock) {
	stockChannel = _stockChannel
	mw = GetMW()

	for {

		if CrawlerOn == false {
			break
		}

		step := debugging.GetMinStep()
		if step != currentStep {
			prvStep = currentStep
			currentStep = step
			triggerChangeStep()
		}

		//call first to init AllRowObject
		mw.UpdateMarketWatch()
		loc := mytime.GetIran()
		now := time.Now().In(loc)
		fmt.Println(now)

		if mw.UpdateCounter == 0 {
			go mw.LoadInstStat()
			go mw.LoadInstHistory()
			go mw.LoadGroups()
			mw.LoadClientType()
		}

		mw.UpdateCounter++

		//Note: don't call LoadClientType and UpdateMarketWatch parallel
		//mw.AllRowsObj updating not
		if mw.UpdateCounter%uint64(30/UpdateSpeed) == 0 || (debugging.IsDebug() && mw.UpdateCounter%uint64(30) == 0) {
			mw.LoadClientType()
		}
		if mw.UpdateCounter%uint64(5/UpdateSpeed) == 0 || (debugging.IsDebug() && mw.UpdateCounter%uint64(5) == 0) {
			if mw.calcedMean == false && mw.getLoadedInstHistory() {
				//concurrent race not occurred
				//because where mw.getLoadedInstHistory() return true , the mw.InstHistory released
				//better method select{} statements
				mw.calcedMean = true

				cfg := settings.LoadConfiguration()
				for insCode, arrayOfHist := range mw.InstHistory {
					var sum float64 = 0
					for i := 0; i < cfg.VolMeanDays; i++ {
						sum += float64(arrayOfHist[i].QTotTran5J)
					}
					tmp, ok := mw.AllRowsObj[insCode]
					if ok {
						tmp.MeanVol = sum / float64(cfg.VolMeanDays)
						mw.AllRowsObj[insCode] = tmp
					}
				}
			}
		}

		//fmt.Println("now", mw.AllRowsObj[43781018754867729].PL, mw.AllRowsObj[43781018754867729].QD3, mw.AllRowsObj[43781018754867729].QO3)
		if debugging.IsDebug() {
			time.Sleep(1 * time.Second)

		} else {
			time.Sleep(time.Duration(UpdateSpeed) * time.Second)
		}

	}
}

func GetHistoryVal(id uint64, key string, minPrv int, typ int) (interface{}, error) {
	history, err := GetHistory(id, minPrv)
	value := reflect.ValueOf(history).FieldByName(key)
	return (interface{})(value), err
}

func GetHistory(id uint64, minPrv int) (StockArchive, error) {
	lockHistoryMap.RLock()
	defer lockHistoryMap.RUnlock()
	res, ok := mw.AllRowsArchiveObj[debugging.GetMinStep()-minPrv][id]
	if !ok {
		return res, &notFoundHistory{id: id, min: debugging.GetMinStep() - minPrv}
	}
	return res, nil
}

func GetHistoryAnalog(id uint64, minPrv int) (StockArchive, error) {
	old, err := GetHistory(id, minPrv)
	if err != nil {
		return old, err
	}
	now := readerAllRowsObj[id]
	analog := StockArchive{
		inscode:     old.inscode,
		PC:          now.PC - old.PC,
		PCC:         now.PCC - old.PCC,
		PCP:         now.PCP - old.PCP,
		PL:          now.PL - old.PL,
		PLC:         now.PLC - old.PLC,
		PLP:         now.PLP - old.PLP,
		TNO:         now.TNO - old.TNO,
		TVOL:        now.TVOL - old.TVOL,
		TVAL:        now.TVAL - old.TVAL,
		PMIN:        now.PMIN - old.PMIN,
		PMAX:        now.PMAX - old.PMAX,
		VisitCount:  now.VisitCount - old.VisitCount,
		ZO1:         now.ZO1 - old.ZO1,
		ZD1:         now.ZD1 - old.ZD1,
		PD1:         now.PD1 - old.PD1,
		PO1:         now.PO1 - old.PO1,
		QD1:         now.QD1 - old.QD1,
		QO1:         now.QO1 - old.QO1,
		ZO2:         now.ZO2 - old.ZO2,
		ZD2:         now.ZD2 - old.ZD2,
		PD2:         now.PD2 - old.PD2,
		PO2:         now.PO2 - old.PO2,
		QD2:         now.QD2 - old.QD2,
		QO2:         now.QO2 - old.QO2,
		ZO3:         now.ZO3 - old.ZO3,
		ZD3:         now.ZD3 - old.ZD3,
		PD3:         now.PD3 - old.PD3,
		PO3:         now.PO3 - old.PO3,
		QD3:         now.QD3 - old.QD3,
		QO3:         now.QO3 - old.QO3,
		ZO4:         now.ZO4 - old.ZO4,
		ZD4:         now.ZD4 - old.ZD4,
		PD4:         now.PD4 - old.PD4,
		PO4:         now.PO4 - old.PO4,
		QD4:         now.QD4 - old.QD4,
		QO4:         now.QO4 - old.QO4,
		ZO5:         now.ZO5 - old.ZO5,
		ZD5:         now.ZD5 - old.ZD5,
		PD5:         now.PD5 - old.PD5,
		PO5:         now.PO5 - old.PO5,
		QD5:         now.QD5 - old.QD5,
		QO5:         now.QO5 - old.QO5,
		BuyCountI:   now.BuyCountI - old.BuyCountI,
		BuyCountN:   now.BuyCountN - old.BuyCountN,
		BuyIVolume:  now.BuyIVolume - old.BuyIVolume,
		BuyNVolume:  now.BuyNVolume - old.BuyNVolume,
		SellCountI:  now.SellCountI - old.SellCountI,
		SellCountN:  now.SellCountN - old.SellCountN,
		SellIVolume: now.SellIVolume - old.SellIVolume,
		SellNVolume: now.SellNVolume - old.SellNVolume,
		Power:       now.Power - old.Power,
		SaraneB:     now.SaraneB - old.SaraneB,
		SaraneS:     now.SaraneS - old.SaraneS,
	}
	return analog, nil
}
