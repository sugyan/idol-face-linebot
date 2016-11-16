package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
	"gopkg.in/redis.v5"
)

type app struct {
	linebot         *linebot.Client
	redis           *redis.Client
	recognizerAdmin *recognizer.Client
	cipherBlock     cipher.Block
	imageDir        string
}

func newApp() (*app, error) {
	// linebot client
	linebotClient, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		return nil, err
	}
	// redis client
	redisURL := os.Getenv("REDIS_URL")
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}
	password, _ := parsedURL.User.Password()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     parsedURL.Host,
		Password: password,
	})
	// recognizer client
	adminEmail := os.Getenv("RECOGNIZER_ADMIN_EMAIL")
	adminToken := os.Getenv("RECOGNIZER_ADMIN_TOKEN")
	recognizerAdminClient, err := recognizer.NewClient(adminEmail, adminToken)
	if err != nil {
		return nil, err
	}
	// cipher
	key, err := hex.DecodeString(os.Getenv("CHANNEL_SECRET"))
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// imageDir
	dirName, err := ioutil.TempDir("", "image")
	if err != nil {
		return nil, err
	}
	return &app{
		linebot:         linebotClient,
		redis:           redisClient,
		recognizerAdmin: recognizerAdminClient,
		cipherBlock:     block,
		imageDir:        dirName,
	}, nil
}
