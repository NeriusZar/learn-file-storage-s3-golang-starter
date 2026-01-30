package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		log.Print(err)
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", err)
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Failed to read the file", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}

	if video.UserID != userID {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Only owner can upload thumbnail", err)
		return
	}

	encodedFile := base64.StdEncoding.EncodeToString(fileBytes)
	dataUrl := fmt.Sprintf("data:%s;base64,%s", mediaType, encodedFile)
	video.ThumbnailURL = &dataUrl
	if err := cfg.db.UpdateVideo(video); err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update the video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
