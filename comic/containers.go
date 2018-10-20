package comic

import "github.com/go-pg/pg"

// PGRepositoryContainer is the container for all the postgres repositories.
type PGRepositoryContainer struct {
	publisherRepository          PublisherRepository
	issueRepository              IssueRepository
	characterRepository          CharacterRepository
	characterIssueRepository     CharacterIssueRepository
	characterSourceRepository    CharacterSourceRepository
	characterSyncLogRepository   CharacterSyncLogRepository
	appearancesByYearsRepository *PGAppearancesByYearsRepository
	statsRepository              StatsRepository
}

// PublisherRepository gets the publisher repository.
func (c *PGRepositoryContainer) PublisherRepository() PublisherRepository {
	return c.publisherRepository
}

// IssueRepository gets the issue repository.
func (c *PGRepositoryContainer) IssueRepository() IssueRepository {
	return c.issueRepository
}

// CharacterRepository gets the character repository.
func (c *PGRepositoryContainer) CharacterRepository() CharacterRepository {
	return c.characterRepository
}

// CharacterIssueRepository gets the character issue repository.
func (c *PGRepositoryContainer) CharacterIssueRepository() CharacterIssueRepository {
	return c.characterIssueRepository
}

// CharacterSourceRepository gets the character source repository.
func (c *PGRepositoryContainer) CharacterSourceRepository() CharacterSourceRepository {
	return c.characterSourceRepository
}

// CharacterSyncLogRepository gets the character sync log repository.
func (c *PGRepositoryContainer) CharacterSyncLogRepository() CharacterSyncLogRepository {
	return c.characterSyncLogRepository
}

// AppearancesByYearsRepository gets the appearances per year repository.
func (c *PGRepositoryContainer) AppearancesByYearsRepository() *PGAppearancesByYearsRepository {
	return c.appearancesByYearsRepository
}

// StatsRepository gets the stats repository.
func (c *PGRepositoryContainer) StatsRepository() StatsRepository {
	return c.statsRepository
}

// NewPGRepositoryContainer creates the new postgres repository container.
func NewPGRepositoryContainer(db *pg.DB) *PGRepositoryContainer {
	return &PGRepositoryContainer{
		publisherRepository:          NewPGPublisherRepository(db),
		issueRepository:              NewPGIssueRepository(db),
		characterRepository:          NewPGCharacterRepository(db),
		characterSyncLogRepository:   NewPGCharacterSyncLogRepository(db),
		characterSourceRepository:    NewPGCharacterSourceRepository(db),
		characterIssueRepository:     NewPGCharacterIssueRepository(db),
		appearancesByYearsRepository: NewPGAppearancesPerYearRepository(db),
		statsRepository:              NewPGStatsRepository(db),
	}
}
