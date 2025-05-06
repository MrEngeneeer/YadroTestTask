package logger

import (
	"fmt"
	"github.com/MrEngeneer/YadroTestTask/compute"
	"github.com/MrEngeneer/YadroTestTask/parser"
	"os"
)

func SaveLog(path string, events []parser.Event) {
	file, _ := os.Create("output/" + path)
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

func SaveResults(results []compute.Result) {
	file, err := os.Create("output/report")
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
		_, err := file.WriteString(fmt.Sprintf("[%s] %d %v %v %d/%d\n", result.TotalTime, result.CompetitorId, result.LapResults, result.PenaltyLapsResult, result.Hit, result.Shot))
		if err != nil {
			fmt.Println("error in report writing: ", err)
		}
	}
}
