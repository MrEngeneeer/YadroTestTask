package main

import (
	"flag"
	"github.com/MrEngeneer/YadroTestTask/compute"
	"github.com/MrEngeneer/YadroTestTask/logger"
	"github.com/MrEngeneer/YadroTestTask/parser"
	"sort"
)

func main() {
	configPath := flag.String("config", "", "")
	inputPath := flag.String("input", "", "")
	flag.Parse()
	config := parser.LoadConfig(*configPath)
	events := parser.LoadEvents(*inputPath)
	events = compute.AppendFinalEvents(events, config)
	sort.Slice(events, func(i, j int) bool { return events[i].Time().Before(events[j].Time()) })
	logger.SaveLog("output.log", events)
	results := compute.ProcessEvents(events, config)
	logger.SaveResults(results)
}
