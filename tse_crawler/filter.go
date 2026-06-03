package tse_crawler

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
	"tsetmc/debugging"
	"tsetmc/ui"
	"tsetmc/utils/logger"
	"tsetmc/utils/settings"
)

// todo: fix size?
var Signals = make([]Signal, 0, 1000)
var SignalsBlockBank = make(map[uint64]map[string]uint64)
var readerAllRowsObj = map[uint64]Stock{}

var StartStep int
var maxTime int

func init() {
	StartStep = debugging.GetMinStep()
}

func getMaxTime() int {
	if maxTime > 0 {
		return maxTime
	}
	cfg := settings.LoadConfiguration()
	var max float64 = 0
	max = math.Max(float64(cfg.CodeBeCodePrvMin), max)
	max = math.Max(float64(cfg.ArzePrvMin), max)
	max = math.Max(float64(cfg.RangeJaheshMin), max)
	max = math.Max(float64(cfg.RangeGhiasiPowerFrom), max)
	max = math.Max(float64(cfg.SafeForooshMin), max)
	max = math.Max(float64(cfg.AstaneMin), max)
	max = math.Max(float64(cfg.SafeKharidMin), max)
	max = math.Max(float64(cfg.TaharokSafeKharidMin), max)
	maxTime = int(max)
	return maxTime
}

func isFilledHistory() bool {
	if debugging.IsDebug() {
		return debugging.GetMinStep()-StartStep > debugging.GetMaxTimeForFillHistory()
	}
	return debugging.GetMinStep()-StartStep > getMaxTime()
}

func CheckFilters(_stockChannel chan map[uint64]Stock) {
	for true {
		channelLen := len(_stockChannel)
		if channelLen > 0 {
			for i := 0; i < channelLen-1; i++ {
				<-_stockChannel
			}
			readerAllRowsObj = <-_stockChannel
			ClearWatchLists()
			for _, stock := range readerAllRowsObj {
				checkStock(stock)
				checkWatchLists(stock)
			}
			sendWatchLists()
			time.Sleep(3 * time.Second)
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func addSignal(sig Signal) {
	if sig.Filled && checkIfNotBlock(sig) {
		if _, ok := SignalsBlockBank[sig.Id]; !ok {
			SignalsBlockBank[sig.Id] = make(map[string]uint64)
		}
		SignalsBlockBank[sig.Id][sig.Signal] = uint64(time.Now().Unix())
		//todo: add log for signall and save parametters
		Signals = append(Signals, sig)

		SendSignal(sig)
	}
}

func SendSignal(sig Signal) {
	data, err := json.Marshal(sig)
	if err != nil {
		return
	}
	ui.EvalJS(fmt.Sprintf("addSignal(%s)", data))
}

func SendGroupCharts(gc GroupCharts) {
	data, err := json.Marshal(gc)
	if err != nil {
		return
	}
	ui.EvalJS(fmt.Sprintf("updateGroupCharts(%s)", data))
}

func SendAskBidCharts(abc AskBidCharts) {
	data, err := json.Marshal(abc)
	if err != nil {
		return
	}
	ui.EvalJS(fmt.Sprintf("updateAskBidCharts(%s)", data))
}

func ifHasOneError(errs ...error) bool {
	for _, err := range errs {
		if err != nil {
			if isFilledHistory() {
				logger.GetFlog().Debug().Err(err).Send()
			}
			return true
		}
	}
	return false
}

func checkIfNotBlock(sig Signal) bool {
	unix, ok := SignalsBlockBank[sig.Id][sig.Signal]
	if sig.Id == 23600798892801694 {
		print(unix, ok)
	}
	if !ok {
		return true
	}
	if sig.Id == 23600798892801694 {
		fmt.Println(sig)
		fmt.Println(uint64(time.Now().Unix()) - unix)
		fmt.Println(uint64(sig.BlockDurationSec))
		fmt.Println(uint64(time.Now().Unix())-unix > uint64(sig.BlockDurationSec))
	}
	if (uint64(time.Now().Unix())-unix)/60 > uint64(sig.BlockDurationSec) {
		return true
	}
	return false
}

func checkStock(stock Stock) {
	var sig Signal

	sig = clacCodeBeCode(stock)
	addSignal(sig)
	sig = clacEhtemalArze(stock)
	addSignal(sig)
	sig = clacRangeMosbat(stock)
	addSignal(sig)
	sig = clacRangeManfi(stock)
	addSignal(sig)

	sig = clacBoxGhodratmand(stock)
	addSignal(sig)

	sig = clacTaharok(stock)
	addSignal(sig)

	sig = clacAstane(stock)
	addSignal(sig)

	sig = clacSafeKharid(stock)
	addSignal(sig)

	sig = clacTaharokSafeKharid(stock)
	addSignal(sig)

}

// it's ok2
func clacCodeBeCode(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	if (stock.PL) != (stock.TMAX) {
		//(stock.Z < uint64(11e9)) &&

		if stock.TVOL == 0 {
			return Signal{}
		}

		hist, err1 := GetHistory(stock.inscode, cfg.CodeBeCodePrvMin)
		histA, err2 := GetHistoryAnalog(stock.inscode, cfg.CodeBeCodePrvMin)
		if (hist.TVOL) == 0 || ifHasOneError(err2, err1) {
			return Signal{}
		}
		var newCcSellNVolume = float64(histA.SellNVolume)
		var newCcBuyIVolume = float64(histA.BuyIVolume)
		var newCcBuyCounti = float64(histA.BuyCountI)
		var oldsnv = float64(hist.SellNVolume)
		var X = float64((stock).SellNVolume) / float64(stock.TVOL) * 100
		var Y = (oldsnv) / float64(hist.TVOL) * 100

		if newCcSellNVolume > 0 && newCcBuyIVolume > 0 && newCcBuyCounti > 0 &&
			float64((stock).SellNVolume)-oldsnv > cfg.CodeBeCodeMinimumHoghooghiVol &&
			int(X-Y) >= cfg.CodeBeCodeHoghooghiJump &&
			(newCcBuyIVolume*float64(stock.PC)/newCcBuyCounti) > cfg.CodeBeCodeSaraneKharidGhiasi &&
			newCcBuyIVolume > cfg.CodeBeCodeZaribKharidHaghighi*float64(newCcSellNVolume) {
			var sig = NewSignal(stock, "codeBeCode", cfg.CodeBeCodeBlock)
			sig.SignalNameFa = "احتمال کد به کد"
			sig.appendInfo("Old SellNVolume", fmt.Sprintf("%.0f", oldsnv))
			sig.appendInfo("SellNVolume", fmt.Sprintf("%d", stock.SellNVolume))
			sig.appendInfo("old % Hoghooghi", fmt.Sprintf("%.2f", Y))
			sig.appendInfo("% Hoghooghi", fmt.Sprintf("%.2f", X))
			sig.appendInfo("Sarane Ghiasi", fmt.Sprintf("%.2f", (newCcBuyIVolume*float64(stock.PC)/newCcBuyCounti)))
			sig.appendInfo("BuyIVolume Ghiasi", fmt.Sprintf("%.0f", newCcBuyIVolume))
			sig.appendInfo("SellNVolume Ghiasi", fmt.Sprintf("%.0f", newCcBuyIVolume))
			sig.appendInfo("SellNVolume Ghiasi*Zarib", fmt.Sprintf("%.2f", cfg.CodeBeCodeZaribKharidHaghighi*float64(newCcSellNVolume)))
			return sig
		}

	}

	return Signal{}
}

// it's ok2
func clacEhtemalArze(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	hist, err1 := GetHistory(stock.inscode, cfg.ArzePrvMin)
	if ifHasOneError(err1) {
		return Signal{}
	}
	var lastQd5 = hist.QD1
	var currentMin = debugging.GetMinStep()

	localDebug := false
	if debugging.IfDebugStock(stock.inscode) && localDebug {
		fmt.Println("--------------------------------------", stock.inscode)
		fmt.Println("(stock.PD1) == (stock.TMAX)", (stock.PD1) == (stock.TMAX), (stock.PD1), (stock.TMAX))
		fmt.Println("(stock.PF) == (stock.TMAX)", (stock.PF) == (stock.TMAX), (stock.PF), (stock.TMAX))
		fmt.Println("(lastQd5)*uint64(stock.TMAX) > cfg.ArzeArzeshSaf", (lastQd5)*uint64(stock.TMAX) > cfg.ArzeArzeshSaf, (lastQd5), uint64(stock.TMAX), (lastQd5)*uint64(stock.TMAX), cfg.ArzeArzeshSaf)
		fmt.Println("(stock.PMIN) == (stock.PMAX)", (stock.PMIN) == (stock.PMAX), (stock.PMIN), (stock.PMAX))
		fmt.Println("(stock.PD1) != (stock.TMAX)", (stock.PD1) != (stock.TMAX), (stock.PD1), (stock.TMAX))
		fmt.Println("(stock.PD1) == (stock.TMAX)", (stock.PD1) == (stock.TMAX), (stock.PD1), (stock.TMAX))
		fmt.Println("(stock.QO1) < 1", (stock.QO1) < 1, (stock.QO1))
	}

	if (stock.PD1) == (stock.TMAX) &&
		(stock.PF) == (stock.TMAX) &&
		(lastQd5)*uint64(stock.TMAX) > cfg.ArzeArzeshSaf &&
		(stock.PMIN) == (stock.PMAX) &&
		currentMin > 121 {
		//&&((stock.QO1) < 1 || )
		//{

		hist1, err1 := GetHistory(stock.inscode, 1)
		if ifHasOneError(err1) {
			return Signal{}
		}

		var last_qd11 = hist1.QD1
		var last_tvol1 = hist1.TVOL

		if debugging.IfDebugStock(stock.inscode) && localDebug {
			fmt.Println("STEP111111111", stock.inscode)
			fmt.Println("(last_qd11) > (last_tvol1) ", (last_qd11) > (last_tvol1), (last_qd11), (last_tvol1))
			fmt.Println("(stock.TVOL) > (stock.QD1)", (stock.TVOL) > (stock.QD1), (stock.TVOL), (stock.QD1))
			fmt.Println("(stock.TVOL) > (last_tvol1)", (stock.TVOL) > (last_tvol1), (stock.TVOL), (last_tvol1))
		}

		if last_qd11 > 0 && last_tvol1 > 0 &&
			(last_qd11) > (last_tvol1) &&
			(stock.TVOL) > (last_tvol1) &&
			(stock.TVOL) > (stock.QD1) {
			var sig = NewSignal(stock, "ehtemalArze", cfg.EhtemalArzeBlock)
			sig.SignalNameFa = "احتمال عرضه"

			sig.appendInfo("Old qd1", fmt.Sprintf("%d", last_qd11))
			sig.appendInfo("qd1", fmt.Sprintf("%d", stock.QD1))
			sig.appendInfo("Old Vol", fmt.Sprintf("%d", last_tvol1))
			sig.appendInfo("Vol", fmt.Sprintf("%d", stock.TVOL))

			return sig
		}

	}

	return Signal{}
}

// it's ok2
func clacRangeMosbat(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin-cfg.RangeJaheshMin <= 120 {
		return Signal{}
	}
	hist, err1 := GetHistory(stock.inscode, cfg.RangeJaheshMin)
	histA, err2 := GetHistoryAnalog(stock.inscode, cfg.RangeJaheshMin)
	histA5, err3 := GetHistoryAnalog(stock.inscode, cfg.RangeGhiasiPowerFrom)

	if ifHasOneError(err3, err2, err1) {
		return Signal{}
	}

	var plp2 = hist.PLP
	var vol_ap5 = histA.TVOL

	var power_ap_55 = (float64(histA5.BuyIVolume) / float64(histA5.BuyCountI)) / (float64(histA5.SellIVolume) / float64(histA5.SellCountI))

	if plp2 > -10 && (plp2 != 0 || (plp2 == 0 && hist.TVOL > 0)) &&
		vol_ap5*uint64(stock.PL) > cfg.RangeArzeshMoamelatGhiasi &&
		(stock.TVOL)*uint64(stock.PL) > cfg.RangeArzeshMoamelat &&
		(stock.PLP)-plp2 > cfg.RangeJaheshPercent &&
		power_ap_55 > cfg.RangeGhiasiPower {
		var sig = NewSignal(stock, "rangeMosbat", cfg.RangeMosbatBlock)
		sig.SignalNameFa = "رنج مثبت"
		sig.appendInfo("Old PricePercent", fmt.Sprintf("%.2f", plp2))
		sig.appendInfo("PricePercent", fmt.Sprintf("%.2f", stock.PLP))
		sig.appendInfo("ArzeshMoamelatGhiasi", fmt.Sprintf("%d", vol_ap5*uint64(stock.PL)))
		sig.appendInfo("ArzeshMoamelat", fmt.Sprintf("%d", (stock.TVOL)*uint64(stock.PL)))
		sig.appendInfo("GhiasiPower", fmt.Sprintf("%.2f", power_ap_55))
		return sig
	}

	return Signal{}
}

// it's ok2
func clacRangeManfi(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin-cfg.RangeJaheshMin <= 120 {
		return Signal{}
	}
	hist, err1 := GetHistory(stock.inscode, cfg.RangeJaheshMin)
	histA, err2 := GetHistoryAnalog(stock.inscode, cfg.RangeJaheshMin)
	histA5, err3 := GetHistoryAnalog(stock.inscode, cfg.RangeGhiasiPowerFrom)
	if ifHasOneError(err3, err2, err1) {
		return Signal{}
	}
	var plp2 = hist.PLP
	var vol_ap5 = histA.TVOL

	var power_ap_55 = (float64(histA5.BuyIVolume) / float64(histA5.BuyCountI)) / (float64(histA5.SellIVolume) / float64(histA5.SellCountI))

	if plp2 > -10 && (plp2 != 0 || (plp2 == 0 && hist.TVOL > 0)) &&
		vol_ap5*uint64(stock.PL) > cfg.RangeArzeshMoamelatGhiasi &&
		(stock.TVOL)*uint64(stock.PL) > cfg.RangeArzeshMoamelat &&
		plp2-(stock.PLP) > cfg.RangeJaheshPercent &&
		power_ap_55 < cfg.RangeGhiasiPower {
		var sig = NewSignal(stock, "rangeManfi", cfg.RangeManfiBlock)
		sig.SignalNameFa = "رنج منفی"
		sig.appendInfo("Old PricePercent", fmt.Sprintf("%.2f", plp2))
		sig.appendInfo("PricePercent", fmt.Sprintf("%.2f", stock.PLP))
		sig.appendInfo("ArzeshMoamelatGhiasi", fmt.Sprintf("%d", vol_ap5*uint64(stock.PL)))
		sig.appendInfo("ArzeshMoamelat", fmt.Sprintf("%d", (stock.TVOL)*uint64(stock.PL)))
		sig.appendInfo("GhiasiPower", fmt.Sprintf("%.2f", power_ap_55))
		return sig
	}

	return Signal{}
}

// it's ok2
func clacBoxGhodratmand(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin <= 120 {
		return Signal{}
	}

	positiveBoundary := uint(float64(stock.PY) + (float64(stock.TMAX)-float64(stock.PY))*cfg.BoxGhodratmandPositivePercent/100)
	negativeBoundary := uint(float64(stock.PY) - (float64(stock.PY)-float64(stock.TMIN))*cfg.BoxGhodratmandNegativePercent/100)

	mashkokMabna := uint64(float64(stock.BVOL) * float64(cfg.BoxGhodratmandMabnaRatio) / 100)
	samteForoosh := false
	samteKharid := false
	rowName := ""
	var rowValue uint64 = 0

	if stock.QD2 > mashkokMabna && stock.QD2*uint64(stock.PD2) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PD2 && stock.PD2 <= positiveBoundary) {
		samteKharid = true
		rowName = "QD2"
		rowValue = stock.QD2
	}
	if stock.QD3 > mashkokMabna && stock.QD3*uint64(stock.PD3) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PD3 && stock.PD3 <= positiveBoundary) {
		samteKharid = true
		rowName = "QD3"
		rowValue = stock.QD3
	}
	if stock.QD4 > mashkokMabna && stock.QD4*uint64(stock.PD4) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PD4 && stock.PD4 <= positiveBoundary) {
		samteKharid = true
		rowName = "QD4"
		rowValue = stock.QD4
	}
	if stock.QD5 > mashkokMabna && stock.QD5*uint64(stock.PD5) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PD5 && stock.PD5 <= positiveBoundary) {
		samteKharid = true
		rowName = "QD5"
		rowValue = stock.QD5
	}

	if stock.QO2 > mashkokMabna && stock.QO2*uint64(stock.PO2) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PO2 && stock.PO2 <= positiveBoundary) {
		samteForoosh = true
		rowName = "QO2"
		rowValue = stock.QO2
	}
	if stock.QO3 > mashkokMabna && stock.QO3*uint64(stock.PO3) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PO3 && stock.PO3 <= positiveBoundary) {
		samteForoosh = true
		rowName = "QO3"
		rowValue = stock.QO3
	}
	if stock.QO4 > mashkokMabna && stock.QO4*uint64(stock.PO4) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PO4 && stock.PO4 <= positiveBoundary) {
		samteForoosh = true
		rowName = "QO4"
		rowValue = stock.QO4
	}
	if stock.QO5 > mashkokMabna && stock.QO5*uint64(stock.PO5) > cfg.BoxGhodratmandMinValue && (negativeBoundary <= stock.PO5 && stock.PO5 <= positiveBoundary) {
		samteForoosh = true
		rowName = "QO5"
		rowValue = stock.QO5
	}

	if !samteForoosh && !samteKharid {
		return Signal{}
	}

	var sig = NewSignal(stock, "boxGhodratmand", cfg.BoxGhodratmandBlock)
	sig.SignalNameFa = "باکس قدرتمند"
	if samteForoosh && samteKharid {
		sig.appendInfo("type", "both")
	} else if samteKharid {
		sig.appendInfo("type", "kharid")
	} else if samteForoosh {
		sig.appendInfo("type", "foroosh")
	}

	sig.appendInfo(rowName, fmt.Sprintf("%d", rowValue))

	return sig

}

// it's ok2
func clacTaharok(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin-cfg.SafeForooshMin <= 120 {
		return Signal{}
	}

	if (stock.PL) > 1000 &&
		(stock.PO1) == (stock.TMIN) {

		var hist, err = GetHistory(stock.inscode, cfg.SafeForooshMin)
		if ifHasOneError(err) {
			return Signal{}
		}
		var old_qo1 = float64(hist.QO1)
		var old_vol = float64(hist.TVOL)

		if float64(stock.QO1) < (old_qo1)*float64(cfg.SafeForooshKaheshRatio) &&
			(float64(stock.TVOL)-old_vol) > float64(cfg.SafeForooshHajmRatio)*(old_qo1) &&
			(cfg.SafeForooshAvalinBar == 0 || (stock.PMIN) == (stock.PMAX)) &&
			(old_qo1) > cfg.SafeForooshMinimumQO {
			var sig = NewSignal(stock, "taharok", cfg.TaharokBlock)
			sig.SignalNameFa = "تحرک صف فروش"
			sig.appendInfo("Price", fmt.Sprintf("%d", stock.PL))
			sig.appendInfo("po1", fmt.Sprintf("%d", stock.PO1))
			sig.appendInfo("Min Mojaz", fmt.Sprintf("%d", stock.TMIN))
			sig.appendInfo("Vol", fmt.Sprintf("%d", stock.TVOL))
			sig.appendInfo("Old qo1", fmt.Sprintf("%.0f", old_qo1))
			sig.appendInfo("Old Vol", fmt.Sprintf("%.0f", old_vol))
			sig.appendInfo("Max Price", fmt.Sprintf("%d", stock.PMAX))
			sig.appendInfo("Min Price", fmt.Sprintf("%d", stock.PMIN))
			return sig
		}
	}

	return Signal{}
}

// it's ok2
func clacAstane(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin <= 120 {
		return Signal{}
	}
	localDebug := false
	if debugging.IfDebugStock(stock.inscode) && localDebug {
		fmt.Println("--------------------------------------", stock.inscode)
		fmt.Println("float64(stock.PLP) >= (((float64(stock.TMAX)/float64(stock.PY))-1)*float64(cfg.AstanePercentAbove))", float64(stock.PLP) >= (((float64(stock.TMAX)/float64(stock.PY))-1)*float64(cfg.AstanePercentAbove)), float64(stock.PLP), ((float64(stock.TMAX)/float64(stock.PY))-1)*float64(cfg.AstanePercentAbove), float64(stock.TMAX), float64(stock.PY))
		fmt.Println("stock.Power > float64(cfg.AstanePower)", stock.Power > float64(cfg.AstanePower), stock.Power, float64(cfg.AstanePower))
		fmt.Println("stock.TVOL > 200e3", stock.TVOL > 200e3, stock.TVOL, 200e3)
		fmt.Println("stock.SaraneB > float64(cfg.AstaneSaraneB)", stock.SaraneB > float64(cfg.AstaneSaraneB), stock.SaraneB, float64(cfg.AstaneSaraneB))
		fmt.Println("stock.PL != stock.TMAX", stock.PL != stock.TMAX, stock.PL, stock.TMAX)
	}

	if float64(stock.PLP) >= (((float64(stock.TMAX)/float64(stock.PY))-1)*float64(cfg.AstanePercentAbove)) &&
		stock.Power > float64(cfg.AstanePower) &&
		stock.SaraneB > float64(cfg.AstaneSaraneB) &&
		stock.TVOL > cfg.AstaneVol &&
		stock.PL != stock.TMAX {

		var hist, err = GetHistory(stock.inscode, cfg.AstaneMin)

		if debugging.IfDebugStock(stock.inscode) && localDebug {
			fmt.Println("fffffffffffffffffffffffffffffffffff")
			fmt.Println("hist.PL != stock.TMAX ", hist.PL != stock.TMAX, hist.PL, stock.TMAX)
		}
		if ifHasOneError(err) {
			return Signal{}
		}

		if hist.PL != stock.TMAX {
			var sig = NewSignal(stock, "astane", cfg.AstaneBlock)
			sig.SignalNameFa = "آستانه صف خرید"
			sig.appendInfo("Price", fmt.Sprintf("%d", stock.PL))
			sig.appendInfo("Price%", fmt.Sprintf("%.2f", stock.PLP))
			sig.appendInfo("Trigger%", fmt.Sprintf("%.2f", (((float64(stock.TMAX)/float64(stock.PY))-1)*float64(cfg.AstanePercentAbove))))
			sig.appendInfo("Power", fmt.Sprintf("%.2f", hist.Power))
			sig.appendInfo("Vol", fmt.Sprintf("%d", hist.TVOL))
			return sig
		}
	}

	return Signal{}
}

// it's ok2
func clacSafeKharid(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin <= 120 {
		return Signal{}
	}

	localDebug := true
	if debugging.IfDebugStock(stock.inscode) && localDebug {
		fmt.Println("stock.QO1 == 0", stock.QO1 == 0, stock.QO1)
		fmt.Println(" stock.PO1>stock.TMAX ", stock.PO1 > stock.TMAX, stock.PO1, stock.TMAX)
		fmt.Println(" stock.PL == stock.TMAX ", stock.PL == stock.TMAX, stock.PL, stock.TMAX)
	}

	if (stock.QO1 == 0 || stock.PO1 > stock.TMAX) && stock.PD1 == stock.TMAX {
		var hist, err = GetHistory(stock.inscode, cfg.SafeKharidMin)
		if ifHasOneError(err) {
			return Signal{}
		}
		if debugging.IfDebugStock(stock.inscode) && localDebug {
			fmt.Println("hist.PL != stock.TMAX", hist.PL != stock.TMAX, hist.PL, stock.TMAX)
		}
		if hist.PL != stock.TMAX {
			var sig = NewSignal(stock, "safeKharid", cfg.SafeKharidBlock)
			sig.SignalNameFa = "صف خرید سریع"
			sig.appendInfo("qd1", fmt.Sprintf("%d", stock.QD1))
			sig.appendInfo("Old Price", fmt.Sprintf("%d", hist.PL))
			sig.appendInfo("Price", fmt.Sprintf("%d", stock.PL))
			sig.appendInfo("Max Mojaz", fmt.Sprintf("%d", stock.TMAX))
			return sig
		}
	}

	return Signal{}
}

// it's ok2
func clacTaharokSafeKharid(stock Stock) Signal {
	cfg := settings.LoadConfiguration()
	var currentMin = debugging.GetMinStep()
	if currentMin-cfg.TaharokSafeKharidMin <= 120 {
		return Signal{}
	}
	localDebug := false

	if debugging.IfDebugStock(stock.inscode) && localDebug {
		fmt.Println("--------------------------------------", stock.inscode)
		fmt.Println("(stock.PL) > 1000", (stock.PL) > 1000, (stock.PL), 1000)
		fmt.Println("(stock.PD1) == (stock.TMAX)", (stock.PD1) == (stock.TMAX), (stock.PD1), (stock.TMAX))
	}

	if (stock.PL) > 1000 &&
		(stock.PD1) == (stock.TMAX) {

		var hist, err = GetHistory(stock.inscode, cfg.TaharokSafeKharidMin)
		if ifHasOneError(err) {
			if debugging.IfDebugStock(stock.inscode) && localDebug {
				fmt.Println("ifHasOneError")
			}
			return Signal{}
		}
		var old_qd1 = float64(hist.QD1)
		var old_vol = float64(hist.TVOL)
		if debugging.IfDebugStock(stock.inscode) && localDebug {
			fmt.Println("float64(stock.QD1) > (old_qd1)*(1+(cfg.TaharokSafeKharidIncreasePercent/100.0))", float64(stock.QD1) > (old_qd1)*(1+(cfg.TaharokSafeKharidIncreasePercent/100.0)), float64(stock.QD1), (old_qd1), (1 + (cfg.TaharokSafeKharidIncreasePercent / 100.0)))
			fmt.Println("(float64(stock.TVOL)-old_vol) < (0.02)*(old_qd1)", (float64(stock.TVOL)-old_vol) < (0.02)*(old_qd1), (float64(stock.TVOL) - old_vol), (0.02), (old_qd1))
			fmt.Println("(old_qd1) > 100000 ", (old_qd1) > 100000, (old_qd1), 100000)
		}
		if float64(stock.QD1) > (old_qd1)*(1+(cfg.TaharokSafeKharidIncreasePercent/100.0)) &&
			//(float64(stock.TVOL)-old_vol) < (0.02)*(old_qd1) &&
			(old_qd1) > cfg.TaharokSafeKharidMinimumQD {
			var sig = NewSignal(stock, "taharokSafeKharid", cfg.TaharokBlock)
			sig.SignalNameFa = "تحرک صف خرید"
			sig.appendInfo("old qd1", fmt.Sprintf("%d", uint64(old_qd1)))
			sig.appendInfo("qd1", fmt.Sprintf("%d", stock.QD1))
			sig.appendInfo("Price", fmt.Sprintf("%d", stock.PL))
			sig.appendInfo("pd1", fmt.Sprintf("%d", stock.PD1))
			sig.appendInfo("Max Mojaz", fmt.Sprintf("%d", stock.TMAX))
			return sig
		}
	}

	return Signal{}
}
