package schematree

import (
	"RecommenderServer/transactions"
	"sort"
	"testing"

	"log"
	"os"

	"github.com/stretchr/testify/assert"
)

var typedTreepath = "../testdata/10M.nt.gz.schemaTree.typed.bin"

// contains checks if a recommendation is in property recommendations with a probability threshold included.
func (ps PropertyRecommendations) contains(str string, prob float64) bool {
	for _, a := range ps {
		if *a.Property.Str == str && a.Probability >= prob {
			return true
		}
	}
	return false
}

func TestRecommend(t *testing.T) {
	f, err := os.Open(typedTreepath)
	if err != nil {
		log.Printf("Encountered error while trying to open the file: %v\n", err)
		log.Panic(err)
	}
	tree, _ := Load(f, false)

	t.Run("one type", func(t *testing.T) {
		list := tree.Recommend([]string{}, []string{"http://www.wikidata.org/entity/Q515"}) // City
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.9))       // country
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.9))      // coordinate location
	})

	t.Run("one property", func(t *testing.T) {
		list := tree.Recommend([]string{"http://www.wikidata.org/prop/direct/P31"}, []string{}) // InstanceOf
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.5))          // country
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.5))         // coordinate location
	})

}

func TestRecommendProperty(t *testing.T) {
	f, err := os.Open(typedTreepath)
	if err != nil {
		log.Printf("Encountered error while trying to open the file: %v\n", err)
		log.Panic(err)
	}
	tree, _ := Load(f, false)
	pMap := tree.PropMap

	t.Run("Only type property", func(t *testing.T) {
		p, ok := pMap.GetIfExisting("t#http://www.wikidata.org/entity/Q515")
		assert.True(t, ok, "Expected property was not in the pmap")
		list := tree.RecommendProperty(IList{p})                                       // City
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.9))  // country
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.9)) // coordinate location
	})

	t.Run("Only common property", func(t *testing.T) {
		p, ok := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P31")
		assert.True(t, ok, "Expected property was not in the pmap")
		list := tree.RecommendProperty(IList{p})                                        // InstanceOf
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.5))  // country
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.5)) // coordinate location
	})

}

func checkRecommendations(
	t *testing.T,
	recs PropertyRecommendations,
	expected []struct { // does not have to be sorted
		id   string
		prob float64
	}) {
	// Check whether the PropertyRecommendations make sense: no duplicates, the right number, and sorted
	ids := make(map[string]bool)
	for _, rec := range recs {
		assert.Falsef(t, ids[*rec.Property.Str], "A property with string value %s occurs at least twice in this recommendation", rec.Property.Str)
		ids[*rec.Property.Str] = true
	}
	assert.Len(t, recs, len(expected), "Incorrect number of recommendations")
	assert.True(t, sort.SliceIsSorted(recs, func(i, j int) bool { return recs[i].Probability > recs[j].Probability }), "recommendations are not correctly sorted")
	// given the recommendations are correctly sorted, we can now find all the expected ones

	for _, expRec := range expected {
		firstOption := sort.Search(len(recs), func(i int) bool { return recs[i].Probability <= expRec.prob }) // note: the recs are descending
		if firstOption == len(recs) {
			// not found
			assert.FailNowf(t, "missing recommedation", "Could not find the recommendation %v in the recommendations %v", expRec, recs)
		}
		// go over all with the correct probability to find a hit
		found := false
		for option := firstOption; recs[option].Probability == expRec.prob; option++ {
			if *recs[option].Property.Str == expRec.id {
				found = true
				break
			}
		}
		assert.True(t, found, "The recommendation %v could not be found in %v", expRec, recs)
	}
}

func TestSpecificRecommendation(t *testing.T) {
	inputDataset := "../testdata/tsv-transaction-test.tsv"
	description := "specific test with tsv-transaction-test.tsv"

	s := transactions.SimpleFileTransactionSource(inputDataset)
	tree := Create(s)

	t.Run(description+" a,b,c", func(t *testing.T) {
		recs := tree.Recommend([]string{"a", "b", "c"}, nil)
		expected := []struct {
			id   string
			prob float64
		}{{id: "d", prob: 0.5}}
		checkRecommendations(t, recs, expected)
	})
	t.Run(description+" b", func(t *testing.T) {
		recs := tree.Recommend([]string{"b"}, nil)
		expected := []struct {
			id   string
			prob float64
		}{
			{id: "a", prob: 0.8},
			{id: "c", prob: 0.6},
			{id: "d", prob: 0.4},
			{id: "e", prob: 0.2},
		}
		checkRecommendations(t, recs, expected)
	})
	t.Run(description+" e", func(t *testing.T) {
		recs := tree.Recommend([]string{"e"}, nil)
		expected := []struct {
			id   string
			prob float64
		}{
			{id: "a", prob: 0.5},
			{id: "c", prob: 1.0},
			{id: "b", prob: 0.5},
		}
		checkRecommendations(t, recs, expected)
	})
	emptyExpected := []struct {
		id   string
		prob float64
	}{
		{id: "a", prob: 5.0 / 6},
		{id: "b", prob: 5.0 / 6},
		{id: "c", prob: 4.0 / 6},
		{id: "d", prob: 2.0 / 6},
		{id: "e", prob: 2.0 / 6},
	}
	t.Run(description+" empty", func(t *testing.T) {
		recs := tree.Recommend(nil, nil)
		checkRecommendations(t, recs, emptyExpected)
	})
	t.Run(description+" root", func(t *testing.T) {
		root := tree.Root.ID.Str
		recs := tree.Recommend([]string{*root}, nil)
		checkRecommendations(t, recs, emptyExpected)
	})
	t.Run(description+" non-existing", func(t *testing.T) {
		recs := tree.Recommend([]string{"f"}, nil)
		checkRecommendations(t, recs, emptyExpected)
	})
}
