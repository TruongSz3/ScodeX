package rerank

type FeatureFlags struct {
	Semantic bool
}

func DefaultFeatureFlags() FeatureFlags {
	return FeatureFlags{Semantic: false}
}

func (f FeatureFlags) SemanticRerankEnabled() bool {
	return f.Semantic
}
