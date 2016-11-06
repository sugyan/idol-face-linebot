package main

import (
	"time"

	"gopkg.in/redis.v5"
)

func (a *app) retrieveUserToken(userID string) (string, error) {
	// retrieve from redis
	key := "token:" + userID
	token, err := a.redis.Get(key).Result()
	if err == redis.Nil {
		// get profile from messaging API
		profile, err := a.linebot.GetProfile(userID).Do()
		if err != nil {
			return "", err
		}
		// register user and get authentication token as admin
		token, err = a.recognizerAdmin.RegisterUser(userID, profile.DisplayName)
		if err != nil {
			return "", err
		}
		// cache to redis (24 hours)
		if err := a.redis.Set(key, token, time.Hour*24).Err(); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return token, nil
}
