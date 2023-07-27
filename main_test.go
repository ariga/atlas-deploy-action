package main

import (
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		expect *Input
		hasErr bool
	}{
		{
			name:   "Missing URL",
			env:    map[string]string{},
			hasErr: true,
		},
		{
			name: "Valid Inputs",
			env: map[string]string{
				"INPUT_URL":         "sqlite://file.db",
				"INPUT_COUNT":       "1",
				"INPUT_TX-MODE":     "all",
				"INPUT_BASELINE":    "1234",
				"INPUT_ALLOW-DIRTY": "true",
			},
			expect: &Input{
				URL:        "sqlite://file.db",
				Count:      1,
				TxMode:     "all",
				Baseline:   "1234",
				AllowDirty: true,
			},
			hasErr: false,
		},
		{
			name: "Illegal TxMode",
			env: map[string]string{
				"INPUT_URL":         "sqlite://file.db",
				"INPUT_COUNT":       "1",
				"INPUT_TX-MODE":     "invalid",
				"INPUT_BASELINE":    "1234",
				"INPUT_ALLOW-DIRTY": "true",
			},
			expect: nil,
			hasErr: true,
		},
		{
			name: "Invalid Dirty",
			env: map[string]string{
				"INPUT_URL":         "sqlite://file.db",
				"INPUT_ALLOW-DIRTY": "notABool",
			},
			expect: nil,
			hasErr: true,
		},
		{
			name: "Invalid Count",
			env: map[string]string{
				"INPUT_URL":   "sqlite://file.db",
				"INPUT_COUNT": "notAnInt",
			},
			expect: nil,
			hasErr: true,
		},
		{
			name: "Dir and CloudDir Exclusion",
			env: map[string]string{
				"INPUT_URL":       "sqlite://file.db",
				"INPUT_DIR":       "dir",
				"INPUT_CLOUD-DIR": "cloud-dir",
			},
			expect: nil,
			hasErr: true,
		},
		{
			name: "Dir",
			env: map[string]string{
				"INPUT_URL": "sqlite://file.db",
				"INPUT_DIR": "dir",
			},
			expect: &Input{
				URL: "sqlite://file.db",
				Dir: "dir",
			},
		},
		{
			name: "CloudDir no Token",
			env: map[string]string{
				"INPUT_URL":       "sqlite://file.db",
				"INPUT_CLOUD-DIR": "dir",
			},
			hasErr: true,
		},
		{
			name: "CloudDir Token",
			env: map[string]string{
				"INPUT_URL":         "sqlite://file.db",
				"INPUT_CLOUD-DIR":   "dir",
				"INPUT_CLOUD-TOKEN": "token",
			},
			expect: &Input{
				URL: "sqlite://file.db",
				Cloud: Cloud{
					Token: "token",
					Dir:   "dir",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := githubactions.New(githubactions.WithGetenv(func(key string) string {
				return tt.env[key]
			}))
			input, err := Load(act)
			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tt.expect, input)
			}
		})
	}
}
