package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var reportIDCounter int

// Определение структуры для хранения отчета
type Report struct {
	Entries []*ReportEntry `json:"entries"`
}

type ReportEntry struct {
	ID           int    `json:"Id"`
	PID          *int   `json:"Pid"`
	OriginalURL  string `json:"OriginalURL,omitempty"`
	ShortURL     string `json:"ShortURL,omitempty"`
	SourceIP     string `json:"SourceIP"`
	TimeInterval string `json:"TimeInterval"`
	Count        int    `json:"Count"`
}

func shortenLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	originalLink := r.FormValue("url")
	if originalLink == "" {
		http.Error(w, "Пустая ссылка", http.StatusBadRequest)
		return
	}

	shortLink := generateShortLink()
	linkMap[shortLink] = originalLink

	// Сохранение данных в файл
	saveLinksToFile()

	// Отправка сокращенной ссылки в ответе
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://217.71.129.139:4712/redirect/%s", shortLink)))
}

func redirectLink(w http.ResponseWriter, r *http.Request) {
	shortLink := r.URL.Path[len("/redirect/"):]
	originalLink, ok := linkMap[shortLink]
	if !ok {
		http.Error(w, "Сокращенная ссылка не найдена", http.StatusNotFound)
		return
	}

	// Генерация отчета
	sourceIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		fmt.Printf("SourceIP: %s\n", sourceIP)
	}

	timeInterval := time.Now().Format("2006-01-02 15:04") // Форматирование времени в нужный формат
	count := 1

	// Создание отчета
	reportEntry := generateReportEntry(shortLink, originalLink, sourceIP, timeInterval, count)

	// Сохранение отчета в JSON файл
	saveReportToFile(reportEntry)

	// Перенаправление пользователя по оригинальной ссылке
	http.Redirect(w, r, originalLink, http.StatusFound)
}

func generateShortLink() string {
	shortLink := ""
	for i := 0; i < linkLength; i++ {
		shortLink += string(characters[rand.Intn(len(characters))])
	}
	return shortLink
}

func loadLinksFromFile() {
	// Загрузка данных из файла
	data, err := ioutil.ReadFile("shortened_urls.txt")
	if err != nil {
		return
	}
	linkMap = make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			linkMap[parts[0]] = parts[1]
		}
	}
}

func saveLinksToFile() {
	// Сохранение данных в файл
	file, err := os.Create("shortened_urls.txt")
	if err != nil {
		log.Println("Ошибка при сохранении данных в файл:", err)
		return
	}
	defer file.Close()

	for shortLink, originalLink := range linkMap {
		_, err := fmt.Fprintf(file, "%s %s\n", shortLink, originalLink)
		if err != nil {
			log.Println("Ошибка при записи в файл:", err)
		}
	}
}
