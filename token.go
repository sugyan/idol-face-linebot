package main

import (
	"net/url"
	"os"
	"time"

	"github.com/sugyan/face-manager-linebot/recognizer"
	"gopkg.in/redis.v5"
)

func (a *app) retrieveUserToken(userID string) (string, error) {
	redisURL := os.Getenv("REDIS_URL") // required
	parsedURL, err := url.Parse(redisURL)
	if err != nil {
		return "", err
	}
	password, _ := parsedURL.User.Password()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     parsedURL.Host,
		Password: password,
	})
	// retrieve from redis
	key := "token:" + userID
	token, err := redisClient.Get(key).Result()
	if err == redis.Nil {
		// get profile from messaging API
		profile, err := a.bot.GetProfile(userID).Do()
		if err != nil {
			return "", err
		}
		// register user and get authentication token as admin
		adminEmail := os.Getenv("RECOGNIZER_ADMIN_EMAIL")
		adminToken := os.Getenv("RECOGNIZER_ADMIN_TOKEN")
		recognizerClient, err := recognizer.NewClient(adminEmail, adminToken)
		if err != nil {
			return "", err
		}
		token, err = recognizerClient.RegisterUser(userID, profile.DisplayName)
		if err != nil {
			return "", err
		}
		// cache to redis (24 hours)
		if err := redisClient.Set(key, token, time.Hour*24).Err(); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return token, nil
}
