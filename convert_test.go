package yaml2json_test

import (
	"bytes"
	_ "embed"
	"errors"
	"io"
	"testing"

	"github.com/MarkRosemaker/yaml2json"
	"github.com/go-json-experiment/json/jsontext"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed example.yaml
	exampleYAML []byte
	//go:embed example.json
	exampleJSON jsontext.Value
)

func TestToJSON(t *testing.T) {
	t.Parallel()

	n := &yaml.Node{}
	if err := yaml.Unmarshal(exampleYAML, n); err != nil {
		t.Fatal(err)
	}

	got, err := yaml2json.Convert(n)
	if err != nil {
		t.Fatal(err)
	}

	equalJSON(t, got, exampleJSON)
}

func TestToJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("invalid node kind", func(t *testing.T) {
		if _, err := yaml2json.Convert(&yaml.Node{
			Kind: yaml.SequenceNode,
			Content: []*yaml.Node{
				{}, // invalid node kind
			},
		}); err == nil {
			t.Fatal("expected error")
		} else if !errors.Is(err, io.EOF) {
			t.Fatalf("got: %q, want: %q", err, io.EOF)
		}
	})

	t.Run("doc with invalid num of content nodes", func(t *testing.T) {
		if _, err := yaml2json.Convert(&yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{Kind: yaml.MappingNode},
				{Kind: yaml.ScalarNode},
			},
		}); err == nil {
			t.Fatal("expected error")
		} else if want := "expected 1 content node, got 2"; err.Error() != want {
			t.Fatalf("got: %q, want: %q", err.Error(), want)
		}
	})

	t.Run("unbalanced mapping node", func(t *testing.T) {
		if _, err := yaml2json.Convert(&yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode},
				},
			}},
		}); err == nil {
			t.Fatal("expected error")
		} else if want := "unbalanced mapping node"; err.Error() != want {
			t.Fatalf("got: %q, want: %q", err.Error(), want)
		}
	})

	t.Run("even mapping child node fails", func(t *testing.T) {
		if _, err := yaml2json.Convert(&yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode},
				{}, // invalid node kind
			},
		}); err == nil {
			t.Fatal("expected error")
		} else if !errors.Is(err, io.EOF) {
			t.Fatalf("got: %q, want: %q", err, io.EOF)
		}
	})

	t.Run("odd mapping child node fails", func(t *testing.T) {
		if _, err := yaml2json.Convert(&yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{}, // invalid node kind
				{Kind: yaml.ScalarNode},
			},
		}); err == nil {
			t.Fatal("expected error")
		} else if !errors.Is(err, io.EOF) {
			t.Fatalf("got: %q, want: %q", err, io.EOF)
		}
	})
}

func equalJSON(t *testing.T, got, want jsontext.Value) {
	t.Helper()

	if err := got.Indent("", "\t"); err != nil {
		t.Fatalf("formatting got: %v", err)
	}

	if err := want.Indent("", "\t"); err != nil {
		t.Fatalf("formatting want: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("got: %q, want: %q", got, want)
	}
}
