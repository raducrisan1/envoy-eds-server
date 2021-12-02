package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func main() {
	server := NewServer()
	router := gin.Default()
	router.GET("/api/:key", func(c *gin.Context) {
		key := c.Param("key")
		if ret, ok := server.Get(key); ok {
			c.JSON(http.StatusOK, ret)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "prometheus target not found"})
		}
	})
	router.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, server.List())
	})

	router.POST("/api", func(c *gin.Context) {
		var target KeyedEdsTarget
		if err := c.ShouldBindJSON(&target); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := server.Post(target.Key, target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "invalid key"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"result": "ok"})
	})

	router.DELETE("/api/:key", func(c *gin.Context) {
		key := c.Param("key")
		err := server.Delete(key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "key does not exist"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"result": "ok"})
	})

	stringHttpPort := os.Getenv("HTTP_LISTEN_PORT")
	//stringGrpcPort := os.Getenv("GRPC_LISTEN_PORT")

	intHttpPort, err := strconv.Atoi(stringHttpPort)
	if err != nil {
		log.Fatal("please provide a valid environment variable HTTP_LISTEN_PORT")
		return
	}

	// intGrpcPort, err := strconv.Atoi(stringGrpcPort)
	// if err != nil {
	// 	log.Fatal("please provide a valid environment variable GRPC_LISTEN_PORT")
	// 	return
	// }
	// stop := make(chan int)
	// ctx := context.Background()
	// cb := &Callbacks{
	// 	Fetches:  0,
	// 	Requests: 0,
	// }
	// cache := cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)
	// grpcServer := serverv3.NewServer(ctx, cache, cb)

	go func() {
		endless.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", intHttpPort), router)
	}()

	// go func() {

	// 	server := serverv3.NewServer(ctx, )
	// 	RunManagementServer(ctx, )
	// 	stop <- 1
	// }()
	// <-stop
}
