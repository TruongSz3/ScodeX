package rerank

import "context"

type Candidate struct {
	DocumentID string
	ChunkID    string
	Content    string
	Score      float64
}

type Engine interface {
	Rerank(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error)
}

type Service struct {
	flags  FeatureFlags
	engine Engine
}

func NewService(flags FeatureFlags, engine Engine) *Service {
	return &Service{flags: flags, engine: engine}
}

func (s *Service) Apply(ctx context.Context, query string, lexical []Candidate) ([]Candidate, error) {
	preserved := append([]Candidate(nil), lexical...)
	if !s.flags.SemanticRerankEnabled() || s.engine == nil {
		return preserved, nil
	}

	return s.engine.Rerank(ctx, query, preserved)
}
