FROM golang:1.18 AS build
WORKDIR /app
COPY . .
RUN env CGO_ENABLED=0 GOBIN=/build go install ./cmd/...

FROM scratch
COPY --from=build /build/* /
ENTRYPOINT ["/ndn-webtrans-gateway"]
