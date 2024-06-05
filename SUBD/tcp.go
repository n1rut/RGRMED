package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

func handleConnection(conn net.Conn, Mute *sync.Mutex) {
	defer conn.Close()

	fmt.Println("Подключено клиентом", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	filename := "DBMS.txt"
	for {
		if !scanner.Scan() {
			break
		}
		command := scanner.Text()
		Mute.Lock()
		response := processCommand(command, filename)
		Mute.Unlock()
		_, err := conn.Write([]byte(response + "\n"))
		if err != nil {
			fmt.Println("Ошибка при отправке ответа клиенту:", err)
			break
		}
	}
}
