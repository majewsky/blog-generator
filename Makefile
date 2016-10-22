blog-generator: *.go assets.go
	go build -o $@ .
assets.go: assets/*
	go run assets/main.go
