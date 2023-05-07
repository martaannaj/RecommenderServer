package schematree

import (
	"RecommenderServer/transactions"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
Test whether the whole system works by
1. creating a schematree from a CSV file
2. serializing it
3. restoring it
4. Comparing the resored tree with the original
5. Checking whether a large set of recommendations on the first schematree correspond to the originally created one
*/

func TestIntegration(t *testing.T) {
	for _, inputDataset := range []string{"../testdata/tsv-transaction-test.tsv", "../testdata/P97.tsv"} {
		description := fmt.Sprintf("integration test with %s", inputDataset)

		t.Run(description,
			func(t *testing.T) {
				s := transactions.SimpleFileTransactionSource(inputDataset)

				original_tree := Create(s)
				allNodesHaveItem(t, &original_tree.Root)
				// store
				proto_file, err := os.CreateTemp("", "schemaTree_test_integration")
				if err != nil {
					log.Fatal(err)
				}
				defer os.Remove(proto_file.Name())

				err = original_tree.SaveProtocolBuffer(proto_file)
				if err != nil {
					log.Panicf("Encountered error while saving protocol buffer the file: %v\n", err)
				}
				// load back
				_, err = proto_file.Seek(0, 0)

				if err != nil {
					log.Printf("Encountered error while seeking in the file: %v\n", proto_file)
					log.Panic(err)
				}
				new_tree, err := LoadProtocolBufferFromReader(proto_file)
				assert.NoError(t, err, "An error occured restring the schematree.")

				// These test are repeated from the saving and loading code.
				assert.EqualValues(t, original_tree.PropMap.Len(), new_tree.PropMap.Len())
				for k, v := range original_tree.PropMap.prop {
					other_v := new_tree.PropMap.prop[k]
					assert.EqualValues(t, v.Str, other_v.Str, "Propmap items do not contain the same strings")
					assert.EqualValues(t, v.TotalCount, other_v.TotalCount, "Propmap items do not preserve Totalcount")
					assert.EqualValues(t, v.SortOrder, other_v.SortOrder, "Propmap items do not preserve the SortOrder")
				}
				// we already know the maps have the same lengths, so no need to check that there are no extra items in the PropMap

				assert.EqualValues(t, original_tree.MinSup, new_tree.MinSup)
				assert.EqualValues(t, original_tree.Typed, new_tree.Typed)

				// Now compare the complete tree
				depthFirstCompare(t, &original_tree.Root, &new_tree.Root)
				// get all properties

				properties := original_tree.PropMap.list_properties()

				rng := rand.New(rand.NewSource(4655454))

				max_prop_count := len(properties)
				if max_prop_count > 30 {
					max_prop_count = 30
				}

				for count := 0; count < max_prop_count; count++ {
					for reps := 0; reps < 1000; reps++ {
						prop_indices := rng.Perm(count)[:count]
						selectedProps := make([]string, 0, count)
						for i := 0; i < count; i++ {
							selectedProps = append(selectedProps, *properties[prop_indices[i]].Str)
						}
						originalRecommendation := original_tree.Recommend(selectedProps, nil)
						restoredRecommendation := new_tree.Recommend(selectedProps, nil)
						assert.Len(t, restoredRecommendation, len(originalRecommendation))
						// the recommendations are ordered by probability, but there could be multiple entries with the same probability.
						//So, to compare, we sort them first by probaility, then lexicographically.
						sort_lexicographically(originalRecommendation)
						sort_lexicographically(restoredRecommendation)

						for i, oRec := range originalRecommendation {
							rRec := restoredRecommendation[i]
							assert.Equal(t, *oRec.Property.Str, *rRec.Property.Str)
							assert.Equal(t, oRec.Probability, rRec.Probability)
						}

					}
				}
			})
	}

}

func sort_lexicographically(recommendation PropertyRecommendations) {
	sort.Slice(recommendation, func(i, j int) bool {
		if recommendation[i].Probability == recommendation[j].Probability {
			return *recommendation[i].Property.Str < *recommendation[j].Property.Str
		}
		return recommendation[i].Probability > recommendation[j].Probability
	})
}
