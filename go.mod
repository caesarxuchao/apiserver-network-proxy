module sigs.k8s.io/apiserver-network-proxy

go 1.12

require (
	github.com/beorn7/perks v1.0.0 // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/common v0.4.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190507164030-5867b95ac084 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/sys v0.0.0-20190225065934-cc5685c2db12 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.27.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/klog v1.0.0
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.0
)

replace sigs.k8s.io/apiserver-network-proxy/konnectivity-client => ./konnectivity-client
