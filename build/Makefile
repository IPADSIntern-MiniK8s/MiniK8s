CMDPATH=../cmd
OUTPATH=./bin
kubectl:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go

kubelet:
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go

apiserver:
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go

scheduler:
	go build -o $(OUTPATH)/scheduler $(CMDPATH)/scheduler.go

controller:
	go build -o $(OUTPATH)/controller $(CMDPATH)/controller.go

kubeproxy:
	go build -o $(OUTPATH)/kubeproxy $(CMDPATH)/kubeproxy.go

serverless:
	go build -o $(OUTPATH)/serverless $(CMDPATH)/serverless.go

all:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go
	go build -o $(OUTPATH)/scheduler $(CMDPATH)/scheduler.go
	go build -o $(OUTPATH)/controller $(CMDPATH)/controller.go
	go build -o $(OUTPATH)/kubeproxy $(CMDPATH)/kubeproxy.go
	go build -o $(OUTPATH)/serverless $(CMDPATH)/serverless.go

clean:
	rm $(OUTPATH)/*
