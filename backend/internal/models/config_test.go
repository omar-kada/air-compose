package models

import (
	"reflect"
	"strings"
	"testing"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/stretchr/testify/assert"
)

func TestConfigPerService_BuildsCorrectArray(t *testing.T) {
	cfg := Config{
		Environment: Environment{
			"GLOBAL": "g",
		},
		Services: map[string]ServiceConfig{
			"svc": {
				"SVC_EXTRA": "s",
			},
		},
	}

	got := cfg.PerService("svc")
	want := orderedmap.NewOrderedMapWithElements(
		&orderedmap.Element[string, string]{Key: "GLOBAL", Value: "g"},
		&orderedmap.Element[string, string]{Key: "SVC_EXTRA", Value: "s"},
	)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ConfigPerService mismatch\nwant=%#v\ngot =%#v", want, got)
	}
}

func TestGetEnabledServices_FiltersCorrectly(t *testing.T) {
	cfg := Config{
		Environment: Environment{
			"GLOBAL": "g",
		},
		Services: map[string]ServiceConfig{
			"svc": {
				"SVC_EXTRA": "s",
			},
			"svc2": {
				"SVC_EXTRA": "s",
			},
		},
	}

	want := []string{"svc", "svc2"}
	assert.ElementsMatch(t, want, cfg.GetEnabledServices())
}

func TestObfuscateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"Empty token", "", ""},
		{"Short token", "123", strings.Repeat("*", 30)},
		{"Long token", "12345678901234567890", "1234567890" + strings.Repeat("*", 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Obfuscate(tt.token))
		})
	}
}

func TestIsObfuscated(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{"Obfuscated token", "1234567890********************", true},
		{"Not obfuscated", "1234567890", false},
		{"Empty token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsObfuscated(tt.token))
		})
	}
}
