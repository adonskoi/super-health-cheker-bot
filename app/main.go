package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/adonskoi/super-health-checker-bot/app/checker"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var TelegramAPIToken string
var ChatID int64
var Path string

func init() {
	flag.StringVar(&TelegramAPIToken, "token", "", "telegram api token")
	flag.Int64Var(&ChatID, "chat", 0, "telegram chat id")
	flag.StringVar(&Path, "path", "", "path to files store")
	flag.Parse()
}

func main() {
	bot, err := tbapi.NewBotAPI(TelegramAPIToken) // interface in bot.go
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	updateConfig := tbapi.NewUpdate(0)
	updateConfig.Timeout = 30

	// listen bots update
	updates := bot.GetUpdatesChan(updateConfig)
	go listenUpdates(bot, updates)

	// channel for send alert messages
	messages := make(chan string)
	go sendMessage(bot, messages)

	for {
		Checker, err := checker.New(Path, messages)
		if err != nil {
			log.Println(err)
		}
		result, ok := Checker.Check()
		log.Printf("result: %s, ok: %t", result, ok)
		if !ok {
			messages <- result
		}
		time.Sleep(120 * time.Second)
	}

}

func listenUpdates(bot *tbapi.BotAPI, updates tbapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		text := "I need .json files"
		if update.Message.Chat.ID != ChatID {
			text = "fuck off" // TODO translate to other lang
		} else if update.Message.Document != nil {
			fileID := update.Message.Document.FileID
			fileUrl, err := bot.GetFileDirectURL(fileID)
			if err != nil {
				log.Println(err)
			}
			filePath := path.Join(Path, update.Message.Document.FileName)
			err = downloadFile(filePath, fileUrl)
			if err != nil {
				log.Println(err)
			}
			text = "file uploaded"
		}

		msg := tbapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func sendMessage(bot *tbapi.BotAPI, messages <-chan string) {
	for text := range messages {
		msg := tbapi.NewMessage(ChatID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	}
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
