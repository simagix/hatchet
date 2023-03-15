/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * http_reader.go
 */

package hatchet

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/simagix/gox"
)

func GetHTTPContent(url, username, password string) (*bufio.Reader, error) {
	var err error
	var req *http.Request

	client := &http.Client{}

	if username != "" && password != "" {
		auth := username + ":" + password
		base64Auth := base64.StdEncoding.EncodeToString([]byte(auth))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Basic "+base64Auth)
	} else {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http failed: %v", resp.Status)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return GetBufioReader(buf)
}

func GetHTTPDigestContent(url, user, secret string) (*bufio.Reader, error) {
	var err error
	var resp *http.Response

	headers := map[string]string{}
	var body []byte
	if resp, err = gox.HTTPDigest("GET", url, user, secret, headers, body); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http failed: %v", resp.Status)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return GetBufioReader(buf)
}
