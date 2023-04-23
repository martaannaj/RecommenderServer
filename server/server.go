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
	if hardLimit < 1 && hardLimit != -1 {
		log.Panic("hardLimit must be positive, or -1")
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
		escapedlogstring := formatForLogging(input)
		log.Println("request received ", escapedlogstring)

		instance := schematree.NewInstanceFromInput(input.Properties, input.Types, model, true)

		// Make a recommendation based on the assessed input and chosen strategy.
		t1 := time.Now()
		recomendations := workflow.Recommend(instance)
		log.Println("request ", escapedlogstring, " answered in ", time.Since(t1))

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

func formatForLogging(input RecommenderRequest) string {
	var jsonstring = fmt.Sprintln(input)
	escapedjsonstring := strings.Replace(jsonstring, "\n", "", -1)
	escapedjsonstring = strings.Replace(escapedjsonstring, "\r", "", -1)
	return escapedjsonstring
}

// Limit the recommendations to contain at most `hardLimit` items and convert to output entries.
// If hardLimit is -1, then no limit is imposed.
func limitRecommendations(recommendations schematree.PropertyRecommendations, hardLimit int) []RecommendationOutputEntry {

	capacity := len(recommendations)
	if hardLimit != -1 {
		if capacity > hardLimit {
			capacity = hardLimit
		}
	}
	outputRecs := make([]RecommendationOutputEntry, 0, capacity)

	for _, recommendation := range recommendations {
		if hardLimit != -1 && len(outputRecs) >= hardLimit {
			break
		}
		if recommendation.Property.IsProp() {
			outputRecs = append(outputRecs, RecommendationOutputEntry{
				PropertyStr: recommendation.Property.Str,
				Probability: recommendation.Probability,
			})
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
