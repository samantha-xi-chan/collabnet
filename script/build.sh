set -e

Ver=2.0-dev
BuildT=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GitCommit=$(git rev-parse --short HEAD)
echo "\$Ver:    "     $Ver        ;
echo "\$BuildT: "     $BuildT     ;
echo "\$GitCommit: "  $GitCommit  ;

CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/server        -ldflags "-X main.Version=$Ver -X main.BuildTime=$BuildT -X main.GitCommit=$GitCommit"   cmd/version.go cmd/server.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/node_manager  -ldflags "-X main.Version=$Ver -X main.BuildTime=$BuildT -X main.GitCommit=$GitCommit"   cmd/version.go cmd/node_manager.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/plugin        -ldflags "-X main.Version=$Ver -X main.BuildTime=$BuildT -X main.GitCommit=$GitCommit"   cmd/version.go cmd/plugin.go ;
