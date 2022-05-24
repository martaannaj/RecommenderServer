package schematree

// This files contains the functions needed to build a schemtree from a wikidata JSON dump

import (
	"context"
	"log"
	"sort"
	"sync/atomic"

	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

// Create creates a new schema tree from given dataset
func Create(dumpfile *mediawiki.ProcessDumpConfig) *SchemaTree {

	schema := New(true, 0)
	schema.TwoPass(wikidataDumpTransactionSource(dumpfile))

	PrintMemUsage()
	return schema
}

type Transaction []string

type TransactionSource func() <-chan Transaction

func wikidataDumpTransactionSource(dumpfile *mediawiki.ProcessDumpConfig) TransactionSource {
	return func() <-chan Transaction {
		channel := make(chan Transaction)
		errE := mediawiki.ProcessWikidataDump(
			context.Background(),
			dumpfile,
			func(_ context.Context, a mediawiki.Entity) errors.E {
				t := make(Transaction, 0)

				for property_name := range a.Claims {
					t = append(t, property_name)
				}
				types_claims := a.Claims["P31"]
				for _, statement := range types_claims {
					if statement.MainSnak.SnakType == mediawiki.Value {
						if statement.MainSnak.DataValue == nil {
							log.Fatal("Found a main snak with type Value, while it does not have a value. This is an error in the dump.")
						}
						val := statement.MainSnak.DataValue.Value
						switch v := val.(type) {
						default:
							log.Printf("unexpected type %T", v)
						case mediawiki.WikiBaseEntityIDValue:
							tokenStr := typePrefix + val.(mediawiki.WikiBaseEntityIDValue).ID
							t = append(t, tokenStr)
						}
					} else {
						log.Printf("Found a type statement without a value: %v", statement)
					}
				}
				channel <- t
				return nil
			},
		)
		if errE != nil {
			log.Panicln("Something went wrong while processing..", errE)
		}
		return channel
	}
}

// TwoPass constructs a SchemaTree from the transactions using a two-pass approach, i.e., the source is called twice to get the transactions
func (tree *SchemaTree) TwoPass(sourceProvider TransactionSource) {
	tree.firstPass(sourceProvider())
	tree.updateSortOrder()
	tree.secondPass(sourceProvider())
}

// first pass: collect IItems and statistics
func (tree *SchemaTree) firstPass(source <-chan Transaction) {
	// log.Printf("Starting first pass for %v\n", dumpfile)
	itemCount := uint64(0)

	for v := range source {
		atomic.AddUint64(&itemCount, uint64(1))
		for _, name := range v {
			predicate := tree.PropMap.get(name)
			predicate.increment()
		}
	}

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

func (tree *SchemaTree) secondPass(source <-chan Transaction) {
	log.Println("Start of the second pass")
	for transaction := range source {
		properties := make([]*IItem, 0)
		for _, name := range transaction {
			predicate := tree.PropMap.get(name)
			properties = append(properties, predicate)
		}
		tree.Insert(properties)
	}
	log.Println("Second Pass ended")
	PrintMemUsage()
}

// updateSortOrder updates iList according to actual frequencies
// calling this directly WILL BREAK non-empty schema trees
// Runtime: O(n*log(n)), Memory: O(n)
func (tree *SchemaTree) updateSortOrder() {
	// make a list of all known properties
	// Runtime: O(n), Memory: O(n)
	iList := make(IList, len(tree.PropMap))
	i := 0
	for _, v := range tree.PropMap {
		iList[i] = v
		i++
	}

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
