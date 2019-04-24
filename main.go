package main

import (
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"k8s-informer-example/k8s"
	"log"
)

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	k8sHelper := k8s.InitK8sHelper()
	if k8sHelper == nil {
		log.Fatal("init k8s helper fail ...")
	}
	// ping 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 获取 pod key 列表
	r.GET("/pod", func(c *gin.Context) {
		podKeys := k8sHelper.PodStore.ListKeys()
		logger.Infof("found pod key list: %v", podKeys)

		c.JSON(200, gin.H{"podKeys": podKeys})
	})
	// 获取 pod 详情信息
	r.GET("/pod/:namespace/:podKey", func(c *gin.Context) {
		// 通过 cache 查找调度器的 pod
		namespace := c.Param("namespace")
		podKey := c.Param("podKey")
		podItem, exists, err := k8sHelper.PodStore.GetByKey(namespace + "/" + podKey)
		if exists && err == nil {
			logger.Infof("found the pod [%v] in cache", podItem)
		}
		c.JSON(200, gin.H{"podDetail": podItem})
	})
	// listen and serve on 0.0.0.0:8080
	log.Fatal(r.Run(":8080"))
}
