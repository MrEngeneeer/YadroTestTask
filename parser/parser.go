package parser

import (
	"encoding/json"
	"fmt"
	"os"
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

func LoadConfig(path string) Config {
	file, _ := os.ReadFile(path)
	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("error in config parsing: ", err)
	}
	return config
}

func LoadEvents(path string) []Event {
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

func FormatDuration(duration time.Duration) string {
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
