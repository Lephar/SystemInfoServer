package timeutil

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type timeUnit struct {
	unitString      string
	unitCoefficient float64
}

var timeUnits = []timeUnit{
	{
		"ns",
		1,
	}, {
		"us",
		1000,
	}, {
		"ms",
		1000000,
	}, {
		"s",
		1000000000,
	}, {
		"m",
		60000000000,
	}, {
		"h",
		3600000000000,
	},
}

func ParseDuration(timeString string) float64 {
	var unit timeUnit

	for _, tempUnit := range timeUnits {
		if strings.Contains(timeString, tempUnit.unitString) {
			unit = tempUnit
			break
		}
	}

	duration, err := strconv.ParseFloat(timeString[:len(timeString)-len(unit.unitString)], 64)

	if err != nil {
		log.Fatalln(err)
	}

	return unit.unitCoefficient * duration
}

func FormatDuration(duration float64) string {
	var unitIndex int

	for unitIndex = 0; unitIndex < len(timeUnits)-1 && duration/timeUnits[unitIndex+1].unitCoefficient > 1; unitIndex++ {
	}

	unit := timeUnits[unitIndex]
	timeValue := duration / unit.unitCoefficient

	return fmt.Sprintf("%G%s", timeValue, unit.unitString)
}
