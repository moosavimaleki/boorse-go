package tse_crawler

import "time"

type Stock struct {
	inscode     uint64
	iid         string
	L18         string
	l30         string
	PY          uint
	heven       uint
	PF          uint
	PC          uint
	PCC         int
	PCP         float32
	MeanVol     float64
	PL          uint
	PLC         int
	PLP         float32
	TNO         uint64
	TVOL        uint64
	TVAL        uint64
	PMIN        uint
	PMAX        uint
	EPS         string
	PE          string
	BVOL        uint
	VisitCount  uint64
	flow        uint
	cs          int
	TMAX        uint
	TMIN        uint
	Z           uint64
	yval        uint
	ZO1         uint
	ZD1         uint
	PD1         uint
	PO1         uint
	QD1         uint64
	QO1         uint64
	ZO2         uint
	ZD2         uint
	PD2         uint
	PO2         uint
	QD2         uint64
	QO2         uint64
	ZO3         uint
	ZD3         uint
	PD3         uint
	PO3         uint
	QD3         uint64
	QO3         uint64
	ZO4         uint
	ZD4         uint
	PD4         uint
	PO4         uint
	QD4         uint64
	QO4         uint64
	ZO5         uint
	ZD5         uint
	PD5         uint
	PO5         uint
	QD5         uint64
	QO5         uint64
	BuyCountI   uint
	BuyCountN   uint
	BuyIVolume  uint64
	BuyNVolume  uint64
	SellCountI  uint
	SellCountN  uint
	SellIVolume uint64
	SellNVolume uint64
	SaraneB     float64
	SaraneS     float64
	Power       float64
	IsBlock     bool
	IsHagh      bool
}

type StockArchive struct {
	inscode     uint64
	PC          uint
	PCC         int
	PCP         float32
	PL          uint
	PLC         int
	PLP         float32
	TNO         uint64
	TVOL        uint64
	TVAL        uint64
	PMIN        uint
	PMAX        uint
	VisitCount  uint64
	ZO1         uint
	ZD1         uint
	PD1         uint
	PO1         uint
	QD1         uint64
	QO1         uint64
	ZO2         uint
	ZD2         uint
	PD2         uint
	PO2         uint
	QD2         uint64
	QO2         uint64
	ZO3         uint
	ZD3         uint
	PD3         uint
	PO3         uint
	QD3         uint64
	QO3         uint64
	ZO4         uint
	ZD4         uint
	PD4         uint
	PO4         uint
	QD4         uint64
	QO4         uint64
	ZO5         uint
	ZD5         uint
	PD5         uint
	PO5         uint
	QD5         uint64
	QO5         uint64
	BuyCountI   uint
	BuyCountN   uint
	BuyIVolume  uint64
	BuyNVolume  uint64
	SellCountI  uint
	SellCountN  uint
	SellIVolume uint64
	SellNVolume uint64
	SaraneB     float64
	SaraneS     float64
	Power       float64
}

func (s Stock) Qo5() uint64 {
	return s.QD5
}
func (s Stock) Qo4() uint64 {
	return s.QD4
}

type History struct {
	PClosing       int
	PDrCotVal      uint64
	PriceFirst     int
	PriceMax       int
	PriceMin       int
	PriceYesterday int
	QTotCap        uint64
	QTotTran5J     uint64
	ZTotTran       uint64
}

type Signal struct {
	NameFa           string
	SignalNameFa     string
	Name             string
	Signal           string
	Id               uint64
	Rid              uint
	SortValue        float64
	Power            float32
	SaraneB          float32
	SaraneS          float32
	Price            uint
	FinalPrice       uint
	Volume           uint64
	MeanVolume       uint64
	VolumeRatio      float32
	Info             map[string]string
	Filled           bool
	CreatedAt        int64
	BlockDurationSec int
}

func NewSignal(stock Stock, signalName string, blockDurationSec int) Signal {

	var rid uint = 0
	var meanVolume uint64 = 0
	var volumeRatio float32 = 0

	return Signal{
		Filled:           true,
		CreatedAt:        time.Now().Unix(),
		SignalNameFa:     stock.L18,
		NameFa:           stock.L18,
		Signal:           signalName,
		Id:               stock.inscode,
		Rid:              rid,
		Power:            float32(stock.Power),
		SaraneB:          float32(stock.SaraneB),
		SaraneS:          float32(stock.SaraneS),
		Price:            stock.PL,
		FinalPrice:       stock.PC,
		Volume:           stock.TVOL,
		MeanVolume:       meanVolume,
		VolumeRatio:      volumeRatio,
		BlockDurationSec: blockDurationSec,
	}
}

func (s *Signal) appendInfo(key string, value string) {
	if s.Info == nil {
		s.Info = map[string]string{}
	}
	s.Info[key] = value
}
