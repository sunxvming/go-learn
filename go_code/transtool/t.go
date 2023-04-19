package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "string list"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type BiTableRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Total int `json:"total"`
		Items []struct {
			ID       string            `json:"id"`
			RecordID string            `json:"record_id"`
			Fields   map[string]string `json:"fields"`
		} `json:"items"`
		PageToken string `json:"page_token"`
		HasMore   bool   `json:"has_more"`
	} `json:"data"`
}

func doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("http response must be 200")
	}
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func checkUpdate() {
	err := doUpdate("https://tools.minervagame.com/downloads/transTool/transTool")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("done 1", runtime.GOOS, runtime.GOARCH)
}

type TranItems []map[string]string
type KeyFunc func(x string)

func downTable(pageToken, docid, tid string) TranItems {
	url := fmt.Sprintf("https://tools.minervagame.com/fei/open-apis/bitable/v1/apps/%s/tables/%s/records", docid, tid)
	url += "?page_token=" + pageToken
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	tbres := BiTableRes{}
	err = json.Unmarshal(s, &tbres)
	if err != nil {
		log.Fatal(err)
	}
	rs := TranItems{}
	for _, e := range tbres.Data.Items {
		rs = append(rs, e.Fields)
	}
	if len(rs) == 0 {
		return rs
	}
	rs2 := downTable(tbres.Data.PageToken, docid, tid)
	log.Println("count:", len(rs2)+len(rs), tbres.Data.Total)
	rs = append(rs, rs2...)
	return rs
}

func Keys(rs TranItems, fun KeyFunc) {
	for k := range rs[0] {
		if k == "key" {
			continue
		}
		fun(k)
	}
}

func CheckFolder(name string) {
	err := os.RemoveAll(name)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(name, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Empty(lang string, e map[string]string) string {
	if e[lang] != "" {
		return e[lang]
	}
	return fmt.Sprintf("%s: %s", lang, e["key"])
}

func Backend(rs TranItems) {
	writefile := func(k string) {
		p := fmt.Sprintf("%s/%s.php", "BackEnd", k)
		f, err := os.Create(p)
		if err != nil {
			log.Fatal(err)
		}
		w := bufio.NewWriter(f)
		w.WriteString(fmt.Sprintf("<?php // Auto Generated %s\n", time.Now().Format("2006-01-02 15:04:05Z07:00")))
		w.WriteString("return [\n")
		for _, e := range rs {
			w.WriteString(fmt.Sprintf("%#v => %#v,\n", e["key"], Empty(k, e)))
		}
		w.WriteString("];\n")
		w.Flush()
		f.Close()
	}
	CheckFolder("BackEnd")
	Keys(rs, writefile)
}

func getName(names TranItems, ln, platform string) string {
	for _, n := range names {
		if n["lang"] == ln && n[platform] != "" {
			return n[platform]
		}
	}
	return ln
}

func IOS(rs, names TranItems) {
	writefile := func(k string) {
		ln := getName(names, k, "IOS")
		CheckFolder(fmt.Sprintf("IOS/%s.lproj", ln))
		p := fmt.Sprintf("IOS/%s.lproj/Localizable.strings", ln)
		f, err := os.Create(p)
		if err != nil {
			log.Fatal(err)
		}
		w := bufio.NewWriter(f)
		w.WriteString(fmt.Sprintf("// %s Auto Generated %s\n", ln, time.Now().Format("2006-01-02 15:04:05Z07:00")))

		for _, e := range rs {
			w.WriteString(fmt.Sprintf("%#v = %#v;\n", e["key"], Empty(k, e)))
		}

		w.Flush()
		f.Close()
	}
	CheckFolder("IOS")
	Keys(rs, writefile)
}

func Android(rs, names TranItems) {
	writefile := func(k string) {
		ln := getName(names, k, "Android")
		CheckFolder(fmt.Sprintf("Android/values-%s", ln))
		p := fmt.Sprintf("Android/values-%s/strings.xml", ln)
		f, err := os.Create(p)
		if err != nil {
			log.Fatal(err)
		}
		w := bufio.NewWriter(f)
		w.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
		w.WriteString(fmt.Sprintf("<!-- %s Auto Generated %s -->\n", ln, time.Now().Format("2006-01-02 15:04:05Z07:00")))
		w.WriteString("<resources>\n")
		for _, e := range rs {
			w.WriteString(fmt.Sprintf("\t<string name=%#v><![CDATA[%v]]></string>\n", e["key"], Empty(k, e)))
		}
		w.WriteString("</resources>")
		w.Flush()
		f.Close()
	}
	CheckFolder("Android")
	Keys(rs, writefile)
}

func contains[T comparable](s []T, str T) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

var act = flag.String("act", "", "act")
var filename = flag.String("f", "", "filename")

var prefixs arrayFlags
var outtypes arrayFlags

func main() {

	flag.Var(&prefixs, "p", "filter text by prefix")
	flag.Var(&outtypes, "o", "output type, in: IOS Android BackEnd")
	flag.Parse()
	if len(outtypes) == 0 {
		outtypes = []string{"IOS", "Android", "BackEnd"}
	}
	log.Println(prefixs, outtypes)
	log.SetOutput(os.Stdout)
	rs := downTable("", "bascnUKPtfk7lMS07rneuDMLIwc", "tblkM8TT4hXlZULW")
	names := downTable("", "bascnUKPtfk7lMS07rneuDMLIwc", "tblflTu1ZlAv8DLb")
	if *act == "back" && *filename != "" {
		s := [2]TranItems{names, rs}
		bs, err := json.Marshal(s)
		if err != nil {
			log.SetOutput(os.Stdout)
			log.Fatal(err)
		}
		var out bytes.Buffer
		json.Indent(&out, bs, "", "\t")
		ioutil.WriteFile(*filename, out.Bytes(), 0644)
		return
	}
	if len(prefixs) > 0 {
		rs2 := TranItems{}
		for _, o := range rs {
			for _, p := range prefixs {
				if strings.HasPrefix(o["key"], p) {
					rs2 = append(rs2, o)
				}
			}
		}
		rs = rs2
	}
	log.Println("total:", len(rs))
	if len(rs) == 0 {
		log.Fatal("error: no record")
	}
	if contains(outtypes, "BackEnd") {
		Backend(rs)
	}
	if contains(outtypes, "IOS") {
		IOS(rs, names)
	}
	if contains(outtypes, "Android") {
		Android(rs, names)
	}
}
