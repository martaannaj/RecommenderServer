package schematree

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"fortio.org/safecast"
)

// A struct capturing
// - a string IRI (str) and
// - its support, i.e. its total number of occurrences (totalCount)
// - an integer indicating sort order
type IItem struct {
	Str              *string
	TotalCount       uint64
	SortOrder        uint32
	traversalPointer *SchemaNode // node traversal pointer
}

func (p *IItem) increment() {
	atomic.AddUint64(&p.TotalCount, 1)
}

// prefix t# identifies properties that represent types
const typePrefix = "t#"

func (p *IItem) IsType() bool {
	return strings.HasPrefix(*p.Str, typePrefix)
}

func (p *IItem) IsProp() bool {
	return !strings.HasPrefix(*p.Str, typePrefix)
}

func (p IItem) String() string {
	return fmt.Sprint(p.TotalCount, "x\t", *p.Str, " (", p.SortOrder, ")")
}

type propMap struct {
	prop     map[string]*IItem
	propLock *sync.RWMutex
}

func NewPropMap() propMap {
	a := propMap{
		prop:     make(map[string]*IItem),
		propLock: new(sync.RWMutex),
	}
	return a
}

// Note: get() is a mutator function. If no property is found with that iri, then it
// will build a new property, mutate the propMap to include it, and return the
// newly created property. The returned `item` is guaranteed to be non-null.
//
// thread-safe
func (m propMap) Get_or_create(iri string) (item *IItem) { // TODO: Implement sameas Mapping/Resolution to single group identifier upon insert!
	m.propLock.RLock()
	item, ok := m.prop[iri]
	m.propLock.RUnlock()
	if !ok {
		m.propLock.Lock()
		defer m.propLock.Unlock()

		// recheck existence - might have been created by other thread
		if item, ok = m.prop[iri]; ok {
			return
		}
		item = &IItem{&iri, 0, safecast.MustConvert[uint32](len(m.prop)), nil}
		m.prop[iri] = item
	}
	return
}

func (m propMap) GetIfExisting(iri string) (item *IItem, ok bool) { // TODO: Implement sameas Mapping/Resolution to single group identifier upon insert!
	m.propLock.RLock()
	item, ok = m.prop[iri]
	m.propLock.RUnlock()
	return
}

// Get the *IItem for the given iri, but does not take care of locking. It is up to the caller to ensure that no concurrent writing is ongoing.
func (m propMap) noWritersGet(iri string) (item *IItem, ok bool) {
	item, ok = m.prop[iri]
	return
}

// Get the number of properties and types (in that order). This does not include the root node.
func (p propMap) count() (int, int) {
	p.propLock.RLock()
	defer p.propLock.RUnlock()
	props := 0
	types := 0
	for _, item := range p.prop {
		if item.IsType() {
			types++
		} else {
			props++
		}
	}
	// remove the root from the properties
	return props - 1, types
}

func (p propMap) Len() (length int) {
	p.propLock.RLock()
	length = len(p.prop)
	p.propLock.RUnlock()
	return
}

// noWritersList_properties returns the list of properties in this propMap
// The caller *must* guarantee that no concurrent write operations are taking place!
func (p propMap) noWritersList_properties() []*IItem {
	all := make([]*IItem, 0, len(p.prop))
	for _, item := range p.prop {
		all = append(all, item)
	}
	return all
}

func (p propMap) list_properties() []*IItem {
	p.propLock.RLock()
	defer p.propLock.RUnlock()
	return p.noWritersList_properties()
}

// An array of pointers to IRI structs
type IList []*IItem

// Sort the list according to the current iList Sort order
func (l IList) Sort() {
	sort.Slice(l, func(i, j int) bool { return l[i].SortOrder < l[j].SortOrder })
}

// inplace sorting and deduplication.
func (l *IList) sortAndDeduplicate() {
	ls := *l

	ls.Sort()

	// inplace deduplication
	j := 0
	for i := 1; i < len(ls); i++ {
		if ls[j] == ls[i] {
			continue
		}
		j++
		ls[i], ls[j] = ls[j], ls[i]
	}
	*l = ls[:j+1]
}

func (l IList) toSet() map[*IItem]bool {
	pSet := make(map[*IItem]bool, len(l))
	for _, p := range l {
		pSet[p] = true
	}
	return pSet
}

func (l IList) String() string {
	//// list representation (includes duplicates)
	o := "[ "
	for i := 0; i < len(l); i++ {
		o += *(l[i].Str) + " "
	}
	return o + "]"

	//// couter presentation (loses order)
	// ctr := make(map[string]int)
	// for i := 0; i < len(p); i++ {
	// 	ctr[p[i].str]++
	// }
	// return fmt.Sprint(ctr)
}
