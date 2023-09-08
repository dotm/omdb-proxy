package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	omdbProxyPB "dotm/omdb-proxy/services/omdbproxy"
)

// flag.Parse() needs to be called in main function
var (
	serverAddress = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

var (
	//GetMovieByIDRequest
	getMovieByIDReq_idHappyPath   = &omdbProxyPB.GetMovieByIDRequest{Id: "tt4853102"}
	getMovieByIDReq_idNotFound    = &omdbProxyPB.GetMovieByIDRequest{Id: "tt4853101"}
	getMovieByIDReq_idUnspecified = &omdbProxyPB.GetMovieByIDRequest{Id: ""}
)

var (
	//SearchMoviesRequest
	searchMoviesReq_queryTooShort       = &omdbProxyPB.SearchMoviesRequest{Query: "Ba", Type: "", Page: 0}
	searchMoviesReq_invalidPageRange    = &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "", Page: 0}
	searchMoviesReq_invalidMovieType    = &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "sinetron", Page: 1}
	searchMoviesReq_tooManyResultsError = &omdbProxyPB.SearchMoviesRequest{Query: "The", Type: "", Page: 1}
	searchMoviesReq_happyPath           = &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "", Page: 1}
	searchMoviesReq_happyPathWithType   = &omdbProxyPB.SearchMoviesRequest{Query: "Bat", Type: "series", Page: 1}
)

func main() {
	flag.Parse()

	connection, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server %s: %v", *serverAddress, err)
	}
	defer connection.Close()

	clientStub := omdbProxyPB.NewOMDBServiceClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("calling GetMovieByID")
	getMovieByIDResp, err := clientStub.GetMovieByID(ctx, getMovieByIDReq_idHappyPath)
	if err != nil {
		log.Fatalf("error calling GetMovieByID: %v", err)
	}
	log.Printf("%+v\n", getMovieByIDResp)
	log.Println("")

	// log.Println("calling SearchMovies")
	// searchMoviesResp, err := clientStub.SearchMovies(ctx, searchMoviesReq_happyPath)
	// if err != nil {
	// 	log.Fatalf("error calling SearchMovies: %v", err)
	// }
	// log.Printf("%+v\n", searchMoviesResp)
	// log.Println("")

	log.Println("terminating client")
}
