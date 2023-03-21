package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var err error

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

}

func main() {
	var condition string
	condition = os.Args[1]
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

	case 1:
		err = clientset.CoreV1().Pods(yespod[0].Namespace).Delete(context.TODO(), yespod[0].Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("已删除pod  %s/%s\n", yespod[0].Namespace, yespod[0].Name)
	default:
		fmt.Printf("符合条件的pod存在 %d pods in the cluster\n", len(yespod))
		for index, pod := range yespod {
			fmt.Printf("%d:  %s   %s\n", index, pod.Namespace, pod.Name)
		}
		var deletenum string

		fmt.Printf("请输入需要删除的pod编号, 删除多个pod以英文逗号(,)进行间隔, 删除所有请输入all:")

		reader := bufio.NewReader(os.Stdin)
		deletenum, _ = reader.ReadString('\n')
		deletenum = strings.TrimSuffix(deletenum, "\n")

		if deletenum == "all" {
			for index, pod := range yespod {
				err = clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err.Error())

				}
				fmt.Printf("已删除pod  %s/%s\n", yespod[index].Namespace, yespod[index].Name)
				time.Sleep(2 * time.Second)
			}
		} else {
			numlist := strings.Split(deletenum, ",")
			var numintlist []int
			for _, num := range numlist {
				// 传入的pod编号格式不正确转换失败
				numint, err := strconv.Atoi(num)
				if err != nil {
					fmt.Println("传入的pod编号格式不为int类型或者传入格式存在特殊字符")
					os.Exit(1)
				}
				if numint >= len(yespod) || numint < 0 {
					fmt.Println("传入的pod编号大于符合条件的pod数量或者小于0")
					os.Exit(1)
				}

				numintlist = append(numintlist, numint)
			}
			for _, numint := range numintlist {
				err = clientset.CoreV1().Pods(yespod[numint].Namespace).Delete(context.TODO(), yespod[numint].Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err.Error())

				} else {
					fmt.Printf("已删除pod  %s/%s\n", yespod[numint].Namespace, yespod[numint].Name)
				}
				time.Sleep(2 * time.Second)
			}

		}

	}

}
