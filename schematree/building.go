package schematree

// This files contains the functions needed to build a schemtree from a wikidata JSON dump

import (
	"log"
	"sort"
	"sync"
	"sync/atomic"

	"RecommenderServer/transactions"
)

// Create creates a new schema tree from given dataset
func Create(sourceProvider transactions.TransactionSource) *SchemaTree {

	schema := New(true, 0)
	schema.TwoPass(sourceProvider)

	PrintMemUsage()
	return schema
}

// TwoPass constructs a SchemaTree from the transactions using a two-pass approach, i.e., the source is called twice to get the transactions
func (tree *SchemaTree) TwoPass(sourceProvider transactions.TransactionSource) {
	tree.firstPass(sourceProvider())
	tree.updateSortOrder()
	tree.secondPass(sourceProvider())
}

// first pass: collect IItems and statistics
func (tree *SchemaTree) firstPass(source <-chan transactions.Transaction) {
	// log.Printf("Starting first pass for %v\n", dumpfile)
	itemCount := uint64(0)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range source {
				atomic.AddUint64(&itemCount, uint64(1))
				amount := atomic.LoadUint64(&itemCount)
				if amount%10000 == 0 {
					log.Printf("Processed %d entities", amount)
				}
				for _, name := range v {
					predicate := tree.PropMap.Get_or_create(name)
					predicate.increment()
				}
			}
		}()
	}
	wg.Wait()
	propCount, typeCount := tree.PropMap.count()

	log.Printf("%v subjects, %v properties, %v types\n", itemCount, propCount, typeCount)

	log.Println("First Pass done")
	PrintMemUsage()

	const MaxUint32 = uint64(^uint32(0))
	if itemCount > MaxUint32 {
		log.Print("\n#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#\n\n")
		log.Printf("WARNING: uint32 OVERFLOW - Processed %v subjects but tree can only track support up to %v!\n", itemCount, MaxUint32)
		log.Print("\n#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#!#\n\n")
		log.Panic("After this overflow all results will be invalid")
	}
}

func (tree *SchemaTree) secondPass(source <-chan transactions.Transaction) {
	log.Println("Start of the second pass")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for transaction := range source {
				properties := make([]*IItem, 0)
				for _, name := range transaction {
					predicate, ok := tree.PropMap.GetIfExisting(name)
					if !ok {
						log.Panic("During the second pass, found a predicate which was not yet in the propmap while this must have been added during the first pass.")
					}
					properties = append(properties, predicate)
				}
				tree.Insert(properties)
			}
		}()
	}
	wg.Wait()
	log.Println("Second Pass ended")
	PrintMemUsage()
}

// updateSortOrder updates iList according to actual frequencies
// calling this directly WILL BREAK non-empty schema trees
// Runtime: O(n*log(n)), Memory: O(n)
func (tree *SchemaTree) updateSortOrder() {
	// make a list of all known properties
	// Runtime: O(n), Memory: O(n)

	iList := tree.PropMap.list_properties()

	// sort by descending support. In case of equal support, lexicographically
	// Runtime: O(n*log(n)), Memory: -
	sort.Slice(
		iList,
		func(i, j int) bool {
			if iList[i].TotalCount != iList[j].TotalCount {
				return iList[i].TotalCount > iList[j].TotalCount
			}
			return *(iList[i].Str) < *(iList[j].Str)
		})

	// update term's internal sortOrder
	// Runtime: O(n), Memory: -
	for i, v := range iList {
		v.SortOrder = uint32(i)
	}
}
