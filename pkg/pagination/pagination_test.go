package pagination

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testItem struct {
	id string
}

func (t testItem) ID() string { return t.id }

func TestPageQuery_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    PageQuery
		expected PageQuery
	}{
		{"defaults", PageQuery{}, PageQuery{Limit: DefaultLimit}},
		{"custom limit", PageQuery{Limit: 50}, PageQuery{Limit: 50}},
		{"cap at max", PageQuery{Limit: 200}, PageQuery{Limit: MaxLimit}},
		{"negative limit", PageQuery{Limit: -1}, PageQuery{Limit: DefaultLimit}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			require.Equal(t, tt.expected.Limit, tt.input.Limit)
		})
	}
}

func TestNewPageResult_HasMore(t *testing.T) {
	items := []testItem{
		{id: "019d65d6-de90-7200-b1cf-4f8745597e01"},
		{id: "019d65d6-de90-7200-b1cf-4f8745597e02"},
		{id: "019d65d6-de90-7200-b1cf-4f8745597e03"},
	}

	result := NewPageResult(items, 2)
	require.True(t, result.HasMore)
	require.Len(t, result.Items, 2)
	require.Equal(t, "019d65d6-de90-7200-b1cf-4f8745597e02", result.NextCursor)
	require.Equal(t, 2, result.Limit)
}

func TestNewPageResult_NoMore(t *testing.T) {
	items := []testItem{
		{id: "019d65d6-de90-7200-b1cf-4f8745597e01"},
		{id: "019d65d6-de90-7200-b1cf-4f8745597e02"},
	}

	result := NewPageResult(items, 5)
	require.False(t, result.HasMore)
	require.Len(t, result.Items, 2)
	require.Equal(t, "019d65d6-de90-7200-b1cf-4f8745597e02", result.NextCursor)
}

func TestNewPageResult_Empty(t *testing.T) {
	items := []testItem{}

	result := NewPageResult(items, 20)
	require.False(t, result.HasMore)
	require.Empty(t, result.Items)
	require.Empty(t, result.NextCursor)
}

func TestParseCursor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", false},
		{"valid uuid v7", "019d65d6-de90-7200-b1cf-4f8745597e0a", false},
		{"too short", "abc", true},
		{"no dashes", "019d65d6de907200b1cf4f8745597e0a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCursor(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewPageResultWithCursor(t *testing.T) {
	items := []testItem{{id: "a"}, {id: "b"}}

	result := NewPageResultWithCursor(items, "cursor-123", true, 10)
	require.True(t, result.HasMore)
	require.Equal(t, "cursor-123", result.NextCursor)
	require.Len(t, result.Items, 2)
	require.Equal(t, 10, result.Limit)
}
