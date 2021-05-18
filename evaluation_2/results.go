package main

import (
	"RecommenderServer/assessment"
	"RecommenderServer/schematree"
	"RecommenderServer/strategy"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	gzip "github.com/klauspost/pgzip"
)

type evalResult struct {
	setSize  uint16 // number of properties used to generate recommendations (both type and non-type)
	numTypes uint16 // number of type properties in both reduced and leftout property sets
	duration int64  // duration (in nanoseconds) of how long the recommendation took
	group    uint16 // extra value that can store values like custom-made groups
	note     string // @TODO: Temporarily added to aid in evaluation debugging
}

// evaluatePair will generate an evalResult for a pair of ( reducedProps , leftoutProps ).
// This function will take a list of reduced properties, run the recommender workflow with
// those reduced properties, generate evaluation result entries by using the recently adquired
// recommendations and the leftout properties.
// The aim is to evaluate how well the leftout properties appear in the recommendations that are
// generated using the reduced set of properties (from where the properties have been left out).
// Note that 'nil' can be returned.
func evaluatePair(
	tree *schematree.SchemaTree,
	workflow *strategy.Workflow,
	reducedProps schematree.IList,
) *evalResult {

	// Evaluator will not generate stats if no properties exist to make a recommendation.
	if len(reducedProps) == 0 {
		return nil
	}

	// Run the recommender with the input properties.
	start := time.Now()
	asm := assessment.NewInstance(reducedProps, tree, true)
	recs := workflow.Recommend(asm)
	duration := time.Since(start).Nanoseconds()

	// hack for wikiEvaluation
	if len(recs) > 500 {
		recs = recs[:500]
	}

	// Calculate the statistics for the evalResult

	// Count the number of properties that are types in both the reduced and leftout sets.
	var numTypeProps uint16
	for _, rp := range reducedProps {
		if rp.IsType() {
			numTypeProps++
		}
	}

	// Prepare the full evalResult by deriving some values.
	result := evalResult{
		setSize:  uint16(len(reducedProps)),
		numTypes: numTypeProps,
		duration: duration,
	}
	return &result
}

// performEvaluation will produce an evaluation CSV, where a test `dataset` is applied on a
// constructed SchemaTree `tree`, by using the strategy `workflow`.
// A parameter `isTyped` is required to provide for reading the dataset and it has to be synchronized
// with the build SchemaTree model.
// `evalMethod` will set which sampling procedures will be used for the test.
func evaluateDataset(
	tree *schematree.SchemaTree,
	workflow *strategy.Workflow,
	isTyped bool,
	filePath string,
	handlerName string,
) []evalResult {

	// Initialize required variables for managing all the results with multiple threads.
	resultList := make([]evalResult, 0)
	resultWaitGroup := sync.WaitGroup{}
	resultQueue := make(chan evalResult, 1000) // collect eval results via channel

	// Start a parellel thread to process and results that are received from the handlers.
	go func() {
		resultWaitGroup.Add(1)
		//var roundID uint16
		for res := range resultQueue {
			//roundID++
			//res.group = roundID
			resultList = append(resultList, res)
		}
		resultWaitGroup.Done()
	}()

	// Depending on the evaluation method, we will use a different handler
	var handler handlerFunc
	handler = handlerAll

	// We also construct the method that will evaluate a pair of property sets.
	evaluator := func(reduced schematree.IList) *evalResult {
		return evaluatePair(tree, workflow, reduced)
	}

	// Build the complete callback function for the subject summary reader.
	// Given a SubjectSummary, we use the handlers to split it into reduced and leftout set.
	// Then we evaluate that pair of property sets. At last, we deliver the result to our
	// resultQueue that will aggregate all results (from multiple sources) in a single list.
	subjectCallback := func(summary *schematree.SubjectSummary) {
		var results []*evalResult = handler(summary, evaluator)
		for _, res := range results {
			// for convenience, this will treat 'nil' results so that old handlers don't need
			// to look out for 'nil' results that can be returned by the evaluator()
			if res != nil {
				resultQueue <- *res // send structs to channel (not pointers)
			}
		}
	}

	// Start the subject summary reader and collect all results into resultList, using the
	// process that is managing the resultQueue.
	schematree.SubjectSummaryReader(filePath, tree.PropMap, subjectCallback, 0, isTyped)
	close(resultQueue)     // mark the end of results channel
	resultWaitGroup.Wait() // wait until the parallel process that manages the queue is terminated

	return resultList
}

// writeResultsToFile will output the entire evalResult array to a CSV file
func writeResultsToFile(filename string, results []evalResult) {
	f, err := os.Create(filename + ".json")
	if err != nil {
		log.Fatalln("Could not open .json file")
	}
	defer f.Close()
	g := gzip.NewWriter(f)
	defer g.Close()
	e := json.NewEncoder(g)
	err = e.Encode(results)
	if err != nil {
		fmt.Println("Failed to write results to file", err)
	}

	// f, _ := os.Create(filename + ".csv")
	// f.WriteString(fmt.Sprintf(
	// 	"%12s,%12s,%12s,%12s,%12s,%12s,%12s,%12s,%12s,%12s, %s\n",
	// 	"setSize", "numTypes", "numLeftOut", "rank", "numTP", "numFP", "numTN", "numFN", "numTP@L", "dur(ys)", "note",
	// ))

	// for _, dr := range results {
	// 	f.WriteString(fmt.Sprintf(
	// 		"%12v,%12v,%12v,%12v,%12v,%12v,%12v,%12v,%12v,%12v, %s\n",
	// 		dr.setSize, dr.numTypes, dr.numLeftOut, dr.rank, dr.numTP, dr.numFP, dr.numTN, dr.numFN, dr.numTPAtL, dr.duration, dr.note,
	// 	))
	// }
	// f.Close()
	return
}

func loadResultsFromFile(filename string) (results []evalResult) {
	f, err := os.Open(filename + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	json.NewDecoder(r).Decode(&results)

	return
}
