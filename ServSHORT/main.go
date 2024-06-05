package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

var characters string
var linkLength = 6

func init() {
	characters = generateCharacters()
}

func generateCharacters() string {
	var chars []rune
	for i := 'a'; i <= 'z'; i++ {
		chars = append(chars, i)
	}
	for i := 'A'; i <= 'Z'; i++ {
		chars = append(chars, i)
	}
	for i := '0'; i <= '9'; i++ {
		chars = append(chars, i)
	}
	return string(chars)
}

func main() {
	listener, err := net.Listen("tcp", "172.17.2.214:9111")
	if err != nil {
		fmt.Println("Ошибка при запуске TCP сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP сервер запущен на порту 9111")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии соединения:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Подключено клиентом", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		command := scanner.Text()
		if strings.HasPrefix(command, "SHORT ") {
			args := strings.Fields(command)[1:]
			if len(args) != 1 {
				fmt.Fprintln(conn, "Неверный формат команды")
				continue
			}

			originalLink := args[0]
			shortenedURL, err := shortenURL(originalLink)
			if err != nil {
				fmt.Fprintln(conn, "Ошибка при сокращении ссылки:", err)
				continue
			}

			response := fmt.Sprintf("Сокращенная ссылка: http://217.71.129.139:4712/redirect/%s", shortenedURL)
			fmt.Fprintln(conn, response)
		} else if command == "EXIT" {
			fmt.Println("Получена команда EXIT. Закрытие соединения.")
			break
		} else {
			fmt.Fprintln(conn, "Неизвестная команда")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка при чтении команд от клиента:", err)
	}
}

func shortenURL(originalURL string) (string, error) {
	shortLink := generateShortLink()
	err := sendToDBService(shortLink, originalURL)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки данных в СУБД: %v", err)
	}
	return shortLink, nil
}

func generateShortLink() string {
	shortLink := ""
	for i := 0; i < linkLength; i++ {
		shortLink += string(characters[rand.Intn(len(characters))])
	}
	return shortLink
}

func sendToDBService(shortLink, originalLink string) error {
	// Отправка данных в СУБД
	conn, err := net.Dial("tcp", "172.17.2.214:6379")
	if err != nil {
		return fmt.Errorf("ошибка подключения к СУБД: %v", err)
	}
	defer conn.Close()

	command := fmt.Sprintf("SHORTLINK %s %s", shortLink, originalLink)
	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		return fmt.Errorf("ошибка отправки команды на СУБД: %v", err)
	}

	return nil
}
