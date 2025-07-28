package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

const listenAddr = "localhost:7421"
const filesDir = "/var/www/i.supa.sh"
const webURL = "https://i.supa.sh/"

type ErrorResponse struct {
	Error string `json:"error"`
}

type UploadResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Ext          string `json:"ext"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	ContentType  string `json:"content_type"`
}

func generateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}

	return string(id)
}

func resError(w http.ResponseWriter, errorMessage string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: errorMessage,
	})
}

func main() {
	http.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		inFile, fileHeader, err := r.FormFile("file")
		if err != nil {
			resError(w, "parsing file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer inFile.Close()

		buf := make([]byte, 512)
		n, _ := inFile.Read(buf)
		inFile.Seek(0, io.SeekStart)

		contentType := http.DetectContentType(buf[:n])
		if contentType == "image/gif" && fileHeader.Size > 100*1024*1024 {
			resError(w, "GIF files larger than 100MiB are not allowed. Please use a more appropriate format.", http.StatusBadRequest)
			return
		}

		ext := filepath.Ext(fileHeader.Filename)
		fileID := generateID(6)
		outPath := path.Join(filesDir, fileID+ext)

		outFile, err := os.Create(outPath)
		if err != nil {
			resError(w, "creating file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(outFile, inFile)
		outFile.Close()
		if err != nil {
			resError(w, "writing file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UploadResponse{
			ID:           fileID,
			Name:         fileHeader.Filename,
			Ext:          ext,
			ContentType:  contentType,
			URL:          webURL + fileID,
			ThumbnailURL: "https://probe.supa.sh/t/?url=" + url.QueryEscape(webURL+fileID+ext),
		})
	})

	log.Println("Server running on " + listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
