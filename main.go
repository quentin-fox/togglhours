package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jason0x43/go-toggl"
)

var apiToken string = os.Getenv("TOGGL_KEY")

const inputDateFormat string = "2006-01-02 MST"

type DayWork struct {
	StartTime    time.Time
	StopTime     time.Time
	Hours        []float64
	Descriptions []string
	Description  string
}

func main() {
	startDate, endDate := getDates(os.Args)
	tt := getEntries(startDate, endDate)

	daysWork := make(map[string]*DayWork)
	var days []string // used for sorting

	for _, t := range tt {
		if t.Duration < 0 { // represents ongoing task
			continue
		}

		start := t.StartTime().Round(time.Hour / 2).Local()
		stop := t.StopTime().Round(time.Hour / 2).Local()


		dur := stop.Sub(start).Hours() // already rounded

		taskStartDate := start.Format("2006-01-02")

		dayWork := daysWork[taskStartDate]

		if dayWork == nil { // null pointer
			daysWork[taskStartDate] = &DayWork{
				StartTime:    start,
				Descriptions: []string{t.Description},
				Hours:        []float64{dur},
			}

			days = append(days, taskStartDate)
		} else {
			var descChanged bool
			for _, desc := range dayWork.Descriptions {
				if desc != t.Description {
					descChanged = true
				}
			}

			if descChanged {
				dayWork.Descriptions = append(dayWork.Descriptions, t.Description)
			}

			dayWork.Hours = append(dayWork.Hours, dur)
		}
	}

	if len(days) == 0 {
		log.Fatal("no days with work found")
	}

	sort.Strings(days)

	for _, k := range days {
		v := daysWork[k]
		var hourSum float64
		for _, h := range v.Hours {
			hourSum += h
		}

		minutes := hourSum * 60
		v.StopTime = v.StartTime.Add(time.Minute * time.Duration(minutes))
		if len(v.Descriptions) == 1 {
			v.Description = v.Descriptions[0]
		} else if len(v.Descriptions) > 1 {
			v.Description = `"` + strings.Join(v.Descriptions, ", ") + `"`
		}

		fmt.Printf("%s,%s,%s,%v,%s\n", k, v.StartTime.Format("3:04 pm"), v.StopTime.Format("3:04 pm"), hourSum, v.Description)
	}
}

func getDates(args []string) (time.Time, time.Time) {
	startStr, endStr := os.Args[1], os.Args[2]

	zone, _ := time.Now().Zone();

	if startStr == "" {
		log.Fatal("you must enter a start date")
	}

	if endStr == "" {
		log.Fatal("you must enter an end date")
	}

	startStr += " " + zone

	startDate, err := time.Parse(inputDateFormat, startStr)

	if err != nil {
		log.Fatalf("could not parse start date: %v", err)
	}

	endStr += " " + zone

	endDate, err := time.Parse(inputDateFormat, endStr)

	if err != nil {
		log.Fatalf("could not parse end date: %v", err)
	}

	return startDate, endDate
}

func getEntries(startDate time.Time, endDate time.Time) []toggl.TimeEntry {
	session, err := toggl.NewSession(apiToken, "api_token")

	if err != nil {
		log.Fatalf(err.Error())
	}

	tt, err := session.GetTimeEntries(startDate, endDate)

	if err != nil {
		log.Fatalf(err.Error())
	}

	return tt
}
