PREFIX := jisui

ifeq ($(OS),Windows_NT)
	CURDIR := $(shell cmd /c "cd" | tr '\\' '/')
else
	CURDIR := $(shell pwd)
endif

container: docker/imagemagick/6/Dockerfile
	docker build -t $(PREFIX)/imagemagick:6 docker/imagemagick/6

run: 
	docker run -it --rm -v $(CURDIR):/go/src/github.com/zuiurs/jisui $(PREFIX)/imagemagick:6
