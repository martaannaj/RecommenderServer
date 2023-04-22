package server

import (
	"RecommenderServer/schematree"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
// That handler will return an array of recommendations with their respective probabilities.
// This array will contain at most `hardlimit` resutls. To not have a limit, set to -1.
func setupLeanRecommender(
	model *schematree.SchemaTree,
	workflow *strategy.Workflow,
	hardLimit int,
) func(http.ResponseWriter, *http.Request) {
	if model == nil {
		log.Panicln("Nil model specified")
	}
	if workflow == nil {
		log.Panicln("Nil workflow specified")
	}
	if hardLimit < 1 || hardLimit != -1 {
		log.Panic("")
	}

	return func(res http.ResponseWriter, req *http.Request) {

		// Decode the JSON input and build a list of input strings
		var input = RecommenderRequest{}

		err := json.NewDecoder(req.Body).Decode(&input)
		if err != nil {
			res.WriteHeader(400)
			log.Println("Malformed Request.") // TODO: Json-Schema helps
			return
		}
		var jsonstring = fmt.Sprintln(input)
		escapedjsonstring := strings.Replace(jsonstring, "\n", "", -1)
		escapedjsonstring = strings.Replace(escapedjsonstring, "\r", "", -1)
		log.Println("request received ", escapedjsonstring)
		instance := schematree.NewInstanceFromInput(input.Properties, input.Types, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		recomendations := workflow.Recommend(instance)
		log.Println("request ", escapedjsonstring, " answered in ", time.Since(t1))

		// Put a hard limit on the recommendations returned
		outputRecs := limitRecommendations(recomendations, hardLimit)

		// Pack everything into the response
		recResp := RecommenderResponse{Recommendations: outputRecs}

		// Write the recommendations as a JSON array.
		res.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(res).Encode(recResp)
		if err != nil {
			log.Println("Malformed Response.", &recResp)
			return
		}
	}
}

func limitRecommendations(recomendations schematree.PropertyRecommendations, hardLimit int) []RecommendationOutputEntry {
	propsCount := 0
	limit := 0
	for i, recommendation := range recomendations {
		if recommendation.Property.IsProp() {
			propsCount += 1
			if propsCount >= hardLimit {
				limit = i
				break
			}
		}
	}
	if limit == 0 {
		limit = len(recomendations) - 1
	}
	if len(recomendations) > limit {
		recomendations = recomendations[:limit]
	}

	outputRecs := make([]RecommendationOutputEntry, propsCount-1)
	i := 0
	for _, rec := range recomendations {
		if rec.Property.IsProp() {
			outputRecs[i].PropertyStr = rec.Property.Str
			outputRecs[i].Probability = rec.Probability
			i += 1
		}
	}
	return outputRecs
}

// SetupEndpoints configures a router with all necessary endpoints and their corresponding handlers.
// func SetupEndpoints(model *schematree.SchemaTree, glossary *glossary.Glossary, workflow *strategy.Workflow, hardLimit int) http.Handler {
func SetupEndpoints(model *schematree.SchemaTree, workflow *strategy.Workflow, hardLimit int) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/recommender", setupLeanRecommender(model, workflow, hardLimit))
	return router
}
