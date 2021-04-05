package server

import (
	"RecommenderServer/schematree"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"RecommenderServer/assessment"
	"RecommenderServer/strategy"
)

type RecommenderRequest struct {
	Types      []string `json:"types"`
	Properties []string `json:"properties"`
}

// RecommenderResponse is the data representation of the json.
type RecommenderResponse struct {
	Recommendations []RecommendationOutputEntry `json:"recommendations"`
}

// RecommendationOutputEntry is each entry that is return from the server.
type RecommendationOutputEntry struct {
	PropertyStr *string `json:"property"`
	Probability float64 `json:"probability"`
}

// setupRecommender will setup a handler to recommend properties based on the list of properties and types.
// It will return an array of recommendations with their respective probabilities.
// No gloassary information is added to the response.
func setupLeanRecommender(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	hardLimit int,
) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		// Decode the JSON input and build a list of input strings
		var input = RecommenderRequest{}
		// var input []string
		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.Write([]byte("Malformed Request.")) // TODO: Json-Schema helps
			return
		}
		fmt.Println(input)
		assessment := assessment.NewInstanceFromInput(input.Properties, input.Types, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		rec := workflow.Recommend(assessment)
		fmt.Println(time.Since(t1))

		// Put a hard limit on the recommendations returned.
		if len(rec) > hardLimit {
			rec = rec[:hardLimit]
		}

		outputRecs := make([]RecommendationOutputEntry, len(rec), len(rec))
		for i, rec := range rec {
			outputRecs[i].PropertyStr = rec.Property.Str
			outputRecs[i].Probability = rec.Probability
		}

		// Pack everything into the response
		recResp := RecommenderResponse{Recommendations: outputRecs}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(recResp)
	}
}

func setupLeanRecommender(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	hardLimit int,
) func(http.ResponseWriter, *http.Request) {

	// Fetch the map of all properties in the SchemaTree
	pMap := model.PropMap

	return func(res http.ResponseWriter, req *http.Request) {

		// Decode the JSON input and build a list of input strings
		var properties []string
		err := json.NewDecoder(req.Body).Decode(&properties)
		if err != nil {
			res.Write([]byte("Malformed Request. Expected an array of property IRIs"))
			return
		}
		fmt.Println(properties)

		// Match the input strings to build a list of input properties.
		list := []*schematree.IItem{}
		for _, pString := range properties {
			p, ok := pMap[pString]
			if ok {
				list = append(list, p)
			}
		}
		// fmt.Println(tree.Support(list), tree.Root.Support)

		// Make an assessment of the input properties.
		assessment := assessment.NewInstance(list, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		rec := workflow.Recommend(assessment)
		fmt.Println(time.Since(t1))

		// Put a hard limit on the recommendations returned.
		if len(rec) > 500 {
			rec = rec[:500]
		}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(rec)
	}
}

// SetupEndpoints configures a router with all necessary endpoints and their corresponding handlers.
// func SetupEndpoints(model *schematree.SchemaTree, glossary *glossary.Glossary, workflow *strategy.Workflow, hardLimit int) http.Handler {
func SetupEndpoints(model *schematree.SchemaTree, workflow *strategy.Workflow, hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/lean-recommender", setupLeanRecommender(model, workflow, hardLimit))
	return router
}
