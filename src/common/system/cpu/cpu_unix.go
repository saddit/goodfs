package cpu

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"time"
)

var cachedCounts [2]int

func getCounts() (logical, physical int, err error) {
	if cachedCounts[0] == 0 && cachedCounts[1] == 0 {
		cachedCounts[0], err = cpu.Counts(true)
		cachedCounts[1], err = cpu.Counts(false)
	}
	logical, physical = cachedCounts[0], cachedCounts[1]
	return
}

func StatInfo() (res Stat, err error) {
	res.LogicalCount, res.PhysicalCount, err = getCounts()
	percents, err := cpu.Percent(time.Millisecond, false)
	if len(percents) > 0 {
		res.UsedPercent = percents[0]
	}
	return
}
