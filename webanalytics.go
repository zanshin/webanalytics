package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zanshin/webanalytics/dbconf"
	"github.com/zanshin/webanalytics/metrics"
	"github.com/zanshin/webanalytics/persist"
)

// Config contains the main configuration settings for the application.
type Config struct {
	BatchInsertSeconds int             `json:"batchInsertSeconds"`
	Port               int             `json:"port"`
	DbConfig           dbconf.DbConfig `json:"database"`
}

var configFilePath string
var hrefClicks []metrics.HrefClick
var pageViews []metrics.PageView

func listenForRecords(db *sql.DB, seconds time.Duration) {
	// Run every x seconds.
	for range time.Tick(seconds) {
		// Handle page views.
		newPageViews := make([]metrics.PageView, len(pageViews))
		copy(newPageViews, pageViews)
		go persist.SetPageViews(db, newPageViews)
		pageViews = pageViews[0:0]

		// Handle href clicks.
		newHrefClicks := make([]metrics.HrefClick, len(hrefClicks))
		copy(newHrefClicks, hrefClicks)
		go persist.SetHrefClicks(db, newHrefClicks)
		hrefClicks = hrefClicks[0:0]
	}
}

func IPAddress(remoteAddr string) string {
	arr := strings.Split(remoteAddr, ":")
	return arr[0]
}

func hrefClickHandler(w http.ResponseWriter, r *http.Request, body []byte) {
	hrefClick := metrics.HrefClick{}
	if err := json.Unmarshal(body, &hrefClick); err != nil {
		log.Println("Unable to unmarshal hrefClick: ", err)
	}
	// Get ip address from http request
	hrefClick.IPAddress = IPAddress(r.RemoteAddr)
	hrefClicks = append(hrefClicks, hrefClick)
	w.WriteHeader(201)
}

func pageViewsHandler(w http.ResponseWriter, r *http.Request, body []byte) {
	pageView := metrics.PageView{}
	if err := json.Unmarshal(body, &pageView); err != nil {
		log.Println("Unable to unmarshal pageView: ", err)
	}
	// Get ip address from http request
	pageView.IPAddress = IPAddress(r.RemoteAddr)
	pageViews = append(pageViews, pageView)
	w.WriteHeader(201)
}

func readConfig(configFilePath string) Config {
	config := Config{}
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal("Unable to read config file: ", err)
	}
	if err = json.Unmarshal(configFile, &config); err != nil {
		log.Fatal("Unable to unmarshal configFile into config: ", err)
	}
	if err = validateConfig(config); err != nil {
		log.Fatal("Unable to validate config: ", err)
	}
	return config
}

func validateConfig(config Config) error {
	if config.BatchInsertSeconds < 1 {
		return fmt.Errorf(
			"BatchInsertSeconds cannot be less than 1, %d was given.",
			config.BatchInsertSeconds,
		)
	}
	if config.Port < 0 || config.Port > 65535 {
		return fmt.Errorf(
			"Port must be between 0 and 65535, %d was given.",
			config.Port,
		)
	}
	if config.DbConfig.Port < 0 || config.DbConfig.Port > 65535 {
		return fmt.Errorf(
			"The database port must be between 0 and 65535, %d was given.",
			config.DbConfig.Port,
		)
	}
	return nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, []byte)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "x-requested-with, x-requested-by, Content-Type")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Unable to read requeset body: ", err)
		}
		fn(w, r, body)
	}
}

func init() {
	goPath := os.Getenv("GOPATH")
	defaultConfigPath := fmt.Sprintf("%s/src/github.com/zanshin/webanalytics/config.json", goPath)
	flag.StringVar(&configFilePath, "config", defaultConfigPath, "path to config.json")
}

func main() {
	// Read the config, initialize the database and listen for records.
	flag.Parse()
	config := readConfig(configFilePath)
	db := persist.Db(config.DbConfig)
	seconds := time.Duration(config.BatchInsertSeconds) * time.Second

	fmt.Println("about to listenForRecords")
	go listenForRecords(db, seconds)

	// Create the handlers for page-view/ and href-click/ POSTs
	fmt.Println("how about some HandleFuncs")
	http.HandleFunc("/page-views/", makeHandler(pageViewsHandler))
	http.HandleFunc("/href-click/", makeHandler(hrefClickHandler))

	fmt.Println("ListenAndServe")
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
