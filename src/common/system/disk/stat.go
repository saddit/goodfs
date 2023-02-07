package disk

import (
	"common/datasize"

	"github.com/shirou/gopsutil/v3/disk"
)

// GetAllDrivesIOStats returns IO stats of all drives found in the machine
func GetAllDrivesIOStats() (info AllDrivesIOStats, err error) {
	ic, err := disk.IOCounters()
	if err != nil {
		return
	}
	info = make(AllDrivesIOStats, len(ic))
	for name, stat := range ic {
		info[DevID(name)] = &IOStats{
			ReadBytes: datasize.DataSize(stat.ReadBytes),
			WriteBytes: datasize.DataSize(stat.WriteBytes),
			ReadCount: stat.ReadCount,
			WriteCount: stat.WriteCount,
			ReadTime: stat.ReadTime,
			WriteTime: stat.WriteTime,
			CurrentIOs: stat.IopsInProgress,
			IoTime: stat.IoTime,
			WeightedIO: stat.WeightedIO,
		}
	}
	return
}

func GetAverageIOStats() (*IOStats, error) {
	allIoStats, err := GetAllDrivesIOStats()
	if err != nil {
		return nil, err
	}
	var ioStat IOStats
	for _, s := range allIoStats {
		ioStat.ReadBytes += s.ReadBytes
		ioStat.WriteBytes += s.WriteBytes
		ioStat.IoTime += s.IoTime
		ioStat.CurrentIOs += s.CurrentIOs
		ioStat.WeightedIO += s.WeightedIO
		ioStat.ReadCount += s.ReadCount
		ioStat.WriteCount += s.WriteCount
		ioStat.ReadTime += ioStat.ReadTime
	}
	total := uint64(len(allIoStats))
	ioStat.WeightedIO /= total
	ioStat.IoTime /= total
	ioStat.CurrentIOs /= total
	ioStat.ReadBytes /= datasize.DataSize(total)
	ioStat.WriteBytes /= datasize.DataSize(total)
	ioStat.ReadCount /= total
	ioStat.WriteCount /= total
	ioStat.ReadTime /= total
	ioStat.WriteTime /= total
	return &ioStat, nil
}