package app

import (
	"time"

	"gopkg.in/redis.v5"
)

func (app *BotApp) retrieveUserToken(userID string) (string, error) {
	// retrieve from redis
	key := "token:" + userID
	token, err := app.redis.Get(key).Result()
	if err == redis.Nil {
		// get profile from messaging API
		profile, err := app.linebot.GetProfile(userID).Do()
		if err != nil {
			return "", err
		}
		// register user and get authentication token as admin
		token, err = app.recognizerAdmin.RegisterUser(userID, profile.DisplayName)
		if err != nil {
			return "", err
		}
		// cache to redis (24 hours)
		if err = app.redis.Set(key, token, time.Hour*24).Err(); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return token, nil
}
