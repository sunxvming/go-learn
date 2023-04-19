package main

import (
	"fmt"
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

func main() {
	db, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ip := net.ParseIP("3.10.15.12") // 81.2.69.142, 221.218.81.219

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	} // Or any appropriate struct

	err = db.Lookup(ip, &record)
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(record.Country.ISOCode)
}
