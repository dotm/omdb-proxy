package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spf13/viper"

	omdbProxyPB "dotm/omdb-proxy/services/omdbproxy"
)

// flag.Parse() needs to be called in main function
var (
	port = flag.Int("port", 50051, "gRPC server port")
)

type omdbProxyServer struct {
	omdbProxyPB.UnimplementedOMDBServiceServer
}

func (s *omdbProxyServer) SearchMovies(
	ctx context.Context, req *omdbProxyPB.SearchMoviesRequest,
) (res *omdbProxyPB.SearchMoviesResponse, err error) {
	/* validate the request */
	query := req.GetQuery()
	if query == "" {
		return res, status.Errorf(codes.InvalidArgument, "please provide a movie title to query")
	}
	if len(query) < 3 {
		return res, status.Errorf(codes.InvalidArgument, "please provide at least 3 characters of movie title to query")
	}
	movieType := req.GetType()
	if movieType != "" && movieType != "movie" && movieType != "series" && movieType != "episode" {
		return res, status.Errorf(codes.InvalidArgument, "allowed movie types you can search are: movie, series, episode")
	}
	page := req.GetPage()
	if page < 1 || page > 100 {
		return res, status.Errorf(codes.InvalidArgument, "page value can only be 1-100")
	}

	/* call OMDB API */
	queryParams := url.Values{}
	queryParams.Add("apikey", viper.GetString("omdb_api_key"))
	queryParams.Add("s", query)
	if movieType != "" {
		queryParams.Add("type", movieType)
	}
	queryParams.Add("page", strconv.FormatUint(page, 10))
	omdbResp, err := http.Get("https://www.omdbapi.com/?" + queryParams.Encode())
	if err != nil {
		return res, status.Errorf(codes.Internal, err.Error())
	}
	defer omdbResp.Body.Close()
	byteSlice, err := io.ReadAll(omdbResp.Body)
	if err != nil {
		return res, err
	}
	var jsonRes map[string]interface{}
	json.Unmarshal(byteSlice, &jsonRes)

	/*
		parse and validate json response from OMDB without json struct tags
		so that we can explicitly raise error for non-existent fields
		and later on implement email alert to developers when the JSON response changes (breaking the proxy)

		but this is still open to feedback.
		if the conciseness of using struct and struct tags is preferable,
		I can change the code to that approach and check for null values for all fields.
	*/
	var val interface{}
	var ok bool

	val, ok = jsonRes["Error"]
	if ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("error SearchMovies: %v", val))
	}

	val, ok = jsonRes["totalResults"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent totalResults field in OMDB SearchMovies proxy")
	}
	totalResultsStr, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for totalResults field in OMDB SearchMovies proxy: %v", val))
	}
	totalResults, err := strconv.ParseUint(totalResultsStr, 10, 64)
	if err != nil {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("can't parse totalResults field into number in OMDB SearchMovies proxy: %v", val))
	}

	val, ok = jsonRes["Search"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Search field in OMDB SearchMovies proxy")
	}
	searchResults, ok := val.([]interface{})
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Search field in OMDB SearchMovies proxy: %v", val))
	}
	movies := []*omdbProxyPB.MovieResult{}
	for _, searchResult := range searchResults {
		movie, ok := searchResult.(map[string]interface{})
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Search field element in OMDB SearchMovies proxy: %v", val))
		}

		val, ok = movie["imdbID"]
		if !ok {
			return res, status.Errorf(codes.Internal, "non-existent imdbID field in OMDB SearchMovies proxy")
		}
		id, ok := val.(string)
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for imdbID field in OMDB SearchMovies proxy: %v", val))
		}

		val, ok = movie["Title"]
		if !ok {
			return res, status.Errorf(codes.Internal, "non-existent Title field in OMDB SearchMovies proxy")
		}
		title, ok := val.(string)
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Title field in OMDB SearchMovies proxy: %v", val))
		}

		val, ok = movie["Year"]
		if !ok {
			return res, status.Errorf(codes.Internal, "non-existent Year field in OMDB SearchMovies proxy")
		}
		year, ok := val.(string)
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Year field in OMDB SearchMovies proxy: %v", val))
		}

		val, ok = movie["Type"]
		if !ok {
			return res, status.Errorf(codes.Internal, "non-existent Type field in OMDB SearchMovies proxy")
		}
		movieType, ok := val.(string)
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Type field in OMDB SearchMovies proxy: %v", val))
		}

		val, ok = movie["Poster"]
		if !ok {
			return res, status.Errorf(codes.Internal, "non-existent Poster field in OMDB SearchMovies proxy")
		}
		posterUrl, ok := val.(string)
		if !ok {
			return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Poster field in OMDB SearchMovies proxy: %v", val))
		}
		movies = append(movies, &omdbProxyPB.MovieResult{
			Id:        id,
			Title:     title,
			Year:      year,
			Type:      movieType,
			PosterUrl: posterUrl,
		})
	}

	return &omdbProxyPB.SearchMoviesResponse{
		TotalResults: totalResults,
		Movies:       movies,
	}, nil
}

func (s *omdbProxyServer) GetMovieByID(
	ctx context.Context, req *omdbProxyPB.GetMovieByIDRequest,
) (res *omdbProxyPB.GetMovieByIDResponse, err error) {
	/* validate the request */
	id := req.GetId()
	if id == "" {
		return res, status.Errorf(codes.InvalidArgument, "please provide a movie id")
	}

	/* call OMDB API */
	queryParams := url.Values{}
	queryParams.Add("apikey", viper.GetString("omdb_api_key"))
	queryParams.Add("i", id)
	omdbResp, err := http.Get("https://www.omdbapi.com/?" + queryParams.Encode())
	if err != nil {
		return res, status.Errorf(codes.Internal, err.Error())
	}
	defer omdbResp.Body.Close()
	byteSlice, err := io.ReadAll(omdbResp.Body)
	if err != nil {
		return res, err
	}
	var jsonRes map[string]interface{}
	json.Unmarshal(byteSlice, &jsonRes)

	/*
		parse and validate json response from OMDB without json struct tags
		so that we can explicitly raise error for non-existent fields
		and later on implement email alert to developers when the JSON response changes (breaking the proxy)

		but this is still open to feedback.
		if the conciseness of using struct and struct tags is preferable,
		I can change the code to that approach and check for null values for all fields.
	*/
	var val interface{}
	var ok bool

	val, ok = jsonRes["Error"]
	if ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("error GetMovieByID: %v", val))
	}

	val, ok = jsonRes["Title"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Title field in OMDB GetMovieByID proxy")
	}
	title, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Title field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Year"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Year field in OMDB GetMovieByID proxy")
	}
	year, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Year field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Rated"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Rated field in OMDB GetMovieByID proxy")
	}
	rated, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Rated field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Genre"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Genre field in OMDB GetMovieByID proxy")
	}
	genre, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Genre field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Plot"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Plot field in OMDB GetMovieByID proxy")
	}
	plot, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Plot field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Director"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Director field in OMDB GetMovieByID proxy")
	}
	director, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Director field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Actors"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Actors field in OMDB GetMovieByID proxy")
	}
	actorsStr, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Actors field in OMDB GetMovieByID proxy: %v", val))
	}
	actors := strings.Split(actorsStr, ", ")

	val, ok = jsonRes["Language"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Language field in OMDB GetMovieByID proxy")
	}
	language, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Language field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Country"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Country field in OMDB GetMovieByID proxy")
	}
	country, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Country field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Type"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Type field in OMDB GetMovieByID proxy")
	}
	movieType, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Type field in OMDB GetMovieByID proxy: %v", val))
	}

	val, ok = jsonRes["Poster"]
	if !ok {
		return res, status.Errorf(codes.Internal, "non-existent Poster field in OMDB GetMovieByID proxy")
	}
	posterUrl, ok := val.(string)
	if !ok {
		return res, status.Errorf(codes.Internal, fmt.Sprintf("unexpected value for Poster field in OMDB GetMovieByID proxy: %v", val))
	}

	return &omdbProxyPB.GetMovieByIDResponse{
		Id:        id,
		Title:     title,
		Year:      year,
		Rated:     rated,
		Genre:     genre,
		Plot:      plot,
		Director:  director,
		Actors:    actors,
		Language:  language,
		Country:   country,
		Type:      movieType,
		PosterUrl: posterUrl,
	}, nil
}

func loadEnv() {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error reading config file: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	loadEnv()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	s := grpc.NewServer()
	omdbProxyPB.RegisterOMDBServiceServer(s, &omdbProxyServer{})

	fmt.Printf("server listening at port %d\n", *port)
	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
