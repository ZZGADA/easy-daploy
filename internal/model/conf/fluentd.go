package conf

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/ZZGADA/easy-deploy/internal/config"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/rest"
)

// 配置模板
const fluentdConfigTemplate = `
<source>
  @type tail
  @id in_tail_container_logs
  path /var/log/containers/*.log
  pos_file /var/log/fluentd-containers.log.pos
  tag kubernetes.*
  read_from_head true
  <parse>
    @type json
    time_format %Y-%m-%dT%H:%M:%S.%NZ
  </parse>
</source>

<filter kubernetes.**>
  @type kubernetes_metadata
</filter>

<filter kubernetes.**>
  @type record_transformer
  <record>
    timestamp ${time}
  </record>
</filter>

<match kubernetes.**>
  @type copy
  <store>
	@type elasticsearch_dynamic
  	host {{.ElasticsearchHost}}
  	port {{.ElasticsearchPort}}
  	user {{.ElasticsearchUser}}
  	password {{.ElasticsearchPassword}}
  	logstash_format true
  	logstash_prefix k8s_${record['kubernetes']['container_name']}
  	include_tag_key true
  	tag_key @log_name
  	retry_max_times 5
  	max_retry_wait 30s
  	disable_retry_limit false
  	reconnect_on_error true
  	reload_on_failure true
  	reload_connections false
  	reload_after -1
  </store>

  <store>
    @type kafka2
    brokers {{.KafkaBrokers}}
	default_topic {{.KafkaTopic}}
    output_data_type json
    <format>
      @type json
    </format>
	<buffer>
	  flush_interval 5s
	</buffer>
  </store>
</match>
`

// FluentdConfig 配置结构体
type FluentdConfig struct {
	ElasticsearchHost     string
	ElasticsearchPort     int
	ElasticsearchUser     string
	ElasticsearchPassword string
	OS                    string
	KafkaBrokers          string
	KafkaTopic            string
}

// InitFluent 初始化Fluentd
func InitFluent() {
	// 获取节点列表
	_, err := KubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("获取节点列表失败: %v", err)
	}

	// 配置 Elasticsearch 信息
	configData := FluentdConfig{
		ElasticsearchHost:     config.GlobalConfig.Elastic.Host,
		ElasticsearchPort:     config.GlobalConfig.Elastic.Port,
		ElasticsearchUser:     config.GlobalConfig.Elastic.Username,
		ElasticsearchPassword: config.GlobalConfig.Elastic.Password,
		KafkaBrokers:          strings.Join(config.GlobalConfig.Kafka.Brokers, ","),
		KafkaTopic:            config.GlobalConfig.Kafka.Topic,
	}

	// 生成 Fluentd 配置文件
	tmpl, err := template.New("fluentdConfig").Parse(fluentdConfigTemplate)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}
	var configBuffer bytes.Buffer
	err = tmpl.Execute(&configBuffer, configData)
	if err != nil {
		log.Fatalf("执行模板失败: %v", err)
	}
	fluentdConfig := configBuffer.String()

	// 创建 ConfigMap 存储 Fluentd 配置
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd-config",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"fluent.conf": fluentdConfig,
		},
	}

	// 检查 ConfigMap 是否已存在
	_, err = KubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "fluentd-config", metav1.GetOptions{})
	if err == nil {
		// 更新 ConfigMap
		_, err = KubeClient.CoreV1().ConfigMaps("kube-system").Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("更新 Fluentd ConfigMap 失败: %v", err)
		}
		log.Info("已更新 Fluentd ConfigMap")
	} else {
		// 创建 ConfigMap
		_, err = KubeClient.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("创建 Fluentd ConfigMap 失败: %v", err)
		}
		log.Info("已创建 Fluentd ConfigMap")
	}

	// 创建 ServiceAccount
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: "kube-system",
		},
	}

	// 检查 ServiceAccount 是否已存在
	_, err = KubeClient.CoreV1().ServiceAccounts("kube-system").Get(context.TODO(), "fluentd", metav1.GetOptions{})
	if err == nil {
		log.Info("Fluentd ServiceAccount 已存在")
	} else {
		// 创建 ServiceAccount
		_, err = KubeClient.CoreV1().ServiceAccounts("kube-system").Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("创建 Fluentd ServiceAccount 失败: %v", err)
		}
		log.Info("已创建 Fluentd ServiceAccount")
	}

	// 创建 ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluentd",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "namespaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// 检查 ClusterRole 是否已存在
	_, err = KubeClient.RbacV1().ClusterRoles().Get(context.TODO(), "fluentd", metav1.GetOptions{})
	if err == nil {
		// 更新 ClusterRole
		_, err = KubeClient.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("更新 Fluentd ClusterRole 失败: %v", err)
		}
		log.Info("已更新 Fluentd ClusterRole")
	} else {
		// 创建 ClusterRole
		_, err = KubeClient.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("创建 Fluentd ClusterRole 失败: %v", err)
		}
		log.Info("已创建 Fluentd ClusterRole")
	}

	// 创建 ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluentd",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "fluentd",
				Namespace: "kube-system",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "fluentd",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// 检查 ClusterRoleBinding 是否已存在
	_, err = KubeClient.RbacV1().ClusterRoleBindings().Get(context.TODO(), "fluentd", metav1.GetOptions{})
	if err == nil {
		// 更新 ClusterRoleBinding
		_, err = KubeClient.RbacV1().ClusterRoleBindings().Update(context.TODO(), clusterRoleBinding, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalf("更新 Fluentd ClusterRoleBinding 失败: %v", err)
		}
		log.Info("已更新 Fluentd ClusterRoleBinding")
	} else {
		// 创建 ClusterRoleBinding
		_, err = KubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("创建 Fluentd ClusterRoleBinding 失败: %v", err)
		}
		log.Info("已创建 Fluentd ClusterRoleBinding")
	}

	// 创建 DaemonSet
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "fluentd",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "fluentd",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "fluentd",
					Containers: []v1.Container{
						{
							Name:            "fluentd",
							Image:           "zzgeda-fluentd:1.0.0",
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								{
									Name:  "FLUENT_ELASTICSEARCH_HOST",
									Value: config.GlobalConfig.Elastic.Host,
								},
								{
									Name:  "FLUENT_ELASTICSEARCH_PORT",
									Value: fmt.Sprintf("%d", config.GlobalConfig.Elastic.Port),
								},
								{
									Name:  "FLUENT_ELASTICSEARCH_USER",
									Value: config.GlobalConfig.Elastic.Username,
								},
								{
									Name:  "FLUENT_ELASTICSEARCH_PASSWORD",
									Value: config.GlobalConfig.Elastic.Password,
								},
								{
									Name: "HOSTNAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name: "K8S_NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/fluentd/etc/fluent.conf",
									SubPath:   "fluent.conf",
								},
								{
									Name:      "varlog",
									MountPath: "/var/log",
								},
								{
									Name:      "varlibdockercontainers",
									MountPath: "/var/lib/docker/containers",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "fluentd-config",
									},
								},
							},
						},
						{
							Name: "varlog",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/var/log",
								},
							},
						},
						{
							Name: "varlibdockercontainers",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/var/lib/docker/containers",
								},
							},
						},
					},
				},
			},
		},
	}

	// 检查 DaemonSet 是否已存在
	_, err = KubeClient.AppsV1().DaemonSets("kube-system").Get(context.TODO(), "fluentd", metav1.GetOptions{})
	if err == nil {
		//// 更新 DaemonSet
		//_, err = KubeClient.AppsV1().DaemonSets("kube-system").Update(context.TODO(), daemonSet, metav1.UpdateOptions{})
		//if err != nil {
		//	log.Fatalf("更新 Fluentd DaemonSet 失败: %v", err)
		//}
		//log.Info("已更新 Fluentd DaemonSet")
	} else {
		// 创建 DaemonSet
		_, err = KubeClient.AppsV1().DaemonSets("kube-system").Create(context.TODO(), daemonSet, metav1.CreateOptions{})
		if err != nil {
			log.Fatalf("创建 Fluentd DaemonSet 失败: %v", err)
		}
		log.Info("已创建 Fluentd DaemonSet")
	}

	// 等待 DaemonSet 就绪
	log.Info("等待 Fluentd DaemonSet 就绪...")
	checkEnd := 30
	for i := 0; i < checkEnd; i++ {
		ds, err := KubeClient.AppsV1().DaemonSets("kube-system").Get(context.TODO(), "fluentd", metav1.GetOptions{})
		if err != nil {
			log.Fatalf("获取 Fluentd DaemonSet 状态失败: %v", err)
		}

		if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
			log.Info("Fluentd DaemonSet 已就绪")
			break
		} else {
			if i == checkEnd-1 {
				panic("Fluentd DaemonSet 启动失败，请查看DaemonSet集群情况")
			}
		}

		log.Infof("Fluentd DaemonSet 就绪状态: %d/%d", ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
		time.Sleep(5 * time.Second)
	}

	log.Info("Fluentd 部署完成")
}
