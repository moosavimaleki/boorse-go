package debugging

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetClientFile() []byte {
	min := GetMinStep()
	sec := time.Now().Second()

	files, err := ioutil.ReadDir("./all_data/")
	if err != nil {
		fmt.Println(err)
	}
	max := -1
	fileName := ""
	for _, file := range files {
		name := strings.Split(file.Name(), "_")
		if len(name) == 3 && name[0] == "c" {
			if name[1] == strconv.Itoa(min) {
				fileSec, _ := strconv.Atoi(strings.Replace(name[2], ".txt", "", 1))
				if fileSec == sec {
					fileName = file.Name()
					break
				}
				if fileSec < sec && fileSec > max {
					max = fileSec
				}
			}
		}
	}
	if fileName == "" && max == -1 {
		max = -1
		back := 1
	getMax:
		for _, file := range files {
			name := strings.Split(file.Name(), "_")
			if len(name) == 3 && name[0] == "c" {
				if name[1] == strconv.Itoa(min-back) {
					fileSec, _ := strconv.Atoi(strings.Replace(name[2], ".txt", "", 1))
					if fileSec > max {
						max = fileSec
					}
				}
			}
		}
		if max == -1 {
			back++
			goto getMax
		}
		fileName = "c_" + strconv.Itoa(min-back) + "_" + strconv.Itoa(max) + ".txt"
	} else if fileName == "" {
		fileName = "c_" + strconv.Itoa(min) + "_" + strconv.Itoa(max) + ".txt"
	}

	filePath := fmt.Sprintf("./all_data/%s", fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []byte("")
	}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Print(err)
		return []byte("")
	}

	return b
}

func GetMarketFile() []byte {
	min := GetMinStep()
	sec := time.Now().Second()
	filePath := fmt.Sprintf("./all_data/%d_%d_.txt", min, sec)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = fmt.Sprintf("./all_data/%d_%d_i.txt", min, sec)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return []byte("")
			//filePath = fmt.Sprintf("./all_data/%d_%d_.txt", min, sec-1)
			//if _, err := os.Stat(filePath); os.IsNotExist(err) {
			//	filePath = fmt.Sprintf("./all_data/%d_%d_i.txt", min, sec-1)
			//	if _, err := os.Stat(filePath); os.IsNotExist(err) {
			//		fmt.Println("not loaded" , min, sec)
			//		return []byte("")
			//	}
			//}
		}
	}

	fmt.Println("loaded", filePath)

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Print(err)
		return []byte("")
	}

	return b
}

func GetFirstMarketFile() []byte {
	min := GetMinStep()

	files, err := ioutil.ReadDir("./all_data")
	if err != nil {
		fmt.Print(err)
		return []byte("")
	}
	maxMin := -1
	maxSec := -1
	maxString := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "_i") {
			splitName := strings.Split(f.Name(), "_")
			fileMin, _ := strconv.Atoi(splitName[0])
			fileSec, _ := strconv.Atoi(splitName[1])
			if fileMin <= min && maxMin <= fileMin {
				if maxMin < fileMin || (maxMin == fileMin && maxSec <= fileSec) {
					maxString = f.Name()
					maxMin = fileMin
					maxSec = fileSec
				}
			}
		}
	}

	b, err := ioutil.ReadFile("./all_data/" + maxString)
	fmt.Println("GetFirstMarketFile", "./all_data/"+maxString)
	if err != nil {
		fmt.Print(err)
		return []byte("")
	}
	return b

	return []byte("")
}
