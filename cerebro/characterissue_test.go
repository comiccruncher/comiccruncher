package cerebro_test

import (
	"github.com/aimeelaplant/comiccruncher/cerebro"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/externalissuesource"
	"github.com/aimeelaplant/externalissuesource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharacterCBParser_Parse(t *testing.T) {
	sourceCtrl := gomock.NewController(t)
	defer sourceCtrl.Finish()
	sourceMock := mock_externalissuesource.NewMockExternalSource(sourceCtrl)
	pages := []*externalissuesource.CharacterPage{
		{
			IssueLinks: []string{
				"test=123",
				"test=1234",
				"test=345",
				"test=999",
			},
		},
		{
			IssueLinks: []string{
				"test=345",
				"test=444",
				"test=1000",
				"test=1884",
			},
		},
	}
	sources := []*comic.CharacterSource{
		{
			IsMain:    true,
			VendorURL: "test",
		},
		{
			IsMain:    false,
			VendorURL: "test2",
		},
	}
	sourceMock.EXPECT().CharacterPage(gomock.Any()).Times(1).Return(pages[0], nil)
	sourceMock.EXPECT().CharacterPage(gomock.Any()).Return(pages[1], nil)
	parser := cerebro.NewCharacterCBParser(sourceMock)
	vi, err := parser.Parse(sources)
	assert.Nil(t, err)
	assert.True(t, vi.MainSources[cerebro.ExternalVendorID("123")])
	assert.True(t, vi.MainSources[cerebro.ExternalVendorID("1234")])
	assert.True(t, vi.MainSources[cerebro.ExternalVendorID("345")])
	assert.True(t, vi.MainSources[cerebro.ExternalVendorID("999")])
	assert.True(t, vi.AltSources[cerebro.ExternalVendorID("345")])
	assert.True(t, vi.AltSources[cerebro.ExternalVendorID("444")])
	assert.True(t, vi.AltSources[cerebro.ExternalVendorID("1000")])
	assert.True(t, vi.AltSources[cerebro.ExternalVendorID("1884")])
}

func TestCharacterCBParser_Parse_NoSources(t *testing.T) {
	sourceCtrl := gomock.NewController(t)
	defer sourceCtrl.Finish()
	sourceMock := mock_externalissuesource.NewMockExternalSource(sourceCtrl)
	sources := []*comic.CharacterSource{}
	parser := cerebro.NewCharacterCBParser(sourceMock)
	_, err := parser.Parse(sources)
	assert.Error(t, err)
}
