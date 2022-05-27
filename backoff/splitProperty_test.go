package backoff

import (
	ST "RecommenderServer/schematree"
	"os"
	"testing"
)

func getTestSchema(t *testing.T, file_name string) (*ST.SchemaTree, error) {
	f, err := os.Open(file_name)
	if err != nil {
		t.Errorf("Encountered error while trying to open the file: %v\n", err)
	}
	return ST.Load(f, false)
}

func TestRecommender(t *testing.T) {

	schema, err := getTestSchema(t, treePath)

	if err != nil {
		t.Errorf("Schematree could not be loaded")
	}
	pMap := schema.PropMap
	b := BackoffSplitPropertySet{}
	b.init(schema, TwoSupportRangesSplitter, DummyMerger)

	prop1, ok1 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P31")
	prop2, ok2 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P21")
	prop3, ok3 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P27")
	if !(ok1 && ok2 && ok3) {
		t.Error("Expected properties were not in the propertyMap")
	}

	props := ST.IList{prop1, prop2, prop3}

	b.Recommend(props)

}

func TestAvgMerger(t *testing.T) {
	schema, err := getTestSchema(t, treePath)

	if err != nil {
		t.Errorf("Schematree could not be loaded")
	}
	pMap := schema.PropMap
	b := BackoffSplitPropertySet{}
	b.init(schema, TwoSupportRangesSplitter, DummyMerger)

	prop1, ok1 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P31")
	prop2, ok2 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P21")
	prop3, ok3 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P27")
	if !(ok1 && ok2 && ok3) {
		t.Error("Expected properties were not in the propertyMap")
	}

	rec1 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop1, Probability: 0.2}, ST.RankedPropertyCandidate{Property: prop2, Probability: 0.5}}
	rec2 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop1, Probability: 0.8}, ST.RankedPropertyCandidate{Property: prop3, Probability: 0.4}}
	rec3 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop2, Probability: 0.2}}
	rec4 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop2, Probability: 0.3}}

	recommendations := []ST.PropertyRecommendations{rec1, rec2, rec3, rec4}

	res := AvgMerger(recommendations)

	for _, r := range res {
		// Test values
		if *(r.Property.Str) == *(prop1.Str) && r.Probability != float64(0.25) {
			t.Errorf("Property 1 should have probability 0.25 but has %f", r.Probability)
		} else if *r.Property.Str == *prop2.Str && r.Probability != float64(0.25) {
			t.Errorf("Property 2 should have probability 0.25 but has %f", r.Probability)
		} else if *r.Property.Str == *prop3.Str && r.Probability != float64(0.1) {
			t.Errorf("Property 3 should have probability 0.1 but has %f", r.Probability)
		}
	}
}

func TestMaxMerger(t *testing.T) {
	schema, err := getTestSchema(t, treePath)

	if err != nil {
		t.Errorf("Schematree could not be loaded")
	}
	pMap := schema.PropMap
	b := BackoffSplitPropertySet{}
	b.init(schema, TwoSupportRangesSplitter, DummyMerger)

	prop1, ok1 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P31")
	prop2, ok2 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P21")
	prop3, ok3 := pMap.GetIfExisting("http://www.wikidata.org/prop/direct/P27")
	if !(ok1 && ok2 && ok3) {
		t.Error("Expected properties were not in the propertyMap")
	}

	rec1 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop1, Probability: 0.2}, ST.RankedPropertyCandidate{Property: prop2, Probability: 0.5}}
	rec2 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop1, Probability: 0.8}, ST.RankedPropertyCandidate{Property: prop3, Probability: 0.4}}
	rec3 := ST.PropertyRecommendations{ST.RankedPropertyCandidate{Property: prop2, Probability: 0.2}}

	recommendations := []ST.PropertyRecommendations{rec1, rec2, rec3}

	res := MaxMerger(recommendations)

	for _, r := range res {
		// Test values
		if *(r.Property.Str) == *(prop1.Str) && r.Probability != 0.8 {
			t.Errorf("Property 1 should have probability 0.8 but has %f", r.Probability)
		} else if *r.Property.Str == *prop2.Str && r.Probability != 0.5 {
			t.Errorf("Property 2 should have probability 0.5 but has %f", r.Probability)
		} else if *r.Property.Str == *prop3.Str && r.Probability != 0.4 {
			t.Errorf("Property 3 should have probability 0.4 but has %f", r.Probability)
		}
	}
}
