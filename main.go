package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/sethvargo/go-githubactions"
)

var (
	//go:embed atlashcl.tmpl
	tmpl   string
	config = template.Must(template.New("atlashcl").Parse(tmpl))
)

func main() {
	act := githubactions.New()
	act.Warningf("This action is deprecated. Please use ariga/atlas-action/migrate/apply instead. " +
		"For details see: https://github.com/ariga/atlas-action#arigaatlas-actionmigrateapply")
	inp, err := Load(act)
	if err != nil {
		act.Fatalf("failed to load input: %v", err)
	}
	run, err := Run(context.Background(), inp)
	if err != nil {
		act.Fatalf("failed to run: %v", err)
	}
	act.SetOutput("error", run.Error)
	act.SetOutput("current", run.Current)
	act.SetOutput("target", run.Target)
	act.SetOutput("pending_count", strconv.Itoa(len(run.Pending)))
	act.SetOutput("applied_count", strconv.Itoa(len(run.Applied)))
	act.Infof("Run complete: +%v", run)
}

type (
	// Input is created from the GitHub Action "with" configuration.
	Input struct {
		URL             string
		Amount          uint64
		TxMode          string
		Baseline        string
		AllowDirty      bool
		Dir             string
		RevisionsSchema string
		Cloud           Cloud
	}
	Cloud struct {
		Dir   string
		Token string
		URL   string
		Tag   string
	}
)

// Load loads the input from the GitHub Action configuration.
func Load(act *githubactions.Action) (*Input, error) {
	i := &Input{
		URL: act.GetInput("url"),
	}
	if i.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if as := act.GetInput("amount"); as != "" {
		a, err := strconv.ParseUint(as, 10, 64)
		if err != nil {
			return nil, err
		}
		i.Amount = a
	}
	if txm := act.GetInput("tx-mode"); txm != "" {
		switch txm {
		case "all", "none", "file":
			i.TxMode = txm
		default:
			return nil, fmt.Errorf("invalid tx-mode %q", txm)
		}
		i.TxMode = act.GetInput("tx-mode")
	}
	i.Baseline = act.GetInput("baseline")
	if ad := act.GetInput("allow-dirty"); ad != "" {
		allowDirty, err := strconv.ParseBool(strings.ToLower(ad))
		if err != nil {
			return nil, fmt.Errorf("invalid allow-dirty %q", ad)
		}
		i.AllowDirty = allowDirty
	}
	i.Dir = act.GetInput("dir")
	i.Cloud.Dir = act.GetInput("cloud-dir")
	if i.Dir != "" && i.Cloud.Dir != "" {
		return nil, fmt.Errorf("dir and cloud-dir are mutually exclusive")
	}
	i.Cloud.Token = act.GetInput("cloud-token")
	if i.Cloud.Dir != "" && i.Cloud.Token == "" {
		return nil, fmt.Errorf("cloud-token is required when cloud-dir is set")
	}
	i.Cloud.URL = act.GetInput("cloud-url")
	i.Cloud.Tag = act.GetInput("cloud-tag")
	return i, nil
}

// Run runs the "migrate apply" for the input.
func Run(ctx context.Context, i *Input) (*atlasexec.ApplyReport, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	client, err := atlasexec.NewClient(wd, "atlas")
	if err != nil {
		return nil, err
	}
	params := &atlasexec.ApplyParams{
		URL:             i.URL,
		Amount:          i.Amount,
		TxMode:          i.TxMode,
		BaselineVersion: i.Baseline,
	}
	if i.Dir != "" {
		params.DirURL = "file://" + i.Dir
	}
	if i.Cloud.Dir != "" {
		var buf bytes.Buffer
		if err := config.Execute(&buf, i); err != nil {
			return nil, err
		}
		cfg, clean, err := atlasexec.TempFile(buf.String(), "hcl")
		if err != nil {
			return nil, err
		}
		// nolint:errcheck
		defer clean()
		params.ConfigURL = cfg
		params.Env = "atlas"
	}
	return client.Apply(ctx, params)
}
