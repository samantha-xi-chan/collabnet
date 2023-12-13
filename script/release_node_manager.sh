
set -e

Ver=v1.9-dev-R04
BuildT=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GitCommit=$(git rev-parse --short HEAD)
echo "\$Ver:    "     $Ver        ;
echo "\$BuildT: "     $BuildT     ;
echo "\$GitCommit: "  $GitCommit  ;

mkdir -p node_manager/config
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o node_manager/node_manager  -ldflags "-X main.Version=$Ver -X main.BuildTime=$BuildT -X main.GitCommit=$GitCommit"   cmd/version.go cmd/node_manager.go ;
cp config/app.yaml  node_manager/config
cp deploy/service/node_manager.service node_manager

tar -cvf release/node_manager_$(date "+%Y%m%d%H%M%S").tar node_manager/

exit 0

HOST=36110_root
ssh $HOST "mkdir -p /opt/"
scp node_manager.tar $HOST:/opt/
ssh $HOST "cd /opt/; tar -xvf node_manager.tar"
ssh $HOST "systemctl stop  node_manager.service || echo " stop service end "
ssh $HOST "cp /opt/node_manager/node_manager.service /etc/systemd/system/"
ssh $HOST "ls /opt/node_manager/"
ssh $HOST "systemctl enable node_manager.service"
ssh $HOST "systemctl restart  node_manager.service"
ssh $HOST "systemctl status  node_manager.service"
ssh $HOST "journalctl -u node_manager -f  -n 100"
