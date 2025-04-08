# 使用基础镜像
FROM fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch

# 设置工作目录
WORKDIR /fluentd

# 切换到root用户，以便安装插件
USER root

# 安装与Kubernetes相关的插件
RUN fluent-gem install fluent-plugin-kubernetes_metadata \
    && fluent-gem install fluent-plugin-kubernetes_metadata_filter \
    && fluent-gem install fluent-plugin-elasticsearch \
    && fluent-gem install fluent-plugin-rewrite-tag-filter \
    && fluent-gem install fluent-plugin-concat \
    && fluent-gem install fluent-plugin-multi-format-parser

# 切换回非root用户（Fluentd默认以非root用户运行）
USER fluent

# 可以在这里添加其他配置或文件复制操作，比如将配置文件复制到镜像中
# COPY your-fluentd-config.conf /fluentd/etc/fluent.conf

# 定义容器启动时执行的命令，这里假设Fluentd的启动命令是fluentd -c /fluentd/etc/fluent.conf
# docker build -t custom-fluentd-k8s-image:1.0.0 .
CMD ["fluentd", "-c", "/fluentd/etc/fluent.conf"]
