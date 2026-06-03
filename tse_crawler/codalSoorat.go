package tse_crawler

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"tsetmc/utils/jalali"
	"tsetmc/utils/logger"
	"tsetmc/utils/mytime"
)

func makeInstLink(id int, len int) string {
	strId := strconv.Itoa(id)
	strLen := strconv.Itoa(len)
	return "https://search.codal.ir/api/search/v2/q?=&Audited=false&AuditorRef=-1&Category=1&Childs=true&CompanyState=-1&CompanyType=1&Consolidatable=true&IsNotAudited=false&Isic=" + strId + "&Length=" + strLen + "&LetterType=6&Mains=true&NotAudited=true&NotConsolidatable=false&PageNumber=1&Publisher=false&TracingNo=-1&search=true"
	return "https://search.codal.ir/api/search/v2/q? &Audited=false&AuditorRef=-1&Category=1&Childs=true&CompanyState=0&CompanyType= 1&Consolidatable=true&IsNotAudited=false&Isic=" + strId + "&Length=" + strLen + "&LetterType=-1&Mains=true&NotAudited=true&NotConsolidatable=true&PageNumber=1&Publisher=false&TracingNo=-1&search=true"
}
func replaceToEnDigit(text string) string {
	text = strings.ReplaceAll(text, "۰", "0")
	text = strings.ReplaceAll(text, "۱", "1")
	text = strings.ReplaceAll(text, "۲", "2")
	text = strings.ReplaceAll(text, "۳", "3")
	text = strings.ReplaceAll(text, "۴", "4")
	text = strings.ReplaceAll(text, "۵", "5")
	text = strings.ReplaceAll(text, "۶", "6")
	text = strings.ReplaceAll(text, "۷", "7")
	text = strings.ReplaceAll(text, "۸", "8")
	text = strings.ReplaceAll(text, "۹", "9")
	return text
}
func replaceToPersianDigit(text string) string {
	text = strings.ReplaceAll(text, "0", "۰")
	text = strings.ReplaceAll(text, "1", "۱")
	text = strings.ReplaceAll(text, "2", "۲")
	text = strings.ReplaceAll(text, "3", "۳")
	text = strings.ReplaceAll(text, "4", "۴")
	text = strings.ReplaceAll(text, "5", "۵")
	text = strings.ReplaceAll(text, "6", "۶")
	text = strings.ReplaceAll(text, "7", "۷")
	text = strings.ReplaceAll(text, "8", "۸")
	text = strings.ReplaceAll(text, "9", "۹")
	return text
}
func replaceArToPer(text string) string {
	text = strings.ReplaceAll(text, "ي", "ی")
	text = strings.ReplaceAll(text, "ك", "ک")
	text = strings.ReplaceAll(text, "ئ", "ی")
	text = strings.ReplaceAll(text, "ء", "")
	text = strings.ReplaceAll(text, "ٔ", "")
	text = strings.ReplaceAll(text, "ة", "ه")
	text = strings.ReplaceAll(text, "آ", "ا")
	text = strings.ReplaceAll(text, "أ", "ا")
	text = strings.ReplaceAll(text, "إ", "ا")
	text = strings.ReplaceAll(text, "ئ", "ی")
	text = strings.ReplaceAll(text, "ؤ", "و")
	return text
}
func replacePerToAr(text string) string {
	text = strings.ReplaceAll(text, "ی", "ي")
	text = strings.ReplaceAll(text, "ک", "ك")
	return text
}

func StartCodalSoorat() {
	firstLink := "https://search.codal.ir/api/search/v2/q?&Audited=false&AuditorRef=-1&Category=1&Childs=false&CompanyState=-1&CompanyType=1&Consolidatable=true&IsNotAudited=false&Length=-1&LetterType=6&Mains=true&NotAudited=true&NotConsolidatable=false&PageNumber=1&Publisher=false&TracingNo=-1&search=true"

	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			err := fmt.Errorf("%v, %s", panicInfo, string(debug.Stack()))
			fmt.Println(err)
			logger.GetFlog().Err(err).Send()
			return
		}
	}()
	lastid := getLastId("S")
	data := get(firstLink)
	if data["Letters"] == nil {
		return
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
		fmt.Println("-------------------------------")
		item := newItems[i].(map[string]interface{})
		if !strings.Contains(item["Url"].(string), "Decision.aspx") {
			continue
		}
		OriginalName := item["Symbol"].(string)
		nameAr := replacePerToAr(OriginalName)
		namePer := replaceArToPer(OriginalName)
		title := item["Title"].(string)
		fmt.Println(title)
		fmt.Println(namePer)

		codalUrl := "https://codal.ir" + item["Url"].(string) + "&sheetId=1"
		request := requestInfo(codalUrl)
		if request != "" {
			currentItemId := makeId(item)
			ok, idInt, newSood, newPriod, prvJy, prvJm, newDateStr := parseHtmlSoorat(request)
			if !ok {
				continue
			}
			fmt.Println("Sood Found ---------------------------")
			mMapper := map[int]int{
				3:  12,
				6:  3,
				9:  6,
				12: 9,
			}
			strOldJy := strconv.Itoa(prvJy)
			strOldJyPer := replaceToPersianDigit(strOldJy)
			strOldJm := fmt.Sprintf("%02d", prvJm)
			strOldJmPer := replaceToPersianDigit(strOldJm)
			oldPeriod := mMapper[newPriod]
			fmt.Println("finding ------", oldPeriod, strOldJyPer+"/"+strOldJmPer)

			instLink := makeInstLink(idInt, oldPeriod)
			fmt.Println(instLink)
			instData := get(instLink)
			if instData["Letters"] == nil {
				logger.GetFlog().Error().Msg(fmt.Sprintf("cant find prv codal %v,%d,%d,%d,%d,%d", ok, idInt, newSood, newPriod, prvJy, prvJm))
				continue
			}
			isntLetters := (instData["Letters"]).([]interface{})
			for _, tmpItem := range isntLetters {
				if tmpItem == nil {
					continue
				}
				instItem := tmpItem.(map[string]interface{})
				if !strings.Contains(instItem["Url"].(string), "Decision.aspx") {
					continue
				}
				oldTitle := instItem["Title"].(string)
				fmt.Println(oldTitle, strings.Contains(oldTitle, strOldJyPer+"/"+strOldJmPer), strOldJyPer+"/"+strOldJmPer)
				if strings.Contains(oldTitle, strOldJyPer+"/"+strOldJmPer) {
					fmt.Println("---------------FOUNDED----------")
					oldUrl := "https://codal.ir" + instItem["Url"].(string) + "&sheetId=1"
					fmt.Println(oldUrl)
					oldRequest := requestInfo(oldUrl)
					if oldRequest != "" {
						fmt.Println("old http ok")
						ok, idInt, oldSood, oldPeriod2, _, _, oldDateStr := parseHtmlSoorat(oldRequest)
						fmt.Println("parse ok", ok, idInt, oldSood, oldPeriod2)
						if ok {
							fmt.Println("idInt", idInt)
							fmt.Println("oldPeriod", oldPeriod2)
							fmt.Println("oldSood", oldSood)
							fmt.Println("newSood", newSood)
							fmt.Println("new", newDateStr+" "+fmt.Sprintf("%d", newPriod))
							fmt.Println("old", oldDateStr+" "+fmt.Sprintf("%d", oldPeriod2))
							regularSood := float32(newPriod) * float32(oldSood) / float32(oldPeriod)
							fmt.Println("regularSood", regularSood)
							diff := float32(newSood) - regularSood
							diffP := ((float32(newSood) / regularSood) - 1) * 100

							tmp := mw.getTmpAllRow()
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
							fmt.Println("searching")
							for _, _stock := range tmp {
								if replaceArToPer(_stock.L18) == namePer {
									stock = _stock
									break
								}
							}
							fmt.Println("finish searching")

							if stock.inscode == 0 {
								infoLink := "https://codal.ir/Company.aspx?Symbol=" + url.QueryEscape(OriginalName)
								infoRequest := requestInfo(infoLink)
								if infoRequest != "" {
									isin := parseHtmlInfo(infoRequest)
									if len(isin) == 12 {
										for _, _stock := range tmp {
											if _stock.iid == isin {
												stock = _stock
												break
											}
										}
									} else {
										infoLink := "https://codal.ir/Company.aspx?Symbol=" + url.QueryEscape(nameAr)
										infoRequest := requestInfo(infoLink)
										if infoRequest != "" {
											isin := parseHtmlInfo(infoRequest)
											if len(isin) == 12 {
												for _, _stock := range tmp {
													if _stock.iid == isin {
														stock = _stock
														break
													}
												}
											}
										}
									}
								}
							}

							var sig = NewSignal(stock, "sood", 0)
							sig.SignalNameFa = "افزایش سرمایه"
							if stock.PY != 0 {
								peOld := float32(stock.PY) / float32(oldSood)
								peNew := float32(stock.PY) / float32(newSood)
								sig.appendInfo("P/E new", fmt.Sprintf("%.1f", peNew))
								sig.appendInfo("P/E old", fmt.Sprintf("%.1f", peOld))
							}
							sig.appendInfo("new", newDateStr+" "+fmt.Sprintf("%d", newPriod))
							sig.appendInfo("old", oldDateStr+" "+fmt.Sprintf("%d", oldPeriod2))
							sig.appendInfo("oldSood", fmt.Sprintf("%d", oldSood))
							sig.appendInfo("newSood", fmt.Sprintf("%d", newSood))
							sig.appendInfo("regularSood", fmt.Sprintf("%0.f", regularSood))
							sig.appendInfo("diff", fmt.Sprintf("%.0f", diff))
							sig.appendInfo("diff%", fmt.Sprintf("%.0f", diffP))
							SendSignal(sig)

						}
						break
					}

				}

			}

			setLastId(currentItemId, "S")
		}
		fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	}

}

func removeNonDigit(text string) string {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	processedString := reg.ReplaceAllString(text, "")
	return processedString
}

func parseHtmlInfo(request string) string {
	r := strings.NewReader(request)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		logger.GetFlog().Err(err).Send()
		return ""
	}

	isin := doc.Find("span#txbISIN").Text()
	return strings.TrimSpace(isin)
}

func parseHtmlSoorat(request string) (bool, int, int, int, int, int, string) {
	r := strings.NewReader(request)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		logger.GetFlog().Err(err).Send()
	}

	sood := getSood(request)
	if sood == 0 {
		return false, 0, 0, 0, 0, 0, ""
	}

	priod := doc.Find("span#ctl00_lblPeriod").Text()
	priod = removeNonDigit(priod)
	priodInt, _ := strconv.Atoi(priod)

	dateStr := doc.Find("span#ctl00_lblPeriodEndToDate bdo").Text()
	date := strings.Split(dateStr, "/")
	id := doc.Find("span#ctl00_lblISIC").Text()
	id = removeNonDigit(id)
	idInt, _ := strconv.Atoi(id)

	jy, _ := strconv.Atoi(date[0])
	jm, _ := strconv.Atoi(date[1])
	jd, _ := strconv.Atoi(date[2])
	my, mm, md := jalali.Jalali_to_gregorian(jy, jm, jd)
	endDate := time.Date(my, time.Month(mm), md, 0, 0, 0, 0, mytime.GetIran())
	endDateUnix := endDate.Unix()
	prvDateUnix := endDateUnix - 100*24*3600
	prvDate := time.Unix(prvDateUnix, 0)
	prvJy, prvJm, _ := jalali.Gregorian_to_jalali(prvDate.Year(), int(prvDate.Month()), prvDate.Day())

	return true, idInt, sood, priodInt, prvJy, prvJm, dateStr
}

func getSood(request string) int {
	//sood_xpath := "//td/span[contains(text(),'سود') and contains(text(),'زیان') and contains(text(),'سهم') and contains(text(),'خالص') ]/parent::td/following-sibling::td[1]/span/text()"
	var re = regexp.MustCompile(`(?m)var[\s]+datasource[\s]+=[\s]+({[^\n]+);`)
	allString := re.FindStringSubmatch(request)
	if allString == nil || len(allString) <= 1 {
		return 0
	}
	datasourceString := strings.TrimSpace(allString[1])
	var datasource map[string]interface{}
	err := json.Unmarshal([]byte(datasourceString), &datasource)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	if datasource["sheets"] == nil {
		return 0
	}
	sheets := datasource["sheets"].([]interface{})
	for _, sheet := range sheets {
		sheetT := sheet.(map[string]interface{})
		if sheetT["title_En"] != nil && sheetT["title_En"].(string) == "Income Statement" {
			tables := sheetT["tables"].([]interface{})
			for _, table := range tables {
				tableT := table.(map[string]interface{})
				if tableT["title_En"] != nil && tableT["title_En"].(string) == "Income Statement" {
					cells := tableT["cells"].([]interface{})
					var address string = ""
					for _, cell := range cells {
						cellT := cell.(map[string]interface{})
						if cellT["value"] == nil || cellT["address"] == nil {
							continue
						}
						tmpValue := cellT["value"].(string)
						if strings.Contains(tmpValue, "سود") && strings.Contains(tmpValue, "زیان") &&
							strings.Contains(tmpValue, "خالص") && strings.Contains(tmpValue, "سهم") {
							address = strings.Replace(cellT["address"].(string), "A", "B", 1)
						}
					}
					for _, cell := range cells {
						cellT := cell.(map[string]interface{})
						if cellT["address"].(string) == address {
							tmp, _ := strconv.Atoi(cellT["value"].(string))
							return tmp
						}
					}
				}
			}
		}
	}
	return 0
}
