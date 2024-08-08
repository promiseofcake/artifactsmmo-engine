package math

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMax(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{0, 1, 1},
		{0, -1, 0},
		{0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d,%d", tt.a, tt.b), func(t *testing.T) {
			res := Max(tt.a, tt.b)
			assert.Equal(t, tt.expected, res)
		})
	}
}
