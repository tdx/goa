package stdlib

//
// Implements Set
// Iterator returns nil values always
//

import (
	"fmt"
)

//
// Set
//
type gtsS struct {
	set map[Term]Term
}

func newSet() Gts {
	t := new(gtsS)
	t.set = make(map[Term]Term)
	return t
}

func (gts *gtsS) Insert(key Term, value Term) {
	gts.set[key] = value
}

func (gts *gtsS) Delete(key Term) {
	delete(gts.set, key)
}

func (gts *gtsS) Lookup(key Term) Term {
	value, found := gts.set[key]
	if found {
		return value
	}
	return nil
}

func (gts *gtsS) Size() int {
	return len(gts.set)
}

func (gts *gtsS) DeleteAllObjects() {
	gts.set = make(map[Term]Term)
}

func (gts *gtsS) Print() {
	fmt.Println(gts.set)
}

//
// Iterator
//
func (gts *gtsS) First() (Term, Term, bool) {
	return nil, nil, false
}

func (gts *gtsS) Last() (Term, Term, bool) {
	return nil, nil, false
}

func (gts *gtsS) Next() (Term, Term, bool) {
	return nil, nil, false
}

func (gts *gtsS) Prev() (Term, Term, bool) {
	return nil, nil, false
}

func (gts *gtsS) ForEach(f GtsForEach) {
	for k, v := range gts.set {
		if !f(k, v) {
			delete(gts.set, k)
		}
	}
}
