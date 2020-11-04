export  GO111MODULE=on
# go mod vendor
export GOPROXY=https://goproxy.io


NAME=ncmdump


TODAY=$(date -u +%Y%m%d)

VERSION=$1_$TODAY
LDFLAGS="-X main.VERSION=$VERSION -w -s"

export GOOS=windows
export GOARCH=amd64
go build  -ldflags "$LDFLAGS" -o ./dist/$NAME-$GOOS-$GOARCH.exe cmd/main.go
upx ./dist/$NAME-$GOOS-$GOARCH.exe
zip -q -r ./dist/$NAME-$GOOS-$GOARCH-$VERSION.zip ./dist/$NAME-$GOOS-$GOARCH.exe

export GOOS=windows
export GOARCH=386
go build  -ldflags "$LDFLAGS"  -o ./dist/$NAME-$GOOS-$GOARCH.exe cmd/main.go
upx ./dist/$NAME-$GOOS-$GOARCH.exe
zip -q -r ./dist/$NAME-$GOOS-$GOARCH-$VERSION.zip ./dist/$NAME-$GOOS-$GOARCH.exe


export GOOS=darwin
export GOARCH=amd64
go build  -ldflags "$LDFLAGS"  -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go
upx ./dist/$NAME-$GOOS-$GOARCH
zip -q -r ./dist/$NAME-$GOOS-$GOARCH-$VERSION.zip ./dist/$NAME-$GOOS-$GOARCH


export GOOS=linux
export GOARCH=amd64
go build  -ldflags "$LDFLAGS"  -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go
upx ./dist/$NAME-$GOOS-$GOARCH
zip -q -r ./dist/$NAME-$GOOS-$GOARCH-$VERSION.zip ./dist/$NAME-$GOOS-$GOARCH


export GOOS=linux
export GOARCH=arm64
go build  -ldflags "$LDFLAGS"  -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go
upx ./dist/$NAME-$GOOS-$GOARCH
zip -q -r ./dist/$NAME-$GOOS-$GOARCH-$VERSION.zip ./dist/$NAME-$GOOS-$GOARCH

