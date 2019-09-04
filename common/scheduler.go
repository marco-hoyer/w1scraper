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
	s.scheduler.Every(30).Seconds().Do(scrape)

	<- s.scheduler.Start()
}

func (s *Schedule) Stop() {
	s.scheduler.Clear()
}

func scrape() {
	log.Printf("scraping 1w sensors")
	i := Influxdb{}
	i.Connect()
	sensors := []string{"28.AA9859501401", "28.AA3D4D501401", "28.AA9B28501401", "28.AAE725501401"}

	for _, element := range sensors {
		url := fmt.Sprintf("http://home.fritz.box:2121/json/%s/temperature", element)
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		var msg []string
		json.NewDecoder(resp.Body).Decode(&msg)
		f, _ := strconv.ParseFloat(msg[0], 64)
	
		fmt.Println(element,f)
	
		i.Send("temperature","HWR", "Grad", element, f)

	}
}
