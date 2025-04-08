# 使用官方的 fluentd-kubernetes-daemonset 镜像作为基础镜像
FROM fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch

# 切换到 root 用户进行安装操作
USER root

# 更新包管理器并安装所需的依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    ruby-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 fluent-plugin-kubernetes_metadata_filter 插件
RUN gem install fluent-plugin-kubernetes_metadata_filter