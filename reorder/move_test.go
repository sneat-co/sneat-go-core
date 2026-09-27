package reorder

import (
	"strings"
	"testing"
)

func TestMoveStr(t *testing.T) {
	arr := []string{"a", "b", "c", "d", "e"}
	for i := 0; i < len(arr); i++ {
		if actual := strings.Join(MoveStr(arr, i, i), ","); actual != "a,b,c,d,e" {
			t.Errorf("Should not mutate if from == to = %v, got: %v", i, actual)
		}
	}
	var test = func(from, to int, expects string) {
		actual := strings.Join(MoveStr(arr, from, to), ",")
		if actual != expects {
			t.Errorf("When moving from %v to %v expects %v, got: %v", from, to, expects, actual)
		}
	}
	test(1, 2, "a,c,b,d,e")
	test(1, 0, "b,a,c,d,e")
	test(2, 0, "c,a,b,d,e")
	test(2, 4, "a,b,d,e,c")
	test(0, 4, "b,c,d,e,a")
	test(4, 0, "e,a,b,c,d")
}

func TestRemoveStr(t *testing.T) {
	arr := []string{"a", "b", "c"}
	res := RemoveStr(arr, "b")
	if strings.Join(res, ",") != "a,c" {
		t.Fatalf("expected a,c, got %v", res)
	}
}

func TestInsertStr(t *testing.T) {
	arr := []string{"a", "c"}
	res := InsertStr(arr, 1, "b")
	if strings.Join(res, ",") != "a,b,c" {
		t.Fatalf("expected a,b,c, got %v", res)
	}
}

func TestIndexOf(t *testing.T) {
	arr := []string{"a", "b", "c"}
	if i := IndexOf(arr, "b"); i != 1 {
		t.Fatalf("expected 1, got %d", i)
	}
	if i := IndexOf(arr, "d"); i != -1 {
		t.Fatalf("expected -1, got %d", i)
	}
}
