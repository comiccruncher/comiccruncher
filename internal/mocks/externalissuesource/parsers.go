// Code generated by MockGen. DO NOT EDIT.
// Source: parsers.go

// Package mock_externalissuesource is a generated GoMock package.
package mock_externalissuesource

import (
	externalissuesource "github.com/aimeelaplant/externalissuesource"
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockIssueParser is a mock of IssueParser interface
type MockIssueParser struct {
	ctrl     *gomock.Controller
	recorder *MockIssueParserMockRecorder
}

// MockIssueParserMockRecorder is the mock recorder for MockIssueParser
type MockIssueParserMockRecorder struct {
	mock *MockIssueParser
}

// NewMockIssueParser creates a new mock instance
func NewMockIssueParser(ctrl *gomock.Controller) *MockIssueParser {
	mock := &MockIssueParser{ctrl: ctrl}
	mock.recorder = &MockIssueParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIssueParser) EXPECT() *MockIssueParserMockRecorder {
	return m.recorder
}

// Parse mocks base method
func (m *MockIssueParser) Parse(body io.Reader) ([]externalissuesource.Issue, error) {
	ret := m.ctrl.Call(m, "Parse", body)
	ret0, _ := ret[0].([]externalissuesource.Issue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Parse indicates an expected call of Parse
func (mr *MockIssueParserMockRecorder) Parse(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Parse", reflect.TypeOf((*MockIssueParser)(nil).Parse), body)
}

// MockExternalIssueParser is a mock of ExternalIssueParser interface
type MockExternalIssueParser struct {
	ctrl     *gomock.Controller
	recorder *MockExternalIssueParserMockRecorder
}

// MockExternalIssueParserMockRecorder is the mock recorder for MockExternalIssueParser
type MockExternalIssueParserMockRecorder struct {
	mock *MockExternalIssueParser
}

// NewMockExternalIssueParser creates a new mock instance
func NewMockExternalIssueParser(ctrl *gomock.Controller) *MockExternalIssueParser {
	mock := &MockExternalIssueParser{ctrl: ctrl}
	mock.recorder = &MockExternalIssueParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExternalIssueParser) EXPECT() *MockExternalIssueParserMockRecorder {
	return m.recorder
}

// Issue mocks base method
func (m *MockExternalIssueParser) Issue(body io.Reader) (*externalissuesource.Issue, error) {
	ret := m.ctrl.Call(m, "Issue", body)
	ret0, _ := ret[0].(*externalissuesource.Issue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Issue indicates an expected call of Issue
func (mr *MockExternalIssueParserMockRecorder) Issue(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Issue", reflect.TypeOf((*MockExternalIssueParser)(nil).Issue), body)
}

// MockExternalCharacterParser is a mock of ExternalCharacterParser interface
type MockExternalCharacterParser struct {
	ctrl     *gomock.Controller
	recorder *MockExternalCharacterParserMockRecorder
}

// MockExternalCharacterParserMockRecorder is the mock recorder for MockExternalCharacterParser
type MockExternalCharacterParserMockRecorder struct {
	mock *MockExternalCharacterParser
}

// NewMockExternalCharacterParser creates a new mock instance
func NewMockExternalCharacterParser(ctrl *gomock.Controller) *MockExternalCharacterParser {
	mock := &MockExternalCharacterParser{ctrl: ctrl}
	mock.recorder = &MockExternalCharacterParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExternalCharacterParser) EXPECT() *MockExternalCharacterParserMockRecorder {
	return m.recorder
}

// Character mocks base method
func (m *MockExternalCharacterParser) Character(body io.Reader) (*externalissuesource.CharacterPage, error) {
	ret := m.ctrl.Call(m, "Character", body)
	ret0, _ := ret[0].(*externalissuesource.CharacterPage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Character indicates an expected call of Character
func (mr *MockExternalCharacterParserMockRecorder) Character(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Character", reflect.TypeOf((*MockExternalCharacterParser)(nil).Character), body)
}

// MockExternalCharacterSearchParser is a mock of ExternalCharacterSearchParser interface
type MockExternalCharacterSearchParser struct {
	ctrl     *gomock.Controller
	recorder *MockExternalCharacterSearchParserMockRecorder
}

// MockExternalCharacterSearchParserMockRecorder is the mock recorder for MockExternalCharacterSearchParser
type MockExternalCharacterSearchParserMockRecorder struct {
	mock *MockExternalCharacterSearchParser
}

// NewMockExternalCharacterSearchParser creates a new mock instance
func NewMockExternalCharacterSearchParser(ctrl *gomock.Controller) *MockExternalCharacterSearchParser {
	mock := &MockExternalCharacterSearchParser{ctrl: ctrl}
	mock.recorder = &MockExternalCharacterSearchParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExternalCharacterSearchParser) EXPECT() *MockExternalCharacterSearchParserMockRecorder {
	return m.recorder
}

// CharacterSearch mocks base method
func (m *MockExternalCharacterSearchParser) CharacterSearch(body io.Reader) (*externalissuesource.CharacterSearchResult, error) {
	ret := m.ctrl.Call(m, "CharacterSearch", body)
	ret0, _ := ret[0].(*externalissuesource.CharacterSearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CharacterSearch indicates an expected call of CharacterSearch
func (mr *MockExternalCharacterSearchParserMockRecorder) CharacterSearch(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CharacterSearch", reflect.TypeOf((*MockExternalCharacterSearchParser)(nil).CharacterSearch), body)
}

// MockExternalSourceParser is a mock of ExternalSourceParser interface
type MockExternalSourceParser struct {
	ctrl     *gomock.Controller
	recorder *MockExternalSourceParserMockRecorder
}

// MockExternalSourceParserMockRecorder is the mock recorder for MockExternalSourceParser
type MockExternalSourceParserMockRecorder struct {
	mock *MockExternalSourceParser
}

// NewMockExternalSourceParser creates a new mock instance
func NewMockExternalSourceParser(ctrl *gomock.Controller) *MockExternalSourceParser {
	mock := &MockExternalSourceParser{ctrl: ctrl}
	mock.recorder = &MockExternalSourceParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExternalSourceParser) EXPECT() *MockExternalSourceParserMockRecorder {
	return m.recorder
}

// Issue mocks base method
func (m *MockExternalSourceParser) Issue(body io.Reader) (*externalissuesource.Issue, error) {
	ret := m.ctrl.Call(m, "Issue", body)
	ret0, _ := ret[0].(*externalissuesource.Issue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Issue indicates an expected call of Issue
func (mr *MockExternalSourceParserMockRecorder) Issue(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Issue", reflect.TypeOf((*MockExternalSourceParser)(nil).Issue), body)
}

// Character mocks base method
func (m *MockExternalSourceParser) Character(body io.Reader) (*externalissuesource.CharacterPage, error) {
	ret := m.ctrl.Call(m, "Character", body)
	ret0, _ := ret[0].(*externalissuesource.CharacterPage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Character indicates an expected call of Character
func (mr *MockExternalSourceParserMockRecorder) Character(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Character", reflect.TypeOf((*MockExternalSourceParser)(nil).Character), body)
}

// CharacterSearch mocks base method
func (m *MockExternalSourceParser) CharacterSearch(body io.Reader) (*externalissuesource.CharacterSearchResult, error) {
	ret := m.ctrl.Call(m, "CharacterSearch", body)
	ret0, _ := ret[0].(*externalissuesource.CharacterSearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CharacterSearch indicates an expected call of CharacterSearch
func (mr *MockExternalSourceParserMockRecorder) CharacterSearch(body interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CharacterSearch", reflect.TypeOf((*MockExternalSourceParser)(nil).CharacterSearch), body)
}

// BaseUrl mocks base method
func (m *MockExternalSourceParser) BaseUrl() string {
	ret := m.ctrl.Call(m, "BaseUrl")
	ret0, _ := ret[0].(string)
	return ret0
}

// BaseUrl indicates an expected call of BaseUrl
func (mr *MockExternalSourceParserMockRecorder) BaseUrl() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseUrl", reflect.TypeOf((*MockExternalSourceParser)(nil).BaseUrl))
}
