package main

import (
	"fmt"
	"os"
	"sort"
)

type evalSummary struct {
	groupBy  int16
	duration float64
	subjects int64
}

// makeStatics receive a list of evaluation results and makes a summary of them.
func makeStatistics(results []evalResult, groupBy string) (statistics []evalSummary) {
	resultsByGroup := make(map[int][]evalResult) // stores grouped results
	const allResults int = -1                    // catch all group
	statistics = make([]evalSummary, 0, len(results)+1)
	groupIds := []int{allResults} // keep track of existing groups
	groupExists := make(map[int]bool)
	// res.numTypes
	for _, res := range results {
		var groupId int
		groupId = int(res.setSize)

		if !groupExists[groupId] {
			groupExists[groupId] = true
			groupIds = append(groupIds, groupId)
		}

		resultsByGroup[allResults] = append(resultsByGroup[allResults], res)
		resultsByGroup[groupId] = append(resultsByGroup[groupId], res)
	}

	// compute statistics
	sort.Ints(groupIds)

	for _, index := range groupIds {
		groupedResults := resultsByGroup[index]
		resCount := len(groupedResults)

		var Duration int64

		for _, result := range groupedResults {

			Duration += result.duration
		}

		newStat := evalSummary{
			groupBy:  int16(index),
			duration: float64(Duration) / float64(resCount),
			subjects: int64(resCount),
		}

		statistics = append(statistics, newStat)
	}
	return
}

func writeStatisticsToFile(filename string, groupBy string, statistics []evalSummary) {
	f, _ := os.Create(filename + ".csv")
	f.WriteString(fmt.Sprintf(
		"%12s,%12s,%12s\n",
		groupBy, "subjects", "duration",
	))

	for _, stat := range statistics {
		f.WriteString(fmt.Sprintf(
			"%12d,%12d,%12.4f\n",
			stat.groupBy, stat.subjects, stat.duration/1000000,
		))
	}
	f.Close()
	return
}
