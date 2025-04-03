package conf

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

var KubeClient *kubernetes.Clientset
var KubeConfig *rest.Config

func InitK8s() {
	// 定义 kubeconfig 文件的路径
	kubeconfigPath := filepath.Join(
		filepath.Dir(clientcmd.RecommendedHomeFile),
		"config",
	)

	// 构建配置对象
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(fmt.Errorf("failed to build kubeconfig: %v", err))
	}

	// 打印 API 服务器的 URL
	logrus.Infof("API Server URL: %s\n", config.Host)

	// 创建客户端集
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create clientset: %v", err))
	}

	KubeClient = clientSet
	KubeConfig = config
}
