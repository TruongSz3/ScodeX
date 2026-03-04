package retrieval

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	lastRequest RepositorySearchRequest
	searchCalls int
	results     []RepositorySearchResult
	err         error
}

func (f *fakeRepository) Search(_ context.Context, req RepositorySearchRequest) ([]RepositorySearchResult, error) {
	f.lastRequest = req
	f.searchCalls++
	if f.err != nil {
		return nil, f.err
	}

	return append([]RepositorySearchResult(nil), f.results...), nil
}

func TestNewServiceRequiresRepository(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil)
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected ErrRepositoryRequired, got %v", err)
	}
}

func TestSearchValidatesInput(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("new service returned error: %v", err)
	}

	if _, err := svc.Search(context.Background(), SearchRequest{Query: "", Limit: 10, Offset: 0}); !errors.Is(err, ErrQueryRequired) {
		t.Fatalf("expected ErrQueryRequired, got %v", err)
	}
	if _, err := svc.Search(context.Background(), SearchRequest{Query: "hello", Limit: 10, Offset: -1}); !errors.Is(err, ErrInvalidOffset) {
		t.Fatalf("expected ErrInvalidOffset, got %v", err)
	}
}

func TestSearchNormalizesLimitAndPassesOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		limit         int
		expectedLimit int
	}{
		{name: "default for zero", limit: 0, expectedLimit: DefaultLimit},
		{name: "default for negative", limit: -10, expectedLimit: DefaultLimit},
		{name: "clamped to max", limit: MaxLimit + 50, expectedLimit: MaxLimit},
		{name: "kept when valid", limit: 15, expectedLimit: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepository{}
			svc, err := NewService(repo)
			if err != nil {
				t.Fatalf("new service returned error: %v", err)
			}

			response, err := svc.Search(context.Background(), SearchRequest{
				Query:  "golang",
				Limit:  tt.limit,
				Offset: 7,
			})
			if err != nil {
				t.Fatalf("search returned error: %v", err)
			}

			if repo.lastRequest.Limit != tt.expectedLimit {
				t.Fatalf("expected repo limit %d, got %d", tt.expectedLimit, repo.lastRequest.Limit)
			}
			if repo.lastRequest.Offset != 7 {
				t.Fatalf("expected repo offset 7, got %d", repo.lastRequest.Offset)
			}
			if response.Limit != tt.expectedLimit {
				t.Fatalf("expected response limit %d, got %d", tt.expectedLimit, response.Limit)
			}
			if response.Offset != 7 {
				t.Fatalf("expected response offset 7, got %d", response.Offset)
			}
		})
	}
}

func TestSearchReturnsScoredResultsInRepositoryOrder(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{results: []RepositorySearchResult{
		{DocumentID: "doc-2", ChunkID: "chunk-2", Content: "second", Score: 9.1},
		{DocumentID: "doc-1", ChunkID: "chunk-1", Content: "first", Score: 8.4},
	}}

	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("new service returned error: %v", err)
	}

	response, err := svc.Search(context.Background(), SearchRequest{Query: "ranking"})
	if err != nil {
		t.Fatalf("search returned error: %v", err)
	}

	if len(response.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(response.Results))
	}
	if response.Results[0].DocumentID != "doc-2" || response.Results[1].DocumentID != "doc-1" {
		t.Fatalf("expected lexical order to be preserved, got %+v", response.Results)
	}
	if response.Results[0].Score != 9.1 || response.Results[1].Score != 8.4 {
		t.Fatalf("scores were not preserved, got %+v", response.Results)
	}
}

func TestSearchPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo failed")
	repo := &fakeRepository{err: repoErr}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("new service returned error: %v", err)
	}

	_, err = svc.Search(context.Background(), SearchRequest{Query: "golang"})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
