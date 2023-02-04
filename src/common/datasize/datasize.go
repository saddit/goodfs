package datasize

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var suffixRegex = regexp.MustCompile(`^(\d+)(B|KB|MB|GB|TB|PB)?$`)

type DataSize uint64

const Step = 1024

const (
	Byte DataSize = 1
	KB            = Byte << 10
	MB            = KB << 10
	GB            = MB << 10
	TB            = GB << 10
	PB            = TB << 10
)

func (d DataSize) Byte() uint64 {
	return uint64(d)
}

func (d DataSize) KiloByte() uint64 {
	return d.Byte() >> 10
}

func (d DataSize) MegaByte() uint64 {
	return d.KiloByte() >> 10
}

func (d DataSize) GigaByte() uint64 {
	return d.MegaByte() >> 10
}

func (d DataSize) TeraByte() uint64 {
	return d.GigaByte() >> 10
}

func (d DataSize) PetaByte() uint64 {
	return d.TeraByte() >> 10
}

func (d DataSize) String() string {
	i := int(math.Floor(math.Log(float64(d)) / math.Log(1024)))
	exceed := 1.0
	if i >= len(units) {
		exceed = math.Pow(Step, float64(i-len(units)+1))
		i = len(units) - 1
	}
	num := float64(d) / math.Pow(Step, float64(i)) * exceed
	return fmt.Sprintf("%.0f%s", num, units[i].name)
}

type unit struct {
	name  string
	value DataSize
}

var units = []unit{
	{"B", Byte},
	{"KB", KB},
	{"MB", MB},
	{"GB", GB},
	{"TB", TB},
	{"PB", PB},
}

var unitNameMap = map[string]DataSize{}

func init() {
	for _, v := range units {
		unitNameMap[v.name] = v.value
	}
}

func Parse(s string) (DataSize, error) {
	s = strings.ToUpper(s)
	res := suffixRegex.FindAllStringSubmatch(s, 1)
	if len(res) == 0 || len(res[0]) < 3 {
		return 0, fmt.Errorf("data size '%v' format doesn't support", s)
	}
	num, e := strconv.Atoi(res[0][1])
	if e != nil {
		return 0, e
	}
	// if it has not unit, see as bytes
	if res[0][2] == "" {
		return DataSize(num), nil
	}
	if u, ok := unitNameMap[res[0][2]]; ok {
		if IsExceedLimit(num, u) {
			return 0, fmt.Errorf("data size '%dPB' exceed limit 1024PB", num)
		}
		return DataSize(num) * u, nil
	}
	return 0, fmt.Errorf("data size doesn't support unit '%v'", res[0][2])
}

func MustParse(s string) DataSize {
	r, e := Parse(s)
	if e != nil {
		panic(e)
	}
	return r
}

// IsExceedLimit Max size is 1023PB
func IsExceedLimit(num int, unit DataSize) bool {
	return unit == PB && num >= Step
}
