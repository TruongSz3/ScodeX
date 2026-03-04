package retrieval

import (
	"context"
	"errors"
)

const (
	DefaultLimit = 20
	MaxLimit     = 200
)

var (
	ErrRepositoryRequired = errors.New("retrieval: repository is required")
	ErrQueryRequired      = errors.New("retrieval: query is required")
	ErrInvalidOffset      = errors.New("retrieval: offset must be greater than or equal to zero")
)

type SearchRequest struct {
	Query  string
	Limit  int
	Offset int
}

type ScoredResult struct {
	DocumentID string
	ChunkID    string
	Content    string
	Score      float64
}

type SearchResponse struct {
	Results []ScoredResult
	Limit   int
	Offset  int
}

type RepositorySearchRequest struct {
	Query  string
	Limit  int
	Offset int
}

type RepositorySearchResult struct {
	DocumentID string
	ChunkID    string
	Content    string
	Score      float64
}

type Repository interface {
	Search(ctx context.Context, req RepositorySearchRequest) ([]RepositorySearchResult, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, ErrRepositoryRequired
	}

	return &Service{repo: repo}, nil
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	if req.Query == "" {
		return SearchResponse{}, ErrQueryRequired
	}
	if req.Offset < 0 {
		return SearchResponse{}, ErrInvalidOffset
	}

	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	repoResults, err := s.repo.Search(ctx, RepositorySearchRequest{
		Query:  req.Query,
		Limit:  limit,
		Offset: req.Offset,
	})
	if err != nil {
		return SearchResponse{}, err
	}

	results := make([]ScoredResult, 0, len(repoResults))
	for _, repoResult := range repoResults {
		results = append(results, ScoredResult{
			DocumentID: repoResult.DocumentID,
			ChunkID:    repoResult.ChunkID,
			Content:    repoResult.Content,
			Score:      repoResult.Score,
		})
	}

	return SearchResponse{
		Results: results,
		Limit:   limit,
		Offset:  req.Offset,
	}, nil
}
