package main

import (
	recIO "RecommenderServer/io"
	"RecommenderServer/schematree"
	"RecommenderServer/strategy"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile := flag.String("memprofile", "", "write memory profile to `file`")
	traceFile := flag.String("trace", "", "write execution trace to `file`")
	trainedModel := flag.String("model", "", "read stored schematree from `file`")
	configPath := flag.String("workflow", "", "Path to workflow config file for single evaluation")
	testFile := flag.String("testSet", "", "the file to parse")
	typedEntities := flag.Bool("typed", false, "Use type information or not")
	handlerType := flag.String("handler", "takeOneButType", "Choose the handler: takeOneButType, takeAllButBest, takeMoreButCommon")
	groupBy := flag.String("groupBy", "setSize", "Choose groupBy: setSize, numTypes, numLeftOut, numNonTypes")
	writeResults := flag.Bool("results", false, "Turn on to write an additional JSON file with all evaluation results")
	loadResults := flag.Bool("loadResults", false, "Turn on to read results back from JSON file instead of running the actual evaluation")
	customName := flag.String("name", "", "Add a custom designation to the generate CSV files")

	// parse commandline arguments/flags
	flag.Parse()

	// write cpu profile to file
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// write cpu profile to file
	if *memprofile != "" {
		defer func() {
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			f.Close()
		}()
	}

	// write cpu profile to file
	if *traceFile != "" {
		f, err := os.Create(*traceFile)
		if err != nil {
			log.Fatal("could not create trace file: ", err)
		}
		if err := trace.Start(f); err != nil {
			log.Fatal("could not start tracing: ", err)
		}
		defer trace.Stop()
	}

	if *testFile == "" {
		log.Fatalln("A test set must be provided!")
	}

	// Calculate the base name of the input file to generate CSVs with similar names.
	// If customName is defined then will use that and, if not, it will use other flags.
	testBase := recIO.TrimExtensions(*testFile)
	if *customName != "" {
		testBase += "-" + *customName
	} else {
		if *typedEntities {
			testBase += "-typed"
		} else {
			testBase += "-standard"
		}
		if *configPath != "" {
			testBase += "-backoff"
		}
		testBase += "-" + *handlerType + "-" + *groupBy
	}

	var datasetResults []evalResult
	if *loadResults {
		datasetResults = loadResultsFromFile(testBase + "-results")
		fmt.Println(datasetResults)
	} else {
		// evaluation
		if *trainedModel == "" {
			log.Fatalln("A model must be provided!")
		}
		tree, err := schematree.Load(*trainedModel)
		if err != nil {
			log.Fatalln(err)
		}

		var wf *strategy.Workflow
		wf = strategy.MakePresetWorkflow("best", tree)

		fmt.Println("Evaluating the dataset...")
		datasetResults = evaluateDataset(tree, wf, *typedEntities, *testFile, *handlerType)
	}

	// When results flag is given, will also write a CSV for evalResult array
	if *writeResults {
		fmt.Printf("Writing results to JSON file...")
		writeResultsToFile(testBase+"-results", datasetResults)
		fmt.Printf(" Complete.\n")
	}

	fmt.Printf("Aggregating the results...")
	datasetStatistics := makeStatistics(datasetResults, *groupBy)
	fmt.Printf(" Complete.\n")

	fmt.Printf("Writing statistics to CSV file...")
	writeStatisticsToFile(testBase+"-stats", *groupBy, datasetStatistics)
	fmt.Printf(" Complete.\n")

	fmt.Printf("%v+\n", datasetStatistics[0])
}
