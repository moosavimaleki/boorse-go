package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	dt "tsetmc/utils/datatype"
	"tsetmc/utils/logger"
	"tsetmc/utils/struct_tools"
)

type Config struct {
	UpdateSpeed           int
	ShowAti               bool
	ShowEkhtiarForoush    bool
	ShowHaghTaghaddom     bool
	ShowHousingFacilities bool
	ShowOraghMosharekat   bool
	ShowPayeFarabourse    bool
	ShowSaham             bool
	ShowSandoogh          bool

	//filter params
	//CodeBeCode
	CodeBeCodePrvMin              int
	CodeBeCodeHoghooghiJump       int
	CodeBeCodeMinimumHoghooghiVol float64
	CodeBeCodeSaraneKharidGhiasi  float64
	CodeBeCodeZaribKharidHaghighi float64
	//arze
	ArzeArzeshSaf uint64
	ArzePrvMin    int
	//range
	RangeJaheshMin            int     //minutes for calculate Jahesh and ArzeshMoamelatGhiasi
	RangeJaheshPercent        float32 //jahesh (Unit: percent)
	RangeGhiasiPowerFrom      int     //minutes for calculate GhiasiPower
	RangeGhiasiPower          float64 //GhiasiPower (RangeGhiasiPowerFrom)
	RangeArzeshMoamelat       uint64  // total arzesh moamelat
	RangeArzeshMoamelatGhiasi uint64  // arzesh moamelat ghiasi (RangeJaheshMin)

	//power box
	BoxGhodratmandMabnaRatio      float32
	BoxGhodratmandMinValue        uint64
	BoxGhodratmandPositivePercent float64
	BoxGhodratmandNegativePercent float64

	//SafeForoosh
	SafeForooshMin         int     //minutes for calculate SafeForoosh
	SafeForooshKaheshRatio float32 // 0.9 means 10% reduce
	SafeForooshHajmRatio   float32 //How much is the reduction of qo from vol
	SafeForooshMinimumQO   float64 // old qo1 must greater than SafeForooshMinimumQO
	SafeForooshAvalinBar   int8    // if 1: avalin bar(tmin=tmax) , if 0 har bar

	//Astane
	AstaneMin          int
	AstanePercentAbove int
	AstanePower        float32
	AstaneSaraneB      float32
	AstaneVol          uint64

	//SafeKharid
	SafeKharidMin int // it is SafeKharid but wasn't SafeKharidMin age

	//TaharokSafe
	TaharokSafeKharidMin             int     //for get old qd1
	TaharokSafeKharidIncreasePercent float64 //percent of increase from old qd1
	TaharokSafeKharidMinimumQD       float64 // old qd1 must greater than TaharokSafeKharidMinimumQD

	//vol
	VolMeanDays int

	//hot money
	HotMoneyValue uint64

	Width                  int
	Height                 int
	CodeBeCodeBlock        int
	EhtemalArzeBlock       int
	RangeMosbatBlock       int
	RangeManfiBlock        int
	BoxGhodratmandBlock    int
	TaharokBlock           int
	AstaneBlock            int
	SafeKharidBlock        int
	TaharokSafeKharidBlock int
}

func defaultSettings() Config {
	return Config{
		UpdateSpeed:           2,
		ShowAti:               false,
		ShowEkhtiarForoush:    false,
		ShowHaghTaghaddom:     true,
		ShowHousingFacilities: false,
		ShowOraghMosharekat:   false,
		ShowPayeFarabourse:    true,
		ShowSaham:             true,
		ShowSandoogh:          false,

		CodeBeCodePrvMin:              1,
		CodeBeCodeHoghooghiJump:       25,
		CodeBeCodeMinimumHoghooghiVol: 800e3,
		CodeBeCodeSaraneKharidGhiasi:  30e7,
		CodeBeCodeZaribKharidHaghighi: 0.3,

		ArzeArzeshSaf: 2e10,
		ArzePrvMin:    2,

		RangeGhiasiPowerFrom:      5,
		RangeGhiasiPower:          1,
		RangeArzeshMoamelat:       10e9,
		RangeArzeshMoamelatGhiasi: 7e9,
		RangeJaheshMin:            5,
		RangeJaheshPercent:        3.7,

		SafeForooshMin:         2,
		SafeForooshKaheshRatio: 0.90,
		SafeForooshHajmRatio:   0.02,
		SafeForooshMinimumQO:   100e3,
		SafeForooshAvalinBar:   1,

		AstaneMin:          2,
		AstanePercentAbove: 83,
		AstanePower:        1.2,
		AstaneSaraneB:      18,
		AstaneVol:          200e3,

		TaharokSafeKharidMin:             2,
		TaharokSafeKharidIncreasePercent: 10.0,
		TaharokSafeKharidMinimumQD:       100e3,

		BoxGhodratmandMabnaRatio:      50,
		BoxGhodratmandMinValue:        5000000000,
		BoxGhodratmandPositivePercent: 60,
		BoxGhodratmandNegativePercent: 60,

		SafeKharidMin: 2,
		VolMeanDays:   22,
		Width:         1024,
		Height:        768,

		CodeBeCodeBlock:        300, //5h
		EhtemalArzeBlock:       300,
		RangeMosbatBlock:       30,
		RangeManfiBlock:        30,
		BoxGhodratmandBlock:    30,
		TaharokBlock:           5,
		AstaneBlock:            5,
		SafeKharidBlock:        300,
		TaharokSafeKharidBlock: 5,
		HotMoneyValue:          10000000000,
	}
}

var loadedConfig Config
var loadedConfigMutex sync.RWMutex
var manualBlockList = make([]uint64, 0)
var manualBlockMutex sync.RWMutex
var trigger func(needInit int)

func getPath() string {
	//_, b, _, _ := runtime.Caller(0)
	//Root := filepath.Join(filepath.Dir(b), "../..")
	//fmt.Println(Root)
	//return Root + "/config.tse"
	return "./config.tse"
}
func LoadForce() {
	var config Config
	loadedConfigMutex.Lock()
	loadedConfig = config
	loadedConfigMutex.Unlock()
	tmp := LoadConfiguration()
	fmt.Println(tmp)
}

func LoadBlock() []uint64 {
	integerLines := make([]uint64, 0)
	b, err := ioutil.ReadFile("./BlockIds.txt")
	logger.GetFlog().Debug().Msg("loaded BlockIds from file")
	if err != nil {
		logger.GetFlog().Err(err).Send()
	} else {
		func() {
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					err := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(err).Send()
				}
			}()
			lines := strings.Split(string(b), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) < 5 {
					continue
				}
				integerLine := dt.ParseUint64(line)
				integerLines = append(integerLines, integerLine)
			}
		}()
	}
	manualBlockMutex.Lock()
	manualBlockList = integerLines
	manualBlockMutex.Unlock()
	func() {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
				logger.GetFlog().Err(err).Send()
			}
		}()
		trigger(100)
	}()
	return integerLines
}
func SetEventForAfterBlockListLoaded(_trigger func(needInit int)) {
	trigger = _trigger
}

func GetBlock() []uint64 {
	manualBlockMutex.RLock()
	integerLines := manualBlockList
	manualBlockMutex.RUnlock()
	if len(integerLines) == 0 {
		integerLines = LoadBlock()
	}

	return integerLines

}

func LoadConfiguration() Config {
	var config Config
	loadedConfigMutex.RLock()
	if loadedConfig != config {
		defer loadedConfigMutex.RUnlock()
		return loadedConfig
	}
	loadedConfigMutex.RUnlock()

	var file = getPath()
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configFile2, createErr := os.Create(file)
			configFile = configFile2
			if createErr != nil {
				fmt.Println(createErr)
				fmt.Println(createErr.Error())
			} else {
				jsonString, jsonErr := json.MarshalIndent(defaultSettings(), "", "\t")
				if jsonErr != nil {
					fmt.Println(jsonErr.Error())
				} else {
					_, err := configFile2.WriteString(string(jsonString))
					if err != nil {
						return defaultSettings()
					}
				}
			}

		} else {
			fmt.Println(err.Error())
		}

	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	loadedConfigMutex.Lock()
	defer loadedConfigMutex.Unlock()
	loadedConfig = config
	return config
}

func SaveConfiguration(config Config) error {
	logger.GetFlog().Debug().Msg("Saving")
	var file = getPath()
	configFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer func(configFile *os.File) {
		closeErr := configFile.Close()
		if closeErr != nil {
			fmt.Println(closeErr)
		}
	}(configFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		configFile, err = os.Create(file)
		if err != nil {
			logger.GetFlog().Err(err).Send()
			return err
		}
	}

	jsonString, jsonErr := json.MarshalIndent(config, "", "\t")
	if jsonErr != nil {
		logger.GetFlog().Err(jsonErr).Send()
		return jsonErr
	}

	_, writeErr := configFile.WriteString(string(jsonString))
	if writeErr != nil {
		logger.GetFlog().Err(writeErr).Send()
		return writeErr
	}

	return nil
}

func SaveLoadedConfiguration() error {
	loadedConfigMutex.RLock()
	defer loadedConfigMutex.RUnlock()
	return SaveConfiguration(loadedConfig)
}

func ChangeAndSaveConfig(name string, val interface{}) {
	ChangeLoadedConfig(name, val)
	SaveLoadedConfiguration()
}

func ChangeLoadedConfig(name string, val interface{}) {
	loadedConfigMutex.Lock()
	defer loadedConfigMutex.Unlock()
	switch val.(type) {
	case bool:
		struct_tools.SetBoolField(&loadedConfig, name, val.(bool))
		break
	case int:
		struct_tools.SetIntField(&loadedConfig, name, int64(val.(int)))
		break
	case int32:
		struct_tools.SetIntField(&loadedConfig, name, int64(val.(int32)))
		break
	case int64:
		struct_tools.SetIntField(&loadedConfig, name, int64(val.(int64)))
		break
	case uint:
		struct_tools.SetUintField(&loadedConfig, name, uint64(val.(uint)))
		break
	case uint32:
		struct_tools.SetUintField(&loadedConfig, name, uint64(val.(uint32)))
		break
	case uint64:
		struct_tools.SetUintField(&loadedConfig, name, uint64(val.(uint64)))
		break
	case string:
		struct_tools.SetStringField(&loadedConfig, name, val.(string))
		break
	}

}
