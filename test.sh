go test -v ./...

go test -coverprofile=ota/.coverprofile github.com/mannkind/mysb/ota
go test -coverprofile=transport/.coverprofile github.com/mannkind/mysb/transport
gover . .coverprofile
go tool cover -html=.coverprofile
find . -name ".coverprofile" -exec rm {} \;
