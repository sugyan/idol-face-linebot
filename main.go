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
	app, err := app.NewBotApp()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(os.Getenv("CALLBACK_PATH")); err != nil {
		log.Fatal(err)
	}
}
