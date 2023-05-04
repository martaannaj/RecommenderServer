package cli

import (
	"bufio"
	"log"
	"os"
	"path/filepath"

	"RecommenderServer/schematree"
	"RecommenderServer/transactions"

	"github.com/spf13/cobra"
	"gitlab.com/tozd/go/mediawiki"
)

func CommandWikiBuild() *cobra.Command {

	var export_format string // used by built-tree

	// subcommand build-tree
	cmdBuildTree := &cobra.Command{
		Use:   "build-tree from-dump|from-tsv <dataset>",
		Short: "Build the SchemaTree model",
		Long: "A SchemaTree model will be built using the file provided in <dataset>." +
			" The dataset should be a wikidata dump file, or a tsv file with one transaction per line. \n" +
			"The output files will be generated in the same directory as <dataset> and with suffixed names, namely:" +
			" '<dataset>.schemaTree.typed.pb'",
	}

	// subcommand build-tree
	buildTreeGeneric := func(fromDump bool) func(cmd *cobra.Command, args []string) {
		return func(cmd *cobra.Command, args []string) {
			inputDataset := filepath.Clean(args[0])

			// Create the tree output file by using the input dataset.
			var s transactions.TransactionSource
			if fromDump {
				dumpFile := &mediawiki.ProcessDumpConfig{
					Path: inputDataset,
				}
				s = transactions.WikidataDumpTransactionSource(dumpFile)
			} else {
				s = transactions.SimpleFileTransactionSource(inputDataset)
			}
			schema := schematree.Create(s)

			outputFileName := inputDataset + ".schemaTree.typed"
			var err error
			switch export_format {
			case "pb":
				outputFileNameFull := filepath.Clean(outputFileName + ".pb")
				outputFile, err := os.Create(outputFileNameFull)
				if err != nil {
					log.Panic(err)
				}
				buf_writer := bufio.NewWriter(outputFile)

				err = schema.SaveProtocolBuffer(buf_writer)
				if err != nil {
					log.Panicln(err)
				}
				err = buf_writer.Flush()
				if err != nil {
					log.Panicln(err)
				}
				err = outputFile.Close()
				if err != nil {
					log.Panicln(err)
				}
			default:
				log.Panic("Format not reconized")

			}
			if err != nil {
				log.Panicln(err)
			}
		}
	}

	cmdBuildTreeDump := &cobra.Command{
		Use:  "from-dump  <dataset>",
		Args: cobra.ExactArgs(1),
		Run:  buildTreeGeneric(true),
	}
	cmdBuildTreeTSV := &cobra.Command{
		Use:  "from-tsv  <dataset>",
		Args: cobra.ExactArgs(1),
		Run:  buildTreeGeneric(false),
	}
	// cmdBuildTree.Flags().StringVarP(&inputDataset, "dataset", "d", "", "`path` to the dataset file to parse")
	// cmdBuildTree.MarkFlagRequired("dataset")
	cmdBuildTree.Flags().StringVar(&export_format, "format", "pb", "The format for the export. Only 'pb' is supported, meaning protocol buffer serialization.")
	cmdBuildTree.AddCommand(cmdBuildTreeDump)
	cmdBuildTree.AddCommand(cmdBuildTreeTSV)
	return cmdBuildTree
}
