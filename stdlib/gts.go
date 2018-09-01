package stdlib

//
// gts - Golang term store. Implementation is not thread-safe
//

//
// GtsKeysComparator is a function for ordered set to compare keys
//
type GtsKeysComparator func(a, b interface{}) int

//
// GtsForEach is an iterator function type to iterate over all keys in table
//
type GtsForEach func(a, b interface{}) bool

//
// GtsIterator is the interface that difines functions to iterater over the table
//
type GtsIterator interface {
	First() (Term, Term, bool)
	Last() (Term, Term, bool)
	Next() (Term, Term, bool)
	Prev() (Term, Term, bool)
	ForEach(GtsForEach)
}

//
// Gts is the interface to manipulate values in the table
//
type Gts interface {
	Insert(key Term, value Term)
	Delete(key Term)
	Lookup(key Term) Term
	Size() int
	DeleteAllObjects()
	Print()

	GtsIterator
}

//
// NewSet makes new set and returns table object to manipulate
//
func NewSet() Gts {
	return newSet()
}

//
// NewOrderedSet makes new ordered set With int keys comparator
//
func NewOrderedSet() Gts {
	return newOrderedSet()
}

//
// NewOrderedSetWith makes new ordered set with specified keys compare function
//
func NewOrderedSetWith(cmp GtsKeysComparator) Gts {
	return newOrderedSetWith(cmp)
}
