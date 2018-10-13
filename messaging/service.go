package messaging

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"go.uber.org/zap"
	"sync"
)

// CharacterMessageService creates the new messenger for a character.
type CharacterMessageService struct {
	messenger                  JSONMessenger
	characterRepository        comic.CharacterRepository
	characterSyncLogRepository comic.CharacterSyncLogRepository
}

func (s *CharacterMessageService) work(characters <-chan *comic.Character, wg *sync.WaitGroup) {
	for character := range characters {
		syncLog := comic.NewSyncLog(character.ID, comic.Pending, comic.YearlyAppearances, nil)
		err := s.characterSyncLogRepository.Create(syncLog)
		if err != nil {
			log.QUEUE().Error("error creating sync log", zap.Error(err))
		} else {
			sMessage := NewSyncMessage(character.Slug.Value(), syncLog.ID.Value())
			err := s.messenger.Send(sMessage)
			if err != nil {
				log.QUEUE().Error("error sending message to queue", zap.Error(err))
			}
		}
		wg.Done()
	}
}

// Send concurrently sends a message about characters from the specified criteria to the queue.
func (s *CharacterMessageService) Send(criteria comic.CharacterCriteria) error {
	characters, err := s.characterRepository.FindAll(criteria)
	if err != nil {
		return err
	}
	characterLen := len(characters)
	characterCh := make(chan *comic.Character, characterLen)
	var wg sync.WaitGroup
	wg.Add(characterLen)
	// do about 20 per job
	for i := 0; i < 20; i++ {
		go s.work(characterCh, &wg)
	}
	for _, character := range characters {
		characterCh <- character
	}
	close(characterCh)
	wg.Wait()
	log.QUEUE().Info("done processing characters", zap.Int("characters", len(characters)))
	return nil
}

// NewCharacterMessageService creates a new message service from params.
func NewCharacterMessageService(messenger JSONMessenger, repositoryContainer *comic.PGRepositoryContainer) *CharacterMessageService {
	return &CharacterMessageService{
		messenger:                  messenger,
		characterRepository:        repositoryContainer.CharacterRepository(),
		characterSyncLogRepository: repositoryContainer.CharacterSyncLogRepository(),
	}
}

// NewCharacterMessageServiceP creates a new character message service from params w/o container.
func NewCharacterMessageServiceP(messenger JSONMessenger, repository comic.CharacterRepository, syncLogRepository comic.CharacterSyncLogRepository) *CharacterMessageService {
	return &CharacterMessageService{
		messenger: messenger,
		characterRepository: repository,
		characterSyncLogRepository: syncLogRepository,
	}
}
