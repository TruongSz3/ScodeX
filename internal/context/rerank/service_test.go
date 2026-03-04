package rerank

import (
	"context"
	"testing"
)

type fakeEngine struct {
	called     bool
	lastQuery  string
	lastInputs []Candidate
	outputs    []Candidate
}

func (f *fakeEngine) Rerank(_ context.Context, query string, candidates []Candidate) ([]Candidate, error) {
	f.called = true
	f.lastQuery = query
	f.lastInputs = append([]Candidate(nil), candidates...)

	return append([]Candidate(nil), f.outputs...), nil
}

func TestDefaultFeatureFlagsAreOff(t *testing.T) {
	t.Parallel()

	flags := DefaultFeatureFlags()
	if flags.SemanticRerankEnabled() {
		t.Fatal("semantic rerank should be disabled by default")
	}
}

func TestApplyPreservesLexicalOrderWhenFeatureIsOff(t *testing.T) {
	t.Parallel()

	engine := &fakeEngine{outputs: []Candidate{{DocumentID: "doc-2"}, {DocumentID: "doc-1"}}}
	svc := NewService(DefaultFeatureFlags(), engine)

	lexical := []Candidate{{DocumentID: "doc-1", Score: 10}, {DocumentID: "doc-2", Score: 9}}
	results, err := svc.Apply(context.Background(), "query", lexical)
	if err != nil {
		t.Fatalf("apply returned error: %v", err)
	}

	if engine.called {
		t.Fatal("expected rerank engine to not be called when feature flag is off")
	}
	if len(results) != 2 || results[0].DocumentID != "doc-1" || results[1].DocumentID != "doc-2" {
		t.Fatalf("expected lexical order to be preserved, got %+v", results)
	}
}

func TestApplyUsesEngineWhenFeatureIsOn(t *testing.T) {
	t.Parallel()

	engine := &fakeEngine{outputs: []Candidate{{DocumentID: "doc-2"}, {DocumentID: "doc-1"}}}
	svc := NewService(FeatureFlags{Semantic: true}, engine)

	lexical := []Candidate{{DocumentID: "doc-1", Score: 10}, {DocumentID: "doc-2", Score: 9}}
	results, err := svc.Apply(context.Background(), "semantic query", lexical)
	if err != nil {
		t.Fatalf("apply returned error: %v", err)
	}

	if !engine.called {
		t.Fatal("expected rerank engine to be called")
	}
	if engine.lastQuery != "semantic query" {
		t.Fatalf("expected query to be forwarded, got %q", engine.lastQuery)
	}
	if len(engine.lastInputs) != 2 || engine.lastInputs[0].DocumentID != "doc-1" {
		t.Fatalf("unexpected engine inputs: %+v", engine.lastInputs)
	}
	if len(results) != 2 || results[0].DocumentID != "doc-2" || results[1].DocumentID != "doc-1" {
		t.Fatalf("expected reranked order, got %+v", results)
	}
}
