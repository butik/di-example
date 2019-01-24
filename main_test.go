package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ SearchService = &mockSearchService{}
type mockSearchService struct {
	mock.Mock
}

func (m *mockSearchService) SearchProductIDs() []int {
	args := m.Called()

	return args.Get(0).([]int)
}

var _ ProductDatastore = &mockDatastore{}
type mockDatastore struct {
	mock.Mock
}

func (m *mockDatastore) Find(id int) Product {
	args := m.Called(id)

	return args.Get(0).(Product)
}

func (m mockDatastore) Save(product Product) {
	panic("implement me")
}

func TestExample(t *testing.T) {
	datastore := &mockDatastore{}
	searchService := &mockSearchService{}

	searchService.On("SearchProductIDs").Return([]int{1})

	expected := Product{
		ID: 1,
		Name: "Test",
	}
	datastore.On("Find", 1).Return(expected)

	feedService := NewFeedService(datastore, searchService)

	products := feedService.GenerateProductFeed()
	require.Len(t, products, 1)
	assert.Equal(t, expected, products[0])
}
