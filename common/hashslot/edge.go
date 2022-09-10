package hashslot

import "fmt"

type Edge struct {
	Start int
	End   int
	Value string
}

func (e Edge) String() string {
	return fmt.Sprintf("%s:(%d-%d)", e.Value, e.Start, e.End)
}

type EdgeList []*Edge

func (el EdgeList) Swap(i, j int) {
	el[i], el[j] = el[j], el[i]
}

func (el EdgeList) Less(i, j int) bool {
	if el[i].Start == el[j].Start {
		return el[i].End < el[j].End
	}
	return el[i].Start < el[j].Start
}

func (el EdgeList) Len() int {
	return len(el)
}

type IEdgeProvider interface {
	get() EdgeList
}

type EdgeProvider struct {
	edges EdgeList
}

func (e *EdgeProvider) get() EdgeList {
	return e.edges
}
