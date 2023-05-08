CMDPATH=../cmd
OUTPATH=./$(CI_COMMIT_BRANCH)/bin
kubectl:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go
kubelet:
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go

apiserver:
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go


all:
	go build -o $(OUTPATH)/kubectl $(CMDPATH)/kubectl.go
	go build -o $(OUTPATH)/kubelet $(CMDPATH)/kubelet.go
	go build -o $(OUTPATH)/apiserver $(CMDPATH)/apiserver.go
clean:
	rm $(OUTPATH)/*
