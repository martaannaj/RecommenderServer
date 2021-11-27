package server

import (
	"RecommenderServer/schematree"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(400)
			fmt.Println("Malformed Request.") // TODO: Json-Schema helps
			return
		}
		fmt.Println(input)
		instance := schematree.NewInstanceFromInput(input.Properties, input.Types, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		rec := workflow.Recommend(instance)
		fmt.Println(time.Since(t1))

		// Put a hard limit on the recommendations returned
		propsCount := 0
		limit := 0
		for i, rec := range rec {
			if rec.Property.IsProp() {
				propsCount += 1
				if propsCount >= hardLimit {
					limit = i
					break
				}
			}
		}
		if limit == 0 {
			limit = len(rec) - 1
		}
		if len(rec) > limit {
			rec = rec[:limit]
		}

		outputRecs := make([]RecommendationOutputEntry, propsCount-1)
		i := 0
		for _, rec := range rec {
			if rec.Property.IsProp() {
				outputRecs[i].PropertyStr = rec.Property.Str
				outputRecs[i].Probability = rec.Probability
				i += 1
			}
		}

		// Pack everything into the response
		recResp := RecommenderResponse{Recommendations: outputRecs}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(recResp)
		if err != nil {
			fmt.Println("Malformed Response.")
			return
		}
	}
}

// SetupEndpoints configures a router with all necessary endpoints and their corresponding handlers.
// func SetupEndpoints(model *schematree.SchemaTree, glossary *glossary.Glossary, workflow *strategy.Workflow, hardLimit int) http.Handler {
func SetupEndpoints(model *schematree.SchemaTree, workflow *strategy.Workflow, hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/recommender", setupLeanRecommender(model, workflow, hardLimit))
	return router
}
