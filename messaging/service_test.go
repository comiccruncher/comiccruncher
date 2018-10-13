package messaging_test

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/messaging"
	"github.com/aimeelaplant/comiccruncher/messaging"
)

func TestCharacterMessageService_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	characterMock := mock_comic.NewMockCharacterRepository(ctrl)
	characters := mockCharacters(200)
	characterMock.EXPECT().FindAll(gomock.Any()).Return(characters, nil)
	characterSyncLogMock := mock_comic.NewMockCharacterSyncLogRepository(ctrl)
	characterSyncLogMock.EXPECT().Create(gomock.Any()).Return(nil).Times(len(characters))
	ctrlr := gomock.NewController(t)
	msngr := mock_messaging.NewMockJSONMessenger(ctrlr)
	msngr.EXPECT().Send(gomock.Any()).Return(nil).Times(len(characters))

	svc := messaging.NewCharacterMessageServiceP(msngr, characterMock, characterSyncLogMock)
	err := svc.Send(comic.CharacterCriteria{})
	assert.Nil(t, err)
}

func mockCharacters(length int) []*comic.Character {
	characters := make([]*comic.Character, 0)
	for i := 1; i <= length; i++ {
		characters = append(characters, &comic.Character{ID: comic.CharacterID(i), Slug: comic.CharacterSlug(fmt.Sprintf("slug-%d", i))})
	}
	return characters
}
