
current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"

docker login -u haidongchen rd.clouditera.com -p 1Haidongchen
VERSION="v2.0-dev-latest" ; echo "\$VERSION: "$VERSION ;

docker build -t rd.clouditera.com/infra/collabnet:$VERSION .
docker push     rd.clouditera.com/infra/collabnet:$VERSION
