build:
	@go build -o dist

run: build
	@./dist	
