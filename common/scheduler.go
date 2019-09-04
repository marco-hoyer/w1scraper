package common

import (
	"github.com/jasonlvhit/gocron"
	"log"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
)

type Schedule struct {
	scheduler *gocron.Scheduler
}

func (s *Schedule) Init() {
	s.scheduler = gocron.NewScheduler()
	s.scheduler.Every(30).Seconds().Do(scrapeThermalHeatingSensors)

	<-s.scheduler.Start()
}

func (s *Schedule) Stop() {
	s.scheduler.Clear()
}

func scrapeThermalHeatingSensors() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("Error occured: ", x)
		}
	}()

	log.Printf("scraping 1w sensors for thermal heating")
	i := Influxdb{}
	i.Connect()

	inletSensors := []string{"28.AA9859501401", "28.AA3D4D501401"}
	outletSensors := []string{"28.AA9B28501401", "28.AAE725501401"}

	var points []DataPoint

	inletSensorPoints, inletSensorAvgValue := readAndAggregate(inletSensors, "SoleInlet")
	outletSensorPoints, outletSensorAvgValue := readAndAggregate(outletSensors, "SoleOutlet")

	points = append(points, inletSensorPoints...)
	points = append(points, outletSensorPoints...)

	deltaValue := outletSensorAvgValue - inletSensorAvgValue
	deltaValuePoint := DataPoint{value: deltaValue, name: "delta_temperature", identifier: "SoleDelta", sensorType: "temperatureSensor", unit: "Degrees Celsius", location: "Sole"}

	points = append(points, deltaValuePoint)
	i.Send(points)
}

func readTemperatureSensor(sensorAddress string) (value float64) {
	url := fmt.Sprintf("http://home.fritz.box:2121/json/%s/temperature", sensorAddress)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var msg []string
	json.NewDecoder(resp.Body).Decode(&msg)
	f, _ := strconv.ParseFloat(msg[0], 64)
	return f
}

func avg(values []float64) (float64) {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func readAndAggregate(addresses []string, location string) ([]DataPoint, float64) {
	var points []DataPoint
	var values []float64

	for _, sensorAddress := range addresses {
		value := readTemperatureSensor(sensorAddress)
		point := DataPoint{value: value, name: "temperature", identifier: sensorAddress, sensorType: "temperatureSensor", unit: "Degrees Celsius", location: location}

		values = append(values, value)
		points = append(points, point)
	}
	averageValue := avg(values)
	averageValuePoint := DataPoint{value: averageValue, name: "average_temperature", identifier: location, sensorType: "temperatureSensor", unit: "Degrees Celsius", location: location}
	points = append(points, averageValuePoint)

	return points, averageValue
}
