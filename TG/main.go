package main

import (
	"bufio"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	botToken      = "6896962024:AAHGF6M47rmMx-zRHG63aYhfAciB3L_8oDQ" // Замените на ваш токен бота
	shortenerHost = "172.17.2.214"
	shortenerPort = "9111"
	logFile       = "logs.log"
)

func main() {
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Ошибка открытия файла логов: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Включить режим отладки

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	log.Printf("[%s] Получено сообщение от %s: %s", time.Now().Format(time.RFC3339), message.From.UserName, message.Text) // Логирование сообщения

	if message.IsCommand() {
		// ... (обработка команд)
	} else {
		// Проверка на валидность ссылки
		if _, err := url.ParseRequestURI(message.Text); err != nil {
			log.Printf("[%s] Невалидная ссылка: %s", time.Now().Format(time.RFC3339), message.Text) // Логирование невалидной ссылки
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Это не ссылка. Пожалуйста, отправьте валидную ссылку."))
			return
		}

		shortenedURL, err := shortenURL(message.Text)
		if err != nil {
			log.Printf("[%s] Ошибка при сокращении ссылки: %v", time.Now().Format(time.RFC3339), err) // Логирование ошибки
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ошибка при сокращении ссылки."))
			return
		}

		log.Printf("[%s] Ссылка сокращена: %s -> %s", time.Now().Format(time.RFC3339), message.Text, shortenedURL) // Логирование успешного сокращения

		replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Сократить ссылку", "shorten"),
				tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
			),
		)

		msg := tgbotapi.NewMessage(message.Chat.ID, shortenedURL)
		msg.ReplyMarkup = replyMarkup
		bot.Send(msg)
	}
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	switch query.Data {
	case "shorten":
		// Здесь можно добавить логику повторного сокращения ссылки
		bot.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, "Ссылка уже сокращена"))
	case "cancel":
		bot.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, "Действие отменено"))
		bot.DeleteMessage(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))
	}
}

func shortenURL(originalURL string) (string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", shortenerHost, shortenerPort))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	fmt.Fprintf(conn, "SHORT %s\n", originalURL)

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		if strings.HasPrefix(response, "Сокращенная ссылка:") {
			return strings.TrimPrefix(response, "Сокращенная ссылка: "), nil
		}
	}

	return "", fmt.Errorf("invalid response from shortener service")
}
