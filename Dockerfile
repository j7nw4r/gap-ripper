FROM golang:1.16-alpine AS builder

ARG GOOS=linux
ARG GOARCH=amd64
ARG name=gap-ripper

ENV GAP-RIPPER=${name}
WORKDIR /go/src/github.com/johnathan-walker/${name}
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ${name} ./cmd/gap-ripper/main.go

FROM alpine:latest 
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin/
COPY --from=builder /go/src/github.com/johnathan-walker/${name}/${name} .
VOLUME [ "/usr/local/bin/gap-ripper/_gap_cache" ]

RUN ["chmod", "a+wx", "./gap-ripper/gap-ripper"]

ENTRYPOINT [ "./gap-ripper/gap-ripper", "632140012" ]