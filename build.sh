export  GO111MODULE=on
# go mod vendor
export GOPROXY=https://goproxy.io

export GOOS=windows
export GOARCH=amd64
export NAME=ncmdump

go build  -ldflags '-w -s' -o ./dist/$NAME-$GOOS-$GOARCH.exe cmd/main.go

export GOARCH=386
go build  -ldflags '-w -s' -o ./dist/$NAME-$GOOS-$GOARCH.exe cmd/main.go



export GOOS=darwin
export GOARCH=amd64
go build  -ldflags '-w -s' -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go

export GOARCH=386
go build  -ldflags '-w -s' -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go


export GOOS=linux
export GOARCH=amd64
go build  -ldflags '-w -s' -o ./dist/$NAME-$GOOS-$GOARCH cmd/main.go


upx dist/ncmdump-*
