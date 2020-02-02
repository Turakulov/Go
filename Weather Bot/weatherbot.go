package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	BotToken   = " "
	WebhookURL = " "
)

var districts = map[string]string{
	"Moscow, Russia":        "https://world-weather.ru/pogoda/russia/moscow/",
	"Samarkand, Uzbekistan": "https://world-weather.ru/pogoda/uzbekistan/samarkand/",
	"Seoul, South Korea":    "https://world-weather.ru/pogoda/south_korea/seoul/",
	"New York, USA":         "https://world-weather.ru/pogoda/usa/new_york/",
	"Dubai, UAE":            "https://world-weather.ru/pogoda/uae/dubai/",
}

func Scraper(url string) (string, error) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return "nil", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "nil", err
	}

	var temperature string
	// Find the review items
	doc.Find(".dw-into").Each(func(i int, s *goquery.Selection) {
		t := strings.Replace(strings.TrimSuffix(s.Text(), "Скрыть"), " Подробнее", " ", 1)
		temperature = strings.Replace(t, "Сегодня", "cегодня", 1)
	})
	return temperature, nil

}

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatal(err)
	}

	//bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		log.Fatal(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":80", nil)
	fmt.Println("start listen :80")

	// получаем все обновления из канала updates
	for update := range updates {

		if update.Message.Text == "/start" {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Cписок городов и стран для вывода погоды: "))
			for v := range districts {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, v))
			}
		}
		if update.Message.Text == "go" || update.Message.Text == "Go" || update.Message.Text == "/go" {
			for v, url := range districts {
				strs, err := Scraper(url)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Sorry, error happend",
					))
				} else {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"В "+v+" "+strs+"\n",
					))
				}
			}
		}
	}
}
