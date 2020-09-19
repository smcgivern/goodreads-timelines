goodreads-timelines: *.go
	@go build

public/ext/style.css: template/css/style.scss
	@sassc -M template/css/style.scss public/ext/style.css

build: goodreads-timelines public/ext/style.css

preview: build
	@./goodreads-timelines

-include *.mk
