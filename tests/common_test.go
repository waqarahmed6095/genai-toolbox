//go:build integration

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

// Package tests contains end to end tests meant to verify the Toolbox Server
// works as expected when executed as a binary.

package tests

import (
	"database/sql"
	"fmt"

	"github.com/googleapis/genai-toolbox/internal/tools"

	bigqueryapi "cloud.google.com/go/bigquery"
)

// GetToolsConfig returns a mock tools config file
func GetToolsConfig(sourceConfig map[string]any, toolKind, param_tool_statement, auth_tool_statement string) map[string]any {
	// Write config into a file and pass it to command
	toolsFile := map[string]any{
		"sources": map[string]any{
			"my-instance": sourceConfig,
		},
		"authServices": map[string]any{
			"my-google-auth": map[string]any{
				"kind":     "google",
				"clientId": ClientId,
			},
		},
		"tools": map[string]any{
			"my-simple-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"description": "Simple tool to test end to end functionality.",
				"statement":   "SELECT 1;",
			},
			"my-param-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"description": "Tool to test invocation with params.",
				"statement":   param_tool_statement,
				"parameters": []any{
					map[string]any{
						"name":        "id",
						"type":        "integer",
						"description": "user ID",
					},
					map[string]any{
						"name":        "name",
						"type":        "string",
						"description": "user name",
					},
				},
			},
			"my-auth-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"description": "Tool to test authenticated parameters.",
				// statement to auto-fill authenticated parameter
				"statement": auth_tool_statement,
				"parameters": []map[string]any{
					{
						"name":        "email",
						"type":        "string",
						"description": "user email",
						"authServices": []map[string]string{
							{
								"name":  "my-google-auth",
								"field": "email",
							},
						},
					},
				},
			},
			"my-auth-required-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"description": "Tool to test auth required invocation.",
				"statement":   "SELECT 1;",
				"authRequired": []string{
					"my-google-auth",
				},
			},
			"my-fail-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"description": "Tool to test statement with incorrect syntax.",
				"statement":   "SELEC 1;",
			},
		},
	}

	return toolsFile
}

// GetHTTPToolsConfig returns a mock HTTP tool's config file
func GetHTTPToolsConfig(sourceConfig map[string]any, toolKind string) map[string]any {
	// Write config into a file and pass it to command
	otherSourceConfig := make(map[string]any)
	for k, v := range sourceConfig {
		otherSourceConfig[k] = v
	}
	otherSourceConfig["headers"] = map[string]string{"X-Custom-Header": "unexpected", "Content-Type": "application/json"}
	otherSourceConfig["queryParams"] = map[string]any{"id": 1, "name": "Sid"}

	toolsFile := map[string]any{
		"sources": map[string]any{
			"my-instance":    sourceConfig,
			"other-instance": otherSourceConfig,
		},
		"authServices": map[string]any{
			"my-google-auth": map[string]any{
				"kind":     "google",
				"clientId": ClientId,
			},
		},
		"tools": map[string]any{
			"my-simple-tool": map[string]any{
				"kind":        toolKind,
				"path":        "/tool0",
				"method":      "POST",
				"source":      "my-instance",
				"requestBody": "{}",
				"description": "Simple tool to test end to end functionality.",
			},
			"my-param-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"method":      "GET",
				"path":        "/tool1",
				"description": "some description",
				"queryParams": []tools.Parameter{
					tools.NewIntParameter("id", "user ID")},
				"requestBody": `{
"age": 36,
"name": "{{.name}}"
}
`,
				"bodyParams": []tools.Parameter{tools.NewStringParameter("name", "user name")},
				"headers":    map[string]string{"Content-Type": "application/json"},
			},
			"my-auth-tool": map[string]any{
				"kind":        toolKind,
				"source":      "my-instance",
				"method":      "GET",
				"path":        "/tool2",
				"description": "some description",
				"requestBody": "{}",
				"queryParams": []tools.Parameter{
					tools.NewStringParameterWithAuth("email", "some description",
						[]tools.ParamAuthService{{Name: "my-google-auth", Field: "email"}}),
				},
			},
			"my-auth-required-tool": map[string]any{
				"kind":         toolKind,
				"source":       "my-instance",
				"method":       "POST",
				"path":         "/tool0",
				"description":  "some description",
				"requestBody":  "{}",
				"authRequired": []string{"my-google-auth"},
			},
			"my-advanced-tool": map[string]any{
				"kind":        toolKind,
				"source":      "other-instance",
				"method":      "get",
				"path":        "/tool3?id=2",
				"description": "some description",
				"headers": map[string]string{
					"X-Custom-Header": "example",
				},
				"queryParams": []tools.Parameter{
					tools.NewIntParameter("id", "user ID"), tools.NewStringParameter("country", "country")},
				"requestBody": `{
"place": "zoo",
"animals": {{json .animalArray }}
}
`,
				"bodyParams":   []tools.Parameter{tools.NewArrayParameter("animalArray", "animals in the zoo", tools.NewStringParameter("animals", "desc"))},
				"headerParams": []tools.Parameter{tools.NewStringParameter("X-Other-Header", "custom header")},
			},
		},
	}
	return toolsFile
}

// GetPostgresSQLParamToolInfo returns statements and param for my-param-tool postgres-sql kind
func GetPostgresSQLParamToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT);", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1), ($2), ($3);", tableName)
	tool_statement := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 OR name = $2;", tableName)
	params := []any{"Alice", "Jane", "Sid"}
	return create_statement, insert_statement, tool_statement, params
}

// GetPostgresSQLAuthToolInfo returns statements and param of my-auth-tool for postgres-sql kind
func GetPostgresSQLAuthToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT, email TEXT);", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name, email) VALUES ($1, $2), ($3, $4)", tableName)
	tool_statement := fmt.Sprintf("SELECT name FROM %s WHERE email = $1;", tableName)
	params := []any{"Alice", SERVICE_ACCOUNT_EMAIL, "Jane", "janedoe@gmail.com"}
	return create_statement, insert_statement, tool_statement, params
}

// GetMssqlParamToolInfo returns statements and param for my-param-tool mssql-sql kind
func GetMssqlParamToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT IDENTITY(1,1) PRIMARY KEY, name VARCHAR(255));", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name) VALUES (@alice), (@jane), (@sid);", tableName)
	tool_statement := fmt.Sprintf("SELECT * FROM %s WHERE id = @id OR name = @p2;", tableName)
	params := []any{sql.Named("alice", "Alice"), sql.Named("jane", "Jane"), sql.Named("sid", "Sid")}
	return create_statement, insert_statement, tool_statement, params
}

// GetMssqlLAuthToolInfo returns statements and param of my-auth-tool for mssql-sql kind
func GetMssqlLAuthToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT IDENTITY(1,1) PRIMARY KEY, name VARCHAR(255), email VARCHAR(255));", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name, email) VALUES (@alice, @aliceemail), (@jane, @janeemail);", tableName)
	tool_statement := fmt.Sprintf("SELECT name FROM %s WHERE email = @email;", tableName)
	params := []any{sql.Named("alice", "Alice"), sql.Named("aliceemail", SERVICE_ACCOUNT_EMAIL), sql.Named("jane", "Jane"), sql.Named("janeemail", "janedoe@gmail.com")}
	return create_statement, insert_statement, tool_statement, params
}

// GetMysqlParamToolInfo returns statements and param for my-param-tool mssql-sql kind
func GetMysqlParamToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255));", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name) VALUES (?), (?), (?);", tableName)
	tool_statement := fmt.Sprintf("SELECT * FROM %s WHERE id = ? OR name = ?;", tableName)
	params := []any{"Alice", "Jane", "Sid"}
	return create_statement, insert_statement, tool_statement, params
}

// GetMysqlLAuthToolInfo returns statements and param of my-auth-tool for mssql-sql kind
func GetMysqlLAuthToolInfo(tableName string) (string, string, string, []any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), email VARCHAR(255));", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (name, email) VALUES (?, ?), (?, ?)", tableName)
	tool_statement := fmt.Sprintf("SELECT name FROM %s WHERE email = ?;", tableName)
	params := []any{"Alice", SERVICE_ACCOUNT_EMAIL, "Jane", "janedoe@gmail.com"}
	return create_statement, insert_statement, tool_statement, params
}

// GetSpannerToolInfo returns statements and param for my-param-tool for spanner-sql kind
func GetSpannerParamToolInfo(tableName string) (string, string, string, map[string]any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT64, name STRING(MAX)) PRIMARY KEY (id)", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (id, name) VALUES (1, @name1), (2, @name2), (3, @name3)", tableName)
	tool_statement := fmt.Sprintf("SELECT * FROM %s WHERE id = @id OR name = @name", tableName)
	params := map[string]any{"name1": "Alice", "name2": "Jane", "name3": "Sid"}
	return create_statement, insert_statement, tool_statement, params
}

// GetSpannerAuthToolInfo returns statements and param of my-auth-tool for spanner-sql kind
func GetSpannerAuthToolInfo(tableName string) (string, string, string, map[string]any) {
	create_statement := fmt.Sprintf("CREATE TABLE %s (id INT64, name STRING(MAX), email STRING(MAX)) PRIMARY KEY (id)", tableName)
	insert_statement := fmt.Sprintf("INSERT INTO %s (id, name, email) VALUES (1, @name1, @email1), (2, @name2, @email2)", tableName)
	tool_statement := fmt.Sprintf("SELECT name FROM %s WHERE email = @email", tableName)
	params := map[string]any{
		"name1":  "Alice",
		"email1": SERVICE_ACCOUNT_EMAIL,
		"name2":  "Jane",
		"email2": "janedoe@gmail.com",
	}
	return create_statement, insert_statement, tool_statement, params
}

// GetBigQueryParamToolInfo returns statements and param for my-param-tool for bigquery kind
func GetBigQueryParamToolInfo(projectID, datasetID, tableName string) (string, string, string, []bigqueryapi.QueryParameter) {
	createStatement := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (id INT64, name STRING);`, tableName)
	insertStatement := fmt.Sprintf(`
		INSERT INTO %s (id, name) VALUES (?, ?), (?, ?), (?, ?);`, tableName)
	toolStatement := fmt.Sprintf(`SELECT * FROM %s WHERE id = ? OR name = ? ORDER BY id;`, tableName)
	params := []bigqueryapi.QueryParameter{
		{Value: int64(1)}, {Value: "Alice"},
		{Value: int64(2)}, {Value: "Jane"},
		{Value: int64(3)}, {Value: "Sid"},
	}
	return createStatement, insertStatement, toolStatement, params
}

// GetBigQueryAuthToolInfo returns statements and param of my-auth-tool for bigquery kind
func GetBigQueryAuthToolInfo(projectID, datasetID, tableName string) (string, string, string, []bigqueryapi.QueryParameter) {
	createStatement := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (id INT64, name STRING, email STRING)`, tableName)
	insertStatement := fmt.Sprintf(`
		INSERT INTO %s (id, name, email) VALUES (?, ?, ?), (?, ?, ?)`, tableName)
	toolStatement := fmt.Sprintf(`
		SELECT name FROM %s WHERE email = ?`, tableName)
	params := []bigqueryapi.QueryParameter{
		{Value: int64(1)}, {Value: "Alice"}, {Value: SERVICE_ACCOUNT_EMAIL},
		{Value: int64(2)}, {Value: "Jane"}, {Value: "janedoe@gmail.com"},
	}
	return createStatement, insertStatement, toolStatement, params
}
