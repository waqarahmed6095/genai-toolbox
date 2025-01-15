// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigtable

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/googleapis/genai-toolbox/internal/sources"
	bigtabledb "github.com/googleapis/genai-toolbox/internal/sources/bigtable"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"google.golang.org/api/iterator"
)

const ToolKind string = "bigtable-sql"

type compatibleSource interface {
	BigtableClient() *bigquery.Client
	DatabaseDialect() string
}

// validate compatible sources are still compatible
var _ compatibleSource = &bigtabledb.Source{}

var compatibleSources = [...]string{bigtabledb.SourceKind}

type Config struct {
	Name         string           `yaml:"name"`
	Kind         string           `yaml:"kind"`
	Source       string           `yaml:"source"`
	Description  string           `yaml:"description"`
	Statement    string           `yaml:"statement"`
	AuthRequired []string         `yaml:"authRequired"`
	Parameters   tools.Parameters `yaml:"parameters"`
}

// validate interface
var _ tools.ToolConfig = Config{}

func (cfg Config) ToolConfigKind() string {
	return ToolKind
}

func (cfg Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	// verify source exists
	rawS, ok := srcs[cfg.Source]
	if !ok {
		return nil, fmt.Errorf("no source named %q configured", cfg.Source)
	}

	// verify the source is compatible
	s, ok := rawS.(compatibleSource)
	if !ok {
		return nil, fmt.Errorf("invalid source for %q tool: source kind must be one of %q", ToolKind, compatibleSources)
	}

	// finish tool setup
	t := Tool{
		Name:         cfg.Name,
		Kind:         ToolKind,
		Parameters:   cfg.Parameters,
		Statement:    cfg.Statement,
		AuthRequired: cfg.AuthRequired,
		Client:       s.BigtableClient(),
		dialect:      s.DatabaseDialect(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: cfg.Parameters.Manifest()},
	}
	return t, nil
}

// validate interface
var _ tools.Tool = Tool{}

type Tool struct {
	Name         string           `yaml:"name"`
	Kind         string           `yaml:"kind"`
	AuthRequired []string         `yaml:"authRequired"`
	Parameters   tools.Parameters `yaml:"parameters"`
	Location     string           `yaml:"location"`

	Client    *bigquery.Client
	Statement string
	dialect   string
	manifest  tools.Manifest
}

func getMapParams(params tools.ParamValues, dialect string) (map[string]interface{}, error) {
	switch strings.ToLower(dialect) {
	case "googlesql":
		return params.AsMap(), nil
	case "sparksql":
		return params.AsMap(), nil
	default:
		return nil, fmt.Errorf("invalid dialect %s", dialect)
	}
}

func (t Tool) Invoke(params tools.ParamValues) (string, error) {
	fmt.Printf("Invoked tool %s\n", t.Name)

	ctx := context.Background()

	q := t.Client.Query(t.Statement)

	// Set legacy SQL dialect
	if t.dialect == "sql" {
		q.UseLegacySQL = true
	}

	// Set job location if specified
	if t.Location != "" {
		q.Location = t.Location
	}

	// Set job query parameters
	q.Parameters = params.AsSliceOfQueryParameters()

	// Run the query and print results when the query job is completed.
	job, err := q.Run(ctx)

	if err != nil {
		return "", fmt.Errorf("unable to run query job: %w", err)

	}
	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("error retrieving job status: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("query job run failure: %w", err)
	}
	it, err := job.Read(ctx)

	// fetch query result
	var out strings.Builder
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("unable to parse row: %w", err)
		}
		out.WriteString(fmt.Sprintf("%s", row))
	}

	return fmt.Sprintf("Stub tool call for %q! Parameters parsed: %q \n Output: %s", t.Name, params, out.String()), nil
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	return tools.ParseParams(t.Parameters, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) Authorized(verifiedAuthSources []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthSources)
}
