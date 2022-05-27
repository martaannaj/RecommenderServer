package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"RecommenderServer/cli"

	"github.com/spf13/cobra"
)

func main() {
	// Program initialization actions
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup the variables where all flags will reside.
	var cpuprofile, memprofile, traceFileName string // used globally
	var measureTime bool                             // used globally

	// Setup helper variables
	var timeCheckpoint time.Time // used globally

	// writeOutPropertyFreqs := flag.Bool("writeOutPropertyFreqs", false, "set this to write the frequency of all properties to a csv after first pass or schematree loading")

	var cpuprofileFile *os.File // used for profiling in case it is enabled
	var traceFile *os.File      // used for tracing in case it is enabled

	// root command
	cmdRoot := &cobra.Command{
		Use: "RecommenderServer",

		// Execute global pre-run activities such as profiling.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			// write cpu profile to file - open file and start profiling
			if cpuprofile != "" {
				f, err := os.Create(filepath.Clean(cpuprofile))
				if err != nil {
					log.Fatal("could not create CPU profile: ", err)
				}
				if err := pprof.StartCPUProfile(f); err != nil {
					log.Fatal("could not start CPU profile: ", err)
				}
				cpuprofileFile = f
			}

			// write trace execution to file - open file and start tracing
			if traceFileName != "" {
				f, err := os.Create(filepath.Clean(traceFileName))
				if err != nil {
					log.Fatal("could not create trace file: ", err)
				}
				if err := trace.Start(f); err != nil {
					log.Fatal("could not start tracing: ", err)
				}
				traceFile = f
			}

			// measure time - start measuring the time
			//   The measurements are done in such a way to not include the time for the profiles operations.
			if measureTime {
				timeCheckpoint = time.Now()
			}

		},

		// Close whatever profiling was running globally.
		PersistentPostRun: func(cmd *cobra.Command, args []string) {

			// measure time - stop time measurement and print the measurements
			if measureTime {
				log.Println("Execution Time:", time.Since(timeCheckpoint))
			}

			// write cpu profile to file - stop profiling
			if cpuprofile != "" {
				pprof.StopCPUProfile()
				err := cpuprofileFile.Close()
				if err != nil {
					log.Fatal(err, "Could not close the cpu profiling file properly")
				}
			}

			// write memory profile to file
			if memprofile != "" {
				f, err := os.Create(filepath.Clean(memprofile))
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
				err = f.Close()
				if err != nil {
					log.Fatal(err)
				}
			}

			// write trace execution to file - stop tracing
			if traceFileName != "" {
				trace.Stop()
				err := traceFile.Close()
				if err != nil {
					log.Fatal(err, "Could not close the trace file properly")
				}
			}

		},
	}

	// global flags for root command
	cmdRoot.PersistentFlags().StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	cmdRoot.PersistentFlags().StringVar(&memprofile, "memprofile", "", "write memory profile to `file`")
	cmdRoot.PersistentFlags().StringVar(&traceFileName, "trace", "", "write execution trace to `file`")
	cmdRoot.PersistentFlags().BoolVarP(&measureTime, "time", "t", false, "measure time of command execution")

	cmdRoot.AddCommand(cli.CommandWikiServe())
	cmdRoot.AddCommand(cli.CommandWikiBuild())
	// Start the CLI application
	err := cmdRoot.Execute()
	if err != nil {
		log.Panicln(err)
	}
}
