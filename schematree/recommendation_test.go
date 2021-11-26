package schematree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
	"fmt"
	"log"
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
		fmt.Printf("Encountered error while trying to open the file: %v\n", err)
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
		fmt.Printf("Encountered error while trying to open the file: %v\n", err)
		log.Panic(err)
	}
	tree, _ := Load(f, false)
	pMap := tree.PropMap

	t.Run("Only type property", func(t *testing.T) {
		list := tree.RecommendProperty(IList{pMap.get("t#http://www.wikidata.org/entity/Q515")}) // City
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.9))            // country
		assert.True(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.9))           // coordinate location
	})

	t.Run("Only common property", func(t *testing.T) {
		list := tree.RecommendProperty(IList{pMap.get("http://www.wikidata.org/prop/direct/P31")}) // InstanceOf
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P17", 0.5))             // country
		assert.False(t, list.contains("http://www.wikidata.org/prop/direct/P625", 0.5))            // coordinate location
	})

}
