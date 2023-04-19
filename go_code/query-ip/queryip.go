package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

type Countrys []CountryElement

type CountryElement struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func GetCountryByIp(ipstr string) (string, error) {
	db, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ip := net.ParseIP(ipstr) // 81.2.69.142, 221.218.81.219

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	} // Or any appropriate struct

	err = db.Lookup(ip, &record)
	if err != nil {
		return "", err
	}
	return strings.ToLower(record.Country.ISOCode), nil
}

func ReadCsvData(fileName string) ([][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

func ReadJsonFromFile(fileName string) (map[string]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data Countrys

	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, err
	}

	returnData := map[string]string{}
	for _, v := range data {
		returnData[v.Code] = v.Name
	}

	return returnData, nil
}

func WriteDataToCsv(fileName string, data [][]string) error {
	f, err := os.Create(fileName)
	f.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM，防止中文乱码

	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteAll(data)
	w.Flush()

	return nil
}

func main() {

	countrys, _ := ReadJsonFromFile("countrys.json")

	records, _ := ReadCsvData("ip.csv")
	result := [][]string{}
	for _, record := range records {
		r := record
		ip := record[3]
		ip = strings.Trim(ip, "\"")
		ip = strings.Split(ip, ",")[0]
		country_code, _ := GetCountryByIp(ip)
		if val, ok := countrys[country_code]; ok {
			r = append(r, val)
		}
		result = append(result, r)
	}
	WriteDataToCsv("result.csv", result)

}
