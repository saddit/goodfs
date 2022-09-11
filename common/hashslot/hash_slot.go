package hashslot

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
)

const MaxSlot = 16383

func formatInt(slot string) (int, int, error) {
	rg := strings.Split(slot, "-")
	if len(rg) != 2 {
		return 0, 0, fmt.Errorf("slot '%s' must format like '0-100'", slot)
	}
	// convert start slot
	start, err := strconv.Atoi(rg[0])
	if err != nil {
		return 0, 0, err
	} else if start < 0 {
		return 0, 0, fmt.Errorf("slot must greater than 0")
	}
	// convert end slot
	end, err := strconv.Atoi(rg[1])
	if err != nil {
		return 0, 0, err
	} else if end < 0 {
		return 0, 0, fmt.Errorf("slot must greater than 0")
	}
	return start, end, nil
}

// WrapSlots slotsMap(key='identify', value=[]string{'0-100','110-221'})
func WrapSlots(slotsMap map[string][]string) (IEdgeProvider, error) {
	if len(slotsMap) == 0 {
		return nil, fmt.Errorf("empty slotsMap")
	}
	res := make(EdgeList, 0, len(slotsMap))
	for value, slots := range slotsMap {
		for _, slot := range slots {
			start, end, err := formatInt(slot)
			if err != nil {
				return nil, err
			}
			res = append(res, &Edge{
				Start: start,
				End:   end,
				Value: value,
			})
		}
	}
	sort.Sort(res)
	// validate overlap
	for i, edge := range res {
		if i == 0 {
			continue
		}
		pre := res[i-1]
		if pre.Start == edge.Start || pre.End > edge.Start {
			return nil, fmt.Errorf("overlap %s and %s", pre, edge)
		}
	}
	return &EdgeProvider{res}, nil
}

func GetStringIdentify(str string, provide IEdgeProvider) (string, error) {
	return GetDataIdentify([]byte(str), provide)
}

func GetDataIdentify(data []byte, provider IEdgeProvider) (string, error) {
	return GetSlotIdentify(CalcBytesSlot(data), provider)
}

func CalcBytesSlot(bt []byte) int {
	return int(crc32.ChecksumIEEE(bt) & MaxSlot)
}

func GetSlotIdentify(slot int, provider IEdgeProvider) (string, error) {
	edges := provider.get()
	idx := sort.Search(len(edges), func(i int) bool {
		return edges[i].Start >= slot
	})
	if idx == len(edges) {
		idx -= 1
	}
	if edges[idx].Start <= slot && edges[idx].End > slot {
		return edges[idx].Value, nil
	}
	return "", fmt.Errorf("slots assigment error, cannot find slot %d", slot)
}

func CopyOfEdges(identify string, provider IEdgeProvider) EdgeList {
	var res EdgeList
	list := provider.get()
	for _, v := range list {
		if v.Value == identify {
			res = append(res, &Edge{
				Start: v.Start,
				End:   v.End,
				Value: v.Value,
			})
		}
	}
	return res
}

func IsSlotInEdges(slot int, edges EdgeList) bool {
	idx := sort.Search(len(edges), func(i int) bool {
		return edges[i].Start >= slot
	})
	if idx == len(edges) {
		idx -= 1
	}
	return edges[idx].Start <= slot && edges[idx].End > slot
}

// FindRangeCurrentData find identifies in range [start,end)
func FindRangeCurrentData(start, end int, provider IEdgeProvider) (res EdgeList) {
	edges := provider.get()
	for _, edge := range edges {
		if edge.Start <= start {
			if edge.End < end {
				res = append(res, &Edge{start, edge.End, edge.Value})
			} else {
				res = append(res, &Edge{start, end, edge.Value})
				return
			}
		} else if edge.Start < end {
			if edge.End < end {
				res = append(res, &Edge{edge.Start, edge.End, edge.Value})
			} else {
				res = append(res, &Edge{edge.Start, end, edge.Value})
			}
		}
	}
	return
}
