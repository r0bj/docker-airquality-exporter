FROM golang:1.15.6 as builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY airquality-exporter.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a --ldflags '-w -extldflags "-static"' -tags netgo -installsuffix netgo -o airquality-exporter .


FROM scratch

COPY --from=builder /workspace/airquality-exporter /

ENTRYPOINT ["/airquality-exporter"]
