package mappers

import (
	"testing"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"

	"github.com/stretchr/testify/assert"
)

func TestFeaturesMapper_Map(t *testing.T) {
	cases := []struct {
		name string
		in   models.Features
		want api.Features
	}{
		{
			name: "basic",
			in: models.Features{
				DisplayConfig: true,
			},
			want: api.Features{
				DisplayConfig: true,
			},
		},
		{
			name: "disabled",
			in: models.Features{
				DisplayConfig: false,
			},
			want: api.Features{
				DisplayConfig: false,
			},
		},
	}

	m := FeaturesMapper{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Map(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
