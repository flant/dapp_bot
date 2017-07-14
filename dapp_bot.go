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
	"strconv"
)

const tips_file_url string = "https://raw.githubusercontent.com/alexey-igrychev/dapp/master/docs/hints"
const chats_file string = "chats"

func handle_update(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var tips []string
	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		add_chat(strconv.Itoa(int(update.Message.Chat.ID)))
		tips = current_tips()

		if update.Message.Text == "/tips" {
			send_tips(bot, update.Message.Chat.ID, tips)
		} else if update.Message.Text == "/tip" && len(tips) > 0 {
			send_tip(bot, update.Message.Chat.ID, random_tip(tips))
		}
	}
}

func handle_notification(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	var tips []string
	var saved_tips []string

	for {
		new_tips := new_tips(saved_tips)
		saved_tips = current_tips()
		if len(new_tips) > 0 {
			tips = new_tips
		} else {
			tips = saved_tips
		}

		for _, chat_id := range current_chat_ids() {
			chat_id_int_64, _ := strconv.ParseInt(string(chat_id), 10, 64)
			send_tip(bot, chat_id_int_64, random_tip(tips))
		}
		time.Sleep(time.Second * 60 * 60)
	}
}

func current_tips() ([]string) {
	resp, err := http.Get(tips_file_url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return strings.Split(strings.Trim(string(body), "---\n"), "---\n")
}

func random_tip(tips []string) (string) {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return tips[r.Intn(len(tips))]
}

func new_tips(old_tips []string) ([]string) {
	return difference(current_tips(), old_tips)
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

func add_chat(new_chat_id string) {
	current_chats_ids := current_chat_ids()

	if array_include(new_chat_id, current_chats_ids) {
		return
	}

	chats_ids := append(current_chats_ids, new_chat_id)
	body := []byte(strings.Join(chats_ids, "\n"))

	err := ioutil.WriteFile(chats_file, body, 0644)
	if err != nil {
		panic(err)
	}
}

func current_chat_ids() ([]string) {
	_, err := os.Stat(chats_file)
	var chats_ids []string

	if !os.IsNotExist(err) {
		body, err := ioutil.ReadFile(chats_file)
		if err != nil {
			panic(err)
		}
		chats_ids = strings.Split(string(body), "\n")
	}
	return chats_ids
}

func array_include(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func send_tips(bot *tgbotapi.BotAPI, chat_id int64, tips []string) {
	for _, tip := range tips {
		send_tip(bot, chat_id, tip)
	}
}

func send_tip(bot *tgbotapi.BotAPI, chat_id int64, tip string) {
	msg := tgbotapi.NewMessage(chat_id, tip)
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

func main() {
	token := flag.String("token", "", "telegram bot token")
	flag.Parse()

	if *token == "" {
		log.Fatal("token flag required")
	}

	go handle_notification(*token)
	handle_update(*token)
}