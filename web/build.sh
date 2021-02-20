
rm web_linux
echo "build web_linux ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o web_linux web.go

docker build -t braidgo/web .
