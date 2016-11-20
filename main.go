package main

import (
	"log"
	"os"

	"github.com/sugyan/idol-face-linebot/app"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	config := &app.Config{
		ChannelSecret:        os.Getenv("CHANNEL_SECRET"),
		ChannelToken:         os.Getenv("CHANNEL_TOKEN"),
		RecognizerAdminEmail: os.Getenv("RECOGNIZER_ADMIN_EMAIL"),
		RecognizerAdminToken: os.Getenv("RECOGNIZER_ADMIN_TOKEN"),
		RedisURL:             os.Getenv("REDIS_URL"),
		AppBaseURL:           os.Getenv("APP_URL"),
		ListenPort:           os.Getenv("PORT"),
	}
	app, err := app.NewBotApp(config)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(os.Getenv("CALLBACK_PATH")); err != nil {
		log.Fatal(err)
	}
}
