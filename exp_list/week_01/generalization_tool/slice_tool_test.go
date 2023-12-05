package generalization_tool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterSet(t *testing.T) {

	testCases := []struct {
		name string
		s1   []int
		s2   []int
		res  []int
	}{
		{
			name: "different number",
			s1:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			s2:   []int{1, 3, 5, 7, 9, 11, 13, 15},
			res:  []int{1, 3, 5, 7, 9},
		},
		{
			name: "repeat number",
			s1:   []int{1, 2, 3, 4},
			s2:   []int{3, 3, 3, 3},
			res:  []int{3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set := InterSet[int](tc.s1, tc.s2)
			assert.Equal(t, tc.res, set)
		})
	}
}

func TestUnionSet(t *testing.T) {

	testCases := []struct {
		name string
		s1   []int
		s2   []int
		res  []int
	}{
		{
			name: "different number",
			s1:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			s2:   []int{1, 3, 5, 7, 9, 11, 13, 15},
			res:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 13, 15},
		},
		{
			name: "repeat number",
			s1:   []int{1, 2, 3, 4},
			s2:   []int{3, 3, 3, 3},
			res:  []int{1, 2, 3, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set := UnionSet[int](tc.s1, tc.s2)
			assert.Equal(t, tc.res, set)
		})
	}
}

func TestDiffSet(t *testing.T) {

	testCases := []struct {
		name string
		s1   []int
		s2   []int
		res  []int
	}{
		{
			name: "SUPPORT A-B || B-A",
			s1:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			s2:   []int{1, 3, 5, 7, 9, 11, 13, 15},
			res:  []int{2, 4, 6, 8},
		},
		{
			name: "repeat number",
			s1:   []int{1, 2, 3, 4},
			s2:   []int{3, 3, 3, 3},
			res:  []int{1, 2, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			set := DiffSet[int](tc.s1, tc.s2)
			assert.Equal(t, tc.res, set)
		})
	}
}
