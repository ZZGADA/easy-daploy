package websocket

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/version"
	"time"

	"github.com/ZZGADA/easy-deploy/internal/model/conf"
	"github.com/gorilla/websocket"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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
		baseProcess(conn, command, data, userID)
	}

	SendSuccess(conn, "command execute success", K8sCommandResponse{
		Command: command,
		Result:  "success",
	})
}

func baseProcess(conn *websocket.Conn, command string, data map[string]interface{}, userID uint) {
	//result, err := executeSSHCommand(command)
	//if err != nil {
	//	SendError(conn, err.Error())
	//} else {
	SendSuccess(conn, "command execute success", K8sCommandResponse{
		Command: command,
		Result:  command,
	})
	//}

}

// 执行 SSH 命令
func executeSSHCommand(command string) (string, error) {
	config := &ssh.ClientConfig{
		User: remoteUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(remotePassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：在生产环境中不要使用此选项，应验证主机密钥
	}

	client, err := ssh.Dial("tcp", remoteHost+":"+remotePort, config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

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
