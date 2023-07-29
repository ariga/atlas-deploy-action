package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ariga.io/atlas/sql/migrate"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
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
				"INPUT_AMOUNT":      "1",
				"INPUT_TX-MODE":     "all",
				"INPUT_BASELINE":    "1234",
				"INPUT_ALLOW-DIRTY": "true",
			},
			expect: &Input{
				URL:        "sqlite://file.db",
				Amount:     1,
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
			name: "Invalid Amount",
			env: map[string]string{
				"INPUT_URL":    "sqlite://file.db",
				"INPUT_AMOUNT": "notAnInt",
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

func TestRun(t *testing.T) {
	dbpath := sqlitedb(t)
	dburl := fmt.Sprintf("sqlite://%s", dbpath)
	run, err := Run(context.Background(), &Input{
		Dir: "./internal/testdata/migrations",
		URL: dburl,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(run.Applied))
}

func TestCloud(t *testing.T) {
	srv := fakeCloud(t)
	defer srv.Close()
	dbpath := sqlitedb(t)
	dburl := fmt.Sprintf("sqlite://%s", dbpath)
	run, err := Run(context.Background(), &Input{
		Cloud: Cloud{
			URL:   srv.URL,
			Dir:   "dir",
			Token: "token",
		},
		URL: dburl,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(run.Applied))
}

// fakeCloud returns a httptest.Server that mocks the cloud endpoint.
func fakeCloud(t *testing.T) *httptest.Server {
	dir := testDir(t, "./internal/testdata/migrations")
	ad, err := migrate.ArchiveDir(&dir)
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		// nolint:errcheck
		fmt.Fprintf(w, `{"data":{"dir":{"content":%q}}}`, base64.StdEncoding.EncodeToString(ad))
	}))
	return srv
}

// testDir returns a migrate.MemDir from the given path.
func testDir(t *testing.T, path string) (d migrate.MemDir) {
	rd, err := os.ReadDir(path)
	require.NoError(t, err)
	for _, f := range rd {
		fp := filepath.Join(path, f.Name())
		b, err := os.ReadFile(fp)
		require.NoError(t, err)
		require.NoError(t, d.WriteFile(f.Name(), b))
	}
	return d
}

// sqlitedb returns a path to an initialized sqlite database file. The file is
// created in a temporary directory and will be deleted when the test finishes.
func sqlitedb(t *testing.T) string {
	td := t.TempDir()
	dbpath := filepath.Join(td, "file.db")
	_, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&_fk=1", dbpath))
	require.NoError(t, err)
	return dbpath
}
