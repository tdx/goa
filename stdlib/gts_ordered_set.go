package stdlib

//
// Implementation of Gts ordeterd set
//

import (
	"fmt"

	"github.com/emirpasic/gods/containers"
	avl "github.com/emirpasic/gods/trees/avltree"
	"github.com/emirpasic/gods/utils"
)

//
// Ordered Set
//
type gtsOs struct {
	tree *avl.Tree
	it   containers.ReverseIteratorWithKey
}

func newOrderedSet() Gts {
	t := new(gtsOs)
	t.tree = avl.NewWithIntComparator()
	return t
}

func newOrderedSetWith(cmp GtsKeysComparator) Gts {
	t := new(gtsOs)
	t.tree = avl.NewWith(utils.Comparator(cmp))
	return t
}

//
// Gts
//
func (gts *gtsOs) Insert(key Term, value Term) {
	gts.tree.Put(key, value)
	gts.it = nil
}

func (gts *gtsOs) Delete(key Term) {
	gts.tree.Remove(key)
	gts.it = nil
}

func (gts *gtsOs) Lookup(key Term) Term {
	value, found := gts.tree.Get(key)
	if found {
		return value
	}
	return nil
}

func (gts *gtsOs) Size() int {
	return gts.tree.Size()
}

func (gts *gtsOs) DeleteAllObjects() {
	gts.tree.Clear()
	gts.it = nil
}

func (gts *gtsOs) Print() {
	fmt.Println(gts.tree)
}

//
// Iterator
//
func (gts *gtsOs) First() (Term, Term, bool) {
	if gts.it == nil {
		gts.it = gts.tree.Iterator()
	}
	if gts.it.First() {
		return gts.it.Key(), gts.it.Value(), true
	}
	return nil, nil, false
}

func (gts *gtsOs) Last() (Term, Term, bool) {
	if gts.it == nil {
		gts.it = gts.tree.Iterator()
	}
	if gts.it.Last() {
		return gts.it.Key(), gts.it.Value(), true
	}
	return nil, nil, false
}

func (gts *gtsOs) Next() (Term, Term, bool) {
	if gts.it == nil {
		gts.it = gts.tree.Iterator()
	}
	if gts.it.Next() {
		return gts.it.Key(), gts.it.Value(), true
	}
	return nil, nil, false
}

func (gts *gtsOs) Prev() (Term, Term, bool) {
	if gts.it == nil {
		gts.it = gts.tree.Iterator()
	}
	if gts.it.Prev() {
		return gts.it.Key(), gts.it.Value(), true
	}
	return nil, nil, false
}

func (gts *gtsOs) ForEach(f GtsForEach) {
	gts.it = nil
	for k, v, ok := gts.First(); ok; k, v, ok = gts.Next() {
		if !f(k, v) {
			gts.Delete(k)
		}
	}

}
