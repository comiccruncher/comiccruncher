// Code generated by MockGen. DO NOT EDIT.
// Source: comic/services.go

// Package mock_comic is a generated GoMock package.
package mock_comic

import (
	comic "github.com/comiccruncher/comiccruncher/comic"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockPublisherServicer is a mock of PublisherServicer interface
type MockPublisherServicer struct {
	ctrl     *gomock.Controller
	recorder *MockPublisherServicerMockRecorder
}

// MockPublisherServicerMockRecorder is the mock recorder for MockPublisherServicer
type MockPublisherServicerMockRecorder struct {
	mock *MockPublisherServicer
}

// NewMockPublisherServicer creates a new mock instance
func NewMockPublisherServicer(ctrl *gomock.Controller) *MockPublisherServicer {
	mock := &MockPublisherServicer{ctrl: ctrl}
	mock.recorder = &MockPublisherServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPublisherServicer) EXPECT() *MockPublisherServicerMockRecorder {
	return m.recorder
}

// Publisher mocks base method
func (m *MockPublisherServicer) Publisher(slug comic.PublisherSlug) (*comic.Publisher, error) {
	ret := m.ctrl.Call(m, "Publisher", slug)
	ret0, _ := ret[0].(*comic.Publisher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Publisher indicates an expected call of Publisher
func (mr *MockPublisherServicerMockRecorder) Publisher(slug interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publisher", reflect.TypeOf((*MockPublisherServicer)(nil).Publisher), slug)
}

// MockIssueServicer is a mock of IssueServicer interface
type MockIssueServicer struct {
	ctrl     *gomock.Controller
	recorder *MockIssueServicerMockRecorder
}

// MockIssueServicerMockRecorder is the mock recorder for MockIssueServicer
type MockIssueServicerMockRecorder struct {
	mock *MockIssueServicer
}

// NewMockIssueServicer creates a new mock instance
func NewMockIssueServicer(ctrl *gomock.Controller) *MockIssueServicer {
	mock := &MockIssueServicer{ctrl: ctrl}
	mock.recorder = &MockIssueServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIssueServicer) EXPECT() *MockIssueServicerMockRecorder {
	return m.recorder
}

// Issues mocks base method
func (m *MockIssueServicer) Issues(ids []comic.IssueID, limit, offset int) ([]*comic.Issue, error) {
	ret := m.ctrl.Call(m, "Issues", ids, limit, offset)
	ret0, _ := ret[0].([]*comic.Issue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Issues indicates an expected call of Issues
func (mr *MockIssueServicerMockRecorder) Issues(ids, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Issues", reflect.TypeOf((*MockIssueServicer)(nil).Issues), ids, limit, offset)
}

// IssuesByVendor mocks base method
func (m *MockIssueServicer) IssuesByVendor(vendorIds []string, vendorType comic.VendorType, limit, offset int) ([]*comic.Issue, error) {
	ret := m.ctrl.Call(m, "IssuesByVendor", vendorIds, vendorType, limit, offset)
	ret0, _ := ret[0].([]*comic.Issue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IssuesByVendor indicates an expected call of IssuesByVendor
func (mr *MockIssueServicerMockRecorder) IssuesByVendor(vendorIds, vendorType, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IssuesByVendor", reflect.TypeOf((*MockIssueServicer)(nil).IssuesByVendor), vendorIds, vendorType, limit, offset)
}

// Create mocks base method
func (m *MockIssueServicer) Create(issue *comic.Issue) error {
	ret := m.ctrl.Call(m, "Create", issue)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockIssueServicerMockRecorder) Create(issue interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockIssueServicer)(nil).Create), issue)
}

// CreateP mocks base method
func (m *MockIssueServicer) CreateP(vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber string, pubDate, saleDate time.Time, isVariant, isMonthUncertain, isReprint bool, format comic.Format) error {
	ret := m.ctrl.Call(m, "CreateP", vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber, pubDate, saleDate, isVariant, isMonthUncertain, isReprint, format)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateP indicates an expected call of CreateP
func (mr *MockIssueServicerMockRecorder) CreateP(vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber, pubDate, saleDate, isVariant, isMonthUncertain, isReprint, format interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateP", reflect.TypeOf((*MockIssueServicer)(nil).CreateP), vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber, pubDate, saleDate, isVariant, isMonthUncertain, isReprint, format)
}

// MockCharacterServicer is a mock of CharacterServicer interface
type MockCharacterServicer struct {
	ctrl     *gomock.Controller
	recorder *MockCharacterServicerMockRecorder
}

// MockCharacterServicerMockRecorder is the mock recorder for MockCharacterServicer
type MockCharacterServicerMockRecorder struct {
	mock *MockCharacterServicer
}

// NewMockCharacterServicer creates a new mock instance
func NewMockCharacterServicer(ctrl *gomock.Controller) *MockCharacterServicer {
	mock := &MockCharacterServicer{ctrl: ctrl}
	mock.recorder = &MockCharacterServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCharacterServicer) EXPECT() *MockCharacterServicerMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockCharacterServicer) Create(character *comic.Character) error {
	ret := m.ctrl.Call(m, "Create", character)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockCharacterServicerMockRecorder) Create(character interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockCharacterServicer)(nil).Create), character)
}

// Character mocks base method
func (m *MockCharacterServicer) Character(slug comic.CharacterSlug) (*comic.Character, error) {
	ret := m.ctrl.Call(m, "Character", slug)
	ret0, _ := ret[0].(*comic.Character)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Character indicates an expected call of Character
func (mr *MockCharacterServicerMockRecorder) Character(slug interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Character", reflect.TypeOf((*MockCharacterServicer)(nil).Character), slug)
}

// Update mocks base method
func (m *MockCharacterServicer) Update(character *comic.Character) error {
	ret := m.ctrl.Call(m, "Update", character)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockCharacterServicerMockRecorder) Update(character interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockCharacterServicer)(nil).Update), character)
}

// UpdateAll mocks base method
func (m *MockCharacterServicer) UpdateAll(characters []*comic.Character) error {
	ret := m.ctrl.Call(m, "UpdateAll", characters)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAll indicates an expected call of UpdateAll
func (mr *MockCharacterServicerMockRecorder) UpdateAll(characters interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAll", reflect.TypeOf((*MockCharacterServicer)(nil).UpdateAll), characters)
}

// CharactersWithSources mocks base method
func (m *MockCharacterServicer) CharactersWithSources(slug []comic.CharacterSlug, limit, offset int) ([]*comic.Character, error) {
	ret := m.ctrl.Call(m, "CharactersWithSources", slug, limit, offset)
	ret0, _ := ret[0].([]*comic.Character)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CharactersWithSources indicates an expected call of CharactersWithSources
func (mr *MockCharacterServicerMockRecorder) CharactersWithSources(slug, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CharactersWithSources", reflect.TypeOf((*MockCharacterServicer)(nil).CharactersWithSources), slug, limit, offset)
}

// Characters mocks base method
func (m *MockCharacterServicer) Characters(slugs []comic.CharacterSlug, limit, offset int) ([]*comic.Character, error) {
	ret := m.ctrl.Call(m, "Characters", slugs, limit, offset)
	ret0, _ := ret[0].([]*comic.Character)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Characters indicates an expected call of Characters
func (mr *MockCharacterServicerMockRecorder) Characters(slugs, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Characters", reflect.TypeOf((*MockCharacterServicer)(nil).Characters), slugs, limit, offset)
}

// CharacterByVendor mocks base method
func (m *MockCharacterServicer) CharacterByVendor(vendorID string, vendorType comic.VendorType, includeIsDisabled bool) (*comic.Character, error) {
	ret := m.ctrl.Call(m, "CharacterByVendor", vendorID, vendorType, includeIsDisabled)
	ret0, _ := ret[0].(*comic.Character)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CharacterByVendor indicates an expected call of CharacterByVendor
func (mr *MockCharacterServicerMockRecorder) CharacterByVendor(vendorID, vendorType, includeIsDisabled interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CharacterByVendor", reflect.TypeOf((*MockCharacterServicer)(nil).CharacterByVendor), vendorID, vendorType, includeIsDisabled)
}

// CharactersByPublisher mocks base method
func (m *MockCharacterServicer) CharactersByPublisher(slugs []comic.PublisherSlug, filterSources bool, limit, offset int) ([]*comic.Character, error) {
	ret := m.ctrl.Call(m, "CharactersByPublisher", slugs, filterSources, limit, offset)
	ret0, _ := ret[0].([]*comic.Character)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CharactersByPublisher indicates an expected call of CharactersByPublisher
func (mr *MockCharacterServicerMockRecorder) CharactersByPublisher(slugs, filterSources, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CharactersByPublisher", reflect.TypeOf((*MockCharacterServicer)(nil).CharactersByPublisher), slugs, filterSources, limit, offset)
}

// CreateSource mocks base method
func (m *MockCharacterServicer) CreateSource(source *comic.CharacterSource) error {
	ret := m.ctrl.Call(m, "CreateSource", source)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSource indicates an expected call of CreateSource
func (mr *MockCharacterServicerMockRecorder) CreateSource(source interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSource", reflect.TypeOf((*MockCharacterServicer)(nil).CreateSource), source)
}

// UpdateSource mocks base method
func (m *MockCharacterServicer) UpdateSource(source *comic.CharacterSource) error {
	ret := m.ctrl.Call(m, "UpdateSource", source)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSource indicates an expected call of UpdateSource
func (mr *MockCharacterServicerMockRecorder) UpdateSource(source interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSource", reflect.TypeOf((*MockCharacterServicer)(nil).UpdateSource), source)
}

// MustNormalizeSources mocks base method
func (m *MockCharacterServicer) MustNormalizeSources(arg0 *comic.Character) {
	m.ctrl.Call(m, "MustNormalizeSources", arg0)
}

// MustNormalizeSources indicates an expected call of MustNormalizeSources
func (mr *MockCharacterServicerMockRecorder) MustNormalizeSources(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MustNormalizeSources", reflect.TypeOf((*MockCharacterServicer)(nil).MustNormalizeSources), arg0)
}

// Source mocks base method
func (m *MockCharacterServicer) Source(id comic.CharacterID, vendorURL string) (*comic.CharacterSource, error) {
	ret := m.ctrl.Call(m, "Source", id, vendorURL)
	ret0, _ := ret[0].(*comic.CharacterSource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Source indicates an expected call of Source
func (mr *MockCharacterServicerMockRecorder) Source(id, vendorURL interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Source", reflect.TypeOf((*MockCharacterServicer)(nil).Source), id, vendorURL)
}

// Sources mocks base method
func (m *MockCharacterServicer) Sources(id comic.CharacterID, vendorType comic.VendorType, isMain *bool) ([]*comic.CharacterSource, error) {
	ret := m.ctrl.Call(m, "Sources", id, vendorType, isMain)
	ret0, _ := ret[0].([]*comic.CharacterSource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Sources indicates an expected call of Sources
func (mr *MockCharacterServicerMockRecorder) Sources(id, vendorType, isMain interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sources", reflect.TypeOf((*MockCharacterServicer)(nil).Sources), id, vendorType, isMain)
}

// TotalSources mocks base method
func (m *MockCharacterServicer) TotalSources(id comic.CharacterID) (int64, error) {
	ret := m.ctrl.Call(m, "TotalSources", id)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TotalSources indicates an expected call of TotalSources
func (mr *MockCharacterServicerMockRecorder) TotalSources(id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalSources", reflect.TypeOf((*MockCharacterServicer)(nil).TotalSources), id)
}

// CreateIssueP mocks base method
func (m *MockCharacterServicer) CreateIssueP(characterID comic.CharacterID, issueID comic.IssueID, appearanceType comic.AppearanceType, importance *comic.Importance) (*comic.CharacterIssue, error) {
	ret := m.ctrl.Call(m, "CreateIssueP", characterID, issueID, appearanceType, importance)
	ret0, _ := ret[0].(*comic.CharacterIssue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateIssueP indicates an expected call of CreateIssueP
func (mr *MockCharacterServicerMockRecorder) CreateIssueP(characterID, issueID, appearanceType, importance interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssueP", reflect.TypeOf((*MockCharacterServicer)(nil).CreateIssueP), characterID, issueID, appearanceType, importance)
}

// CreateIssue mocks base method
func (m *MockCharacterServicer) CreateIssue(issue *comic.CharacterIssue) error {
	ret := m.ctrl.Call(m, "CreateIssue", issue)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateIssue indicates an expected call of CreateIssue
func (mr *MockCharacterServicerMockRecorder) CreateIssue(issue interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssue", reflect.TypeOf((*MockCharacterServicer)(nil).CreateIssue), issue)
}

// CreateIssues mocks base method
func (m *MockCharacterServicer) CreateIssues(issues []*comic.CharacterIssue) error {
	ret := m.ctrl.Call(m, "CreateIssues", issues)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateIssues indicates an expected call of CreateIssues
func (mr *MockCharacterServicerMockRecorder) CreateIssues(issues interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssues", reflect.TypeOf((*MockCharacterServicer)(nil).CreateIssues), issues)
}

// Issue mocks base method
func (m *MockCharacterServicer) Issue(characterID comic.CharacterID, issueID comic.IssueID) (*comic.CharacterIssue, error) {
	ret := m.ctrl.Call(m, "Issue", characterID, issueID)
	ret0, _ := ret[0].(*comic.CharacterIssue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Issue indicates an expected call of Issue
func (mr *MockCharacterServicerMockRecorder) Issue(characterID, issueID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Issue", reflect.TypeOf((*MockCharacterServicer)(nil).Issue), characterID, issueID)
}

// RemoveIssues mocks base method
func (m *MockCharacterServicer) RemoveIssues(ids ...comic.CharacterID) (int, error) {
	varargs := []interface{}{}
	for _, a := range ids {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RemoveIssues", varargs...)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveIssues indicates an expected call of RemoveIssues
func (mr *MockCharacterServicerMockRecorder) RemoveIssues(ids ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveIssues", reflect.TypeOf((*MockCharacterServicer)(nil).RemoveIssues), ids...)
}

// CreateSyncLogP mocks base method
func (m *MockCharacterServicer) CreateSyncLogP(id comic.CharacterID, status comic.CharacterSyncLogStatus, syncType comic.CharacterSyncLogType, syncedAt *time.Time) (*comic.CharacterSyncLog, error) {
	ret := m.ctrl.Call(m, "CreateSyncLogP", id, status, syncType, syncedAt)
	ret0, _ := ret[0].(*comic.CharacterSyncLog)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSyncLogP indicates an expected call of CreateSyncLogP
func (mr *MockCharacterServicerMockRecorder) CreateSyncLogP(id, status, syncType, syncedAt interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSyncLogP", reflect.TypeOf((*MockCharacterServicer)(nil).CreateSyncLogP), id, status, syncType, syncedAt)
}

// CreateSyncLog mocks base method
func (m *MockCharacterServicer) CreateSyncLog(syncLog *comic.CharacterSyncLog) error {
	ret := m.ctrl.Call(m, "CreateSyncLog", syncLog)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSyncLog indicates an expected call of CreateSyncLog
func (mr *MockCharacterServicerMockRecorder) CreateSyncLog(syncLog interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSyncLog", reflect.TypeOf((*MockCharacterServicer)(nil).CreateSyncLog), syncLog)
}

// UpdateSyncLog mocks base method
func (m *MockCharacterServicer) UpdateSyncLog(syncLog *comic.CharacterSyncLog) error {
	ret := m.ctrl.Call(m, "UpdateSyncLog", syncLog)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSyncLog indicates an expected call of UpdateSyncLog
func (mr *MockCharacterServicerMockRecorder) UpdateSyncLog(syncLog interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSyncLog", reflect.TypeOf((*MockCharacterServicer)(nil).UpdateSyncLog), syncLog)
}

// MockRankedServicer is a mock of RankedServicer interface
type MockRankedServicer struct {
	ctrl     *gomock.Controller
	recorder *MockRankedServicerMockRecorder
}

// MockRankedServicerMockRecorder is the mock recorder for MockRankedServicer
type MockRankedServicerMockRecorder struct {
	mock *MockRankedServicer
}

// NewMockRankedServicer creates a new mock instance
func NewMockRankedServicer(ctrl *gomock.Controller) *MockRankedServicer {
	mock := &MockRankedServicer{ctrl: ctrl}
	mock.recorder = &MockRankedServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRankedServicer) EXPECT() *MockRankedServicerMockRecorder {
	return m.recorder
}

// AllPopular mocks base method
func (m *MockRankedServicer) AllPopular(cr comic.PopularCriteria) ([]*comic.RankedCharacter, error) {
	ret := m.ctrl.Call(m, "AllPopular", cr)
	ret0, _ := ret[0].([]*comic.RankedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllPopular indicates an expected call of AllPopular
func (mr *MockRankedServicerMockRecorder) AllPopular(cr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllPopular", reflect.TypeOf((*MockRankedServicer)(nil).AllPopular), cr)
}

// DCPopular mocks base method
func (m *MockRankedServicer) DCPopular(cr comic.PopularCriteria) ([]*comic.RankedCharacter, error) {
	ret := m.ctrl.Call(m, "DCPopular", cr)
	ret0, _ := ret[0].([]*comic.RankedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DCPopular indicates an expected call of DCPopular
func (mr *MockRankedServicerMockRecorder) DCPopular(cr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DCPopular", reflect.TypeOf((*MockRankedServicer)(nil).DCPopular), cr)
}

// MarvelPopular mocks base method
func (m *MockRankedServicer) MarvelPopular(cr comic.PopularCriteria) ([]*comic.RankedCharacter, error) {
	ret := m.ctrl.Call(m, "MarvelPopular", cr)
	ret0, _ := ret[0].([]*comic.RankedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarvelPopular indicates an expected call of MarvelPopular
func (mr *MockRankedServicerMockRecorder) MarvelPopular(cr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarvelPopular", reflect.TypeOf((*MockRankedServicer)(nil).MarvelPopular), cr)
}

// MarvelTrending mocks base method
func (m *MockRankedServicer) MarvelTrending(limit, offset int) ([]*comic.RankedCharacter, error) {
	ret := m.ctrl.Call(m, "MarvelTrending", limit, offset)
	ret0, _ := ret[0].([]*comic.RankedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarvelTrending indicates an expected call of MarvelTrending
func (mr *MockRankedServicerMockRecorder) MarvelTrending(limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarvelTrending", reflect.TypeOf((*MockRankedServicer)(nil).MarvelTrending), limit, offset)
}

// DCTrending mocks base method
func (m *MockRankedServicer) DCTrending(limit, offset int) ([]*comic.RankedCharacter, error) {
	ret := m.ctrl.Call(m, "DCTrending", limit, offset)
	ret0, _ := ret[0].([]*comic.RankedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DCTrending indicates an expected call of DCTrending
func (mr *MockRankedServicerMockRecorder) DCTrending(limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DCTrending", reflect.TypeOf((*MockRankedServicer)(nil).DCTrending), limit, offset)
}

// MockExpandedServicer is a mock of ExpandedServicer interface
type MockExpandedServicer struct {
	ctrl     *gomock.Controller
	recorder *MockExpandedServicerMockRecorder
}

// MockExpandedServicerMockRecorder is the mock recorder for MockExpandedServicer
type MockExpandedServicerMockRecorder struct {
	mock *MockExpandedServicer
}

// NewMockExpandedServicer creates a new mock instance
func NewMockExpandedServicer(ctrl *gomock.Controller) *MockExpandedServicer {
	mock := &MockExpandedServicer{ctrl: ctrl}
	mock.recorder = &MockExpandedServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExpandedServicer) EXPECT() *MockExpandedServicerMockRecorder {
	return m.recorder
}

// Character mocks base method
func (m *MockExpandedServicer) Character(slug comic.CharacterSlug) (*comic.ExpandedCharacter, error) {
	ret := m.ctrl.Call(m, "Character", slug)
	ret0, _ := ret[0].(*comic.ExpandedCharacter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Character indicates an expected call of Character
func (mr *MockExpandedServicerMockRecorder) Character(slug interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Character", reflect.TypeOf((*MockExpandedServicer)(nil).Character), slug)
}

// MockCharacterThumbServicer is a mock of CharacterThumbServicer interface
type MockCharacterThumbServicer struct {
	ctrl     *gomock.Controller
	recorder *MockCharacterThumbServicerMockRecorder
}

// MockCharacterThumbServicerMockRecorder is the mock recorder for MockCharacterThumbServicer
type MockCharacterThumbServicerMockRecorder struct {
	mock *MockCharacterThumbServicer
}

// NewMockCharacterThumbServicer creates a new mock instance
func NewMockCharacterThumbServicer(ctrl *gomock.Controller) *MockCharacterThumbServicer {
	mock := &MockCharacterThumbServicer{ctrl: ctrl}
	mock.recorder = &MockCharacterThumbServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCharacterThumbServicer) EXPECT() *MockCharacterThumbServicerMockRecorder {
	return m.recorder
}

// Upload mocks base method
func (m *MockCharacterThumbServicer) Upload(c *comic.Character) (*comic.CharacterThumbnails, error) {
	ret := m.ctrl.Call(m, "Upload", c)
	ret0, _ := ret[0].(*comic.CharacterThumbnails)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upload indicates an expected call of Upload
func (mr *MockCharacterThumbServicerMockRecorder) Upload(c interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockCharacterThumbServicer)(nil).Upload), c)
}
