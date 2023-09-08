package main

import (
	"context"
	omdbProxyPB "dotm/omdb-proxy/services/omdbproxy"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMovieByIDHappyPath(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	getMovieByIDReq_idHappyPath := &omdbProxyPB.GetMovieByIDRequest{Id: "tt4853102"}
	//when
	_, err := server.GetMovieByID(ctx, getMovieByIDReq_idHappyPath)
	//then
	assert.NoError(t, err)
}

func TestGetMovieByIDNotFound(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	getMovieByIDReq_idNotFound := &omdbProxyPB.GetMovieByIDRequest{Id: "tt4853101"}
	//when
	_, err := server.GetMovieByID(ctx, getMovieByIDReq_idNotFound)
	//then
	assert.Error(t, err)
}

func TestGetMovieByIDUnspecified(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	getMovieByIDReq_idUnspecified := &omdbProxyPB.GetMovieByIDRequest{Id: ""}
	//when
	_, err := server.GetMovieByID(ctx, getMovieByIDReq_idUnspecified)
	//then
	assert.Error(t, err)
}

func TestSearchMoviesHappyPath(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_happyPath := &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "", Page: 1}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_happyPath)
	//then
	assert.NoError(t, err)
}

func TestSearchMoviesHappyPathWithMovieType(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_happyPathWithType := &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "series", Page: 1}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_happyPathWithType)
	//then
	assert.NoError(t, err)
}

func TestSearchMoviesQueryTooShort(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_queryTooShort := &omdbProxyPB.SearchMoviesRequest{Query: "Ba", Type: "", Page: 0}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_queryTooShort)
	//then
	assert.Error(t, err)
}

func TestSearchMoviesInvalidPageRange(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_invalidPageRange := &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "", Page: 0}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_invalidPageRange)
	//then
	assert.Error(t, err)
}

func TestSearchMoviesInvalidMovieType(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_invalidMovieType := &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "sinetron", Page: 1}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_invalidMovieType)
	//then
	assert.Error(t, err)
}

func TestSearchMoviesTooManyResultsError(t *testing.T) {
	loadEnv("../../.env")
	server := &omdbProxyServer{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//given
	searchMoviesReq_tooManyResultsError := &omdbProxyPB.SearchMoviesRequest{Query: "The", Type: "", Page: 1}
	//when
	_, err := server.SearchMovies(ctx, searchMoviesReq_tooManyResultsError)
	//then
	assert.Error(t, err)
}
