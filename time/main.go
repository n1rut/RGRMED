package main

import (
    "fmt"
    "net"
    "time"
)

func main() {
    serverAddr := "172.17.2.214:6379"
    message := "SENDJSON"

    for {
        conn, err := net.Dial("tcp", serverAddr)
        if err != nil {
            fmt.Println("Error connecting:", err)
            time.Sleep(time.Minute) // Повторная попытка через минуту
            continue
        }

        _, err = fmt.Fprintf(conn, message+"\n") // Отправка сообщения с переводом строки
        if err != nil {
            fmt.Println("Error sending:", err)
        }

        conn.Close()
        time.Sleep(time.Minute) // Пауза на минуту перед следующим циклом
    }
}
