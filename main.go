package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

const webAddr = "localhost:7421"
const dirPath = "/var/www/i.supa.sh"
const webUrl = "https://i.supa.sh/"

type ErrorResponse struct {
	Error string `json:"error"`
}

type UploadResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Ext          string `json:"ext"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
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
		inFile, header, err := r.FormFile("file")
		if err != nil {
			resError(w, "parsing file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer inFile.Close()

		ext := filepath.Ext(header.Filename)
		fileID := generateID(6)
		outPath := path.Join(dirPath, fileID+ext)

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
			Name:         header.Filename,
			Ext:          ext,
			URL:          webUrl + fileID,
			ThumbnailURL: "https://probe.supa.sh/t/?url=" + url.QueryEscape(webUrl+fileID+ext),
		})
	})

	fmt.Println("Server running on " + webAddr)
	log.Fatal(http.ListenAndServe(webAddr, nil))
}
