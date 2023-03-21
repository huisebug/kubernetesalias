package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var err error
var clientset *kubernetes.Clientset

type PodContainer struct {
	Namespace, Name, Container string
}

func init() {
	switch {
	case len(os.Args) == 1:
		fmt.Println("1.确保存在kubernetes管理认证文件/etc/kubernetes/admin.conf")
		fmt.Println("2.请输入pod过滤条件,仅允许传递一个条件")
		os.Exit(0)
	case len(os.Args) > 2:
		fmt.Println("1.请输入pod过滤条件,仅允许传递一个条件")
		os.Exit(0)
	}
	clientset = clientsetCreate()

}

func clientsetCreate() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/kubernetes/admin.conf")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return clientset
}

func main() {
	var condition string
	condition = os.Args[1]

	// 获取集群中所有pod
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// 过滤出符合条件的pod

	var yespod []corev1.Pod

	for _, pod := range pods.Items {
		// 符合条件的pod加入到新的pod列表中
		if strings.Contains(pod.Name, condition) {
			yespod = append(yespod, pod)
		}
	}

	switch len(yespod) {
	case 0:
		fmt.Println("未找到符合条件的pod")

	default:
		fmt.Println("符合条件的pod和pod中对应的容器名如下:")
		var podcontainerlist []PodContainer
		for _, pod := range yespod {
			for _, container := range pod.Spec.Containers {
				PC := PodContainer{
					Namespace: pod.Namespace,
					Name:      pod.Name,
					Container: container.Name,
				}
				podcontainerlist = append(podcontainerlist, PC)
			}

		}
		for index, PC := range podcontainerlist {
			fmt.Printf("%d:  %s  %s  %s\n", index, PC.Namespace, PC.Name, PC.Container)
		}

		var deletenum string

		fmt.Printf("请输入需要查看日志的容器对应编号:")

		reader := bufio.NewReader(os.Stdin)
		deletenum, _ = reader.ReadString('\n')
		deletenum = strings.TrimSuffix(deletenum, "\n")
		// 传入的pod编号格式不正确转换失败
		numint, err := strconv.Atoi(deletenum)
		if err != nil {
			fmt.Println("传入的pod编号格式不为int类型或者传入格式存在特殊字符")
			os.Exit(1)
		}
		if numint >= len(podcontainerlist) || numint < 0 {
			fmt.Println("传入的pod编号大于符合条件的pod数量或者小于0")
			os.Exit(1)
		}
		logprint(podcontainerlist[numint])

	}

}

func logprint(PC PodContainer) {
	opts := &corev1.PodLogOptions{
		Follow:    true, // 对应kubectl logs -f参数
		Container: PC.Container,
	}
	request := clientset.CoreV1().Pods(PC.Namespace).GetLogs(PC.Name, opts)

	readCloser, err := request.Stream(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
	defer readCloser.Close()
	r := bufio.NewReader(readCloser)

	for {
		bytes, err := r.ReadBytes('\n')
		fmt.Println(strings.TrimSuffix(string(bytes), "\n"))
		if err != nil {
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

		}
	}
}
