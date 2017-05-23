package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

var (
	pointChan       chan *client.Point
	batchPoints     client.BatchPoints
	batchPointsLock sync.Mutex

	noinflux bool
)

func main() {
	initFlags()
	flag.Parse()

	config, err := LoadConfig()

	log.Println("--no-influx:", noinflux)

	if err != nil {
		fmt.Println("Error encountered while loading configuration:", err.Error())
		os.Exit(1)
	} else {
		confBytes, _ := json.MarshalIndent(config, "", "    ")
		fmt.Printf("Environment Configuration:\n%s\n", string(confBytes))
	}

	var influxClient client.Client
	if !noinflux {
		influxClient, err = client.NewHTTPClient(client.HTTPConfig{
			Addr:     fmt.Sprintf("http://%s:%s", config.InfluxAddr, config.InfluxPort),
			Username: config.InfluxUsername,
			Password: config.InfluxPassword,
		})
		if err != nil {
			fmt.Println("Error connecting to InfluxDB:", err.Error())
			os.Exit(2)
		}
	}

	publishChan := make(chan PublishEvent)
	pointChan = make(chan *client.Point)

	consumers := getConsumers(config, publishChan)
	producer := NewSatoriProducer(config, publishChan)
	if err = producer.Initialize(); err != nil {
		log.Println("Error initializing Satori producer:", err.Error())
	}

	go producer.Produce()

	shutdown := make(chan bool, 1)
	go listenForShutdown(producer, consumers, shutdown)

	if !noinflux {
		go reportInfluxStats(&influxClient, config.InfluxDatabase, shutdown)
	}

	var wg sync.WaitGroup
	for _, c := range consumers {
		wg.Add(1)
		go func(consumer TwitchAPIConsumer) {
			defer wg.Done()
			consumer.Consume()
		}(c)
	}

	wg.Wait()

	close(shutdown)

	log.Println("Shutting down...")
}

func initFlags() {
	flag.BoolVar(&noinflux, "no-influx", false, "Make a connection to InfluxDB via the environment config (default: false)")
}

func getConsumers(config *Config, pub chan PublishEvent) []TwitchAPIConsumer {
	return []TwitchAPIConsumer{
		DefaultClipsConsumer(pub, config),
		DefaultGamesConsumer(pub, config),
		DefaultStreamsConsumer(pub, config),
		DefaultVideosConsumer(pub, config),
	}
}

func listenForShutdown(producer *SatoriProducer, consumers []TwitchAPIConsumer, shutdown chan bool) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for s := range sig {
		log.Println("Recieved signal:", s.String())

		// halt all consumption of events
		for _, consumer := range consumers {
			consumer.Shutdown()
		}

		// halt all production
		producer.Shutdown = true

		shutdown <- true
	}

	return nil
}

func reportInfluxStats(influxClient *client.Client, database string, shutdown chan bool) {
	log.Println("Reporting stats to InfluxDB...")
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	batchPointsLock.Lock()
	batchPoints, _ = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	batchPointsLock.Unlock()

	go createBatchPoints()

	go createPointChanPoint()

	for range ticker.C {
		log.Println("Writing points to InfluxDB")
		err := (*influxClient).Write(batchPoints)
		if err != nil {
			log.Println("Error writing batch points to InfluxDB:", err.Error())
		}

		batchPointsLock.Lock()
		batchPoints, _ = client.NewBatchPoints(client.BatchPointsConfig{
			Database:  database,
			Precision: "s",
		})
		batchPointsLock.Unlock()

		select {
		case <-shutdown:
			return
		default:
			// continue to process influx points
		}
	}
}

func createBatchPoints() {
	for point := range pointChan {
		batchPointsLock.Lock()
		batchPoints.AddPoint(point)
		batchPointsLock.Unlock()
	}
}

func createPointChanPoint() {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fields := map[string]interface{}{
			"channel_len": len(pointChan),
		}
		tags := map[string]string{"channel": "pointChan"}
		sendStatsToInflux("internal_go_channels", tags, fields)
	}
}
