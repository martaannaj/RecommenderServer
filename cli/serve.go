package cli

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

	"github.com/spf13/cobra"
)

func CommandWikiServe() *cobra.Command {

	var serveOnPort uint    // used by serve
	var workflowFile string // used by serve
	var certFile string     // used by serve
	var keyFile string      // used by serve
	var hardLimit int       // used by serve, the hard limit on the output length. -1 for no limit.

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

			log.Printf("Loading schema (from file %v): ", cleanedmodelBinary)

			/// file handling
			f, err := os.Open(cleanedmodelBinary)
			if err != nil {
				log.Printf("Encountered error while trying to open the file: %v\n", err)
				log.Panic(err)
			}

			model, err := schematree.LoadProtocolBufferFromReader(f)
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
				workflow = strategy.MakePresetWorkflow("best", model)
				log.Printf("Run best Recommender ")
			}

			// Initiate the HTTP server. Make it stop on <Enter> press.
			router := server.SetupEndpoints(model, workflow, hardLimit)

			var server = &http.Server{
				Addr:              fmt.Sprintf("0.0.0.0:%v", serveOnPort),
				ReadHeaderTimeout: 5 * time.Second,
				Handler:           router,
			}

			log.Printf("Now listening for https requests on 0.0.0.0:%v\n", serveOnPort)

			if certFile != "" && keyFile != "" {
				err = server.ListenAndServeTLS(certFile, keyFile)
				if err != nil {
					log.Panicln(err)
				}
			} else {
				// we do not want semgrep to catch this because the option to use the server with TLS is provided, but not necessary in all environments
				err = server.ListenAndServe() // nosemgrep: go.lang.security.audit.net.use-tls.use-tls
				if err != nil {
					log.Panicln(err)
				}
			}
		},
	}
	cmdServe.Flags().UintVarP(&serveOnPort, "port", "p", 8080, "`port` of http server")
	cmdServe.Flags().StringVarP(&certFile, "cert", "c", "", "the location of the certificate file (for TLS)")
	cmdServe.Flags().StringVarP(&keyFile, "key", "k", "", "the location of the private key file (for TLS)")
	cmdServe.Flags().StringVarP(&workflowFile, "workflow", "w", "./configuration/Workflow.json", "`path` to config file that defines the workflow")
	cmdServe.Flags().IntVar(&hardLimit, "hard_limit", 500, "The hard limit of the number of results returned by this server, specify -1 for no limit")
	return cmdServe
}
