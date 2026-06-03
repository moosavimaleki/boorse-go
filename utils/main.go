package main

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

func main() {

	start()

}

func get(url string) map[string]interface{} {
	var result map[string]interface{}

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
		func() {
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					parseError := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(parseError).Send()
					//return
				}
			}()
			err := json.Unmarshal(resp.Body(), &result)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	return result
}

func setLastId(id string, ltype string) {
	err := ioutil.WriteFile("./CodalDownloader"+ltype+".txt", []byte(id), 0644)
	if err != nil {
		fmt.Print(err)
	}
}

func getLastId(ltype string) string {
	b, err := ioutil.ReadFile("./CodalDownloader" + ltype + ".txt")
	if err != nil {
		fmt.Print(err)
		return "0"
	}
	str := string(b)
	return str
}

func makeId(item interface{}) string {
	mappedItem := item.(map[string]interface{})
	return fmt.Sprintf("%.0f", mappedItem["TracingNo"]) + mappedItem["PublishDateTime"].(string)
}

func start() {
	link := map[string]string{}
	link["P"] = "https://search.codal.ir/api/search/v2/q?&Audited=true&AuditorRef=-1&Category=7&Childs=true&CompanyState=-1&CompanyType=-1&Consolidatable=true&IsNotAudited=false&Length=-1&LetterType=55&Mains=true&NotAudited=true&NotConsolidatable=true&PageNumber=1&Publisher=false&TracingNo=-1&search=true"
	link["M"] = "https://search.codal.ir/api/search/v2/q?&Audited=true&AuditorRef=-1&Category=7&Childs=true&CompanyState=-1&CompanyType=-1&Consolidatable=true&IsNotAudited=false&Length=-1&LetterType=45&Mains=true&NotAudited=true&NotConsolidatable=true&PageNumber=1&Publisher=false&TracingNo=-1&search=true"

	for _, ltype := range [2]string{"P", "M"} {
		lastid := getLastId(ltype)
		fmt.Println(lastid)
		data := get(link[ltype])
		letters := (data["Letters"]).([]interface{})
		newItems := make([]interface{}, 20)
		for _, item := range letters {
			currentItemId := makeId(item)
			if lastid == currentItemId {
				break
			} else {
				//fmt.Println("--------------------------------------------------------------------")
				//fmt.Println(item)
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
				name := item["Symbol"]
				fmt.Println(name, percent, ltype)
				setLastId(currentItemId, ltype)
			}
		}
	}
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
	percent := int(((float32(intToAmount) / float32(intFromAmount)) - 1) * 100)
	fmt.Println(intFromAmount, intToAmount, percent)

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
		func() {
			defer func() {
				if panicInfo := recover(); panicInfo != nil {
					err := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
					logger.GetFlog().Err(err).Send()
					//return
				}
			}()
			result = string(resp.Body())
		}()
	}
	return result
}
