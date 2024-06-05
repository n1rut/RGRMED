package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type CustomTime struct {
	time.Time
}

const ctLayout = "2006-01-02 15:04"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(`"`+ctLayout+`"`, s)
	return
}

type Report struct {
	ID           int        `json:"Id"`
	PID          *int       `json:"Pid"`
	OriginalURL  string     `json:"OriginalURL"`
	ShortURL     string     `json:"ShortURL"`
	SourceIP     string     `json:"SourceIP"`
	TimeInterval CustomTime `json:"TimeInterval"`
	Count        int        `json:"Count"`
}

type Reports struct {
	Entries []Report `json:"entries"`
}

var (
	reportsMutex sync.RWMutex
	reportsData  Reports
)

func main() {
	go fetchAndUpdateData()
	go startTCPServer()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/report", reportHandler)
	log.Println("Starting web server on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func startTCPServer() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer ln.Close()
	log.Println("TCP server listening on port 9090...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("Error reading data: %v", err)
		return
	}

	// Log the raw data received
	log.Printf("Raw data received: %s", data)

	var reports Reports
	err = json.Unmarshal(data, &reports)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return
	}

	reportsMutex.Lock()
	reportsData = reports
	reportsMutex.Unlock()

	log.Println("Data successfully updated from main server.")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("short_url")
	if shortURL == "" || len(shortURL) != 6 {
		http.Error(w, "Invalid short URL", http.StatusBadRequest)
		return
	}

	log.Printf("Received request for short URL: %s", shortURL)

	reportsMutex.RLock()
	filteredReports := filterReports(reportsData.Entries, shortURL)
	reportsMutex.RUnlock()

	if len(filteredReports) == 0 {
		http.Error(w, "No data for the provided short URL", http.StatusNotFound)
		return
	}

	// Преобразование дат в строковый формат
	for i := range filteredReports {
		formattedTime, err := time.Parse(ctLayout, filteredReports[i].TimeInterval.Time.Format(ctLayout))
		if err != nil {
			log.Printf("Error parsing time: %v", err)
			continue
		}
		filteredReports[i].TimeInterval = CustomTime{formattedTime}
	}

	jsonData, err := json.Marshal(filteredReports)
	if err != nil {
		http.Error(w, "Error marshaling JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	log.Printf("Sent data: %s", jsonData)
}

func filterReports(reports []Report, shortURL string) []Report {
	var filtered []Report
	for _, report := range reports {
		if report.ShortURL == shortURL {
			filtered = append(filtered, report)
		}
	}
	return filtered
}

func fetchAndUpdateData() {
	for {
		log.Println("Attempting to fetch data...")

		conn, err := net.Dial("tcp", "172.17.2.214:6379")
		if err != nil {
			log.Printf("Error dialing TCP server: %v", err)
			time.Sleep(10 * time.Second) // Если произошла ошибка, подождем 10 секунд перед повторной попыткой
			continue
		}

		_, err = conn.Write([]byte("SENDJSON\n"))
		if err != nil {
			log.Printf("Error sending request: %v", err)
			conn.Close()
			time.Sleep(10 * time.Second) // Если произошла ошибка, подождем 10 секунд перед повторной попыткой
			continue
		}

		data, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			conn.Close()
			time.Sleep(10 * time.Second) // Если произошла ошибка, подождем 10 секунд перед повторной попыткой
			continue
		}

		var reports Reports
		err = json.Unmarshal(data, &reports)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			conn.Close()
			time.Sleep(10 * time.Second) // Если произошла ошибка, подождем 10 секунд перед повторной попыткой
			continue
		}

		reportsMutex.Lock()
		reportsData = reports
		reportsMutex.Unlock()

		log.Println("Data successfully updated from main server.")
		conn.Close() // Закрываем соединение после успешного чтения данных и обновления

		// Обратный отсчет до следующего запроса
		for i := 60; i > 0; i-- {
			log.Printf("Next data fetch in %d seconds...", i)
			time.Sleep(1 * time.Second)
		}
	}
}
