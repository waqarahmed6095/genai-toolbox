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

package alloydbainl

import (
	"context"
	"fmt"
	"strings"

	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/sources/alloydbpg"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ToolKind string = "alloydb-ai-nl"

type compatibleSource interface {
	PostgresPool() *pgxpool.Pool
}

// validate compatible sources are still compatible
var _ compatibleSource = &alloydbpg.Source{}

var compatibleSources = [...]string{alloydbpg.SourceKind}

type Config struct {
	Name               string           `yaml:"name" validate:"required"`
	Kind               string           `yaml:"kind" validate:"required"`
	Source             string           `yaml:"source" validate:"required"`
	Description        string           `yaml:"description" validate:"required"`
	NLConfig           string           `yaml:"nlConfig" validate:"required"`
	AuthRequired       []string         `yaml:"authRequired"`
	NLConfigParameters tools.Parameters `yaml:"nlConfigParameters"`
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

	numParams := len(cfg.NLConfigParameters)
	quotedNameParts := make([]string, 0, numParams)
	placeholderParts := make([]string, 0, numParams)

	for i, paramDef := range cfg.NLConfigParameters {
		name := paramDef.GetName()
		escapedName := strings.ReplaceAll(name, "'", "''") // Escape for SQL literal
		quotedNameParts = append(quotedNameParts, fmt.Sprintf("'%s'", escapedName))
		placeholderParts = append(placeholderParts, fmt.Sprintf("$%d", i+3)) // $1, $2 reserved
	}

	var paramNamesSQL string
	var paramValuesSQL string

	if numParams > 0 {
		paramNamesSQL = fmt.Sprintf("ARRAY[%s]", strings.Join(quotedNameParts, ", "))
		paramValuesSQL = fmt.Sprintf("ARRAY[%s]", strings.Join(placeholderParts, ", "))
	} else {
		paramNamesSQL = "ARRAY[]::TEXT[]"
		paramValuesSQL = "ARRAY[]::TEXT[]"
	}

	// execute_nl_query is the AlloyDB AI function that executes the natural language query
	// The first parameter is the natural language query, which is passed as $1
	// The second parameter is the NLConfig, which is passed as a $2
	// The following params are the list of PSV values passed to the NLConfig
	// Example SQL statement being executed:
	// SELECT alloydb_ai_nl.execute_nl_query('How many tickets do I have?', 'cymbal_air_nl_config', param_names => ARRAY ['user_email'], param_values => ARRAY ['hailongli@google.com']);
	stmtFormat := "SELECT alloydb_ai_nl.execute_nl_query($1, $2, param_names => %s, param_values => %s);"
	stmt := fmt.Sprintf(stmtFormat, paramNamesSQL, paramValuesSQL)

	newQuestionParam := tools.NewStringParameter(
		"question",                              // name
		"The natural language question to ask.", // description
	)

	cfg.NLConfigParameters = append([]tools.Parameter{newQuestionParam}, cfg.NLConfigParameters...)

	mcpManifest := tools.McpManifest{
		Name:        cfg.Name,
		Description: cfg.Description,
		InputSchema: cfg.NLConfigParameters.McpManifest(),
	}

	t := Tool{
		Name:         cfg.Name,
		Kind:         ToolKind,
		Parameters:   cfg.NLConfigParameters,
		Statement:    stmt,
		NLConfig:     cfg.NLConfig,
		AuthRequired: cfg.AuthRequired,
		Pool:         s.PostgresPool(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: cfg.NLConfigParameters.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest:  mcpManifest,
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

	Pool        *pgxpool.Pool
	Statement   string
	NLConfig    string
	manifest    tools.Manifest
	mcpManifest tools.McpManifest
}

func (t Tool) Invoke(params tools.ParamValues) ([]any, error) {
	sliceParams := params.AsSlice()
	allParamValues := make([]any, len(sliceParams)+1)
	allParamValues[0] = fmt.Sprintf("%s", sliceParams[0]) // nl_question
	allParamValues[1] = t.NLConfig                        // nl_config
	for i, param := range sliceParams[1:] {
		allParamValues[i+2] = fmt.Sprintf("%s", param)
	}

	results, err := t.Pool.Query(context.Background(), t.Statement, allParamValues...)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w. Query: %v , Values: %v", err, t.Statement, allParamValues)
	}

	fields := results.FieldDescriptions()

	var out []any
	for results.Next() {
		v, err := results.Values()
		if err != nil {
			return nil, fmt.Errorf("unable to parse row: %w", err)
		}
		vMap := make(map[string]any)
		for i, f := range fields {
			vMap[f.Name] = v[i]
		}
		out = append(out, vMap)
	}

	return out, nil
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	return tools.ParseParams(t.Parameters, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t Tool) Authorized(verifiedAuthServices []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthServices)
}
