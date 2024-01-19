FROM golang:1.21.3
LABEL authors="liangyongbin"
LABEL name="DuckBox"
LABEL container="DuckBox"
LABEL tag="0.1.1"

ENV GOPATH /golang
WORKDIR $GOPATH/src/DuckBox
COPY . $GOPATH/src/DuckBox
RUN go mod tidy
RUN go build

ENTRYPOINT ["./DuckBox"]
