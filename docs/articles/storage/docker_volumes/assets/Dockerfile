FROM ubuntu:oracular AS builder
RUN apt-get update && apt-get install golang-go ca-certificates -y
COPY . /build
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -tags netgo,osusergo -o myplugin
CMD ["/build/myplugin"]

FROM ubuntu:oracular
RUN mkdir -p /run/docker/plugins /var/lib/myplugin/volumes
COPY --from=builder /build/myplugin /myplugin
ENTRYPOINT ["/myplugin"]

