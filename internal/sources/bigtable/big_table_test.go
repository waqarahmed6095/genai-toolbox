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

package bigtable_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/sources/bigtable"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"gopkg.in/yaml.v3"
)

func TestParseFromYamlSpannerDb(t *testing.T) {
	tcs := []struct {
		desc string
		in   string
		want server.SourceConfigs
	}{
		{
			desc: "basic example",
			in: `
			sources:
				my-bigtable-instance:
					kind: bigtable
					project: my-project
			`,
			want: map[string]sources.SourceConfig{
				"my-bigtable-instance": bigtable.Config{
					Name:    "my-bigtable-instance",
					Kind:    bigtable.SourceKind,
					Project: "my-project",
					Dialect: "googlesql",
				},
			},
		},
		{
			desc: "gsql dialect",
			in: `
			sources:
				my-bigtable-instance:
					kind: bigtable
					project: my-project
					dialect: Googlesql 
			`,
			want: map[string]sources.SourceConfig{
				"my-bigtable-instance": bigtable.Config{
					Name:    "my-bigtable-instance",
					Kind:    bigtable.SourceKind,
					Project: "my-project",
					Dialect: "googlesql",
				},
			},
		},
		{
			desc: "postgresql dialect",
			in: `
			sources:
				my-bigtable-instance:
					kind: bigtable
					project: my-project
					dialect: sql
			`,
			want: map[string]sources.SourceConfig{
				"my-bigtable-instance": bigtable.Config{
					Name:    "my-bigtable-instance",
					Kind:    bigtable.SourceKind,
					Project: "my-project",
					Dialect: "sql",
				},
			},
		},
		{
			desc: "invalid dialect",
			in: `
			sources:
				my-bigtable-instance:
					kind: bigtable
					project: my-project
                    dialect: fail
			`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Sources server.SourceConfigs `yaml:"sources"`
			}{}
			// Parse contents
			err := yaml.Unmarshal(testutils.FormatYaml(tc.in), &got)
			if err != nil {
				if tc.want == nil {
					return
				}
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if !cmp.Equal(tc.want, got.Sources) {
				if tc.want == nil {
					return
				}
				t.Fatalf("incorrect parse: want %v, got %v", tc.want, got.Sources)
			}
		})
	}

}
