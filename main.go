package main

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/net/context"
)

type options struct {
	Environment string
	AmqpURL     string `default:"amqp://localhost"`
}

type dockerHubPushMessage struct {
	CallbackURL string `json:"callback_url"`
	PushData    struct {
		Tag string `json:"tag"`
	} `json:"push_data"`
	Repository struct {
		Name string `json:"repo_name"`
	} `json:"repository"`
}

func main() {
	var opt options
	err := envconfig.Process("listener", &opt)
	if err != nil {
		log.Fatal(err)
	}

	if opt.Environment == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	ctx, done := context.WithCancel(context.Background())
	subscriber, err := NewSubscriber(ctx, done, opt.AmqpURL,
		"holmescode.deployments", "topic", "holmescode.deploymentsQueue", "", "")
	if err != nil {
		log.Fatal(err)
	}

	subscriber.Consume()
	for {
		select {
		case msg := <-subscriber.Messages:
			push := &dockerHubPushMessage{}
			err := json.Unmarshal(msg.Body, push)
			if err != nil {
				log.WithFields(log.Fields{
					"raw_message": string(msg.Body[:]),
					"error":       err,
				}).Error("Could not parse message")
				break
			}

			log.WithFields(log.Fields{
				"callback_url": push.CallbackURL,
				"tag":          push.PushData.Tag,
				"repo_name":    push.Repository.Name,
			}).Info("Received push notification")
		case <-ctx.Done():
			subscriber.Close()
			return
		}
	}
}
