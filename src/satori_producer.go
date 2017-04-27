package main

import (
	"log"

	"github.com/satori-com/satori-rtm-sdk-go/rtm"
	"github.com/satori-com/satori-rtm-sdk-go/rtm/auth"
)

type SatoriProducer struct {
	ChannelName  string
	Endpoint     string
	AppKey       string
	Role         string
	Secret       string
	PublishChan  chan PublishEvent
	Shutdown     bool
	SatoriClient *rtm.RTM
}

func NewSatoriProducer(config *Config, pub chan PublishEvent) *SatoriProducer {
	return &SatoriProducer{
		ChannelName: config.SatoriChannelName,
		Endpoint:    config.SatoriEndpoint,
		AppKey:      config.SatoriAppKey,
		Role:        config.SatoriRole,
		Secret:      config.SatoriSecret,
		PublishChan: pub,
		Shutdown:    false,
	}
}

func (producer *SatoriProducer) Initialize() error {
	client, err := rtm.New(producer.Endpoint, producer.AppKey, rtm.Options{
		AuthProvider: auth.New(producer.Role, producer.Secret),
	})

	producer.SatoriClient = client

	producer.SatoriClient.Start()

	return err
}

func (producer *SatoriProducer) Produce() {
	// continue to publish events while we have not recieved a kill signal or we
	// have events that need to be published
	for !producer.Shutdown || len(producer.PublishChan) > 0 {
		select {
		case event := <-producer.PublishChan:
			if producer.SatoriClient.IsConnected() {
				err := producer.SatoriClient.Publish(producer.ChannelName, event)
				if err != nil {
					log.Println("Error publishing event to Satori channel:", err.Error())
				}
			}
		default:
			// do nothing
		}
	}

	producer.SatoriClient.Stop()
	log.Println("Satori event production has shutdown")
}
