docker login -u haidongchen rd.clouditera.com -p 1Haidongchen
VERSION="v2.0-dev-latest" ; echo "\$VERSION: "$VERSION ;

docker build -t rd.clouditera.com/infra/golang_bizcache:1.18.10 . -f ./DockerfileCache
docker push     rd.clouditera.com/infra/golang_bizcache:1.18.10
