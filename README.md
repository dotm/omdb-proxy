## Setup

- git init
- create .proto file in services and compile it
  - instruction to compile is available as a comment at the top of respective proto files.
- go mod init dotm/omdb-proxy
- go mod tidy
- create server and client for the service

## Testing Server and Client

- go run .\server\omdbproxy\main.go
- go run .\client\omdbproxy\main.go
  - you can choose between the provided requests as test case

## TODO

- move apiKey in server/omdbproxy/main.go to .env file
  (DONE; probably need to be added to Additional Considerations)
- trim down server/omdbproxy/main.go by separating the method implementations into multiple files
- complete unit test
  - might be better to prioritize end-to-end test (?)
    - since the proxy doesn't contain complicated business logic and is mostly just CRUD
- considerations for caching mechanism:
  - cache GetMovieByID response with id to reduce call for movies that are currently searched a lot
    - invalidate cache everyday
      - probably OK since the movie data rarely changes
        and getting the latest data is assumed to not be critical to the client
  - we can also cache SearchMovies
    - but to avoid large and unused cache, we probably should only cache page 1 since those are the most accessed page.
    - invalidate cache everyday
      - probably OK since the movie data rarely changes
        and getting the latest data is assumed to not be critical to the client
- send email when json response from OMDB changes (causing the proxy to break)

## Assumptions

- the client does not require the response to be in JSON format
- minimum query length for SearchMovies is 3 characters

## Question

- should genre be list of string?