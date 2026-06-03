package tse_crawler

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"tsetmc/utils/logger"
)

func StartCodal() {
	link := map[string]string{}
	link["P"] = "https://search.codal.ir/api/search/v2/q?&Audited=true&AuditorRef=-1&Category=7&Childs=true&CompanyState=-1&CompanyType=-1&Consolidatable=true&IsNotAudited=false&Length=-1&LetterType=55&Mains=true&NotAudited=true&NotConsolidatable=true&PageNumber=1&Publisher=false&TracingNo=-1&search=true"
	link["M"] = "https://search.codal.ir/api/search/v2/q?&Audited=true&AuditorRef=-1&Category=7&Childs=true&CompanyState=-1&CompanyType=-1&Consolidatable=true&IsNotAudited=false&Length=-1&LetterType=45&Mains=true&NotAudited=true&NotConsolidatable=true&PageNumber=1&Publisher=false&TracingNo=-1&search=true"

	for _, ltype := range [2]string{"P", "M"} {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				err := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
				logger.GetFlog().Err(err).Send()
				return
			}
		}()
		lastid := getLastId(ltype)
		data := get(link[ltype])
		if data["Letters"] == nil {
			continue
		}
		letters := (data["Letters"]).([]interface{})
		newItems := make([]interface{}, 0, 20)
		for _, item := range letters {
			if item == nil {
				continue
			}
			currentItemId := makeId(item)
			if lastid == currentItemId {
				break
			} else {
				newItems = append(newItems, item)
			}
		}
		for i := len(newItems) - 1; i >= 0; i-- {
			if newItems[i] == nil {
				continue
			}
			item := newItems[i].(map[string]interface{})
			url := "https://codal.ir" + item["Url"].(string)
			request := requestInfo(url)
			if request != "" {
				currentItemId := makeId(item)
				percent := parseHtml(request, ltype)
				stock := Stock{
					L18:     item["Symbol"].(string),
					inscode: 0,
					Power:   0,
					SaraneB: 0,
					SaraneS: 0,
					PL:      0,
					PC:      0,
					TVOL:    0,
				}
				var sig = NewSignal(stock, "afzayeshSarmaye", 0)
				sig.SignalNameFa = "افزایش سرمایه"
				sig.appendInfo("percent", fmt.Sprintf("%d", percent))
				sig.appendInfo("type", ltype)
				SendSignal(sig)
				setLastId(currentItemId, ltype)
			}
		}
	}
}

func get(url string) map[string]interface{} {
	var result map[string]interface{}

	client := resty.New()
	client.SetTimeout(40 * time.Second)
	resp, requestErr := client.R().
		SetHeader("Accept", "application/json, text/plain, */*").
		SetHeader("Referer", "https://codal.ir/").
		SetHeader("Accept-Encoding", "gzip, deflate, br").
		SetHeader("Accept-Language", "en-US,en;q=0.9,fa;q=0.8").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
		Get(url)
	if requestErr == nil && resp.StatusCode() == 200 {
		err := json.Unmarshal(resp.Body(), &result)
		if err != nil {
			logger.GetFlog().Err(err).Send()
		}
	}
	return result
}

func setLastId(id string, ltype string) {
	err := ioutil.WriteFile("./CodalDownloader"+ltype+".txt", []byte(id), 0644)
	if err != nil {
		logger.GetFlog().Err(err).Send()
	}
}

func getLastId(ltype string) string {
	b, err := ioutil.ReadFile("./CodalDownloader" + ltype + ".txt")
	if err != nil {
		logger.GetFlog().Err(err).Send()
		return "0"
	}
	str := string(b)
	return str
}

func makeId(item interface{}) string {
	mappedItem := item.(map[string]interface{})
	return fmt.Sprintf("%.0f", mappedItem["TracingNo"]) + mappedItem["PublishDateTime"].(string)
}

func parseHtml(request string, ltype string) int {
	r := strings.NewReader(request)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		logger.GetFlog().Err(err).Send()
	}

	toId := ""
	fromId := ""
	if ltype == "P" {
		toId = "lblToAmount"
		fromId = "lblFromAmount"
	} else {
		toId = "txbCapitalIncreaseAmountAccept"
		fromId = "ucCapitalIncreaseLicense_lblCompanyCapital"
	}

	stringFromAmount := doc.Find("span#" + fromId).Text()
	stringFromAmount = strings.Replace(stringFromAmount, ",", "", -1)
	intFromAmount, _ := strconv.Atoi(stringFromAmount)

	stringToAmount := doc.Find("span#" + toId).Text()
	stringToAmount = strings.Replace(stringToAmount, ",", "", -1)
	intToAmount, _ := strconv.Atoi(stringToAmount)
	var percent int
	if ltype == "P" {
		percent = int(((float32(intToAmount) / float32(intFromAmount)) - 1) * 100)
	} else {
		percent = int(((float32(intToAmount+intFromAmount) / float32(intFromAmount)) - 1) * 100)
	}
	logger.GetFlog().Debug().Int("afzayesh sarmaye", percent)

	return percent
}

func requestInfo(url string) string {
	result := ""

	client := resty.New()
	client.SetTimeout(40 * time.Second)
	resp, requestErr := client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Referer", "https://codal.ir/").
		SetHeader("Accept-Encoding", "gzip, deflate, br").
		SetHeader("Accept-Language", "en-US,en;q=0.9,fa;q=0.8").
		SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36").
		Get(url)
	if requestErr == nil && resp.StatusCode() == 200 {
		result = string(resp.Body())
	}
	return result
}
