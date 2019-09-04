package common

import (
	"log"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"fmt"
)

const (
	url      = "http://127.0.0.1:8086"
	database = "sole"
	username = "root"
	password = "root"
)

type DataPoint struct {
	name       string
	location   string
	identifier string
	sensorType string
	unit       string
	value      float64
}

type Influxdb struct {
	client client.Client
}

func (i *Influxdb) Connect() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}

	i.client = c

	i.queryDB(fmt.Sprintf("CREATE DATABASE %s", database))
}

func (i *Influxdb) Disconnect() {
	i.client.Close()
}

func (i *Influxdb) queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := i.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func (i *Influxdb) Send(dataPoints []DataPoint) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range dataPoints {
		// Create a point and add to batch
		tags := map[string]string{
			"location":   p.location,
			"identifier": p.identifier,
			"unit":       p.unit,
		}

		fields := map[string]interface{}{"value": p.value,}

		pt, err := client.NewPoint(p.name, tags, fields, time.Now())
		if err != nil {
			fmt.Println(err)
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	if err := i.client.Write(bp); err != nil {
		fmt.Println(err)
	}

	// Close client resources
	if err := i.client.Close(); err != nil {
		log.Fatal(err)
	}
}
