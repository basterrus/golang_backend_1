package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileListGet(t *testing.T) {
	req, err := http.NewRequest("GET", "/filelist", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := &FileList{}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestFileListPost(t *testing.T) {
	hts := httptest.NewServer(&FileList{})

	req, err := http.NewRequest("POST", hts.URL+"/filelist", strings.NewReader(`{"ext":"txt"}`))
	if err != nil {
		t.Fatal(err)
	}

	cli := hts.Client()

	resp, err := cli.Do(req)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			resp.StatusCode, http.StatusOK)
	}
}

func TestUploadHandler(t *testing.T) {
	file, _ := os.Open("textfile.txt")
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		HostAddr:  ts.URL,
	}

	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `File with this name already exists, please rename the file`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
