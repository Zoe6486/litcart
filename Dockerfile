# FROM golang:alpine AS builder

# # 为我们的镜像设置必要的环境变量
# ENV GO111MODULE=on \
#     CGO_ENABLED=0 \
#     GOOS=linux \
#     GOARCH=amd64

# # 移动到工作目录：/build
# WORKDIR /build

# # 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
# COPY go.mod .
# COPY go.sum .
# RUN go mod download

# # 将代码复制到容器中
# COPY . .

# # 将我们的代码编译成二进制可执行文件 litcart_app
# RUN go build -o litcart_app .

# ###################
# # 接下来创建一个小镜像
# ###################
# # FROM debian:stretch-slim太老旧了报错
# FROM debian:bookworm-slim


# COPY ./wait-for.sh /
# COPY ./templates /templates
# COPY ./static /static
# COPY ./conf /conf

# # 从builder镜像中把/dist/app 拷贝到当前目录
# COPY --from=builder /build/litcart_app /

# # RUN set -eux; \
# # 	apt-get update; \
# # 	apt-get install -y \
# # 		--no-install-recommends \
# # 		netcat; \在 debian:bookworm-slim 里，netcat 是虚拟包，没有直接可安装的候选，所以 build 会失败。解决办法就是安装具体提供者，比如 netcat-openbsd
# #         chmod 755 wait-for.sh
# RUN set -eux; \
#     apt-get update; \
#     apt-get install -y --no-install-recommends netcat-openbsd; \
#     rm -rf /var/lib/apt/lists/*; \
#     chmod 755 /wait-for.sh
# # 声明服务端口
# EXPOSE 8084

# # 需要运行的命令
# #ENTRYPOINT ["/litcart_app", "conf/config.yaml"]


FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o litcart_app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/litcart_app .
COPY conf/ conf/

EXPOSE 8084
# CMD ["./litcart_app", "./conf/config.yaml"]
CMD ["./litcart_app"]