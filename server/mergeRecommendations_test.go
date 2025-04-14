package server

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestWithID struct {
	ID       int
	priority string
}

func (t TestWithID) getID() interface{} {
	return t.ID
}

func _testSomeMerges(t *testing.T, merger func(higherPriority, lowerPriority []TestWithID) []TestWithID) {
	low := []TestWithID{
		{1, "low"},
		{2, "low"},
		{3, "low"},
		{4, "low"},
		{5, "low"},
	}
	t.Run("Merge Empty in", func(t *testing.T) {
		high := []TestWithID{}
		res := merger(high, low)
		assert.Len(t, res, 5)
	})

	t.Run("Merge One at the start", func(t *testing.T) {
		high := []TestWithID{{1, "high"}}
		res := merger(high, low)
		assert.Len(t, res, 5)
		expected := []TestWithID{
			{1, "high"},
			{2, "low"},
			{3, "low"},
			{4, "low"},
			{5, "low"},
		}
		assert.Equal(t, res, expected)

	})

	t.Run("Merge One at the end", func(t *testing.T) {
		high := []TestWithID{{5, "high"}}
		res := merger(high, low)
		assert.Len(t, res, 5)
		expected := []TestWithID{
			{5, "high"},
			{1, "low"},
			{2, "low"},
			{3, "low"},
			{4, "low"},
		}
		assert.Equal(t, res, expected)

	})

	t.Run("Merge One at start, middle and end", func(t *testing.T) {
		high := []TestWithID{{5, "high"}, {3, "high"}, {1, "high"}}
		res := merger(high, low)
		assert.Len(t, res, 5)
		expected := []TestWithID{
			{5, "high"},
			{3, "high"},
			{1, "high"},
			{2, "low"},
			{4, "low"},
		}
		assert.Equal(t, res, expected)

	})

	t.Run("Merge two in the middle", func(t *testing.T) {
		high := []TestWithID{{4, "high"}, {2, "high"}}
		res := merger(high, low)
		assert.Len(t, res, 5)
		expected := []TestWithID{
			{4, "high"},
			{2, "high"},
			{1, "low"},
			{3, "low"},
			{5, "low"},
		}
		assert.Equal(t, res, expected)

	})
}

func TestSimpleMerge(t *testing.T) {
	_testSomeMerges(t, SimpleMerge[TestWithID])
}

func TestFastMerge(t *testing.T) {
	_testSomeMerges(t, FasterMerge[TestWithID])

	// We test by comparing a lot of merges with the outcome of the simple merge. They must be the same.
	t.Run("Compare Fast to Slow", func(t *testing.T) {
		rng := rand.New(rand.NewSource(4567890))
		for low_size := 0; low_size < 50; low_size++ {
			low := make([]TestWithID, 0, low_size)
			for i := 0; i < low_size; i++ {
				low = append(low, TestWithID{i, "low"})
			}
			max_high := int(math.Min(float64(low_size), 20))
			for high_size := 0; high_size <= max_high; high_size++ {
				high_all := make([]TestWithID, 0, high_size)
				for i := 0; i < low_size; i++ {
					high_all = append(high_all, TestWithID{i, "high"})
				}
				rng.Shuffle(len(high_all), func(i, j int) { high_all[i], high_all[j] = high_all[j], high_all[i] })
				high := high_all[:high_size]

				expected := SimpleMerge(high, low)
				got := FasterMerge(high, low)
				assert.Equal(t, expected, got)
			}
		}
	})
}

func _benchMarkMerge(b *testing.B, merger func(higherPriority, lowerPriority []TestWithID) []TestWithID) {
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(4567890))
		rng2 := rand.New(rand.NewSource(45675679876))
		for low_size := 0; low_size < 500; low_size++ {
			low := make([]TestWithID, 0, low_size)
			for i := 0; i < low_size; i++ {
				low = append(low, TestWithID{i, "low"})
			}
			max_high := int(math.Min(float64(low_size), 200))
			for high_size := 0; high_size <= max_high; high_size++ {
				high_all := make([]TestWithID, 0, high_size)
				for i := 0; i < low_size; i++ {
					high_all = append(high_all, TestWithID{i, "high"})
				}
				rng.Shuffle(len(high_all),
					func(i, j int) {
						// We only really shuffle sometimes
						p := rng2.Float32()
						if p < 1.00 {
							high_all[i], high_all[j] = high_all[j], high_all[i]
						}
					})
				high := high_all[:high_size]

				merger(high, low)
			}
		}
	}
}

func BenchmarkSimpleMerge(b *testing.B) {
	_benchMarkMerge(b, SimpleMerge[TestWithID])
}

func BenchmarkFastMerge(b *testing.B) {
	_benchMarkMerge(b, FasterMerge[TestWithID])
}
