mkdir -p .coverage/$1
go test ./$1 -coverprofile=.coverage/$1/coverage.txt -covermode=atomic
go tool cover -html=.coverage/$1/coverage.txt
