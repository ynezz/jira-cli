package view

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestValidAttachmentColumns(t *testing.T) {
	t.Parallel()

	columns := ValidAttachmentColumns()
	expected := []string{"ID", "FILENAME", "SIZE", "AUTHOR", "CREATED", "MIMETYPE"}

	assert.Equal(t, expected, columns)
}

func TestFormatSize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			size:     1024,
			expected: "1.0 KB",
		},
		{
			name:     "kilobytes with decimal",
			size:     2560,
			expected: "2.5 KB",
		},
		{
			name:     "megabytes",
			size:     1048576,
			expected: "1.0 MB",
		},
		{
			name:     "megabytes with decimal",
			size:     5242880,
			expected: "5.0 MB",
		},
		{
			name:     "gigabytes",
			size:     1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "gigabytes with decimal",
			size:     2684354560,
			expected: "2.5 GB",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, formatSize(tc.size))
		})
	}
}

func TestAttachmentListTableData(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
		{
			ID:       "10001",
			Filename: "image.png",
			Author:   jira.User{DisplayName: "Another User"},
			Created:  "2024-01-16T11:00:00.000+0000",
			Size:     2097152,
			MimeType: "image/png",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
		},
	}

	data := list.tableData()

	// Should have header + 2 rows
	assert.Len(t, data, 3)

	// Check headers
	assert.Equal(t, []string{"ID", "FILENAME", "SIZE", "AUTHOR", "CREATED", "MIMETYPE"}, data[0])

	// Check first row
	assert.Equal(t, "10000", data[1][0])
	assert.Equal(t, "test.txt", data[1][1])
	assert.Equal(t, "1.0 KB", data[1][2])
	assert.Equal(t, "Test User", data[1][3])

	// Check second row
	assert.Equal(t, "10001", data[2][0])
	assert.Equal(t, "image.png", data[2][1])
	assert.Equal(t, "2.0 MB", data[2][2])
	assert.Equal(t, "Another User", data[2][3])
}

func TestAttachmentListTableDataWithCustomColumns(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
			Columns:   []string{"id", "filename", "size"},
		},
	}

	data := list.tableData()

	// Should have header + 1 row
	assert.Len(t, data, 2)

	// Check headers (only selected columns)
	assert.Equal(t, []string{"ID", "FILENAME", "SIZE"}, data[0])

	// Check row (only selected columns)
	assert.Len(t, data[1], 3)
	assert.Equal(t, "10000", data[1][0])
	assert.Equal(t, "test.txt", data[1][1])
	assert.Equal(t, "1.0 KB", data[1][2])
}

func TestAttachmentListTableDataPlainNoHeaders(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     true,
			NoHeaders: true,
		},
	}

	data := list.tableData()

	// Should have only 1 row (no header)
	assert.Len(t, data, 1)
	assert.Equal(t, "10000", data[0][0])
}

func TestAttachmentListEmptyData(t *testing.T) {
	t.Parallel()

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   []*jira.Attachment{},
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
		},
	}

	data := list.tableData()

	// Should have only header row
	assert.Len(t, data, 1)
	assert.Equal(t, []string{"ID", "FILENAME", "SIZE", "AUTHOR", "CREATED", "MIMETYPE"}, data[0])
}

func TestAttachmentListTableDataWithAllInvalidColumns(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
			Columns:   []string{"invalid", "bad", "wrong"},
		},
	}

	data := list.tableData()

	// Should fallback to all valid columns
	assert.Len(t, data, 2)
	assert.Equal(t, ValidAttachmentColumns(), data[0])
	assert.Len(t, data[1], 6)
}

func TestAttachmentListTableDataWithMixedValidInvalidColumns(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
			Columns:   []string{"id", "invalid", "filename"},
		},
	}

	data := list.tableData()

	// Should only include valid columns (id, filename), invalid is silently ignored
	assert.Len(t, data, 2)
	assert.Equal(t, []string{"ID", "FILENAME"}, data[0])
	assert.Len(t, data[1], 2)
	assert.Equal(t, "10000", data[1][0])
	assert.Equal(t, "test.txt", data[1][1])
}

func TestAttachmentListTableDataWithCaseInsensitiveColumns(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     false,
			NoHeaders: false,
			Columns:   []string{"ID", "FileName", "SIZE"},
		},
	}

	data := list.tableData()

	// Column names should be normalized to uppercase
	assert.Len(t, data, 2)
	assert.Equal(t, []string{"ID", "FILENAME", "SIZE"}, data[0])
}

func TestAttachmentListPlainNoHeadersWithAllInvalidColumns(t *testing.T) {
	t.Parallel()

	attachments := []*jira.Attachment{
		{
			ID:       "10000",
			Filename: "test.txt",
			Author:   jira.User{DisplayName: "Test User"},
			Created:  "2024-01-15T10:30:00.000+0000",
			Size:     1024,
			MimeType: "text/plain",
		},
	}

	list := &AttachmentList{
		Server: "https://example.atlassian.net",
		Data:   attachments,
		Display: DisplayFormat{
			Plain:     true,
			NoHeaders: true,
			Columns:   []string{"invalid"},
		},
	}

	data := list.tableData()

	// Should have 1 data row (no header), using all valid columns as fallback
	assert.Len(t, data, 1)
	assert.Len(t, data[0], 6)
}
