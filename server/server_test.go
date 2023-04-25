package server

import (
	"RecommenderServer/schematree"
	"RecommenderServer/strategy"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var treePath = "../testdata/10M.nt.gz.schemaTree.bin"

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	m.Run()
}

func TestLimitRecommendations(t *testing.T) {

	t.Run("Limit the number of recommendations", func(t *testing.T) {
		recs := make([]schematree.RankedPropertyCandidate, 0, 2)
		propStr := "Void"
		rec := schematree.RankedPropertyCandidate{
			Property: &schematree.IItem{
				Str: &propStr,
			},
			Probability: 0.1,
		}
		recs = append(recs, rec)
		recs = append(recs, rec)
		assert.Len(t, limitRecommendations(schematree.PropertyRecommendations(recs), 1), 1)
		assert.Len(t, limitRecommendations(schematree.PropertyRecommendations(recs), 3), 2)
		assert.Len(t, limitRecommendations(schematree.PropertyRecommendations(recs), 0), 0)
		assert.Len(t, limitRecommendations(schematree.PropertyRecommendations(recs), -1), 2)
	})
}

func TestGETRecommendations(t *testing.T) {
	t.Run("Test limits with empty request", func(t *testing.T) {

		f, err := os.Open(treePath)
		if err != nil {
			log.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, _ := schematree.Load(f, false)

		workflow := strategy.MakePresetWorkflow("best", tree)

		request := RecommenderRequest{
			Types:      make([]string, 0),
			Properties: make([]string, 0),
		}

		request_body_builder := new(strings.Builder)

		err = json.NewEncoder(request_body_builder).Encode(request)
		if err != nil {
			t.Fatal(err)
		}

		limits := []int{-1, 1, 2, 20, 1250}
		expected := []int{1242, 1, 2, 20, 1242}

		for i, limit := range limits {

			recommender_server := setupLeanRecommender(tree, workflow, limit)

			request_body := strings.NewReader(request_body_builder.String())
			http_request, _ := http.NewRequest(http.MethodGet, "/recommender", request_body)

			http_response := httptest.NewRecorder()
			recommender_server(http_response, http_request)
			response := RecommenderResponse{}
			err = json.NewDecoder(http_response.Body).Decode(&response)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, response.Recommendations, expected[i], "Incorrect number of recommendations obtained, expected %d, got %d", expected[i], len(response.Recommendations))
		}
	})

	t.Run("Test limits with full request", func(t *testing.T) {

		f, err := os.Open(treePath)
		if err != nil {
			log.Printf("Encountered error while trying to open the file: %v\n", err)
			log.Panic(err)
		}
		tree, _ := schematree.Load(f, false)

		workflow := strategy.MakePresetWorkflow("best", tree)

		request := RecommenderRequest{
			Types:      make([]string, 0),
			Properties: tree.AllProperties(),
		}

		request_body_builder := new(strings.Builder)

		err = json.NewEncoder(request_body_builder).Encode(request)
		if err != nil {
			t.Fatal(err)
		}

		limits := []int{-1, 1, 2, 20, 1250}
		expected := []int{0, 0, 0, 0, 0}

		for i, limit := range limits {

			recommender_server := setupLeanRecommender(tree, workflow, limit)

			request_body := strings.NewReader(request_body_builder.String())
			http_request, _ := http.NewRequest(http.MethodGet, "/recommender", request_body)

			http_response := httptest.NewRecorder()
			recommender_server(http_response, http_request)
			response := RecommenderResponse{}
			err = json.NewDecoder(http_response.Body).Decode(&response)
			if err != nil {
				t.Fatal(err)
			}

			assert.Len(t, response.Recommendations, expected[i], "Incorrect number of recommendations obtained, expected %d, got %d", expected[i], len(response.Recommendations))
		}
	})
}
