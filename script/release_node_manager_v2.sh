
set -e

HOST=36107_root
echo $HOST && sleep 2

Ver=v2.0
BuildT=$(date "+%Y%m%d%H%M%S")
GitCommit=$(git rev-parse --short HEAD)
echo "\$Ver:    "     $Ver        ;
echo "\$BuildT: "     $BuildT     ;
echo "\$GitCommit: "  $GitCommit  ;

mkdir -p node_manager/config
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o node_manager/node_manager  -ldflags "-X main.Version=$Ver -X main.BuildTime=$BuildT -X main.GitCommit=$GitCommit"   cmd/version.go cmd/node_manager.go ;
md5sum node_manager/node_manager > node_manager/node_manager.md5sum
cp config/app.yaml  node_manager/config
cp deploy/service/node_manager.service node_manager

Tar=node_manager_$BuildT.tar
echo $Tar
tar -cvf release/$Tar node_manager/

#exit 0

ssh $HOST "mkdir -p /opt/"
scp release/$Tar $HOST:/opt/node_manager.tar

cat << 'EOF' > deploy.sh
set -e
echo "DefaultLimitNOFILE=1048576" >> /etc/systemd/system.conf
systemctl stop  node_manager.service || echo 'stop service end' && \
cd /opt/ && tar -xvf node_manager.tar && \
cp /opt/node_manager/node_manager.service /etc/systemd/system/ && \
systemctl enable node_manager.service && \
systemctl restart  node_manager.service && \
systemctl status  node_manager.service
# journalctl -u node_manager -f  -n 100
EOF
scp deploy.sh $HOST:/root/
ssh $HOST "sh /root/deploy.sh"

exit 0