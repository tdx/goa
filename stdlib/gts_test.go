package stdlib

import (
	"testing"
)

type timer struct {
	time uint64
	ref  string
}

func cmpTimer(a1, b1 interface{}) int {

	a := a1.(*timer)
	b := b1.(*timer)

	switch {
	case a.time > b.time:
		return 1
	case a.time < b.time:
		return -1
	default:
		switch {
		case a.ref > b.ref:
			return 1
		case a.ref < b.ref:
			return -1
		default:
			return 0
		}
	}
}

func TestGtsOs(t *testing.T) {

	tab := NewOrderedSet()
	if tab == nil {
		t.Fatalf("failed to create new gts")
	}

	v10t := "test10"
	tab.Insert(10, v10t)
	tab.Insert(2, "test2")

	v1 := tab.Lookup(10)
	if v1 != v10t {
		t.Fatalf("expected value '%s', actual '%s'", v10t, v1)
	}

	v2 := tab.Lookup(2)
	if v2 != "test2" {
		t.Fatalf("expected value 'test2', actual '%s'", v2)
	}

	tab.Delete(2)

	v2 = tab.Lookup(2)
	if v2 != nil {
		t.Fatalf("expected value does not exist, actual '%s'", v2)
	}
}

func TestGtsOsTraverseMany(t *testing.T) {

	tab := NewOrderedSet()
	if tab == nil {
		t.Fatalf("failed to create new gts")
	}

	test := []struct {
		key int
		val string
	}{
		{1, "t1"},
		{2, "t2"},
		{3, "t3"},
		{4, "t4"},
		{5, "t5"},
		{6, "t6"},
		{7, "t7"},
		{8, "t8"},
		{9, "t9"},
		{10, "t10"},
	}

	for k, v := range test {
		tab.Insert(k, v)
	}

	tab.Print()

	for k, v, ok := tab.First(); ok; k, v, ok = tab.Next() {
		t.Log(k, v)
	}

	for k, v, ok := tab.First(); ok; k, v, ok = tab.First() {
		t.Log(k, v)
		tab.Delete(k)
	}
	if tab.Size() != 0 {
		t.Fatalf("expected tab size 0, actual %d", tab.Size())
	}
}

func TestGtsOsDeleteInLoop(t *testing.T) {

	tab := NewOrderedSet()
	if tab == nil {
		t.Fatalf("failed to create new gts")
	}

	test := []struct {
		key int
		val string
	}{
		{1, "t1"},
		{2, "t2"},
		{3, "t3"},
		{4, "t4"},
		{5, "t5"},
		{6, "t4"},
		{7, "t7"},
		{8, "t8"},
		{9, "t4"},
		{10, "t10"},
	}

	for _, v := range test {
		tab.Insert(v.key, v.val)
	}

	tab.Print()

	for k, v, ok := tab.First(); ok; k, v, ok = tab.Next() {
		if v == "t4" {
			tab.Delete(k)
		}
	}

	tab.Print()

	if tab.Size() != 7 {
		t.Fatalf("expected tab size 0, actual %d", tab.Size())
	}

}

func TestGtsSetDeleteInLoop(t *testing.T) {

	tab := NewSet()
	if tab == nil {
		t.Fatalf("failed to create new gts")
	}

	test := []struct {
		key int
		val string
	}{
		{1, "t1"},
		{2, "t2"},
		{3, "t3"},
		{4, "t4"},
		{5, "t5"},
		{6, "t4"},
		{7, "t7"},
		{8, "t8"},
		{9, "t4"},
		{10, "t10"},
	}

	for _, v := range test {
		tab.Insert(v.key, v.val)
	}

	f := func(k, v interface{}) bool {
		if v == "t4" {
			return false
		}
		return true
	}
	tab.ForEach(f)

	if tab.Size() != 7 {
		t.Fatalf("expected tab size 0, actual %d", tab.Size())
	}

}

func TestGtsOsStruct(t *testing.T) {
	t1 := timer{51, "a51"}
	t2 := timer{13, "b13"}
	t3 := timer{12, "c12"}
	t4 := timer{36, "d36"}
	t5 := timer{12, "y12"}
	t6 := timer{13, "b12"}

	tab := NewOrderedSetWith(cmpTimer)
	if tab == nil {
		t.Fatalf("tab==nil")
	}

	tab.Insert(&t1, 1)
	tab.Insert(&t2, 2)
	tab.Insert(&t3, 3)
	tab.Insert(&t4, 4)
	tab.Insert(&t5, 5)
	tab.Insert(&t6, 6)

	tab.Print()

	for k, v, ok := tab.First(); ok; k, v, ok = tab.First() {
		t.Log(k, v)
		tab.Delete(k)
	}
	if tab.Size() != 0 {
		t.Fatalf("expected tab size 0, actual %d", tab.Size())
	}

}

func TestGtsSet(t *testing.T) {
	tab := NewSet()
	if tab == nil {
		t.Fatalf("failed to create new gts")
	}

	v10t := "test10"
	tab.Insert(10, v10t)
	tab.Insert(2, "test2")

	v1 := tab.Lookup(10)
	if v1 != v10t {
		t.Fatalf("expected value '%s', actual '%s'", v10t, v1)
	}

	v2 := tab.Lookup(2)
	if v2 != "test2" {
		t.Fatalf("expected value 'test2', actual '%s'", v2)
	}

	tab.Delete(2)

	v2 = tab.Lookup(2)
	if v2 != nil {
		t.Fatalf("expected value does not exist, actual '%s'", v2)
	}

	tab.Print()
	tab.DeleteAllObjects()
	tab.Print()
}

func BenchmarkGtsOs(b *testing.B) {

	tab := NewOrderedSet()

	insert := func(i int) { tab.Insert(i, i) }
	lookup := func(i int) { tab.Lookup(i) }
	remove := func(i int) { tab.Delete(i) }

	ops := []struct {
		name string
		fun  func(int)
	}{
		{"insert", insert},
		{"lookup", lookup},
		{"delete", remove},
	}

	b.ResetTimer()

	for _, op := range ops {
		b.Run(op.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				op.fun(i)
			}
		})
	}
}

func BenchmarkGtsSet(b *testing.B) {
	tab := NewSet()

	insert := func(i int) { tab.Insert(i, i) }
	lookup := func(i int) { tab.Lookup(i) }
	remove := func(i int) { tab.Delete(i) }

	ops := []struct {
		name string
		fun  func(int)
	}{
		{"insert", insert},
		{"lookup", lookup},
		{"delete", remove},
	}

	b.ResetTimer()

	for _, op := range ops {
		b.Run(op.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				op.fun(i)
			}
		})
	}
}
