package cli

import (
	"log"

	"RecommenderServer/schematree"
	"RecommenderServer/transactions"

	"github.com/spf13/cobra"
	"gitlab.com/tozd/go/mediawiki"
)

func CommandWikiBuild() *cobra.Command {

	var firstNsubjects int64       // used by build-tree
	var writeOutPropertyFreqs bool // used by build-tree
	var export_format string       // used by built-tree

	// subcommand build-tree
	cmdBuildTree := &cobra.Command{
		Use:   "build-tree <dataset>",
		Short: "Build the SchemaTree model",
		Long: "A SchemaTree model will be built using the file provided in <dataset>." +
			" The dataset should be a N-Triple of Items.\nTwo output files will be" +
			" generated in the same directory as <dataset> and with suffixed names, namely:" +
			" '<dataset>.firstPass.bin' and '<dataset>.schemaTree.bin'",
		Args: cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			inputDataset := &args[0]

			// Create the tree output file by using the input dataset.
			dumpFile := &mediawiki.ProcessDumpConfig{
				Path: *inputDataset,
			}

			s := transactions.WikidataDumpTransactionSource(dumpFile)

			schema := schematree.Create(s)
			outputFileName := *inputDataset + ".schemaTree.typed"
			var err error
			switch export_format {
			case "pb":
				err = schema.SaveProtocolBuffer(outputFileName + ".pb")
			default:
				log.Panic("Format not reconized")

			}
			if err != nil {
				log.Panicln(err)
			}
			// if writeOutPropertyFreqs {
			// 	propFreqsPath := *inputDataset + ".propertyFreqs.csv"
			// 	schema.WritePropFreqs(propFreqsPath)
			// 	fmt.Printf("Wrote PropertyFreqs to %s\n", propFreqsPath)

			// 	if typed_tree {
			// 		typeFreqsPath := *inputDataset + ".typeFreqs.csv"
			// 		schema.WriteTypeFreqs(typeFreqsPath)
			// 		fmt.Printf("Wrote PropertyFreqs to %s\n", typeFreqsPath)
			// 	}
			// }

		},
	}
	// cmdBuildTree.Flags().StringVarP(&inputDataset, "dataset", "d", "", "`path` to the dataset file to parse")
	// cmdBuildTree.MarkFlagRequired("dataset")
	cmdBuildTree.Flags().Int64VarP(&firstNsubjects, "first", "n", 0, "only parse the first `n` subjects") // TODO: handle negative inputs
	cmdBuildTree.Flags().BoolVarP(
		&writeOutPropertyFreqs, "write-frequencies", "f", false,
		"write all property frequencies to a csv file named '<dataset>.propertyFreqs.csv' after the SchemaTree is built",
	)
	cmdBuildTree.Flags().StringVar(&export_format, "format", "pb", "The format for the export. Only 'pb' is supported, meaning protocol buffer serialization.")

	return cmdBuildTree
}
