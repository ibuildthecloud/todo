FROM golang:1.19 as base
WORKDIR /src
RUN go install github.com/mitranim/gow@latest
COPY . .

FROM base as dev
RUN go mod vendor && \
    go build -o app .
CMD ["gow", "run", "."]

FROM base as default
RUN --mount=type=cache,target=/go/pkg --mount=type=cache,target=/root/.cache/go-build go build -o app .
ENTRYPOINT ["./app"]
