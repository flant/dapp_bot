package main

import (
	"log"
	"flag"
	"net/http"
	"gopkg.in/telegram-bot-api.v4"
	"math/rand"
	"strings"
	"time"
	"io/ioutil"
	"os"
)

const tips_file_url string = "https://raw.githubusercontent.com/flant/dapp/master/docs/hints"
const chats_file string = "chats"
var tips []string

func new_tips() ([]string) {
	resp, err := http.Get(tips_file_url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	file_tips := strings.Split(strings.Trim(string(body), "---\n"), "---\n")
	new_tips := difference(file_tips, tips)
	tips = file_tips
	return new_tips
}

func random_tip() (string) {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return tips[r.Intn(len(tips))]
}

func send_tips(bot *tgbotapi.BotAPI, chat_id int64, tips []string) {
	for _, tip := range tips {
		msg := tgbotapi.NewMessage(chat_id, tip)
		msg.ParseMode = "markdown"
		bot.Send(msg)
	}
}

func add_chat(new_chat_id string) {
	_, err := os.Stat(chats_file)
	var chats_ids []string

	if os.IsExist(err) {
		body, err := ioutil.ReadFile(chats_file)
		if err != nil {
			panic(err)
		}
		chats_ids = strings.Split(string(body), "\n")
	}

	if array_include(new_chat_id, chats_ids) {
		return
	}

	chats_ids = append(chats_ids, new_chat_id)
	body := []byte(strings.Join(chats_ids, "\n"))
	err = ioutil.WriteFile(chats_file, body, 0644)
	if err != nil {
		panic(err)
	}
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func array_include(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	token := flag.String("token", "", "telegram bot token")
	flag.Parse()

	if *token == "" {
		log.Fatal("token flag required")
	}

	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		add_chat(string(update.Message.Chat.ID))
		send_tips(bot, update.Message.Chat.ID, new_tips())

		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/tips" {
			send_tips(bot, update.Message.Chat.ID, tips)
		} else if update.Message.Text == "/tip" && len(tips) > 0 {
			send_tips(bot, update.Message.Chat.ID, []string { random_tip() })
		}
	}
}