package transactions

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

type Transaction []string

type TransactionSource func() <-chan Transaction

func WikidataDumpTransactionSource(dumpfile *mediawiki.ProcessDumpConfig) TransactionSource {
	return func() <-chan Transaction {
		channel := make(chan Transaction)

		go func() {
			errE := mediawiki.ProcessWikidataDump(
				context.Background(),
				dumpfile,
				func(_ context.Context, a mediawiki.Entity) errors.E {
					t := make(Transaction, 0)

					for property_name := range a.Claims {
						t = append(t, property_name)
					}
					types_claims := a.Claims["P31"]
					for _, statement := range types_claims {
						if statement.MainSnak.SnakType == mediawiki.Value {
							if statement.MainSnak.DataValue == nil {
								log.Fatal("Found a main snak with type Value, while it does not have a value. This is an error in the dump.")
							}
							val := statement.MainSnak.DataValue.Value
							switch v := val.(type) {
							default:
								log.Printf("unexpected type %T", v)
							case mediawiki.WikiBaseEntityIDValue:
								tokenStr := "t#" + val.(mediawiki.WikiBaseEntityIDValue).ID
								t = append(t, tokenStr)
							}
						} else {
							log.Printf("Found a type statement without a value: %v", statement)
						}
					}
					channel <- t
					return nil
				},
			)
			if errE != nil {
				log.Panicln("Something went wrong while processing..", errE, errE.Details())
			}
			close(channel)
		}()
		return channel
	}
}

func SimpleFileTransactionSource(inputFile string) TransactionSource {
	return SimpleReaderTransactionSource(
		func() io.Reader {
			f, err := os.Open(path.Clean(inputFile))
			if err != nil {
				log.Panic("File could not be opened", err)
			}
			return f
		},
	)
}

// SimpleFileTransactionSource creates a TransactionSource from a reader with on each line a transaction
// Note that this needs a function providing readers because the two pass algorithm will pass twice
func SimpleReaderTransactionSource(input func() io.Reader) TransactionSource {
	return func() <-chan Transaction {
		channel := make(chan Transaction)

		go func() {
			scanner := bufio.NewScanner(input())
			for scanner.Scan() {
				line := scanner.Text()
				line = strings.TrimSpace(line)
				if line != "" {
					names := strings.Fields(line)
					t := Transaction(names)
					channel <- t
				}
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			close(channel)
		}()
		return channel
	}
}
