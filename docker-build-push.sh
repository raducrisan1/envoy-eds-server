export IMAGE_TAG=1.0.0
docker stop $(docker ps -aqf "name=envoy-eds-server")
docker rm $(docker ps -aqf "name=envoy-eds-server")
docker build --no-cache -t envoy-eds-server .
docker tag envoy-eds-server envoy-eds-server:$IMAGE_TAG
docker tag envoy-eds-server raducrisan/envoy-eds-server:$IMAGE_TAG
docker tag envoy-eds-server raducrisan/envoy-eds-server:latest
docker push raducrisan/envoy-eds-server:$IMAGE_TAG
docker push raducrisan/envoy-eds-server:latest
