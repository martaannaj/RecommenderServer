package main

import (
	"RecommenderServer/schematree"
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/spf13/cobra"
)

func main() {
	// Program initialization actions
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup the variables where all flags will reside.
	var cpuprofile, memprofile, traceFile string // used globally
	var measureTime bool                         // used globally
	// var firstNsubjects int64                     // used by build-tree
	// var writeOutPropertyFreqs bool               // used by build-tree
	// var serveOnPort int                          // used by serve
	// var workflowFile string                      // used by serve
	// var contiguousInput bool                     // used by split-dataset:by-type
	// var everyNthSubject uint                     // used by split-dataset:1-in-n

	// Setup helper variables
	var timeCheckpoint time.Time // used globally

	// writeOutPropertyFreqs := flag.Bool("writeOutPropertyFreqs", false, "set this to write the frequency of all properties to a csv after first pass or schematree loading")

	// root command
	cmdRoot := &cobra.Command{
		Use: "recommender",

		// Execute global pre-run activities such as profiling.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			// write cpu profile to file - open file and start profiling
			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					log.Fatal("could not create CPU profile: ", err)
				}
				if err := pprof.StartCPUProfile(f); err != nil {
					log.Fatal("could not start CPU profile: ", err)
				}
			}

			// write trace execution to file - open file and start tracing
			if traceFile != "" {
				f, err := os.Create(traceFile)
				if err != nil {
					log.Fatal("could not create trace file: ", err)
				}
				if err := trace.Start(f); err != nil {
					log.Fatal("could not start tracing: ", err)
				}
			}

			// measure time - start measuring the time
			//   The measurements are done in such a way to not include the time for the profiles operations.
			if measureTime == true {
				timeCheckpoint = time.Now()
			}

		},

		// Close whatever profiling was running globally.
		PersistentPostRun: func(cmd *cobra.Command, args []string) {

			// measure time - stop time measurement and print the measurements
			if measureTime == true {
				fmt.Println("Execution Time:", time.Since(timeCheckpoint))
			}

			// write cpu profile to file - stop profiling
			if cpuprofile != "" {
				pprof.StopCPUProfile()
			}

			// write memory profile to file
			if memprofile != "" {
				f, err := os.Create(memprofile)
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
				f.Close()
			}

			// write trace execution to file - stop tracing
			if traceFile != "" {
				trace.Stop()
			}

		},
	}

	// global flags for root command
	cmdRoot.PersistentFlags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	cmdRoot.PersistentFlags().StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
	cmdRoot.PersistentFlags().StringVar(&traceFile, "trace", "", "write execution trace to `file`")
	cmdRoot.PersistentFlags().BoolVarP(&measureTime, "time", "t", false, "measure time of command execution")

	// subcommand visualize
	cmdBuildDot := &cobra.Command{
		Use:   "build-dot <tree>",
		Short: "Build a DOT file from a schematree binary",
		Long: "Load the schematree binary stored in path given by <tree> and build a DOT file using" +
			" the GraphViz toolbox.\n" +
			"Will create a file in the same directory as <tree>, with the name: '<tree>.dot'\n",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			treeBinary := &args[0]

			// Load the schematree from the binary file.
			schema, err := schematree.Load(*treeBinary)
			if err != nil {
				log.Panicln(err)
			}
			// schematree.PrintMemUsage()

			// Write the dot file and open it with visualizer.
			// TODO: output a GraphViz visualization of the tree to `tree.png
			// TODO: Println could show the real file name
			f, err := os.Create(*treeBinary + ".dot")
			if err == nil {
				defer f.Close()
				f.WriteString(fmt.Sprint(schema))
				fmt.Println("Run e.g. `dot -Tsvg tree.dot -o tree.svg` to visualize!")
			}

		},
	}
	cmdRoot.AddCommand(cmdBuildDot)

	// Start the CLI application
	cmdRoot.Execute()
}

func waitForReturn() {
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(sentence))
	}
}
