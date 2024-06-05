package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func processCommand(command string, filename string) string {
	// Разбиваем команду на токены, разделенные пробелами
	tokens := strings.Fields(command)

	if len(tokens) < 1 {
		return "Пустая команда"
	}

	cmd := tokens[0]

	switch cmd {
	case "SENDJSON":
		// Обработка команды SENDJSON
		sendReportsJSONToStatisticService()
		return "SENDJSON выполнено"

	case "SHORTLINK":
		// Обработка команды SHORTLINK
		if len(tokens) < 3 {
			return "Недостаточно аргументов для команды SHORTLINK"
		}
		shortLink := tokens[1]
		originalLink := strings.Join(tokens[2:], " ")
		linkMap[shortLink] = originalLink
		saveLinksToFile()

		return fmt.Sprintf("SHORTLINK выполнено: %s -> %s", shortLink, originalLink)

	default:
		return "Неизвестная команда"
	}
}

// Функция для чтения строк из файла
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Функция для записи строк в файл
func writeLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
