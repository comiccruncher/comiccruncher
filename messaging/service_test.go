package messaging

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// MockJsonMessenger is a mock of JsonMessenger interface
type MockJsonMessenger struct {
	ctrl     *gomock.Controller
	recorder *MockJsonMessengerMockRecorder
}

// MockJsonMessengerMockRecorder is the mock recorder for MockJsonMessenger
type MockJsonMessengerMockRecorder struct {
	mock *MockJsonMessenger
}

// NewMockJsonMessenger creates a new mock instance
func NewMockJsonMessenger(ctrl *gomock.Controller) *MockJsonMessenger {
	mock := &MockJsonMessenger{ctrl: ctrl}
	mock.recorder = &MockJsonMessengerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJsonMessenger) EXPECT() *MockJsonMessengerMockRecorder {
	return m.recorder
}

// Send mocks base method
func (m *MockJsonMessenger) Send(arg0 JsonMessage) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockJsonMessengerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockJsonMessenger)(nil).Send), arg0)
}

func TestCharacterMessageService_Send(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	characterMock := mock_comic.NewMockCharacterRepository(ctrl)
	characters := mockCharacters(200)
	characterMock.EXPECT().FindAll(gomock.Any()).Return(characters, nil)
	characterSyncLogMock := mock_comic.NewMockCharacterSyncLogRepository(ctrl)
	characterSyncLogMock.EXPECT().Create(gomock.Any()).Return(nil).Times(len(characters))
	messenger := NewMockJsonMessenger(ctrl)
	messenger.EXPECT().Send(gomock.Any()).Return(nil).Times(len(characters))
	svc := CharacterMessageService{
		characterRepository:        characterMock,
		characterSyncLogRepository: characterSyncLogMock,
		messenger:                  messenger,
	}

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
