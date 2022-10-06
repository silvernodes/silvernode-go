go mod init github.com/silvernodes/silvernode-go

go mod edit -replace github.com/coreos/bbolt=go.etcd.io/bbolt@v1.3.4

go mod edit -replace google.golang.org/grpc=google.golang.org/grpc@v1.26.0
 
go mod tidy