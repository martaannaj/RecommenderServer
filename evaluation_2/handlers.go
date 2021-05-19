package main

import "RecommenderServer/schematree"

type handlerFunc func(*schematree.SubjectSummary, func(schematree.IList) *evalResult) []*evalResult

func handlerAll(
	summary *schematree.SubjectSummary,
	evaluator func(schematree.IList) *evalResult,
) []*evalResult {
	results := make([]*evalResult, 0, 1)

	// Create and fill both subsets
	props := make(schematree.IList, 0, len(summary.Properties))
	for property := range summary.Properties {
		props = append(props, property)
	}

	// Only one result is generated for this handler. If no types properties exist, then
	// the evaluator will return nil.
	res := evaluator(props)
	if res != nil {
		res.note = summary.Str // @TODO: Temporarily added to aid in evaluation debugging
		results = append(results, res)
	}
	return results // return an array of one or zero results
}
