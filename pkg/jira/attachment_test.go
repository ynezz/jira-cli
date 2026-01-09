package jira

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAttachments(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1", r.URL.Path)
		assert.Equal(t, "fields=attachment", r.URL.RawQuery)

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			resp, err := os.ReadFile("./testdata/attachment.json")
			assert.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write(resp)
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.GetAttachments("TEST-1")
	assert.NoError(t, err)
	assert.Len(t, actual, 2)

	assert.Equal(t, "10000", actual[0].ID)
	assert.Equal(t, "test.txt", actual[0].Filename)
	assert.Equal(t, "Test User", actual[0].Author.DisplayName)
	assert.Equal(t, int64(1024), actual[0].Size)
	assert.Equal(t, "text/plain", actual[0].MimeType)

	assert.Equal(t, "10001", actual[1].ID)
	assert.Equal(t, "image.png", actual[1].Filename)
	assert.Equal(t, int64(2048), actual[1].Size)
	assert.Equal(t, "image/png", actual[1].MimeType)

	unexpectedStatusCode = true

	_, err = client.GetAttachments("TEST-1")
	assert.Error(t, err)
}

func TestAddAttachment(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1/attachments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "no-check", r.Header.Get("X-Atlassian-Token"))
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20)
		assert.NoError(t, err)

		// Verify file was sent
		file, fileHeader, err := r.FormFile("file")
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Equal(t, "testfile.txt", fileHeader.Filename)
		_ = file.Close()

		resp, err := os.ReadFile("./testdata/attachment-added.json")
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(resp)
	}))
	defer server.Close()

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "testfile.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0o644)
	assert.NoError(t, err)

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.AddAttachment("TEST-1", testFilePath)
	assert.NoError(t, err)
	assert.Len(t, actual, 1)
	assert.Equal(t, "10002", actual[0].ID)
	assert.Equal(t, "uploaded.txt", actual[0].Filename)

	unexpectedStatusCode = true

	_, err = client.AddAttachment("TEST-1", testFilePath)
	assert.Error(t, err)
}

func TestAddAttachmentFileNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called when file doesn't exist")
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	_, err := client.AddAttachment("TEST-1", "/nonexistent/file.txt")
	assert.Error(t, err)
}

func TestDownloadAttachment(t *testing.T) {
	var unexpectedStatusCode bool
	testContent := []byte("downloaded file content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/attachment/content/10000", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(404)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(200)
			_, _ = w.Write(testContent)
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "downloaded.txt")

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.DownloadAttachment("10000", destPath)
	assert.NoError(t, err)

	// Verify file was written
	content, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.Equal(t, testContent, content)

	unexpectedStatusCode = true

	err = client.DownloadAttachment("10000", destPath)
	assert.Error(t, err)
}

func TestDeleteAttachment(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/attachment/10000", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		if unexpectedStatusCode {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(204)
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.DeleteAttachment("10000")
	assert.NoError(t, err)

	unexpectedStatusCode = true

	err = client.DeleteAttachment("10000")
	assert.Error(t, err)
}
