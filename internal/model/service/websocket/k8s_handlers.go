package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZZGADA/easy-deploy/internal/define"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/dao"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
)

const (
	GetAllNameSpace  = "kubectl get namespace"
	GetAllPods       = "kubectl get pod -A"
	GetAllService    = "kubectl get svc -A"
	GetAllDeployment = "kubectl get deployment -A"
	GetClusterInfo   = "kubectl cluster-info"
	GetNodes         = "kubectl get nodes"
	ResourceApply    = "kubectl apply -f"
	ResourceDelete   = "kubectl delete"
)

// 远程服务器配置
const (
	remoteHost     = "your-remote-host"
	remotePort     = "22"
	remoteUsername = "your-username"
	remotePassword = "your-password"
)

type K8sCommandResponse struct {
	Command string `json:"command"`
	Result  string `json:"result"`
}

func (s *SocketService) HandleKubeCommand(conn *websocket.Conn, command string, data map[string]interface{}, userID uint) {
	switch command {
	case GetAllNameSpace:
		{
			namespaces, err := conf.KubeClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to get namespaces: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatNamespace(namespaces),
			})
			return
		}
	case GetAllPods:
		{
			pods, err := conf.KubeClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to get pods: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatPods(pods),
			})
			return
		}
	case GetAllService:
		{
			services, err := conf.KubeClient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to get services: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatServices(services),
			})
			return
		}
	case GetAllDeployment:
		{
			deployments, err := conf.KubeClient.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to get deployments: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatDeployments(deployments),
			})
			return
		}
	case GetClusterInfo:
		{
			version, err := conf.KubeClient.Discovery().ServerVersion()
			if err != nil {
				panic(fmt.Errorf("failed to get cluster info: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatClusterInfo(version),
			})
			return
		}
	case GetNodes:
		{
			nodes, err := conf.KubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to get nodes: %v", err))
			}

			SendSuccess(conn, "command execute success", K8sCommandResponse{
				Command: command,
				Result:  formatNodes(nodes),
			})
			return
		}
	default:
		s.baseProcess(conn, command, data, userID)
	}
}

func (s *SocketService) baseProcess(conn *websocket.Conn, command string, data map[string]interface{}, userID uint) {
	switch command {
	case ResourceApply:
		s.resourceApply(conn, command, data, userID)
		return
	case ResourceDelete:
		s.resourceDelete(conn, command, data, userID)
		return
	default:
		SendSuccess(conn, "command execute success", K8sCommandResponse{
			Command: command,
			Result:  command,
		})
	}
}

func (s *SocketService) resourceApply(conn *websocket.Conn, command string, data map[string]interface{}, userID uint) {
	logrus.Info("resource apply ", "data: ", data)
	k8sResourceID, exist := data["k8s_resource_id"].(float64)
	if !exist {
		SendError(conn, "缺少k8s_resource_id 参数")
		return
	}

	// 从数据库查询资源信息
	resource, err := s.userK8sResourceDao.QueryById(uint32(k8sResourceID))
	if err != nil {
		SendError(conn, fmt.Sprintf("查询资源失败: %v", err))
		return
	}

	// 创建本地 k8s 目录（如果不存在）
	k8sDir := "k8s"
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		SendError(conn, fmt.Sprintf("创建 k8s 目录失败: %v", err))
		return
	}

	// 生成本地文件路径
	localFilePath := filepath.Join(k8sDir, fmt.Sprintf("%d_%s", resource.Id, resource.FileName))

	// 从 OSS 下载文件
	ossClient, exist := conf.WSServer.OssClient[userID]
	if !exist {
		SendError(conn, fmt.Sprintf("获取 OSS 客户端失败: %v", err))
		return
	}

	// 从 URL 中提取 object-name
	objectNameUrl := strings.TrimPrefix(resource.OssURL, "https://")
	objectNameS := strings.Split(objectNameUrl, "/")[1:] // 去掉域名部分
	objectName := strings.Join(objectNameS, "/")

	// 下载文件
	err = ossClient.GetObjectToFile(objectName, localFilePath)
	if err != nil {
		SendError(conn, fmt.Sprintf("从 OSS 下载文件失败: %v", err))
		return
	}

	// 确保函数结束时删除本地文件
	defer func() {
		if err := os.Remove(localFilePath); err != nil {
			logrus.Errorf("删除本地文件失败: %v", err)
		}
	}()

	// 读取 YAML 文件内容
	yamlContent, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		SendError(conn, fmt.Sprintf("读取 YAML 文件失败: %v", err))
		return
	}

	// 解析 YAML 文件，获取 namespace
	namespace, err := s.extractNamespaceFromYAML(yamlContent)
	if err != nil {
		SendError(conn, fmt.Sprintf("解析 YAML 文件失败: %v", err))
		return
	}

	// 如果 namespace 为空，使用默认的 "default"
	if namespace == "" {
		namespace = "default"
	}

	// 检查 namespace 是否存在，如果不存在则创建
	if err := s.ensureNamespaceExists(namespace); err != nil {
		SendError(conn, fmt.Sprintf("创建 namespace 失败: %v", err))
		return
	}

	// 解析 YAML 文件，获取资源名称和标签
	resourceName, labels, err := s.extractResourceInfo(yamlContent)
	if err != nil {
		SendError(conn, fmt.Sprintf("解析资源信息失败: %v", err))
		return
	}

	// 创建资源
	if err := s.createResourceFromYAML(conf.KubeClient, localFilePath, namespace); err != nil {
		SendError(conn, fmt.Sprintf("创建资源失败: %v", err))
		return
	}

	// 记录操作日志
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		logrus.Errorf("序列化标签失败: %v", err)
		// 继续执行，不中断流程
	}

	// 构建完整的 kubectl 命令
	fullCommand := fmt.Sprintf("kubectl apply -f %s -n %s", localFilePath, namespace)

	// 检查资源状态
	var status int
	if resource.ResourceType == "deployment" {
		// 等待一段时间，让部署有时间启动
		time.Sleep(5 * time.Second)

		// 检查部署状态
		deployment, err := conf.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if err != nil {
			// 部署失败
			status = define.K8sResourceStatusStop // 运行停止
		} else {
			// 检查部署状态
			if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
				status = define.K8sResourceStatusRun // 运行正常
			} else if deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
				status = define.K8sResourceStatusRestart // 容器重启
			} else {
				status = define.K8sResourceStatusStop // 运行停止
			}
		}
	} else if resource.ResourceType == "service" {
		// 服务创建后默认为运行正常
		status = define.K8sResourceStatusRun
	}

	// 创建操作日志
	operationLog := &dao.UserK8sResourceOperationLog{
		K8sResourceID:  uint(k8sResourceID),
		UserID:         userID,
		Namespace:      namespace,
		MetadataName:   resourceName,
		MetadataLabels: string(labelsJSON),
		OperationType:  "create",
		Status:         status,
		Command:        fullCommand,
	}

	// 保存操作日志
	if err := s.userK8sResourceOperationLogDao.Create(operationLog); err != nil {
		logrus.Errorf("保存操作日志失败: %v", err)
		// 继续执行，不中断流程
	}

	// 发送成功响应
	SendSuccess(conn, "command execute success", K8sCommandResponse{
		Command: command,
		Result:  fmt.Sprintf("资源 %s 已成功部署到 namespace %s", resource.FileName, namespace),
	})
}

func (s *SocketService) resourceDelete(conn *websocket.Conn, command string, data map[string]interface{}, userID uint) {
	logrus.Info("resource delete ", "data: ", data)
	k8sResourceID, exist := data["k8s_resource_id"].(float64)
	if !exist {
		SendError(conn, "缺少k8s_resource_id 参数")
		return
	}

	// 从数据库查询资源信息
	resource, err := s.userK8sResourceDao.QueryById(uint32(k8sResourceID))
	if err != nil {
		SendError(conn, fmt.Sprintf("查询资源失败: %v", err))
		return
	}

	// 查询最新的操作日志，获取资源信息
	logs, err := s.userK8sResourceOperationLogDao.QueryByK8sResourceID(uint(k8sResourceID))
	if err != nil || len(logs) == 0 {
		SendError(conn, fmt.Sprintf("查询资源操作日志失败: %v", err))
		return
	}

	// 获取最新的操作日志
	latestLog := logs[0]
	namespace := latestLog.Namespace
	metadataName := latestLog.MetadataName

	// 检查资源是否正在运行
	resourceType := resource.ResourceType

	if resourceType == "deployment" {
		_, err = conf.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
	} else if resourceType == "service" {
		_, err = conf.KubeClient.CoreV1().Services(namespace).Get(context.TODO(), metadataName, metav1.GetOptions{})
	} else {
		SendError(conn, fmt.Sprintf("不支持的资源类型: %s", resourceType))
		return
	}

	// 如果资源不存在，说明已经停止运行
	if err != nil {
		if k8sErr, ok := err.(*k8serrors.StatusError); ok && k8sErr.Status().Code == 404 {
			SendError(conn, fmt.Sprintf("%s 已经关闭", metadataName))
			return
		}
		SendError(conn, fmt.Sprintf("检查资源状态失败: %v", err))
		return
	}

	// 资源正在运行，执行删除操作
	var deleteCommand string
	var deleteErr error

	if resourceType == "deployment" {
		deleteCommand = fmt.Sprintf("kubectl delete deployment %s -n %s", metadataName, namespace)
		deleteErr = conf.KubeClient.AppsV1().Deployments(namespace).Delete(context.TODO(), metadataName, metav1.DeleteOptions{})
	} else if resourceType == "service" {
		deleteCommand = fmt.Sprintf("kubectl delete service %s -n %s", metadataName, namespace)
		deleteErr = conf.KubeClient.CoreV1().Services(namespace).Delete(context.TODO(), metadataName, metav1.DeleteOptions{})
	}

	// 检查删除操作是否成功
	if deleteErr != nil {
		SendError(conn, fmt.Sprintf("删除资源失败: %v", deleteErr))
		return
	}

	// 记录操作日志
	operationLog := &dao.UserK8sResourceOperationLog{
		K8sResourceID:  uint(k8sResourceID),
		UserID:         userID,
		Namespace:      namespace,
		MetadataName:   metadataName,
		MetadataLabels: latestLog.MetadataLabels,
		OperationType:  "delete",
		Status:         define.K8sResourceStatusStop, // 运行停止
		Command:        deleteCommand,
	}

	// 保存操作日志
	if err := s.userK8sResourceOperationLogDao.Create(operationLog); err != nil {
		logrus.Errorf("保存操作日志失败: %v", err)
		// 继续执行，不中断流程
	}

	// 发送成功响应
	SendSuccess(conn, "command execute success", K8sCommandResponse{
		Command: command,
		Result:  fmt.Sprintf("资源 %s 已成功停止运行", metadataName),
	})
}

// 从 YAML 内容中提取 namespace
func (s *SocketService) extractNamespaceFromYAML(yamlContent []byte) (string, error) {
	// 创建一个通用的 map 来解析 YAML
	var resource map[string]interface{}
	if err := yaml.Unmarshal(yamlContent, &resource); err != nil {
		return "", fmt.Errorf("解析 YAML 失败: %v", err)
	}

	// 检查 metadata.namespace 字段
	if metadata, ok := resource["metadata"].(map[string]interface{}); ok {
		if namespace, ok := metadata["namespace"].(string); ok {
			return namespace, nil
		}
	}

	// 如果没有找到 namespace，返回空字符串
	return "", nil
}

// 确保 namespace 存在，如果不存在则创建
func (s *SocketService) ensureNamespaceExists(namespace string) error {
	// 检查 namespace 是否存在
	_, err := conf.KubeClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil {
		// namespace 已存在
		return nil
	}

	// 检查错误是否为 "not found"
	if k8sErr, ok := err.(*k8serrors.StatusError); ok && k8sErr.Status().Code == 404 {
		// namespace 不存在，创建它
		ns := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = conf.KubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("创建 namespace 失败: %v", err)
		}
		return nil
	}

	// 其他错误
	return fmt.Errorf("检查 namespace 失败: %v", err)
}

// GetOssClient 获取 OSS 客户端
func (s *SocketService) GetOssClient(userId uint) (*oss.Bucket, error) {
	// 从配置或数据库获取 OSS 配置
	// 这里假设您有一个方法来获取 OSS 配置

	userOss, err := s.userOssDao.QueryByUserID(userId)
	if err != nil {
		return nil, err
	}

	// 创建 OSS 客户端
	client, err := oss.New(fmt.Sprintf("https://%s.aliyuncs.com", userOss.Region), userOss.AccessKeyID, userOss.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建 OSS 客户端失败: %v", err)
	}

	// 获取 bucket
	bucket, err := client.Bucket(userOss.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取 bucket 失败: %v", err)
	}

	return bucket, nil
}

func (s *SocketService) createResourceFromYAML(client *kubernetes.Clientset, yamlPath string, namespace string) error {
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %v", err)
	}

	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add apps/v1 to scheme: %v", err)
	}
	if err := v1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add core/v1 to scheme: %v", err)
	}

	codecFactory := serializer.NewCodecFactory(scheme)
	decoder := codecFactory.UniversalDeserializer()

	obj, _, err := decoder.Decode(yamlFile, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %v", err)
	}

	switch resource := obj.(type) {
	case *appsv1.Deployment:
		_, err = client.AppsV1().Deployments(namespace).Create(context.TODO(), resource, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Deployment: %v", err)
		}
		fmt.Printf("Deployment %s created successfully.\n", resource.Name)
	case *v1.Service:
		_, err = client.CoreV1().Services(namespace).Create(context.TODO(), resource, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Service: %v", err)
		}
		fmt.Printf("Service %s created successfully.\n", resource.Name)
	default:
		return fmt.Errorf("unsupported resource type: %T", resource)
	}

	return nil
}

// 执行 SSH 命令
//func executeSSHCommand(command string) (string, error) {
//	config := &ssh.ClientConfig{
//		User: remoteUsername,
//		Auth: []ssh.AuthMethod{
//			ssh.Password(remotePassword),
//		},
//		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：在生产环境中不要使用此选项，应验证主机密钥
//	}
//
//	client, err := ssh.Dial("tcp", remoteHost+":"+remotePort, config)
//	if err != nil {
//		return "", err
//	}
//	defer client.Close()
//
//	session, err := client.NewSession()
//	if err != nil {
//		return "", err
//	}
//	defer session.Close()
//
//	output, err := session.CombinedOutput(command)
//	if err != nil {
//		return "", err
//	}
//
//	return string(output), nil
//}

func formatNamespace(namespaces *v1.NamespaceList) string {
	var result string

	result = fmt.Sprintf("%-20s %-10s %-10s\n", "NAME", "STATUS", "AGE")
	now := time.Now()
	for _, ns := range namespaces.Items {
		// 打印每个命名空间的信息
		age := now.Sub(ns.CreationTimestamp.Time)
		ageStr := formatDuration(age)
		result += fmt.Sprintf("%-20s %-10s %-10s\n", ns.Name, string(ns.Status.Phase), ageStr)

	}
	return result
}

func formatPods(pods *v1.PodList) string {
	var result string

	result = fmt.Sprintf("%-20s %-40s %-10s %-10s %-10s %-10s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE")
	now := time.Now()
	for _, pod := range pods.Items {
		cnt := 0
		for _, cond := range pod.Status.ContainerStatuses {
			if cond.Ready {
				cnt++
			}
		}

		ready := fmt.Sprintf("%d/%d", cnt, len(pod.Spec.Containers))
		age := now.Sub(pod.CreationTimestamp.Time)
		ageStr := formatDuration(age)
		restarts := 0
		if len(pod.Status.ContainerStatuses) > 0 {
			restarts = int(pod.Status.ContainerStatuses[0].RestartCount)
		}
		result += fmt.Sprintf("%-20s %-40s %-10s %-10s %-10d %-10s\n",
			pod.Namespace, pod.Name, ready, string(pod.Status.Phase), restarts, ageStr)
	}
	return result
}

func formatServices(services *v1.ServiceList) string {
	var result string

	result = fmt.Sprintf("%-20s %-20s %-10s %-20s %-10s\n", "NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "PORT(S)")
	for _, svc := range services.Items {
		ports := ""
		for i, port := range svc.Spec.Ports {
			if i > 0 {
				ports += ", "
			}
			ports += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}
		result += fmt.Sprintf("%-20s %-20s %-10s %-20s %-10s\n",
			svc.Namespace, svc.Name, string(svc.Spec.Type), svc.Spec.ClusterIP, ports)
	}
	return result
}

func formatDeployments(deployments *appsv1.DeploymentList) string {
	var result string

	result = fmt.Sprintf("%-20s %-20s %-10s %-10s %-10s %-10s\n", "NAMESPACE", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")
	now := time.Now()
	for _, dep := range deployments.Items {
		ready := fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas)
		age := now.Sub(dep.CreationTimestamp.Time)
		ageStr := formatDuration(age)
		result += fmt.Sprintf("%-20s %-20s %-10s %-10d %-10d %-10s\n",
			dep.Namespace, dep.Name, ready, dep.Status.UpdatedReplicas, dep.Status.AvailableReplicas, ageStr)
	}
	return result
}

func formatClusterInfo(version *version.Info) string {
	var result string

	result = fmt.Sprintf("Kubernetes control plane is running at %s\n", conf.KubeConfig.Host)
	result += fmt.Sprintf("Kubernetes version: %s\n", version.String())
	return result
}

func formatNodes(nodes *v1.NodeList) string {
	var result string

	result = fmt.Sprintf("%-20s %-10s %-10s %-10s %-10s %-10s\n", "NAME", "STATUS", "ROLES", "AGE", "VERSION", "INTERNAL-IP")
	now := time.Now()
	for _, node := range nodes.Items {
		age := now.Sub(node.CreationTimestamp.Time)
		ageStr := formatDuration(age)
		roles := ""
		for key := range node.Labels {
			if key == "node-role.kubernetes.io/master" || key == "node-role.kubernetes.io/control-plane" {
				roles = "master"
				break
			} else if key == "node-role.kubernetes.io/worker" {
				roles = "worker"
				break
			}
		}
		if roles == "" {
			roles = "<none>"
		}
		internalIP := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				internalIP = addr.Address
				break
			}
		}
		result += fmt.Sprintf("%-20s %-10s %-10s %-10s %-10s %-10s\n",
			node.Name, string(node.Status.Conditions[len(node.Status.Conditions)-1].Type), roles, ageStr, node.Status.NodeInfo.KubeletVersion, internalIP)
	}
	return result
}

// formatDuration 将时间间隔格式化为类似 "19h" 的字符串
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}

// 从 YAML 内容中提取资源名称和标签
func (s *SocketService) extractResourceInfo(yamlContent []byte) (string, map[string]string, error) {
	// 创建一个通用的 map 来解析 YAML
	var resource map[string]interface{}
	if err := yaml.Unmarshal(yamlContent, &resource); err != nil {
		return "", nil, fmt.Errorf("解析 YAML 失败: %v", err)
	}

	// 获取资源名称
	var resourceName string
	if metadata, ok := resource["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			resourceName = name
		}
	}

	// 获取资源标签
	labels := make(map[string]string)
	if metadata, ok := resource["metadata"].(map[string]interface{}); ok {
		if labelsMap, ok := metadata["labels"].(map[string]interface{}); ok {
			for key, value := range labelsMap {
				if strValue, ok := value.(string); ok {
					labels[key] = strValue
				}
			}
		}
	}

	return resourceName, labels, nil
}
