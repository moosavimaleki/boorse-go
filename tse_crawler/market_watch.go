package tse_crawler

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"tsetmc/debugging"
	dt "tsetmc/utils/datatype"
	"tsetmc/utils/logger"
	"tsetmc/utils/settings"
	"tsetmc/utils/slice"
)

type GroupCharts struct {
	ListOfGroupName  []string
	ListOfGroupPower []string
	ListOfGroupRange []string
}
type AskBidCharts struct {
	StepOfOrderValue []int
	BidValueOfOrders []uint64
	AskValueOfOrders []uint64
}

type MarketWatch struct {
	client            *resty.Client
	heven             int
	refid             int
	UpdateCounter     uint64
	RoundNo           int //use only for sort by col, increase by each request
	needInit          int //if ID not found in AllRows
	NeedInitLock      *sync.RWMutex
	blockedStock      map[uint64]bool //like Oragh mosharekat
	preOpen           bool
	AllRowsObj        map[uint64]Stock
	TmpAllRowsObj     map[uint64]Stock
	AllRowsArchiveObj map[int]map[uint64]StockArchive
	askBidCharts      AskBidCharts
	groupCharts       GroupCharts
	InstHistory       map[uint64][]History
	InstStat          map[uint64]map[int]float64
	groups            map[int]string
	LoadedLock        *sync.RWMutex
	TmpAllRowLock     *sync.RWMutex
	LoadedInstHistory bool
	LoadedInstStat    bool
	LoadedGroups      bool
	calcedMean        bool
}

func (mw *MarketWatch) setLoadedInstHistory(val bool) {
	mw.LoadedLock.Lock()
	defer mw.LoadedLock.Unlock()
	mw.LoadedInstHistory = val
}
func (mw *MarketWatch) getLoadedInstHistory() bool {
	mw.LoadedLock.RLock()
	defer mw.LoadedLock.RUnlock()
	return mw.LoadedInstHistory
}
func (mw *MarketWatch) setLoadedInstStat(val bool) {
	mw.LoadedLock.Lock()
	defer mw.LoadedLock.Unlock()
	mw.LoadedInstStat = val
}
func (mw *MarketWatch) getLoadedInstStat() bool {
	mw.LoadedLock.RLock()
	defer mw.LoadedLock.RUnlock()
	return mw.LoadedInstStat
}
func (mw *MarketWatch) setLoadedGroups(val bool) {
	mw.LoadedLock.Lock()
	defer mw.LoadedLock.Unlock()
	mw.LoadedGroups = val
}
func (mw *MarketWatch) getLoadedGroups() bool {
	mw.LoadedLock.RLock()
	defer mw.LoadedLock.RUnlock()
	return mw.LoadedGroups
}

func (mw *MarketWatch) getTmpAllRow() map[uint64]Stock {
	mw.TmpAllRowLock.RLock()
	defer mw.TmpAllRowLock.RUnlock()
	return mw.TmpAllRowsObj
}
func (mw *MarketWatch) setTmpAllRow(tmp map[uint64]Stock) {
	mw.TmpAllRowLock.Lock()
	defer mw.TmpAllRowLock.Unlock()
	mw.TmpAllRowsObj = tmp
}

var baseUrl = "https://old.tsetmc.com"

var singletonMw *MarketWatch

func createMarketWatch() {
	fmt.Println("create")

	singletonMw = &MarketWatch{
		client:            resty.New(),
		heven:             0,
		refid:             0,
		UpdateCounter:     0,
		RoundNo:           1,
		AllRowsObj:        make(map[uint64]Stock, 1000),
		TmpAllRowsObj:     make(map[uint64]Stock, 1000),
		AllRowsArchiveObj: make(map[int]map[uint64]StockArchive, 300),
		InstHistory:       make(map[uint64][]History),
		InstStat:          make(map[uint64]map[int]float64),
		askBidCharts: AskBidCharts{
			AskValueOfOrders: make([]uint64, 0, 300),
			BidValueOfOrders: make([]uint64, 0, 300),
			StepOfOrderValue: make([]int, 0, 300),
		},
		preOpen:           true,
		NeedInitLock:      &sync.RWMutex{},
		LoadedLock:        &sync.RWMutex{},
		TmpAllRowLock:     &sync.RWMutex{},
		blockedStock:      map[uint64]bool{},
		LoadedInstHistory: false,
		LoadedInstStat:    false,
		LoadedGroups:      false,
		calcedMean:        false,
	}

	(*singletonMw).client.SetTimeout(10 * time.Second)
	(*singletonMw).client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			logger.GetFlog().Err(v.Err).Send()
		}
		// Log the error, increment a metric, etc...
	})

	settings.SetEventForAfterBlockListLoaded((*singletonMw).SetNeedInit)
}

func init() {
	createMarketWatch()
}

func GetMW() *MarketWatch {
	//Note: singleton pattern :
	//https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f
	//if mw == nil {
	// <--- NOT THREAD SAFE USE init() approach instant
	//
	//	createMarketWatch()
	//}
	return singletonMw
}

func PrintMW() {
	fmt.Println("MW is: ", singletonMw)
}

func AdvRoundSting(num float32, precision int) string {
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f", num)
}

func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

func (mw *MarketWatch) addNewRowToStoreObj(RowID uint64, data Stock) {
	if _, ok := mw.AllRowsObj[RowID]; !ok {
		mw.AllRowsObj[RowID] = data
	} else {
		//fmt.Println("exsits", mw.AllRowsObj[RowID])
	}
}

func (mw *MarketWatch) LoadInstStat() {

	var parseError error
sendAgain:
	url := baseUrl + "/tsev2/data/InstValue.aspx"
	fmt.Println(url)

	client := resty.New()
	client.SetTimeout(40 * time.Second)
	resp, requestErr := client.R().
		SetHeader("Accept", "text/html").
		SetHeader("Referer", "http://www.tsetmc.com/Loader.aspx?ParTree=15131F").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
		Get(url)
	if requestErr == nil && resp.StatusCode() == 200 {
		func() {
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					parseError = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(parseError).Send()
					return
				}
			}()
			data := string(resp.Body())
			var InsCode uint64 = 0
			var rows = strings.Split(data, ";")
			var cols []string
			for qpos := 0; qpos < len(rows); qpos++ {
				cols = strings.Split(rows[qpos], ",")
				if len(cols) == 3 {
					InsCode, requestErr = strconv.ParseUint(cols[0], 10, 64)
					if _, ok := mw.InstStat[InsCode]; !ok {
						mw.InstStat[InsCode] = make(map[int]float64)
					}
					mw.InstStat[InsCode][dt.ParseInt(cols[1])] = dt.ParseFloat64(cols[2])
				} else if len(cols) > 1 {
					if _, ok := mw.InstStat[InsCode]; !ok {
						mw.InstStat[InsCode] = make(map[int]float64)
					}
					mw.InstStat[InsCode][dt.ParseInt(cols[0])] = dt.ParseFloat64(cols[1])
				}
			}
			mw.setLoadedInstStat(true)
		}()
	}
	if requestErr != nil || parseError != nil {
		goto sendAgain
	}
}

func (mw *MarketWatch) LoadClientType() {
	url := baseUrl + "/tsev2/data/ClientTypeAll.aspx"
	fmt.Println(url)
	if debugging.IsDebug() {
		res := debugging.GetClientFile()
		if len(res) > 1 {
			err := mw.parseClientResponse(string(res))
			if err != nil {
				logger.GetFlog().Err(err).Send()
			}
		}

	} else {
		client := resty.New()
		client.SetTimeout(10 * time.Second)
		resp, err := client.R().
			SetHeader("Accept", "text/html").
			SetHeader("Referer", "http://www.tsetmc.com/Loader.aspx?ParTree=15131F").
			SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
			Get(url)
		if err == nil && resp.StatusCode() == 200 {
			err = mw.parseClientResponse(string(resp.Body()))
			if err != nil {
				logger.GetFlog().Err(err).Send()
			}
			step := debugging.GetMinStep()
			if step >= 120 && step <= 330 && !debugging.IsDebug() {
				mw.saveToFileClient(resp)
			}
		}
	}

}

func (mw *MarketWatch) parseClientResponse(resp string) error {
	var err error
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
			return
		}
	}()
	data := resp
	if len([]rune(data)) == 0 {
		return fmt.Errorf("response empty")
	}
	var rows = strings.Split(data, ";")
	var cols []string
	var RowID uint64 = 0

	for qpos := 0; qpos < len(rows); qpos++ {
		cols = strings.Split(rows[qpos], ",")
		RowID, err = strconv.ParseUint(cols[0], 10, 64)
		if err != nil {
			logger.GetFlog().Warn().Str("why client not RowID", cols[0]).Send()
			continue
		}
		tmpRow, ok := mw.AllRowsObj[RowID]
		if ok {
			tmpRow.BuyCountI = uint(convertStringToUint(cols[1], 32))
			tmpRow.BuyCountN = uint(convertStringToUint(cols[2], 32))
			tmpRow.BuyIVolume = uint64(convertStringToUint(cols[3], 64))
			tmpRow.BuyNVolume = uint64(convertStringToUint(cols[4], 64))
			tmpRow.SellCountI = uint(convertStringToUint(cols[5], 32))
			tmpRow.SellCountN = uint(convertStringToUint(cols[6], 32))
			tmpRow.SellIVolume = uint64(convertStringToUint(cols[7], 64))
			tmpRow.SellNVolume = uint64(convertStringToUint(cols[8], 64))

			var power float64 = 0
			var saraneB float64 = 0
			var saraneS float64 = 0
			if (tmpRow).BuyIVolume > 0 && (tmpRow).SellCountI > 0 && (tmpRow).SellIVolume > 0 && (tmpRow).BuyCountI > 0 {
				power = (float64((tmpRow).BuyIVolume) * float64((tmpRow).SellCountI)) / (float64((tmpRow).SellIVolume) * float64((tmpRow).BuyCountI))
			}
			if (tmpRow).BuyCountI > 0 {
				saraneB = (float64((tmpRow).BuyIVolume) * float64(tmpRow.PC) / float64((tmpRow).BuyCountI)) / 1e7
			}
			if (tmpRow).SellCountI > 0 {
				saraneS = (float64((tmpRow).SellIVolume) * float64(tmpRow.PC) / float64((tmpRow).SellCountI)) / 1e7
			}

			tmpRow.Power = power
			tmpRow.SaraneB = saraneB
			tmpRow.SaraneS = saraneS

			mw.AllRowsObj[RowID] = tmpRow
		}
	}

	return err
}

func (mw *MarketWatch) LoadInstHistory() {

	var parseError error
sendAgain:
	url := baseUrl + "/tsev2/data/ClosingPriceAll.aspx"
	fmt.Println(url)
	client := resty.New()
	client.SetTimeout(50 * time.Second)

	resp, requestError := client.R().
		SetHeader("Accept", "text/html").
		SetHeader("Referer", "http://www.tsetmc.com/Loader.aspx?ParTree=15131F").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
		Get(url)

	if requestError == nil && resp.StatusCode() == 200 {
		func() {
			//parse response
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					parseError = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(parseError).Send()
					return
				}
			}()
			data := string(resp.Body())
			var InsCode uint64 = 0
			var rows = strings.Split(data, ";")
			var cols []string
			var offset int
			var days int
			for qpos := 0; qpos < len(rows); qpos++ {
				cols = strings.Split(rows[qpos], ",")
				if len(cols) == 11 {
					InsCode, requestError = strconv.ParseUint(cols[0], 10, 64)
					offset = 1
				} else {
					offset = 0
				}

				if InsCode == 0 {
					fmt.Println("why code is 0")
					continue
				}

				days = dt.ParseInt(cols[offset])
				if _, ok := mw.InstHistory[InsCode]; !ok {
					mw.InstHistory[InsCode] = make([]History, 60)
				}

				if days > 59 {
					continue
				}

				mw.InstHistory[InsCode][days] = History{
					PClosing:       dt.ParseInt(cols[offset+1]),
					PDrCotVal:      dt.ParseUint64(cols[offset+2]),
					ZTotTran:       dt.ParseUint64(cols[offset+3]),
					QTotTran5J:     dt.ParseUint64(cols[offset+4]),
					QTotCap:        dt.ParseUint64(cols[offset+5]),
					PriceMin:       dt.ParseInt(cols[offset+6]),
					PriceMax:       dt.ParseInt(cols[offset+7]),
					PriceYesterday: dt.ParseInt(cols[offset+8]),
					PriceFirst:     dt.ParseInt(cols[offset+9]),
				}
			}
			mw.setLoadedInstHistory(true)
		}()
	}

	if requestError != nil || parseError != nil {
		logger.GetFlog().Err(requestError).Send()
		goto sendAgain
	}
}

func sendToChannel(msg map[uint64]Stock) {
	if len(stockChannel) > 1 {
		<-stockChannel
	}
	//copy it
	newMsg := make(map[uint64]Stock, len(msg))
	for id, stock := range msg {
		newMsg[id] = stock
	}
	stockChannel <- newMsg
}

func (mw *MarketWatch) saveToFile(resp *resty.Response, isInit bool) error {
	var err error
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
			logger.GetFlog().Err(err).Send()
			return
		}
	}()
	var isInitString string = ""
	if isInit {
		isInitString = "i"
	}
	err = ioutil.WriteFile(fmt.Sprintf("./all_data/%d_%d_%s.txt", debugging.GetMinStep(), time.Now().Second(), isInitString), resp.Body(), 0644)
	if err != nil {
		logger.GetFlog().Err(err).Send()
	}

	return err
}

func (mw *MarketWatch) saveToFileClient(resp *resty.Response) error {
	var err error
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
			logger.GetFlog().Err(err).Send()
			return
		}
	}()

	err = ioutil.WriteFile(fmt.Sprintf("./all_data/c_%d_%d.txt", debugging.GetMinStep(), time.Now().Second()), resp.Body(), 0644)
	if err != nil {
		logger.GetFlog().Err(err).Send()
	}

	return err
}

func (mw *MarketWatch) GetNeedInit() int {
	mw.NeedInitLock.RLock()
	defer mw.NeedInitLock.RUnlock()
	needInit := mw.needInit
	return needInit
}

func (mw *MarketWatch) SetNeedInit(needInit int) {
	mw.NeedInitLock.Lock()
	defer mw.NeedInitLock.Unlock()
	mw.needInit = needInit
}
func (mw *MarketWatch) UpdateMarketWatch() {
	url := baseUrl + "/tsev2/data/MarketWatchInit.aspx?r=0&h=0"
	isInitForce := false
	isInit := true
	if mw.UpdateCounter > 0 {
		isInit = false
	}
	if mw.GetNeedInit() > 10 {
		isInitForce = true
		mw.SetNeedInit(0)
		mw.heven = 0
		mw.refid = 0
		mw.RoundNo = 1
	} else if mw.heven != 0 {
		isInit = false
		isInitForce = false
		h := 5 * (mw.heven / 5)   // int/int -> floor
		r := 25 * (mw.refid / 25) // int/int -> floor
		url = baseUrl + "/tsev2/data/MarketWatchPlus.aspx?h=" + strconv.Itoa(h) + "&r=" + strconv.Itoa(r)
	}
	isInitForce = isInitForce

	if debugging.IsDebug() {
		logger.GetFlog().Debug().Str("Req Url", url).Send()
		var res []byte
		if isInit {
			res = debugging.GetFirstMarketFile()
		} else {
			res = debugging.GetMarketFile()
		}
		if len(res) > 1 {
			start := time.Now().UnixMicro()
			mw.parseTseResponse(string(res))
			sendToChannel(mw.AllRowsObj)
			logger.GetFlog().Debug().Int64("finish parse, Duration(ms):", (time.Now().UnixMicro()-start)/1000).Send()
		} else {
			fmt.Println("file peida nashod")
			start := time.Now().UnixMicro()
			sendToChannel(mw.AllRowsObj)
			logger.GetFlog().Debug().Int64("finish parse, Duration(ms):", (time.Now().UnixMicro()-start)/1000).Send()
		}
		mw.setTmpAllRow(mw.AllRowsObj)
	} else {
		logger.GetFlog().Debug().Str("Req Url", url).Send()
		resp, err := mw.client.R().
			SetHeader("Accept", "text/html").
			SetHeader("Referer", "http://www.tsetmc.com/Loader.aspx?ParTree=15131F").
			SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
			Get(url)
		if err == nil && resp.StatusCode() == 200 {
			start := time.Now().UnixMicro()
			mw.parseTseResponse(string(resp.Body()))
			sendToChannel(mw.AllRowsObj)
			step := debugging.GetMinStep()
			if step >= 120 && step <= 330 && !debugging.IsDebug() {
				mw.saveToFile(resp, isInit)
			}
			mw.setTmpAllRow(mw.AllRowsObj)
			logger.GetFlog().Debug().Int64("finish parse, Duration(ms):", (time.Now().UnixMicro()-start)/1000).Send()
		} else {
			sendToChannel(mw.AllRowsObj)
		}
	}

}

func (mw *MarketWatch) parseTseResponse(resp string) error {
	var err error
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
			logger.GetFlog().Err(err).Send()
			return
		}
	}()

	manualBlocked := settings.GetBlock()

	mw.RoundNo++
	if mw.RoundNo > 8 {
		mw.RoundNo = 1
	}

	all := strings.Split(resp, "@")
	InstPrice := strings.Split(all[2], ";")
	for ipos := 0; ipos < len(InstPrice); ipos++ {
		col := strings.Split(InstPrice[ipos], ",")
		RowIDStr := col[0]
		RowID, _ := strconv.ParseUint(RowIDStr, 10, 64)
		//if RowID == 24254843881948059{
		//	fmt.Println("update stock" ,col)
		//}
		if len(col) == 10 {
			//new val
			if _, ok := mw.AllRowsObj[RowID]; !ok {
				if mw.blockedStock[RowID] == false {
					logger.GetFlog().Warn().Str("RowID not found", strconv.FormatUint(RowID, 10)+","+strconv.Itoa(mw.needInit)).Send()
					mw.needInit++
				}
			} else {
				if _, founded := slice.FindUint64(manualBlocked, RowID); founded {
					delete(mw.AllRowsObj, RowID)
					mw.blockedStock[RowID] = true
					continue
				}
				//if RowID == 43531447361624437 {
				//	fmt.Println("---------------------hamed")
				//	fmt.Println(col)
				//}
				tmpRow := mw.AllRowsObj[RowID]
				var py = float32(tmpRow.PY)
				var eps = tmpRow.EPS

				tmpRow.heven = uint(convertStringToUint(col[1], 32))
				tmpRow.PF = uint(convertStringToUint(col[2], 32))
				pc := float32(convertStringToFloat(col[3], 32))
				tmpRow.PC = uint(pc)
				diffPc := pc - py
				tmpRow.PCC = int(diffPc)
				tmpRow.PCP = divide32(diffPc, py, 4) * 100
				pl := float32(convertStringToFloat(col[4], 32))
				tmpRow.PL = uint(pl)
				if col[5] != "0" {
					diff := pl - py
					tmpRow.PLC = int(diff)
					tmpRow.PLP = divide32(diff, py, 4) * 100
				} else {
					tmpRow.PLC = 0
					tmpRow.PLP = 0
				}
				tmpRow.TNO = uint64(convertStringToUint(col[5], 64))
				tmpRow.TVOL = uint64(convertStringToUint(col[6], 64)) //
				tmpRow.TVAL = uint64(convertStringToUint(col[7], 64))
				tmpRow.PMIN = uint(convertStringToUint(col[8], 32))
				tmpRow.PMAX = uint(convertStringToUint(col[9], 32))
				if eps == "" {
					tmpRow.PE = fmt.Sprintf("%.2f", divideStrings(col[4], eps, 2))
				} else {
					tmpRow.PE = ""
				}

				mw.AllRowsObj[RowID] = tmpRow

				//mw.ChangeRowList = append(mw.ChangeRowList, RowID)
				if mw.preOpen == true && col[5] != "0" {
					mw.preOpen = false
				}
				if mw.heven < dt.ParseInt(col[1]) {
					mw.heven = dt.ParseInt(col[1])
				}
			}
		} else if len(col) >= 22 {
			yval := uint(convertStringToUint(col[22], 64))
			l18 := col[2]
			if strings.HasPrefix(l18, "تسه") || strings.HasPrefix(l18, "تملي") ||
				(yval == 306 || yval == 301 || yval == 706 || yval == 208) || //ShowOraghMosharekat
				//(yval == 400 || yval == 403 || yval == 404) || //ShowHaghTaghaddom
				(yval == 263) || //ShowAti
				//(yval == 305 || yval == 380) || //ShowSandoogh
				(yval == 600 || yval == 602 || yval == 605 || yval == 311 || yval == 312 || yval == 321) || //ShowEkhtiarForoush
				(yval == 320 || yval == 603) || //ShowEkhtiarKharid
				(yval == 308 || yval == 701) || //ShowKala
				bool(false) {
				mw.blockedStock[RowID] = true
				continue
			}

			if _, founded := slice.FindUint64(manualBlocked, RowID); founded {
				delete(mw.AllRowsObj, RowID)
				mw.blockedStock[RowID] = true
				continue
			}

			//if RowID == 43531447361624437 {
			//	fmt.Println("---------------------hamed")
			//	fmt.Println(col[22])
			//	fmt.Println(yval)
			//}

			stock := Stock{
				inscode:    RowID,
				iid:        col[1],
				L18:        l18,
				l30:        col[3],
				PY:         uint(convertStringToUint(col[13], 32)),
				BVOL:       uint(convertStringToUint(col[15], 32)),
				VisitCount: uint64(convertStringToUint(col[16], 64)),
				flow:       uint(convertStringToUint(col[17], 32)),
				cs:         int(convertStringToUint(col[18], 32)),
				TMAX:       uint(convertStringToUint(col[19], 64)),
				TMIN:       uint(convertStringToUint(col[20], 64)),
				Z:          uint64(convertStringToUint(col[21], 64)),
				yval:       yval,
			}
			if strings.HasSuffix(l18, "2") {
				stock.IsBlock = true
			}
			if strings.HasSuffix(l18, "ح") {
				stock.IsHagh = true
			}
			if strings.HasSuffix(l18, "ح2") {
				stock.IsHagh = true
				stock.IsBlock = true
			}
			if stock.inscode == 59363563131789466 {
				fmt.Println("inited 59363563131789466", "////////////////////////")
			}

			//in addNewRowToStoreObj function we check if id is exists then skip them
			//if we force to init middle of matket time no problems occur
			mw.addNewRowToStoreObj(RowID, stock)

			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.heven = uint(convertStringToUint(col[4], 32))
			tmpRow.PF = uint(convertStringToUint(col[5], 32))

			py := float32(tmpRow.PY)
			pc := float32(convertStringToFloat(col[6], 32))
			pl := float32(convertStringToFloat(col[7], 32))
			tmpRow.PL = uint(pl)
			tmpRow.PC = uint(pc)
			//tmpRow.PY = uint(PY) //?

			diffPc := pc - py
			tmpRow.PCC = int(diffPc)
			tmpRow.PCP = divide32(diffPc, py, 4) * 100

			if col[8] != "0" {
				diff := pl - py
				tmpRow.PLC = int(diff)
				tmpRow.PLP = divide32(diff, py, 4) * 100
			} else {
				tmpRow.PLC = 0
				tmpRow.PLP = 0
			}

			tmpRow.TNO = uint64(convertStringToUint(col[8], 64))
			tmpRow.TVOL = uint64(convertStringToUint(col[9], 64))
			//fmt.Println("TVOL first time", tmpRow.TVOL)
			tmpRow.TVAL = uint64(convertStringToUint(col[10], 64))
			tmpRow.PMIN = uint(convertStringToUint(col[11], 32))
			tmpRow.PMAX = uint(convertStringToUint(col[12], 32))
			tmpRow.EPS = col[14]
			if tmpRow.EPS == "" {
				tmpRow.PE = fmt.Sprintf("%.2f", divideStrings(col[6], col[14], 2))
			} else {
				tmpRow.PE = ""
			}

			mw.AllRowsObj[RowID] = tmpRow

			//mw.ChangeRowList = append(mw.ChangeRowList, RowID)
			if mw.preOpen == true && col[8] != "0" {
				mw.preOpen = false
			}
			if mw.heven < dt.ParseInt(col[4]) {
				mw.heven = dt.ParseInt(col[4])
			}
		}
	}
	BestLimit := strings.Split(all[3], ";")
	for ipos := 0; ipos < len(BestLimit); ipos++ {
		var col = strings.Split(BestLimit[ipos], ",")
		RowIDStr := col[0]
		RowID, _ := strconv.ParseUint(RowIDStr, 10, 64)

		//if _, ok := mw.AllRows[RowID]; !ok {
		//	continue
		//}
		if _, ok := mw.AllRowsObj[RowID]; !ok {
			continue
		}

		//var data map[string]string
		switch col[1] {
		case "1":
			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.ZO1 = uint(convertStringToUint(col[2], 32))
			tmpRow.ZD1 = uint(convertStringToUint(col[3], 32))
			tmpRow.PD1 = uint(convertStringToUint(col[4], 32))
			tmpRow.PO1 = uint(convertStringToUint(col[5], 32))
			tmpRow.QD1 = uint64(convertStringToUint(col[6], 64))
			tmpRow.QO1 = uint64(convertStringToUint(col[7], 64))
			mw.AllRowsObj[RowID] = tmpRow
			break
		case "2":
			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.ZO2 = uint(convertStringToUint(col[2], 32))
			tmpRow.ZD2 = uint(convertStringToUint(col[3], 32))
			tmpRow.PD2 = uint(convertStringToUint(col[4], 32))
			tmpRow.PO2 = uint(convertStringToUint(col[5], 32))
			tmpRow.QD2 = uint64(convertStringToUint(col[6], 64))
			tmpRow.QO2 = uint64(convertStringToUint(col[7], 64))
			mw.AllRowsObj[RowID] = tmpRow
			break
		case "3":
			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.ZO3 = uint(convertStringToUint(col[2], 32))
			tmpRow.ZD3 = uint(convertStringToUint(col[3], 32))
			tmpRow.PD3 = uint(convertStringToUint(col[4], 32))
			tmpRow.PO3 = uint(convertStringToUint(col[5], 32))
			tmpRow.QD3 = uint64(convertStringToUint(col[6], 64))
			tmpRow.QO3 = uint64(convertStringToUint(col[7], 64))
			mw.AllRowsObj[RowID] = tmpRow
			break
		case "4":
			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.ZO4 = uint(convertStringToUint(col[2], 32))
			tmpRow.ZD4 = uint(convertStringToUint(col[3], 32))
			tmpRow.PD4 = uint(convertStringToUint(col[4], 32))
			tmpRow.PO4 = uint(convertStringToUint(col[5], 32))
			tmpRow.QD4 = uint64(convertStringToUint(col[6], 64))
			tmpRow.QO4 = uint64(convertStringToUint(col[7], 64))
			mw.AllRowsObj[RowID] = tmpRow
			break
		case "5":
			//data = map[string]string{
			//	"ZO5":     col[2],
			//	"ZD5":     col[3],
			//	"PD5":     col[4],
			//	"PO5":     col[5],
			//	"QD5":     col[6],
			//	"QO5":     col[7],
			//	"render":  "",
			//	"preview": "",
			//}
			tmpRow := mw.AllRowsObj[RowID]
			tmpRow.ZO5 = uint(convertStringToUint(col[2], 32))
			tmpRow.ZD5 = uint(convertStringToUint(col[3], 32))
			tmpRow.PD5 = uint(convertStringToUint(col[4], 32))
			tmpRow.PO5 = uint(convertStringToUint(col[5], 32))
			tmpRow.QD5 = uint64(convertStringToUint(col[6], 64))
			tmpRow.QO5 = uint64(convertStringToUint(col[7], 64))
			mw.AllRowsObj[RowID] = tmpRow
			break
		}
		//mw.addDataToStore(RowID, data)

		//mw.ChangeRowList = append(mw.ChangeRowList, RowID)
	}
	if all[4] != "0" && dt.ParseInt(all[4]) > mw.refid {
		mw.refid = dt.ParseInt(all[4])
	}

	//fmt.Println("---------------------hamed")
	//fmt.Println(time.Now())
	//fmt.Println(mw.AllRowsObj[24254843881948059].PL,mw.AllRowsObj[24254843881948059].PLP)
	return err
}

func (mw *MarketWatch) LoadGroups() {

	var parseError error
sendAgain:
	url := baseUrl + "/tsev2/res/loader.aspx?t=g&_510="
	fmt.Println(url)
	client := resty.New()
	client.SetTimeout(50 * time.Second)

	resp, requestError := client.R().
		SetHeader("Accept", "text/html").
		SetHeader("Referer", "http://www.tsetmc.com/Loader.aspx?ParTree=15131F").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
		Get(url)

	if requestError == nil && resp.StatusCode() == 200 {
		func() {
			//parse response
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					parseError = fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(parseError).Send()
					return
				}
			}()
			data := string(resp.Body())
			data = strings.Replace(data, "var Sectors=", "", 1)
			data = strings.Replace(data, ";", "", 1)
			data = strings.Replace(data, "'", string('"'), -1)
			groups := make([][2]string, 0, 100)
			parseError = json.Unmarshal([]byte(data), &groups)
			if parseError != nil {
				return
			}
			_groups := make(map[int]string, len(groups))
			_groups[25] = "لاستیک"
			_groups[35] = "حمل‌نقل"
			_groups[58] = "واسطه‌سایر"
			_groups[41] = "آب"
			_groups[51] = "ح.ن.هوایی"
			_groups[76] = "فکری"
			_groups[42] = "غذایی"
			_groups[47] = "خرده"
			_groups[50] = "موتور"
			_groups[14] = "معادن"
			_groups[43] = "دارویی"
			_groups[44] = "شیمیایی"
			_groups[56] = "سرمایه‌گذاری"
			_groups[65] = "واسطه‌پولی"
			_groups[70] = "مسکن"
			_groups[28] = "م‌فلزی"
			_groups[49] = "کاشی"
			_groups[52] = "انبارداری"
			_groups[57] = "بانک"
			_groups[33] = "پزشکی"
			_groups[46] = "تجارت"
			_groups[93] = "ورزشی"
			_groups[2] = "جنگل"
			_groups[17] = "منسوجات"
			_groups[27] = "فلزات"
			_groups[29] = "ماشین‌آلات"
			_groups[67] = "کمکی"
			_groups[40] = "برق‌گاز‌بخار"
			_groups[82] = "پشتیبانی"
			_groups[1] = "زراعت"
			_groups[31] = "کاغذ"
			_groups[32] = "مبل"
			_groups[63] = "پ‌حمل‌نقل"
			_groups[23] = "نفتی"
			_groups[26] = "کامپیوتری"
			_groups[60] = "حمل‌نقل‌ارتباطات"
			_groups[61] = "حمل‌ن آبی"
			_groups[71] = "مهندسی"
			_groups[73] = "ارتباطات"
			_groups[77] = "لیزینگ"
			_groups[90] = "هنری"
			_groups[31] = "ماشین‌آلات‌برقی"
			_groups[33] = "وسایل‌ارتباطی"
			_groups[34] = "خودرو"
			_groups[39] = "چند‌رشته"
			_groups[53] = "سیمان"
			_groups[54] = "کانی"
			_groups[72] = "رایانه"
			_groups[13] = "استخراج فلزی"
			_groups[20] = "چوبی"
			_groups[38] = "قند"
			_groups[68] = "صندوق"
			_groups[74] = "مهندسی"
			_groups[22] = "چاپ"
			_groups[55] = "هتل"
			_groups[10] = "زغال‌سنگ"
			_groups[11] = "نفت گاز"
			_groups[19] = "دباغی"
			_groups[66] = "بیمه"
			for _, val := range groups {
				strId := strings.TrimSpace(val[0])
				if strId == "" {
					continue
				}
				id, err := strconv.Atoi(strId)
				if err == nil {
					if _, ok := _groups[id]; !ok {
						// if has not default value
						_groups[id] = val[1]
					}
				} else {
					//like X1
					continue
				}
			}
			mw.groups = _groups
			mw.setLoadedGroups(true)
		}()
	}

	if requestError != nil {
		logger.GetFlog().Err(requestError).Send()
		goto sendAgain
	}
	if parseError != nil {
		logger.GetFlog().Err(parseError).Send()
		goto sendAgain
	}
}
