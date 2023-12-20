
set -e

Ver=v1.9-dev-R04
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

HOST=36107_root
ssh $HOST "mkdir -p /opt/"
scp release/$Tar $HOST:/opt/node_manager.tar

cat << 'EOF' > deploy.sh
set -e
systemctl stop  node_manager.service || echo 'stop service end' && \
cd /opt/ && tar -xvf node_manager.tar && \
cp /opt/node_manager/node_manager.service /etc/systemd/system/ && \
systemctl enable node_manager.service && \
systemctl restart  node_manager.service && \
systemctl status  node_manager.service
EOF
scp deploy.sh $HOST:/root/
ssh $HOST "sh /root/deploy.sh"

exit 0

ssh $HOST "cd /opt/; tar -xvf node_manager.tar"
ssh $HOST "systemctl stop  node_manager.service || echo 'stop service end' "
ssh $HOST "cp /opt/node_manager/node_manager.service /etc/systemd/system/"
ssh $HOST "ls /opt/node_manager/"
ssh $HOST "systemctl enable   node_manager.service"
ssh $HOST "systemctl restart  node_manager.service"
ssh $HOST "systemctl status   node_manager.service"
#ssh $HOST "journalctl -u node_manager -f  -n 100"

# 前置条件 (注意 config/app.yaml 文件中的IP必须是准确的 K8S地址 )
# root@node-013:/opt/node_manager# ls -alh
# total 25M
# drwxr-xr-x 3 root root 4.0K Dec 19 06:16 .
# drwxr-xr-x 4 root root 4.0K Dec 19 06:16 ..
# drwxr-xr-x 2 root root 4.0K Dec 19 06:16 config
# -rwxr-xr-x 1 root root  25M Dec 19 06:16 node_manager
# -rw-r--r-- 1 root root   60 Dec 19 06:16 node_manager.md5sum
# -rw-r--r-- 1 root root  272 Dec 19 06:16 node_manager.service

# 系统优化与服务部署
echo "DefaultLimitNOFILE=1048576" >> /etc/systemd/system.conf
systemctl stop  node_manager.service || echo 'stop service end'
cp /opt/node_manager/node_manager.service /etc/systemd/system/
ls /opt/node_manager/
systemctl enable   node_manager.service
systemctl restart  node_manager.service
systemctl status   node_manager.service

# 查看最近日志
journalctl -u node_manager -f  -n 100

