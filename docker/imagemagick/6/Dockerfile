FROM fedora:26
LABEL maintainer="zuiurs"

RUN dnf install -y ImageMagick ImageMagick-devel

RUN dnf install -y go git
RUN mkdir /go

ENV GOPATH /go
ENV PATH $PATH:$GOPATH/bin

#### Workaround of failing compile ####
# https://github.com/gographics/imagick/issues/155
ENV CGO_CFLAGS_ALLOW -fopenmp
#######################################

RUN go get gopkg.in/gographics/imagick.v2/imagick

ARG repodir="github.com/zuiurs/jisui"

RUN mkdir -p /go/src/${repodir}
VOLUME /go/src/${repodir}

WORKDIR /go/src/${repodir}
CMD ["/bin/bash", "-c", "go install && bash"]]

