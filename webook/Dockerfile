# 基础镜像
FROM ubuntu:20.04
# 将编译后的webook打包进镜像，放到工作目录 /app
COPY webook /app/webook
WORKDIR /app
# CMD 是执行命令，ENTRYPOINT 也是，相对来说ENTRYPOINT最佳
ENTRYPOINT ["/app/webook"]
