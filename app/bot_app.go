package app

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sugyan/idol-face-linebot/recognizer"
	"gopkg.in/redis.v5"
)

// BotApp type
type BotApp struct {
	linebot         *linebot.Client
	redis           *redis.Client
	recognizerAdmin *recognizer.Client
	cipherBlock     cipher.Block
	imageDir        string
	baseURL         string
	port            string
}

// Config type
type Config struct {
	ChannelSecret        string
	ChannelToken         string
	RecognizerAdminEmail string
	RecognizerAdminToken string
	RedisURL             string
	AppBaseURL           string
	ListenPort           string
}

// Run method
func (app *BotApp) Run(callbackPath string) error {
	http.HandleFunc(callbackPath, app.callbackHandler)
	http.HandleFunc("/image", app.imageHandler)
	http.HandleFunc("/face", app.faceHandler)
	if err := http.ListenAndServe(":"+app.port, nil); err != nil {
		return err
	}
	return nil
}

// NewBotApp function returns app inctance
func NewBotApp(config *Config) (*BotApp, error) {
	// linebot client
	linebotClient, err := linebot.New(config.ChannelSecret, config.ChannelToken)
	if err != nil {
		return nil, err
	}
	// recognizer client
	recognizerAdminClient, err := recognizer.NewClient(config.RecognizerAdminEmail, config.RecognizerAdminToken)
	if err != nil {
		return nil, err
	}
	// redis client
	parsedURL, err := url.Parse(config.RedisURL)
	if err != nil {
		return nil, err
	}
	password, _ := parsedURL.User.Password()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     parsedURL.Host,
		Password: password,
	})
	// cipher
	key, err := hex.DecodeString(config.ChannelSecret)
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
	return &BotApp{
		linebot:         linebotClient,
		recognizerAdmin: recognizerAdminClient,
		redis:           redisClient,
		cipherBlock:     block,
		imageDir:        dirName,
		baseURL:         config.AppBaseURL,
		port:            config.ListenPort,
	}, nil
}
