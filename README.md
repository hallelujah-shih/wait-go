# wait-go
```
使用docker-compose的时候经常存在服务依赖问题

前人使用了shell的解决方案，但是对于多服务难以操作，我这儿取巧使用了golang版本的wait-go

直接引用的是已经存在的github项目，地址为：https://github.com/adrian-gheorghe/wait-go

若后续需要增强功能，直接上面扩展即可
```

## 使用方法
```
关于 wait-go的使用方法，参见adrian-gheorghe/wait-go

Dockerfile:
    FROM xxx AS builder
    XXX XXX
    XXXX XXXX

    FROM shihan/wait-go AS waiter

    FROM alpine AS runner
    COPY --from=waiter /go/bin/wait-go /go/bin/wait-go
    COPY --from=builder src_path dst_path
    CMD [xxx xxx]

docker-compose.yml:

version: '3'
services:
  web:
    build: .
    ports:
      - "9090:8080"
    links:
      - mongo
      - es
    depends_on:
      - mongo
      - es
    restart: on-failure
    command: ["/go/bin/wait-go", "--wait", "es:9200", "--wait", "mongo:27017", "--command", "/go/bin/web_app -c config.yml"]

  mongo:
    image: mongo
    command: mongod --smallfiles

  es:
    image: es-ik:6.7.2

```
