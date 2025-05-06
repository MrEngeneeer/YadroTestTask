package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CustomTime struct{ time.Time }

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("15:04:05.000", s)
	if err != nil {
		t, err = time.Parse("15:04:05", s)
		if err != nil {
			return err
		}
	}
	ct.Time = t
	return nil
}

type CustomDuration struct{ time.Duration }

func (cd *CustomDuration) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("15:04:05.000", s)
	if err != nil {
		t, err = time.Parse("15:04:05", s)
		if err != nil {
			return err
		}
	}
	cd.Duration = time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond())
	return nil
}

type Config struct {
	Laps        int            `json:"laps"`
	LapLen      int            `json:"lapLen"`
	PenaltyLen  int            `json:"penaltyLen"`
	FiringLines int            `json:"firingLines"`
	Start       CustomTime     `json:"start"`
	StartDelta  CustomDuration `json:"startDelta"`
}

type Event interface {
	Time() time.Time
	EventID() int
	CompetitorID() int
	ExtraParams() interface{}
	String() string
}

type event struct {
	time         time.Time
	eventID      int
	competitorID int
	extraParams  interface{}
}

var eventComments = map[int]string{
	1:  "The competitor(%d) registered",
	2:  "The start time for the competitor(%d) was set by a draw to %s",
	3:  "The competitor(%d) is on the start line",
	4:  "The competitor(%d) has started",
	5:  "The competitor(%d) is on the firing range(%d)",
	6:  "The target(%d) has been hit by competitor(%d)",
	7:  "The competitor(%d) left the firing range",
	8:  "The competitor(%d) entered the penalty laps",
	9:  "The competitor(%d) left the penalty laps",
	10: "The competitor(%d) ended the main lap",
	11: "The competitor(%d) can`t continue: %s",
	32: "The competitor(%d) is disqualified: %s",
	33: "The competitor(%d) finished",
}

func (e *event) Time() time.Time          { return e.time }
func (e *event) EventID() int             { return e.eventID }
func (e *event) CompetitorID() int        { return e.competitorID }
func (e *event) ExtraParams() interface{} { return e.extraParams }
func (e *event) String() string {
	switch e.eventID {
	case 2:
		return fmt.Sprintf(eventComments[e.eventID], e.competitorID, e.extraParams.(time.Time).Format("15:04:05.000"))
	case 5:
		return fmt.Sprintf(eventComments[e.eventID], e.competitorID, e.extraParams.(int))
	case 6:
		n := e.extraParams.(int)
		return fmt.Sprintf(eventComments[e.eventID], n, e.competitorID)
	case 11, 32:
		return fmt.Sprintf(eventComments[e.eventID], e.competitorID, e.extraParams.(string))
	default:
		return fmt.Sprintf(eventComments[e.eventID], e.competitorID)
	}
}

func ParseEvent(line string) (Event, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid event format: %s", line)
	}
	t, err := time.Parse("15:04:05.000", strings.Trim(parts[0], "[]"))
	if err != nil {
		t, err = time.Parse("15:04:05", strings.Trim(parts[0], "[]"))
		if err != nil {
			return nil, err
		}
	}
	eventID, _ := strconv.Atoi(parts[1])
	competitorID, _ := strconv.Atoi(parts[2])
	var extra interface{}
	switch eventID {
	case 2:
		st, _ := time.Parse("15:04:05.000", parts[3])
		extra = st
	case 5, 6:
		n, _ := strconv.Atoi(parts[3])
		extra = n
	case 11:
		extra = strings.Join(parts[3:], " ")
	}
	return &event{t, eventID, competitorID, extra}, nil
}

type LapResult struct {
	time  string
	speed float64
}

type Result struct {
	totalTime         string
	competitorId      int
	lapResults        []LapResult
	penaltyLapsResult LapResult
	hit               int
	shot              int
}

func main() {
	configPath := flag.String("config", "", "")
	inputPath := flag.String("input", "", "")
	flag.Parse()
	cfg := loadConfig(*configPath)
	events := loadEvents(*inputPath)
	events = appendFinalEvents(events, cfg)
	sort.Slice(events, func(i, j int) bool { return events[i].Time().Before(events[j].Time()) })
	saveLog("output.log", events)
	results := processEvents(events, cfg)
	saveResults(results)
}

func loadConfig(path string) Config {
	file, _ := os.ReadFile(path)
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("error in config parsing: ", err)
	}
	return config
}

func loadEvents(path string) []Event {
	b, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	var evs []Event
	for _, l := range lines {
		if e, err := ParseEvent(l); err == nil {
			evs = append(evs, e)
		}
	}
	return evs
}

func appendFinalEvents(events []Event, config Config) []Event {
	byCompetitors := make(map[int][]Event)
	for _, event := range events {
		byCompetitors[event.CompetitorID()] = append(byCompetitors[event.CompetitorID()], event)
	}
	for id, evs := range byCompetitors {
		laps := 0
		var scheduledStartTime time.Time
		started := false
		for _, e := range evs {
			if e.EventID() == 2 {
				scheduledStartTime = e.ExtraParams().(time.Time)
			}
			if e.EventID() == 4 {
				started = e.Time().Sub(scheduledStartTime) <= config.StartDelta.Duration
			}
			if e.EventID() == 10 {
				laps++
			}
		}
		last := evs[len(evs)-1].Time()
		if !started {
			events = append(events, &event{last, 32, id, "Not started"})
		} else if laps == config.Laps {
			events = append(events, &event{last, 33, id, nil})
		} else {
			events = append(events, &event{last, 32, id, "Not finished"})
		}
	}
	return events
}

func saveLog(path string, events []Event) {
	file, _ := os.Create(path)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	for _, event := range events {
		_, err := file.WriteString(fmt.Sprintf("[%s] %s\n", event.Time().Format("15:04:05.000"), event.String()))
		if err != nil {
			fmt.Println("error in log writing: ", err)
		}
	}
}

func processEvents(events []Event, config Config) []Result {
	byCompetitor := make(map[int][]Event)
	for _, event := range events {
		byCompetitor[event.CompetitorID()] = append(byCompetitor[event.CompetitorID()], event)
	}
	var results []Result
	for id, evs := range byCompetitor {
		sort.Slice(evs, func(i, j int) bool { return evs[i].Time().Before(evs[j].Time()) })
		results = append(results, compute(id, evs, config))
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].totalTime == "Not finished" || results[i].totalTime == "Not started" {
			return false
		}
		return results[i].totalTime < results[j].totalTime
	})
	return results
}

func compute(id int, events []Event, config Config) Result {
	var status string
	if events[len(events)-1].EventID() == 33 {
		status = ""
	} else {
		status = events[len(events)-1].ExtraParams().(string)
	}

	var scheduledStartTime time.Time
	var startFiringTime time.Time
	var firingDuration time.Duration
	var currentLapStartTime time.Time
	var currentPenaltyStartTime time.Time
	allPenaltyDuration := time.Duration(0)
	penaltyLaps := 0
	hits := 0
	shots := 0
	lapResults := make([]LapResult, 0)
	var totalTime string
	for _, event := range events {
		switch event.EventID() {
		case 2:
			scheduledStartTime = event.ExtraParams().(time.Time)
		case 4:
			currentLapStartTime = event.Time()
		case 5:
			startFiringTime = event.Time()
			shots += 5
		case 6:
			hits++
		case 7:
			firingDuration = event.Time().Sub(startFiringTime)
		case 8:
			currentPenaltyStartTime = event.Time()
			penaltyLaps += shots - hits
		case 9:
			allPenaltyDuration += event.Time().Sub(currentPenaltyStartTime)
		case 10:
			lapResults = append(lapResults, LapResult{time: formatDuration(event.Time().Sub(currentLapStartTime)), speed: float64(config.LapLen) / (event.Time().Sub(currentLapStartTime) - firingDuration).Seconds()})
			currentLapStartTime = event.Time()
		case 11:
			for len(lapResults) < config.Laps {
				lapResults = append(lapResults, LapResult{})
			}

		}
	}
	penaltyLapsResult := LapResult{time: formatDuration(allPenaltyDuration), speed: float64(config.PenaltyLen*penaltyLaps) / allPenaltyDuration.Seconds()}
	if status == "" {
		totalTime = formatDuration(events[len(events)-1].Time().Sub(scheduledStartTime))
	} else {
		totalTime = status
	}

	return Result{totalTime, id, lapResults, penaltyLapsResult, hits, shots}
}

func saveResults(results []Result) {
	file, err := os.Create("report")
	if err != nil {
		fmt.Println("error creating file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	for _, result := range results {
		_, err := file.WriteString(fmt.Sprintf("[%s] %d %v %v %d/%d\n", result.totalTime, result.competitorId, result.lapResults, result.penaltyLapsResult, result.hit, result.shot))
		if err != nil {
			fmt.Println("error in report writing: ", err)
		}
	}
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = -duration
	}
	hours := int(duration / time.Hour)
	duration -= time.Duration(hours) * time.Hour
	minutes := int(duration / time.Minute)
	duration -= time.Duration(minutes) * time.Minute
	seconds := int(duration / time.Second)
	duration -= time.Duration(seconds) * time.Second
	milliseconds := int(duration / time.Millisecond)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
