package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type kwWriter struct {
	Path string
	buf  bytes.Buffer
}

var (
	errParseContent = errors.New("parse content error")
	errParseResp    = errors.New("parse response error")
)

func (k *kwWriter) Write(b []byte) (nr int, err error) {
	return k.buf.Write(b)
}

func (k *kwWriter) Close() error {
	dec := json.NewDecoder(&k.buf)

	var resp struct {
		Status interface{}
		Data   map[string]interface{}
	}
	if err := dec.Decode(&resp); err != nil {
		return err
	}

	content, ok := resp.Data["content"].(string)
	if !ok {
		return errParseContent
	}

	title, ok := resp.Data["title"].(string)
	if !ok {
		return errParseContent
	}

	date, ok := resp.Data["releasedate"].(string)
	if !ok {
		return errParseContent
	}

	title, err := genFileName(title, date)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(k.Path, 0755); err != nil {
		return err
	}

	// Overwrite the old
	file, err := os.Create(filepath.Join(k.Path, title))
	if err != nil {
		return err
	}

	defer file.Close()

	file.WriteString("<meta charset=\"utf-8\">")
	file.WriteString((genContent(content)))

	return nil
}

func genFileName(title, date string) (string, error) {
	title = strings.ReplaceAll(title, " ", "")
	if title == "" {
		return "", errParseResp
	}

	reg, err := regexp.Compile("[0-9][0-9]*-[0-9][0-9]*-[0-9][0-9]*")
	if err != nil {
		return "", err
	}
	date = reg.FindString(date)

	if date == "" {
		return "", errParseResp
	}

	return date + title + ".html", nil
}

func genContent(content string) string {
	// FIXME: data-src not in <img/> or not png
	reg, err := regexp.Compile("(?U)data-src=\"http[^>]*png\"")
	if err != nil {
		return ""
	}

	return reg.ReplaceAllStringFunc(content, func(repl string) string {
		url := strings.TrimPrefix(repl, "data-src=\"")
		url = strings.TrimSuffix(url, "\"")

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Sprintf("src=\"%s\"", url)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Sprintf("src=\"%s\"", url)
		}

		enc := base64.StdEncoding.EncodeToString(body)

		return "src=\"data:image/png;base64, " + enc + "\""
	})
}
