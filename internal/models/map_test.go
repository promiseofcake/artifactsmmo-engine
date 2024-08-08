package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		one      Coords
		two      Coords
		expected int
	}{
		{Coords{0, 0}, Coords{0, 0}, 0},
		{Coords{1, 1}, Coords{0, 0}, 2},
		{Coords{-5, 1}, Coords{0, 0}, 6},
	}

	for n, tt := range tests {
		t.Run(fmt.Sprintf("%d", n), func(t *testing.T) {
			res := CalculateDistance(tt.one, tt.two)
			assert.Equal(t, tt.expected, res)
		})
	}
}
