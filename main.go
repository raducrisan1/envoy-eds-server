package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
	edsResource := EdsResource{
		ListenerPort: uint32(intGrpcPort),
		RouteName:    "local_route",
		ListenerName: "listener_0",
		ClusterName:  "hpc_cluster",
		WebServer:    httpServer,
		NodeId:       nodeID,
	}

	stop := make(chan int)

	go func() {
		endless.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", intHttpPort), router)
	}()

	go func() {
		l := Logger{}
		datacache := cache.NewSnapshotCache(false, cache.IDHash{}, l)
		snapshot := edsResource.GenerateSnapshot()
		if err := snapshot.Consistent(); err != nil {
			l.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
			os.Exit(1)
		}
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
		grpcServer := server.NewServer(ctx, datacache, cb)
		RunGrpcServer(ctx, grpcServer, uint(intGrpcPort))
		stop <- 1
	}()
	<-stop
}
