package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

func main() {
	config, err := LoadConfig()

	if err != nil {
		fmt.Println("Error encountered while loading configuration:", err.Error())
		os.Exit(1)
	} else {
		confBytes, _ := json.MarshalIndent(config, "", "    ")
		fmt.Printf("Environment Configuration:\n%s\n", string(confBytes))
	}

	publishChan := make(chan PublishEvent)

	consumers := getConsumers(config, publishChan)
	producer := NewSatoriProducer(config, publishChan)
	if err = producer.Initialize(); err != nil {
		log.Println("Error initializing Satori producer:", err.Error())
	}

	go producer.Produce()

	var wg sync.WaitGroup
	for _, c := range consumers {
		wg.Add(1)
		go func(consumer TwitchAPIConsumer) {
			defer wg.Done()
			consumer.Consume()
		}(c)
	}

	wg.Wait()

	log.Println("Shutting down...")
}

func getConsumers(config *Config, pub chan PublishEvent) []TwitchAPIConsumer {
	return []TwitchAPIConsumer{
		DefaultClipsConsumer(pub, config),
		DefaultGamesConsumer(pub, config),
		DefaultStreamsConsumer(pub, config),
		DefaultVideosConsumer(pub, config),
	}
}
