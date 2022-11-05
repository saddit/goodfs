package cpu

type Stat struct {
	UsedPercent   float64 `msg:"used_percent" json:"usedPercent"`
	LogicalCount  int     `msg:"logical_count" json:"logicalCount"`
	PhysicalCount int     `msg:"physical_count" json:"physicalCount"`
}
