package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var N int
var Host string
var GetUrl string
var SyncUrlTmpl string

type Page struct {
	Next_page next_page
	Data      Data
	Prev_page next_page
}

type Syncitem struct {
	Id      int
	Is_sync int
}

type Data []struct {
	Id, DateModified string
}

type next_page struct {
	Path, Uri, Offset string
}

func getUrl(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return []byte(body)
}

func getPage(url string, c chan Page) {
	var p Page
	var jsonPage = getUrl(url)

	jerr := json.Unmarshal(jsonPage, &p)
	if jerr != nil {
		fmt.Println("error:", jerr)
	}

	go getPage(p.Next_page.Uri, c)
	//go processPage(p, s)

	c <- p
}

func processPage(p Page, s chan string) {
	d := 2
	step := len(p.Data) / d
	for i := 0; i < d; i++ {
		tmpdata := p.Data[i*step : i*step+step]
		go processSubPage(tmpdata, s)
	}
}

func processSubPage(tmpdata Data, s chan string) {
	for _, v := range tmpdata {
		//s <- v.Id
		s <- syncItem(v.Id)
	}
}

func syncItem(id string) string {
	start := time.Now().String()

	url := fmt.Sprintf(SyncUrlTmpl, Host, id)

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return start + string(body) + time.Now().String()

	//var si Syncitem

	//jerr := json.Unmarshal([]byte(body), &si)
	//if jerr != nil {
	//fmt.Println("error:", jerr)
	//}
}

func showHelp() {
	fmt.Println("")
	fmt.Println("Pro turbo sync - prozorro sync system")
	fmt.Println("")
	fmt.Println("use:")
	fmt.Println("              pro-turbo-sync  cdb1  localhost    prod    100")
	fmt.Println("                                |         |        |       |")
	fmt.Println("                                |         |        |       |")
	fmt.Println("database (cdb1 or cdb2) --------|         |        |       |")
	fmt.Println("                                          |        |       |")
	fmt.Println("host -------------------------------------|        |       |")
	fmt.Println("                                                   |       |")
	fmt.Println("stage (prod or test) ------------------------------|       |")
	fmt.Println("                                                           |")
	fmt.Println("pages to sync (100 item per page) -------------------------|")
	os.Exit(1)
}

func main() {
	if len(os.Args) >= 4 {
		if os.Args[1] == "cdb1" {
			SyncUrlTmpl = "http://%s/prozorrosale/auctions/get-one-data?id=%s&json=1"
			if os.Args[3] == "prod" {
				GetUrl = "https://public.api.ea.openprocurement.org/api/2/auctions?descending=1&feed=changes"
			} else if os.Args[3] == "test" {
				GetUrl = "https://public.api-sandbox.ea.openprocurement.org/api/2/auctions?descending=1&feed=changes"
			} else {
				showHelp()
			}
		} else if os.Args[1] == "cdb2" {
			SyncUrlTmpl = "http://%s/prozorrosale2/auctions/get-one-data?id=%s&json=1"
			if os.Args[3] == "prod" {
				GetUrl = "https://public.api.ea2.openprocurement.net/api/2/auctions?descending=1&feed=changes"
			} else if os.Args[3] == "test" {
				GetUrl = "https://public.api-sandbox.ea2.openprocurement.net/api/2/auctions?descending=1&feed=changes"
			} else {
				showHelp()
			}
		} else {
			showHelp()
		}
		Host = os.Args[2]
		N, _ = strconv.Atoi(os.Args[4])
	} else {
		showHelp()
	}

	c := make(chan Page)
	s := make(chan string)

	go getPage(GetUrl, c)
	for i := 0; i < N; i++ {
		val, ok := <-c
		if ok == false {
			fmt.Println(val.Next_page.Uri, ok, "<--loop broke!")
			break
		} else {
			fmt.Println(val.Next_page.Uri, ok)
			go processPage(val, s)
		}
	}
	for i := 0; i < N*100; i++ {
		val, ok := <-s
		if ok == false {
			fmt.Println(i, val, ok, "<--loop broke!")
			break
		} else {
			fmt.Println(i, val)
		}
	}
}
