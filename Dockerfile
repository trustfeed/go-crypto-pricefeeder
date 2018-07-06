FROM golang:1.9.4 as build
WORKDIR /go/src/github.com/trustfeed/go-crypto-pricefeeder
COPY . .
RUN mv -vn config_example.json config.json \
 && go get -v -d \
 && GOARCH=386 GOOS=linux CGO_ENABLED=0 go install -v \
 && mv /go/bin/linux_386 /go/bin/go-crypto-pricefeeder

FROM alpine:latest
COPY --from=build /go/bin/go-crypto-pricefeeder /app/
COPY --from=build /go/src/github.com/trustfeed/go-crypto-pricefeeder/config.json /app/
EXPOSE 9050
CMD ["/app/go-crypto-pricefeeder"]
