
 HOST="3266_root"
 ssh $HOST "systemctl list-units --type=service --state=running"
 ssh $HOST "rm -rf /opt/node/; mkdir -p /opt/node/; "
 scp deploy/node_manager.service $HOST:/etc/systemd/system/
 ssh $HOST "systemctl stop  node_manager.service"
 scp release/node_manager  $HOST:/opt/node/;
 scp -r config  $HOST:/opt/node/;
 ssh $HOST "systemctl enable node_manager.service"
 ssh $HOST "systemctl start  node_manager.service"
 ssh $HOST "systemctl status  node_manager.service"
 ssh $HOST "journalctl -u node_manager -f  -n 100"
