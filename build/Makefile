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

all:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go
	go build -o $(OUTPATH)/scheduler $(CMDPATH)/scheduler.go

clean:
	rm $(OUTPATH)/*
