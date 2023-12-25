FROM golang:1.18.10

WORKDIR /app
COPY . .

RUN cd collab-net-v2 && \
    go env -w GO111MODULE=on && \
    go env && \
    go env -w  GOPROXY=https://goproxy.cn,direct && \
    go mod tidy && \
    sh script/build.sh && mv release/server /app/main && cd .. && rm -rf collab-net-v2 && rm -rf /go && ls -alh /app

EXPOSE 80 1080 2080 8081
CMD ["./main"]
