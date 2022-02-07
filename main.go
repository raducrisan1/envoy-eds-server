package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func main() {
	httpServer := NewHttpServer()
	router := gin.Default()
	router.GET("/api/:key", func(c *gin.Context) {
		key := c.Param("key")
		if ret, ok := httpServer.Get(key); ok {
			c.JSON(http.StatusOK, ret)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "prometheus target not found"})
		}
	})
	router.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, httpServer.List())
	})

	router.POST("/api", func(c *gin.Context) {
		var target KeyedEdsTarget
		if err := c.ShouldBindJSON(&target); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := httpServer.Post(target.Key, target)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": fmt.Sprintf("Error: %v", err)})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"result": "ok"})
	})

	router.DELETE("/api/:key", func(c *gin.Context) {
		key := c.Param("key")
		err := httpServer.Delete(key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"result": "key does not exist"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"result": "ok"})
	})

	stringHttpPort := os.Getenv("HTTP_LISTEN_PORT")
	intHttpPort, err := strconv.Atoi(stringHttpPort)
	if err != nil {
		log.Fatal("please provide a valid environment variable HTTP_LISTEN_PORT")
		return
	}
	nodeID := os.Getenv("NODE_ID")
	stringGrpcPort := os.Getenv("GRPC_LISTEN_PORT")
	intGrpcPort, err := strconv.Atoi(stringGrpcPort)
	if err != nil {
		log.Fatal("please provide a valid environment variable GRPC_LISTEN_PORT")
		return
	}
	stringEvictionTimeout := os.Getenv("EVICTION_TIMEOUT_IN_SEC")
	var intEvictionTimeout int
	if stringEvictionTimeout == "" {
		intEvictionTimeout = 42
	} else {
		intEvictionTimeout, err = strconv.Atoi(stringEvictionTimeout)
		if err != nil {
			log.Fatal("please provide a valid numeric value of seconds for environment variable EVICTION_TIMEOUT_IN_SEC")
			return
		}
	}
	httpServer.EvictionTimeout = intEvictionTimeout
	edsResource := EdsResource{
		ClusterName:     "cluster",
		WebServer:       httpServer,
		NodeId:          nodeID,
		SnapshotVersion: 1,
	}

	stopChan := make(chan int)
	completionChan := make(chan int, 2)
	evictionTicker := time.NewTicker(time.Second)

	var edsServer server.Server
	customEdsServer := CustomEdsServer{}

	go func() {
		log.Printf("EDS Server is listening for incoming HTTP requests on port %d", intHttpPort)
		endless.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", intHttpPort), router)
		customEdsServer.Shutdown()
		stopChan <- 1
	}()

	go func() {
		l := Logger{}
		datacache := cache.NewSnapshotCache(false, cache.IDHash{}, l)
		snapshot := edsResource.GenerateSnapshot()
		httpServer.DataCache = &datacache
		httpServer.Eds = &edsResource
		ctx := context.Background()
		if err := datacache.SetSnapshot(ctx, nodeID, snapshot); err != nil {
			l.Errorf("snapshot error %q for %+v", err, snapshot)
			os.Exit(1)
		}
		cb := &test.Callbacks{
			Fetches:  0,
			Requests: 0,
		}
		edsServer = server.NewServer(ctx, datacache, cb)
		customEdsServer.Initialize()
		customEdsServer.RunGrpcServer(ctx, edsServer, uint(intGrpcPort))
		completionChan <- 1
	}()

	go func() {
		exit := false
		for !exit {
			select {
			case <-evictionTicker.C:
				if intEvictionTimeout > 0 {
					httpServer.EvictHeartbeatTimeout()
				}
			case <-stopChan:
				exit = true
				evictionTicker.Stop()
			}
		}
		completionChan <- 1
	}()

	<-completionChan
	<-completionChan
}
