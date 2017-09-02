PREFIX := jisui

container: docker/imagemagick/6/Dockerfile
	docker build -t $(PREFIX)/imagemagick:6 docker/imagemagick/6

run: 
	docker run -it --rm -v $$(pwd):/go/src/github.com/zuiurs/jisui $(PREFIX)/imagemagick:6
