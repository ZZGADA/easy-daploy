# 基于Kubernetes的一体化DevOps云平台

## 项目概述

本项目是一个基于**Kubernetes**的综合性DevOps云平台，旨在通过整合Kubernetes和DevOps理念，简化软件开发与部署流程。项目采用**Go**语言（后端，基于Gin框架）和**TypeScript**（前端，基于React框架）开发，实现了全自动化的**CI/CD流水线**，支持持续集成、持续交付和持续部署。平台利用**Docker**进行容器化，**Kubernetes**进行容器编排，并通过**EFK**（Elasticsearch、Fluentd、Kibana）技术实现日志收集与监控，为管理容器化应用提供了高效的解决方案。

本项目的核心目标是帮助学生和开发者将项目部署到服务器上，超越本地IDE环境的限制。通过模拟企业级的开发工作流，用户能够深入理解和实践DevOps理念，使项目部署变得简单高效。项目源码及详细的技术白皮书已开源至GitHub，以促进Kubernetes和DevOps技术的学习与普及。

## 功能特性

- **自动化CI/CD流水线**：
  - **持续集成（CI）**：通过Docker和Shell脚本，自动完成GitHub代码仓库的编译、测试和镜像构建。
  - **持续交付/部署（CD）**：将Docker镜像推送至远程仓库，并通过一键操作部署到Kubernetes集群。
- **Kubernetes资源管理**：
  - 支持Kubernetes资源配置文件（例如Deployment、Service）的创建、更新、删除和渲染。
  - 提供控制面板，支持部署、停止、查询Pod、Service、Deployment、Namespace及集群/节点状态等操作。
- **日志收集与监控告警**：
  - 利用EFK技术实现容器日志的实时收集、存储和可视化。
  - 基于Kafka的监控告警系统，对日志等级高于“Info”的记录通过邮件通知相关人员。
- **用户与团队管理**：
  - 提供用户注册登录、GitHub账号绑定、Docker/OSS账号管理及团队协作功能。
  - 支持团队创建、成员管理及通过邮件审批的入队/离队流程。
- **Web界面**：
  - 采用TypeScript和React开发，提供响应式、用户友好的前端界面。
  - 通过可视化组件支持CI/CD操作、代码仓库管理和Kubernetes资源监控。
- **数据管理**：
  - 使用MySQL存储结构化数据，Redis进行高速缓存，阿里云OSS存储Kubernetes配置文件等非结构化数据。
  - 数据库设计合理，支持功能模块的数据存储与交互。

## 技术栈

| **组件**           | **技术**             | **版本**     | **用途**                     |
|--------------------|---------------------|-------------|-----------------------------|
| 后端               | Go (Gin框架)        | v1.23       | API服务与业务逻辑          |
| 前端               | TypeScript, React   | Node v18.20.5 | 用户界面与交互            |
| 数据库             | MySQL               | v8.0.22     | 结构化数据存储             |
| 缓存               | Redis               | v6.2.14     | 高速缓存                   |
| 日志存储           | Elasticsearch       | v7.17.10    | 日志存储与搜索             |
| 日志收集           | Fluentd             | v1.18       | 容器日志收集               |
| 日志可视化         | Kibana              | v7.17.10    | 日志查询与可视化           |
| 消息队列           | Kafka, Zookeeper    | v7.3.3      |システム

| 容器化             | Docker              | v27.4.0     | 镜像构建与管理             |
| 容器编排           | Kubernetes          | v1.30.5     | 容器部署与扩展             |
| 对象存储           | 阿里云OSS           | -           | Kubernetes配置文件存储     |

## 系统架构

平台采用**B/S（浏览器/服务器）**架构，前后端分离，分为六层：

1. **表示层**：基于TypeScript和React，构建直观的用户界面。
2. **网关层**：处理API路由分发和用户鉴权，支持HTTP和WebSocket协议。
3. **应用层**：后端核心逻辑，包括用户管理、GitHub集成、Docker镜像构建和Kubernetes编排等模块。
4. **监控层**：通过EFK实现日志收集，基于Kafka实现实时监控与告警。
5. **基建层**：集成Docker、Kubernetes、MySQL、Redis、OSS和Elasticsearch，提供稳定的运行环境。
6. **第三方服务**：利用阿里云ECS部署中间件，GitHub托管代码，阿里云容器镜像服务存储镜像。

## 安装与配置

### 前置条件
- Go (v1.23)
- Node.js (v18.20.5)
- Docker (v27.4.0)
- Kubernetes (v1.30.5)
- MySQL (v8.0.22)
- Redis (v6.2.14)
- Elasticsearch (v7.17.10), Fluentd (v1.18), Kibana (v7.17.10)
- Kafka (v7.3.3), Zookeeper (v7.3.3)
- 阿里云账号（用于OSS和ECS）

### 安装步骤
1. **克隆仓库**：
   ```bash
   git clone https://github.com/your-repo/devops-kubernetes-platform.git
   cd devops-kubernetes-platform
   ```

2. **后端配置**：
   - 进入后端目录并安装依赖：
     ```bash
     cd backend
     go mod tidy
     ```
   - 在`.env`文件中配置环境变量（例如MySQL、Redis、OSS凭证）。
   - 启动Go服务：
     ```bash
     go run main.go
     ```

3. **前端配置**：
   - 进入前端目录并安装依赖：
     ```bash
     cd frontend
     npm install
     ```
   - 启动React开发服务器：
     ```bash
     npm start
     ```

4. **中间件部署**：
   - 在阿里云ECS或本地部署MySQL、Redis、Elasticsearch、Kafka等中间件。
   - 在Kubernetes集群中以DaemonSet形式配置Fluentd以收集日志。
   - 配置Kibana连接Elasticsearch以实现日志可视化。

5. **Kubernetes集群**：
   - 确保Kubernetes集群正常运行（可使用Minikube或云服务提供商）。
   - 应用Fluentd DaemonSet及其他资源配置。

6. **访问平台**：
   - 在浏览器中打开前端URL（默认：`http://localhost:3000`）。
   - 注册/登录并绑定GitHub/Docker/OSS账号以开始使用。

## 使用指南

1. **用户注册/登录**：
   - 使用邮箱和密码注册，通过验证码验证后登录平台。
2. **账号绑定**：
   - 绑定GitHub以访问代码仓库，绑定Docker以存储镜像，绑定OSS以存储Kubernetes配置文件。
3. **团队管理**：
   - 创建或加入团队进行项目协作，通过邮件审批入队/离队。
4. **代码仓库管理**：
   - 查看和管理GitHub仓库，包括分支、文件树和文件详情。
5. **Docker镜像构建**：
   - 创建/导入Dockerfile，触发CI/CD流水线以构建并推送镜像至远程仓库。
6. **Kubernetes部署**：
   - 管理Kubernetes资源（Deployment、Service），执行部署/停止操作，监控集群状态。
7. **日志监控**：
   - 使用Kibana查询容器日志，通过Kafka接收关键问题的邮件告警。

## 项目动机

本项目作为本科毕业设计，旨在弥合学术学习与企业级开发之间的差距。通过提供一个开源的DevOps平台，项目希望：
- 帮助学生将项目部署到服务器，体验真实的DevOps工作流。
- 通过用户友好的界面简化Kubernetes和Docker的复杂性。
- 通过实践促进对CI/CD、容器编排和监控的理解。
- 通过开源代码和文档为社区做出贡献。

## 未来改进

- **性能优化**：优化Kubernetes资源调度算法和数据库查询效率。
- **功能扩展**：集成人工智能实现预测性扩展，支持更多云平台。
- **安全性增强**：加强容器安全检查和数据加密。
- **易用性提升**：简化操作流程，提供更详细的文档和优化的用户界面。

## 许可证

本项目采用MIT许可证，详情见[LICENSE](LICENSE)文件。

## 致谢


- 技术支持：Kubernetes、Docker、Go、TypeScript、React、EFK、Kafka
- 云服务：阿里云（ECS、OSS、容器镜像服务）
- 开源社区：提供工具、库和灵感

如需详细技术说明，请参阅[毕业设计](./docs/基于Kubernetes的DevOps云平台系统设计-v4-匿名.pdf)。欢迎贡献和反馈！