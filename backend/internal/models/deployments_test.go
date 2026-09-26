package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        Deployment
		b        Deployment
		expected int
	}{
		{"less than", Deployment{ID: 1}, Deployment{ID: 2}, -1},
		{"equal", Deployment{ID: 5}, Deployment{ID: 5}, 0},
		{"greater than", Deployment{ID: 10}, Deployment{ID: 5}, 1},
		{"zero ids", Deployment{ID: 0}, Deployment{ID: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.a.Compare(tt.b))
		})
	}
}
