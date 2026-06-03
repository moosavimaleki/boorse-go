package tse_crawler

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"tsetmc/ui"
)

var Vol = map[uint64]Signal{}
var SafeKharidValue = map[uint64]Signal{}
var SafeForooshValue = map[uint64]Signal{}

func updateList(sig Signal, list *map[uint64]Signal) {
	if sig.Filled {
		(*list)[sig.Id] = sig
	}
}

func ClearWatchLists() {
	//clear old list
	Vol = map[uint64]Signal{}
	SafeKharidValue = map[uint64]Signal{}
	SafeForooshValue = map[uint64]Signal{}
}

func convertToSlice(_map map[uint64]Signal) []Signal {
	tmp := make([]Signal, len(_map))
	idx := 0
	for _, value := range _map {
		tmp[idx] = value
		idx++
	}
	return tmp
}
func sortSignals(sigs *[]Signal) {
	sort.Slice(*sigs, func(i, j int) bool {
		return (*sigs)[i].SortValue > (*sigs)[j].SortValue
	})
}

func sendWatchLists() {
	if len(Vol) > 1 {
		tmp := convertToSlice(Vol)
		sortSignals(&tmp)
		volJson, _ := json.Marshal(tmp)
		ui.EvalJS(fmt.Sprintf("updateLists('vol',%s)", volJson))
	}
	if len(SafeKharidValue) > 1 {
		tmp := convertToSlice(SafeKharidValue)
		sortSignals(&tmp)
		safeKharidValueJson, _ := json.Marshal(tmp)
		ui.EvalJS(fmt.Sprintf("updateLists('safeKharidValue',%s)", safeKharidValueJson))
	}
	if len(SafeForooshValue) > 1 {
		tmp := convertToSlice(SafeForooshValue)
		sortSignals(&tmp)
		safeForooshValueJson, _ := json.Marshal(tmp)
		//fmt.Println(safeForooshValueJson)
		//fmt.Println(aa)
		ui.EvalJS(fmt.Sprintf("updateLists('safeForooshValue',%s)", safeForooshValueJson))
	}

}

func checkWatchLists(stock Stock) {
	var sig Signal

	sig = clacVol(stock)
	updateList(sig, &Vol)
	sig = clacSafeKharidValue(stock)
	updateList(sig, &SafeKharidValue)
	sig = clacSafeForooshValue(stock)
	updateList(sig, &SafeForooshValue)
}

func clacSafeKharidValue(stock Stock) Signal {
	if stock.PL == stock.TMAX && stock.QD1*uint64(stock.PL) > 1e9 {
		sig := NewSignal(stock, "safeKharidValue", 0)
		tmp := stock.QD1 * uint64(stock.PL)
		sig.SortValue = float64(tmp)
		sig.appendInfo("value", strconv.FormatUint(tmp, 10))
		return sig
	}
	return Signal{}
}

func clacSafeForooshValue(stock Stock) Signal {
	if stock.PL == stock.TMIN && stock.QO1*uint64(stock.PL) > 1e9 {
		sig := NewSignal(stock, "safeForooshValue", 0)
		tmp := stock.QO1 * uint64(stock.PL)
		sig.SortValue = float64(tmp)
		sig.appendInfo("value", strconv.FormatUint(tmp, 10))
		return sig
	}
	return Signal{}
}

func clacVol(stock Stock) Signal {
	if stock.MeanVol > 0 {
		sig := NewSignal(stock, "vol", 0)
		sig.MeanVolume = uint64(stock.MeanVol)
		sig.SortValue = float64(stock.TVOL) / stock.MeanVol
		return sig
	}
	return Signal{}
}
