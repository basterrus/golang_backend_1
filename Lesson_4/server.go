package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type UploadHandler struct {
	UploadDir string
	HostAddr  string
}

type FileList struct {
	Extension string `json:"ext"`
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	filePath := h.UploadDir + "/" + header.Filename

	files, err := ioutil.ReadDir("./upload")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.Name() == header.Filename {
			fmt.Fprintf(w, "File with this name already exists, please rename the file\n")
			return
			//filePath = h.UploadDir + "/" + "copy-" + header.Filename
		}
	}

	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fileLink := h.HostAddr + "/" + header.Filename
	fmt.Fprintf(w, "File %s has been successfully uploaded\n", header.Filename)
	fmt.Fprintln(w, fileLink)
}

func (h *FileList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		files, err := ioutil.ReadDir("./upload")
		if err != nil {
			log.Fatal(err)
		}
		for i, f := range files {
			ext := after(f.Name(), ".")
			printFileInfo(w, i, f.Name(), ext, f.Size())
		}
	case http.MethodPost:
		err := json.NewDecoder(r.Body).Decode(&h)
		if err != nil {
			http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
			return
		}

		files, err := ioutil.ReadDir("./upload")
		if err != nil {
			log.Fatal(err)
		}

		var extList int
		if h.Extension != "" {
			for i, f := range files {
				ext := after(f.Name(), ".")
				if h.Extension == ext {
					printFileInfo(w, i, f.Name(), ext, f.Size())
					extList++
				}
			}
		} else {
			fmt.Fprintln(w, "Please, Enter any extension")
		}
		if extList == 0 {
			fmt.Fprintln(w, "There are no files with this extension")
		}
	}
}

func main() {
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
	}
	fileList := &FileList{}

	http.Handle("/upload", uploadHandler)
	http.Handle("/filelist", fileList)

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		dirToServe := http.Dir(uploadHandler.UploadDir)
		fs := &http.Server{
			Addr:         ":8000",
			Handler:      http.FileServer(dirToServe),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
		fmt.Println("file server started")
		err := fs.ListenAndServe()
		if err != nil {
			log.Fatal("File server ListenAndServe: ", err)
		}
	}()

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Server ListenAndServe: ", err)
	}

}

func after(value string, a string) string {

	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}

func printFileInfo(w http.ResponseWriter, i int, name string, ext string, size int64) {
	fmt.Fprintln(w, "ID:", i)
	fmt.Fprintln(w, "Name:", name)
	fmt.Fprintln(w, "Extension:", ext)
	fmt.Fprintln(w, "Size:", size)
}
