package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
)

func sendReportsJSONToStatisticService() {
	// Устанавливаем соединение с сервисом статистики
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		log.Printf("Ошибка при установлении соединения с сервисом статистики: %v\n", err)
		return
	}
	defer conn.Close()

	// Читаем данные из файла "reports.json"
	data, err := ioutil.ReadFile("reports.json")
	if err != nil {
		log.Printf("Ошибка при чтении файла reports.json: %v\n", err)
		return
	}

	// Отправляем данные в сервис статистики
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Ошибка при отправке данных в сервис статистики: %v\n", err)
		return
	}

	log.Println("Данные успешно отправлены в сервис статистики.")
}

func generateReportEntry(shortLink, originalLink, sourceIP, timeInterval string, count int) *ReportEntry {
	// Заменяем ::1 на 127.0.0.1, если адрес является локальным
	if sourceIP == "::1" {
		sourceIP = "127.0.0.1"
	}

	reportIDCounter++
	saveCounterToFile()
	return &ReportEntry{
		ID:           reportIDCounter,
		PID:          nil,
		OriginalURL:  originalLink,
		ShortURL:     shortLink,
		SourceIP:     sourceIP,
		TimeInterval: timeInterval,
		Count:        count,
	}
}

func saveCounterToFile() {
	data := []byte(fmt.Sprintf("%d", reportIDCounter))
	err := ioutil.WriteFile("report_counter.txt", data, 0644)
	if err != nil {
		log.Println("Ошибка при сохранении счетчика в файл:", err)
	}
}

func loadCounterFromFile() {
	data, err := ioutil.ReadFile("report_counter.txt")
	if err == nil {
		savedCounter, err := strconv.Atoi(string(data))
		if err == nil {
			reportIDCounter = savedCounter
		}
	}
}

// Сохранение данных в файл
func saveReportToFile(entry *ReportEntry) {
	fileName := "reports.json"

	// Чтение существующего файла
	existingData, err := ioutil.ReadFile(fileName)
	if err != nil && !os.IsNotExist(err) {
		log.Println("Ошибка при чтении существующего файла отчета:", err)
		return
	}

	// Распаковка существующего файла в массив отчетов
	var report Report
	if len(existingData) > 0 {
		err := json.Unmarshal(existingData, &report)
		if err != nil {
			log.Println("Ошибка при разборе существующего файла отчета:", err)
			return
		}
	}

	// Добавление нового отчета
	report.Entries = append(report.Entries, entry)

	// Сохранение обновленного отчета в файл
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Println("Ошибка при маршалинге отчета:", err)
		return
	}

	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		log.Println("Ошибка при записи отчета в файл:", err)
	}
}
