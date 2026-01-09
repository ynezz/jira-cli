package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// GetAttachments fetches attachments for an issue using GET /issue/{key} endpoint.
func (c *Client) GetAttachments(key string) ([]*Attachment, error) {
	path := fmt.Sprintf("/issue/%s?fields=attachment", key)

	res, err := c.GetV2(context.Background(), path, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out struct {
		Fields struct {
			Attachment []*Attachment `json:"attachment"`
		} `json:"fields"`
	}

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Fields.Attachment, nil
}

// AddAttachment uploads a file to an issue using POST /issue/{key}/attachments endpoint.
func (c *Client) AddAttachment(key, filePath string) ([]*Attachment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/issue/%s/attachments", key)

	res, err := c.postMultipart(context.Background(), path, body.Bytes(), writer.FormDataContentType())
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out []*Attachment
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}

// DownloadAttachment downloads an attachment content using GET /attachment/content/{id} endpoint.
func (c *Client) DownloadAttachment(id, destPath string) error {
	path := fmt.Sprintf("/attachment/content/%s", id)

	res, err := c.GetV2(context.Background(), path, nil)
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return formatUnexpectedResponse(res)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, res.Body)
	return err
}

// DeleteAttachment deletes an attachment using DELETE /attachment/{id} endpoint.
func (c *Client) DeleteAttachment(id string) error {
	path := fmt.Sprintf("/attachment/%s", id)

	res, err := c.DeleteV2(context.Background(), path, nil)
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNoContent {
		return formatUnexpectedResponse(res)
	}
	return nil
}

// postMultipart sends POST request with multipart/form-data to v2 version of the jira api.
func (c *Client) postMultipart(ctx context.Context, path string, body []byte, contentType string) (*http.Response, error) {
	endpoint := c.server + baseURLv2 + path

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Atlassian-Token", "no-check")

	// Set default auth type to `basic`.
	if c.authType == nil {
		basic := AuthTypeBasic
		c.authType = &basic
	}

	switch c.authType.String() {
	case string(AuthTypeMTLS):
		if c.token != "" {
			req.Header.Add("Authorization", "Bearer "+c.token)
		}
	case string(AuthTypeBearer):
		req.Header.Add("Authorization", "Bearer "+c.token)
	case string(AuthTypeBasic):
		req.SetBasicAuth(c.login, c.token)
	}

	httpClient := &http.Client{Transport: c.transport}

	return httpClient.Do(req.WithContext(ctx))
}
