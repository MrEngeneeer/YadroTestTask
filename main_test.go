package main

import (
	"github.com/MrEngeneer/YadroTestTask/compute"
	"github.com/MrEngeneer/YadroTestTask/parser"
	"testing"
	"time"
)

func TestCustomTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
		wantErr  bool
	}{
		{`"12:34:56.789"`, time.Date(0, 1, 1, 12, 34, 56, 789000000, time.UTC), false},
		{`"07:08:09"`, time.Date(0, 1, 1, 7, 8, 9, 0, time.UTC), false},
		{`"invalid"`, time.Time{}, true},
	}
	for _, tt := range tests {
		var ct parser.CustomTime
		err := ct.UnmarshalJSON([]byte(tt.input))
		if (err != nil) != tt.wantErr {
			t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !ct.Time.Equal(tt.expected) {
			t.Errorf("got %v, want %v", ct.Time, tt.expected)
		}
	}
}

func TestCustomDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{`"01:02:03.004"`, time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond, false},
		{`"00:10:20"`, 10*time.Minute + 20*time.Second, false},
		{`"bad"`, 0, true},
	}
	for _, tt := range tests {
		var cd parser.CustomDuration
		err := cd.UnmarshalJSON([]byte(tt.input))
		if (err != nil) != tt.wantErr {
			t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && cd.Duration != tt.expected {
			t.Errorf("got %v, want %v", cd.Duration, tt.expected)
		}
	}
}

func TestParseEvent(t *testing.T) {
	line := "[12:00:01.234] 2 5 12:05:00.000"
	event, err := parser.ParseEvent(line)
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if event.EventID() != 2 || event.CompetitorID() != 5 {
		t.Errorf("EventID or CompetitorID wrong: got %d, %d", event.EventID(), event.CompetitorID())
	}
	extra, ok := event.ExtraParams().(time.Time)
	if !ok {
		t.Fatalf("ExtraParams type incorrect")
	}
	wantTime, _ := time.Parse("15:04:05.000", "12:05:00.000")
	if !extra.Equal(wantTime) {
		t.Errorf("ExtraParams time wrong: got %v, want %v", extra, wantTime)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		dur      time.Duration
		expected string
	}{
		{time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond, "01:02:03.004"},
		{-3*time.Second - 50*time.Millisecond, "00:00:03.050"},
	}
	for _, tt := range tests {
		got := parser.FormatDuration(tt.dur)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.dur, got, tt.expected)
		}
	}
}

func TestEvent_String(t *testing.T) {
	timeTemp := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	e := parser.NewEvent(timeTemp, 2, 1, timeTemp)
	got := e.String()
	want := "The start time for the competitor(1) was set by a draw to 09:00:00.000"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestProcessEvents_NotStarted(t *testing.T) {
	config := parser.Config{Laps: 1, LapLen: 100, PenaltyLen: 50, FiringLines: 1,
		Start:      parser.CustomTime{time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		StartDelta: parser.CustomDuration{time.Second},
	}

	timeTemp, _ := time.Parse("15:04:05.000", "09:00:00.000")
	event := parser.NewEvent(timeTemp, 2, 1, timeTemp)
	eList := []parser.Event{event}
	final := compute.AppendFinalEvents(eList, config)

	last := final[len(final)-1]
	if last.EventID() != 32 {
		t.Errorf("Expected disqualification eventID 32, got %d", last.EventID())
	}
}
