package strategy

import (
	"RecommenderServer/schematree"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var treePath = "../testdata/10M.nt.gz.schemaTree.bin"

func TestConditions(t *testing.T) {
	/// file handling
	f, err := os.Open(treePath)
	if err != nil {
		log.Printf("Encountered error while trying to open the file: %v\n", err)
		log.Panic(err)
	}
	schema, err := schematree.Load(f, false)
	if err != nil {
		t.Errorf("Schematree could not be loaded")
	}
	pMap := schema.PropMap
	// create properties
	item1, ok1 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P31") // large number (1224) recommendations after executing on the schema tree ../testdata/10M.nt.gz.schemaTree.bin
	item2, ok2 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P21") // small number (487)
	assert.True(t, ok1 && ok2, "Expected properties wre not in the propmap")
	// create assessments
	asm1 := schematree.NewInstance(schematree.IList{item1}, schema, true)
	asm2 := schematree.NewInstance(schematree.IList{item2}, schema, true)
	asm21 := schematree.NewInstance(schematree.IList{item2, item1}, schema, true)

	// check all strategies
	countTooLessProperties := MakeTooFewRecommendationsCondition(500)
	if countTooLessProperties(asm1) || !countTooLessProperties(asm2) {
		t.Errorf("'TooLessRecommendationsCondition' failed.")
	}

	countTooManyProperties := MakeTooManyRecommendationsCondition(500)
	if !countTooManyProperties(asm1) || countTooManyProperties(asm2) {
		t.Errorf("'TooManyRecommendationsCondition' failed.")
	}

	aboveThreshholdCondition := MakeAboveThresholdCondition(1)
	if aboveThreshholdCondition(asm1) || !aboveThreshholdCondition(asm21) {
		t.Errorf("'aboveThreshholdCondition' failed.")
	}

}
