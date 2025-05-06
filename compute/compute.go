package compute

import (
	"github.com/MrEngeneer/YadroTestTask/parser"
	"sort"
	"time"
)

type LapResult struct {
	time  string
	speed float64
}

type Result struct {
	TotalTime         string
	CompetitorId      int
	LapResults        []LapResult
	PenaltyLapsResult LapResult
	Hit               int
	Shot              int
}

func AppendFinalEvents(events []parser.Event, config parser.Config) []parser.Event {
	byCompetitors := make(map[int][]parser.Event)
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
			events = append(events, parser.NewEvent(last, 32, id, "Not started"))
		} else if laps == config.Laps {
			events = append(events, parser.NewEvent(last, 33, id, nil))
		} else {
			events = append(events, parser.NewEvent(last, 32, id, "Not finished"))
		}
	}
	return events
}

func compute(id int, events []parser.Event, config parser.Config) Result {
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
			lapResults = append(lapResults, LapResult{time: parser.FormatDuration(event.Time().Sub(currentLapStartTime)), speed: float64(config.LapLen) / (event.Time().Sub(currentLapStartTime) - firingDuration).Seconds()})
			currentLapStartTime = event.Time()
		case 11:
			for len(lapResults) < config.Laps {
				lapResults = append(lapResults, LapResult{})
			}

		}
	}
	penaltyLapsResult := LapResult{time: parser.FormatDuration(allPenaltyDuration), speed: float64(config.PenaltyLen*penaltyLaps) / allPenaltyDuration.Seconds()}
	if status == "" {
		totalTime = parser.FormatDuration(events[len(events)-1].Time().Sub(scheduledStartTime))
	} else {
		totalTime = status
	}

	return Result{totalTime, id, lapResults, penaltyLapsResult, hits, shots}
}

func ProcessEvents(events []parser.Event, config parser.Config) []Result {
	byCompetitor := make(map[int][]parser.Event)
	for _, event := range events {
		byCompetitor[event.CompetitorID()] = append(byCompetitor[event.CompetitorID()], event)
	}
	var results []Result
	for id, evs := range byCompetitor {
		sort.Slice(evs, func(i, j int) bool { return evs[i].Time().Before(evs[j].Time()) })
		results = append(results, compute(id, evs, config))
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].TotalTime == "Not finished" || results[i].TotalTime == "Not started" {
			return false
		}
		return results[i].TotalTime < results[j].TotalTime
	})
	return results
}
