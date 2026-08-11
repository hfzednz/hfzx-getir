package postgres

import "database/sql"

// Repos groups search persistence adapters (lexical/vector stay in-memory).
type Repos struct {
	Docs     *DocumentRepo
	Synonyms *SynonymRepo
	Boosts   *BoostRepo
	Jobs     *IndexJobRepo
	Trends   *TrendRepo
	Suggests *SuggestRepo
	Outbox   *OutboxRepo
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Docs:     &DocumentRepo{DB: db},
		Synonyms: &SynonymRepo{DB: db},
		Boosts:   &BoostRepo{DB: db},
		Jobs:     &IndexJobRepo{DB: db},
		Trends:   &TrendRepo{DB: db},
		Suggests: &SuggestRepo{DB: db},
		Outbox:   &OutboxRepo{DB: db},
	}
}
