package parser

import (
	"fmt"
	"time"
)

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

func NewEvent(t time.Time, eventID, compID int, params interface{}) Event {
	return &event{time: t, eventID: eventID, competitorID: compID, extraParams: params}
}
