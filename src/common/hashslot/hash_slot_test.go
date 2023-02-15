package hashslot

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCombineEdges(t *testing.T) {
	slots1 := []string{
		"0-100",
		"120-500",
	}
	slots2 := []string{
		"100-110",
		"115-120",
	}
	e1, err := WrapSlotsToEdges(slots1, "A")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := WrapSlotsToEdges(slots2, "A")
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).EqualValues([]string{"0-110", "115-500"}, CombineEdges(e1, e2).Strings())
}

func TestRemoveEdges(t *testing.T) {
	slots1 := []string{
		"0-100",
		"120-500",
	}
	slots2 := []string{
		"10-100",
		"125-200",
	}
	e1, err := WrapSlotsToEdges(slots1, "A")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := WrapSlotsToEdges(slots2, "A")
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).EqualValues([]string{"0-10", "120-125", "200-500"}, RemoveEdges(e1, e2).Strings())
}

func TestRemoveEdges2(t *testing.T) {
	slots1 := []string{
		"0-16384",
	}
	slots2 := []string{
		"3000-4000",
		"6000-10000",
	}
	e1, err := WrapSlotsToEdges(slots1, "A")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := WrapSlotsToEdges(slots2, "A")
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).EqualValues([]string{"0-3000", "4000-6000", "10000-16384"}, RemoveEdges(e1, e2).Strings())
}

func TestRemoveEdges3(t *testing.T) {
	slots1 := []string{
		"0-8000",
		"9000-9500",
	}
	slots2 := []string{
		"7000-9200",
	}
	e1, err := WrapSlotsToEdges(slots1, "A")
	if err != nil {
		t.Fatal(err)
	}
	e2, err := WrapSlotsToEdges(slots2, "A")
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).EqualValues([]string{"0-7000", "9200-9500"}, RemoveEdges(e1, e2).Strings())
}

func TestFindRangeCurrentData(t *testing.T) {
	sm := map[string][]string{
		"A": {"0-100", "110-115"},
		"B": {"120-500"},
		"C": {"115-120"},
	}
	p, err := WrapSlots(sm)
	if err != nil {
		t.Fatal(err)
	}
	// assert A:(110-115) C:(115-120) B:(120-130)
	expect := EdgeList{
		{110, 115, "A"},
		{115, 120, "C"},
		{120, 130, "B"},
	}
	actual, _ := FindRangeCurrentData(110, 130, p)
	assert.New(t).EqualValues(fmt.Sprint(expect), fmt.Sprint(actual))
}

func TestIsValidEdge(t *testing.T) {
	sm := map[string][]string{
		"A": {"0-100", "110-115"},
		"B": {"120-500"},
		"C": {"115-120"},
	}
	p, err := WrapSlots(sm)
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).True(IsValidEdge(&Edge{110, 115, "A"}, p))
	assert.New(t).False(IsValidEdge(&Edge{100, 115, "A"}, p))
	assert.New(t).True(IsValidEdge(&Edge{115, 120, "C"}, p))
	assert.New(t).False(IsValidEdge(&Edge{114, 116, "C"}, p))
}

func TestIsSlotInEdges(t *testing.T) {
	p, err := WrapSlotsToEdges([]string{"0-100", "110-115"}, "A")
	if err != nil {
		t.Fatal(err)
	}

	assert.New(t).True(IsSlotInEdges(0, p))
	assert.New(t).True(IsSlotInEdges(110, p))
	assert.New(t).True(IsSlotInEdges(50, p))
	assert.New(t).True(IsSlotInEdges(112, p))

	assert.New(t).False(IsSlotInEdges(115, p))
	assert.New(t).False(IsSlotInEdges(100, p))
	assert.New(t).False(IsSlotInEdges(105, p))
	assert.New(t).False(IsSlotInEdges(120, p))
}
