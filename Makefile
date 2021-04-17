test:
	go run main.go

clean_cache:
	rm -rf ./_gap_cache/

build:
	go build -o gap_ripper main.go