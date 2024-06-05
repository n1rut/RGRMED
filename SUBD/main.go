package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var linkMap map[string]string
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
	linkMap = make(map[string]string)
	rand.Seed(time.Now().UnixNano())

	go startDBServer("172.17.2.214:6379")

	// Загрузка ссылок и счетчика
	loadLinksFromFile()
	loadCounterFromFile()

	// Маршрут для создания сокращенной ссылки
	http.HandleFunc("/shorten", shortenLink)

	// Маршрут для перенаправления
	http.HandleFunc("/redirect/", redirectLink)

	// Маршрут для отображения HTML-страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "Client.html")
	})

	port := 8080
	fmt.Printf("Сервер слушает на порту %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("Ошибка при запуске сервера: ", err)
	}
}

func startDBServer(address string) {
	var mutex sync.Mutex
	fmt.Printf("Запуск СУБД на %s...\n", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера СУБД на %s: %v\n", address, err)
		return
	}
	defer ln.Close()

	fmt.Printf("Сервер СУБД на %s запущен. Ожидание подключений...\n", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при подключении клиента СУБД:", err)
			continue
		}

		fmt.Println("Подключение от", conn.RemoteAddr().String())

		go handleConnection(conn, &mutex)
	}
}
