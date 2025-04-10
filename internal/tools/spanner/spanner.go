// Copyright 2024 Google LLC
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

package spanner

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/googleapis/genai-toolbox/internal/sources"
	spannerdb "github.com/googleapis/genai-toolbox/internal/sources/spanner"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"google.golang.org/api/iterator"
)

const ToolKind string = "spanner-sql"

type compatibleSource interface {
	SpannerClient() *spanner.Client
	DatabaseDialect() string
}

// validate compatible sources are still compatible
var _ compatibleSource = &spannerdb.Source{}

var compatibleSources = [...]string{spannerdb.SourceKind}

type Config struct {
	Name         string           `yaml:"name" validate:"required"`
	Kind         string           `yaml:"kind" validate:"required"`
	Source       string           `yaml:"source" validate:"required"`
	Description  string           `yaml:"description" validate:"required"`
	Statement    string           `yaml:"statement" validate:"required"`
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

	mcpManifest := tools.McpManifest{
		Name:        cfg.Name,
		Description: cfg.Description,
		InputSchema: cfg.Parameters.McpManifest(),
	}

	// Determin the type of the statement
	readOnly, err := isReadOnlyTransaction(cfg.Statement)
	if err != nil {
		return nil, fmt.Errorf("failed to determine the statement type %s", err)
	}

	// finish tool setup
	t := Tool{
		Name:         cfg.Name,
		Kind:         ToolKind,
		Parameters:   cfg.Parameters,
		Statement:    cfg.Statement,
		AuthRequired: cfg.AuthRequired,
		Client:       s.SpannerClient(),
		dialect:      s.DatabaseDialect(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: cfg.Parameters.Manifest()},
		mcpManifest:  mcpManifest,
		isReadOnly:   readOnly,
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

	Client      *spanner.Client
	Statement   string
	dialect     string
	manifest    tools.Manifest
	mcpManifest tools.McpManifest

	isReadOnly bool
}

func getMapParams(params tools.ParamValues, dialect string) (map[string]interface{}, error) {
	switch strings.ToLower(dialect) {
	case "googlesql":
		return params.AsMap(), nil
	case "postgresql":
		return params.AsMapByOrderedKeys(), nil
	default:
		return nil, fmt.Errorf("invalid dialect %s", dialect)
	}
}

func (t Tool) Invoke(params tools.ParamValues) ([]any, error) {
	mapParams, err := getMapParams(params, t.dialect)
	if err != nil {
		return nil, fmt.Errorf("fail to get map params: %w", err)
	}

	var out []any

	_, err = t.Client.ReadWriteTransaction(context.Background(), func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL:    t.Statement,
			Params: mapParams,
		}
		var iter *RowIterator
		switch t.isReadOnly 
		case true:
			iter := txn.Query(ctx, stmt)
		case false:
			iter := txn.Update(ctx, stmt)
		defer iter.Stop()

		for {
			row, err := iter.Next()
			if err == iterator.Done {
				return nil
			}
			if err != nil {
				return fmt.Errorf("unable to parse row: %w", err)
			}

			vMap := make(map[string]any)
			cols := row.ColumnNames()
			for i, c := range cols {
				vMap[c] = row.ColumnValue(i)
			}

			out = append(out, vMap)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute client: %w", err)
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

// The following helper functions are copied and adopted from github.com/googleapis/go-sql-spanner/
// The purpose is to determine the read/write transaction type of the
var readKeywords = map[string]bool{"SELECT": true, "WITH": true, "GRAPH": true, "FROM": true}
var writeKeywords = map[string]bool{"INSERT": true, "UPDATE": true, "DELETE": true}
var allKeywords = map[string]bool{"SELECT": true, "WITH": true, "GRAPH": true, "FROM": true, "INSERT": true, "UPDATE": true, "DELETE": true}

// RemoveCommentsAndTrim removes any comments in the query string and trims any
// spaces at the beginning and end of the query. This makes checking what type
// of query a string is a lot easier, as only the first word(s) need to be
// checked after this has been removed.
func removeCommentsAndTrim(sql string) (string, error) {
	const singleQuote = '\''
	const doubleQuote = '"'
	const backtick = '`'
	const hyphen = '-'
	const dash = '#'
	const slash = '/'
	const asterisk = '*'
	isInQuoted := false
	isInSingleLineComment := false
	isInMultiLineComment := false
	var startQuote rune
	lastCharWasEscapeChar := false
	isTripleQuoted := false
	res := strings.Builder{}
	res.Grow(len(sql))
	index := 0
	runes := []rune(sql)
	for index < len(runes) {
		c := runes[index]
		if isInQuoted {
			if (c == '\n' || c == '\r') && !isTripleQuoted {
				return "", fmt.Errorf("statement contains an unclosed literal: %s", sql)
			} else if c == startQuote {
				if lastCharWasEscapeChar {
					lastCharWasEscapeChar = false
				} else if isTripleQuoted {
					if len(runes) > index+2 && runes[index+1] == startQuote && runes[index+2] == startQuote {
						isInQuoted = false
						startQuote = 0
						isTripleQuoted = false
						res.WriteRune(c)
						res.WriteRune(c)
						index += 2
					}
				} else {
					isInQuoted = false
					startQuote = 0
				}
			} else if c == '\\' {
				lastCharWasEscapeChar = true
			} else {
				lastCharWasEscapeChar = false
			}
			res.WriteRune(c)
		} else {
			// We are not in a quoted string.
			if isInSingleLineComment {
				if c == '\n' {
					isInSingleLineComment = false
					// Include the line feed in the result.
					res.WriteRune(c)
				}
			} else if isInMultiLineComment {
				if len(runes) > index+1 && c == asterisk && runes[index+1] == slash {
					isInMultiLineComment = false
					index++
				}
			} else {
				if c == dash || (len(runes) > index+1 && c == hyphen && runes[index+1] == hyphen) {
					// This is a single line comment.
					isInSingleLineComment = true
				} else if len(runes) > index+1 && c == slash && runes[index+1] == asterisk {
					isInMultiLineComment = true
					index++
				} else {
					if c == singleQuote || c == doubleQuote || c == backtick {
						isInQuoted = true
						startQuote = c
						// Check whether it is a triple-quote.
						if len(runes) > index+2 && runes[index+1] == startQuote && runes[index+2] == startQuote {
							isTripleQuoted = true
							res.WriteRune(c)
							res.WriteRune(c)
							index += 2
						}
					}
					res.WriteRune(c)
				}
			}
		}
		index++
	}
	if isInQuoted {
		return "", fmt.Errorf("statement contains an unclosed literal: %s", sql)
	}
	trimmed := strings.TrimSpace(res.String())
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == ';' {
		return trimmed[:len(trimmed)-1], nil
	}
	return trimmed, nil
}

// Removes any statement hints at the beginning of the statement.
// It assumes that any comments have already been removed.
func removeStatementHint(sql string) string {
	// Return quickly if the statement does not start with a hint.
	if len(sql) < 2 || sql[0] != '@' {
		return sql
	}

	// Valid statement hints at the beginning of a query statement can only contain a fixed set of
	// possible values. Although it is possible to add a @{FORCE_INDEX=...} as a statement hint, the
	// only allowed value is _BASE_TABLE. This means that we can safely assume that the statement
	// hint will not contain any special characters, for example a closing curly brace or one of the
	// keywords SELECT, UPDATE, DELETE, WITH, and that we can keep the check simple by just
	// searching for the first occurrence of a keyword that should be preceded by a closing curly
	// brace at the end of the statement hint.
	startStatementHintIndex := strings.Index(sql, "{")
	// Statement hints are allowed for both queries and DML statements.
	startQueryIndex := -1
	upperCaseSql := strings.ToUpper(sql)
	for keyword := range allKeywords {
		if startQueryIndex = strings.Index(upperCaseSql, keyword); startQueryIndex > -1 {
			break
		}
	}
	// The startQueryIndex can theoretically be larger than the length of the SQL string,
	// as the length of the uppercase SQL string can be different from the length of the
	// lower/mixed case SQL string. This is however only the case for specific non-ASCII
	// characters that are not allowed in a statement hint, so in that case we can safely
	// assume the statement to be invalid.
	if startQueryIndex > -1 && startQueryIndex < len(sql) {
		endStatementHintIndex := strings.LastIndex(sql[:startQueryIndex], "}")
		if startStatementHintIndex == -1 || startStatementHintIndex > endStatementHintIndex || endStatementHintIndex >= len(sql)-1 {
			// Looks like an invalid statement hint. Just ignore at this point
			// and let the caller handle the invalid query.
			return sql
		}
		return strings.TrimSpace(sql[endStatementHintIndex+1:])
	}
	// Seems invalid, just return the original statement.
	return sql
}

// Return True of the SQL statement is read-only
func isReadOnlyTransaction(sql string) (bool, error) {
	sql, err := removeCommentsAndTrim(sql)
	if err != nil {
		return false, err
	}
	sql = removeStatementHint(sql)

	// We can safely check if the string starts with a specific string, as we
	// have already removed all leading spaces, and there are no keywords that
	// start with the same substring as one of the keywords.
	for keyword := range writeKeywords {
		if len(sql) >= len(keyword) && strings.EqualFold(sql[:len(keyword)], keyword) {
			return false, nil
		}
	}
	return true, nil
}
