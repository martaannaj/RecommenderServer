package main

import (
	"RecommenderServer/configuration"
	"RecommenderServer/schematree"
	"RecommenderServer/server"
	"RecommenderServer/strategy"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	var serveOnPort int                          // used by serve
	var workflowFile string                      // used by serve
	var stripURIs bool                           // used globally
	var certFile string                          // used by serve
	var keyFile string                           // used by serve

	// Setup helper variables
	var timeCheckpoint time.Time // used globally

	// writeOutPropertyFreqs := flag.Bool("writeOutPropertyFreqs", false, "set this to write the frequency of all properties to a csv after first pass or schematree loading")

	// root command
	cmdRoot := &cobra.Command{
		Use: "RecommenderServer",

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
			if measureTime {
				timeCheckpoint = time.Now()
			}

		},

		// Close whatever profiling was running globally.
		PersistentPostRun: func(cmd *cobra.Command, args []string) {

			// measure time - stop time measurement and print the measurements
			if measureTime {
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
				err = f.Close()
				if err != nil {
					log.Panic(err)
				}
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

	// subcommand serve
	cmdServe := &cobra.Command{
		Use:   "serve <model>",
		Short: "Serve a SchemaTree model via an HTTP Server",
		Long: "Load the <model> (schematree binary) and the recommendation" +
			" endpoint using an HTTP Server.\nAvailable endpoints are stated in the server README.",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			if (keyFile == "") != (certFile == "") {
				log.Panicln("Either both --cert and --key must be set, or neither of them")
			}

			modelBinary := args[0]
			cleanedmodelBinary := filepath.Clean(modelBinary)
			// glossaryBinary := &args[1]

			// Load the schematree from the binary file.

			fmt.Printf("Loading schema (from file %v): ", cleanedmodelBinary)

			/// file handling
			f, err := os.Open(cleanedmodelBinary)
			if err != nil {
				fmt.Printf("Encountered error while trying to open the file: %v\n", err)
				log.Panic(err)
			}

			model, err := schematree.Load(f, stripURIs)
			if err != nil {
				log.Panicln(err)
			}
			schematree.PrintMemUsage()

			// read config file if given as parameter, test if everything needed is there, create a workflow
			// if no config file is given, the standard recommender is set as workflow.
			var workflow *strategy.Workflow
			if workflowFile != "" {
				config, err := configuration.ReadConfigFile(&workflowFile)
				if err != nil {
					log.Panicln(err)
				}
				err = config.Test()
				if err != nil {
					log.Panicln(err)
				}
				workflow, err = configuration.ConfigToWorkflow(config, model)
				if err != nil {
					log.Panicln(err)
				}
				log.Printf("Run Config Workflow %v", workflowFile)
			} else {
				workflow = strategy.MakePresetWorkflow("direct", model)
				fmt.Printf("Run Standard Recommender ")
			}

			// Initiate the HTTP server. Make it stop on <Enter> press.
			router := server.SetupEndpoints(model, workflow, 500)

			if certFile != "" && keyFile != "" {
				fmt.Printf("Now listening for https requests on 0.0.0.0:%v\n", serveOnPort)
				err = http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%v", serveOnPort), certFile, keyFile, router)
				if err != nil {
					log.Panicln(err)
				}
			} else {
				fmt.Printf("Now listening for http requests on 0.0.0.0:%v\n", serveOnPort)
				err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", serveOnPort), router)
				if err != nil {
					log.Panicln(err)
				}
			}
		},
	}
	cmdServe.Flags().IntVarP(&serveOnPort, "port", "p", 8080, "`port` of http server")
	cmdServe.Flags().StringVarP(&certFile, "cert", "c", "", "the location of the certificate file (for TLS)")
	cmdServe.Flags().StringVarP(&keyFile, "key", "k", "", "the location of the private key file (for TLS)")
	cmdServe.Flags().StringVarP(&workflowFile, "workflow", "w", "./configuration/Workflow.json", "`path` to config file that defines the workflow")
	cmdServe.Flags().BoolVarP(&stripURIs, "stripURI", "s", true, "flag set to true if the recommendations should be served without URIs")

	cmdRoot.AddCommand(cmdServe)
	// Start the CLI application
	err := cmdRoot.Execute()
	if err != nil {
		log.Panicln(err)
	}
}
