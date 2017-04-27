package main

import (
	"log"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

func sendStatsToInflux(name string, tags map[string]string, fields map[string]interface{}) {
	point, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		log.Println("Error creating point for InfluxDB:", err.Error())
	}
	pointChan <- point
}
