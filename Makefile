default: darwin

darwin:
	env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./build/$@/gofm
linux:
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./build/$@/gofm
linux32:
	env GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o ./build/$@/gofm