package client

import (
	"context"
	"flag"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	pb "github.com/az58740/grpc-bidirectional-maxservice-proto/golang/max"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	numbers string
)

func init() {
	flag.StringVar(&numbers, "numbers", "", "provide comma separated numbers to compute max")
}

type App struct {
	clientconn *grpc.ClientConn
}

func NewApp() *App {
	return &App{}
}
func (a *App) start() {
	var err error
	flag.Parse()

	servAddr := "localhost:50053"
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	a.clientconn, err = grpc.NewClient(servAddr, opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	maxServiceClient := pb.NewMaxServiceClient(a.clientconn)
	numberArr := make([]int, 0)

	strNumbers := strings.Split(strings.TrimSpace(numbers), ",")
	for _, strNum := range strNumbers {
		intNum, _ := strconv.Atoi(strings.TrimSpace(strNum))
		numberArr = append(numberArr, intNum)
	}
	doBidirectionalStreaming(maxServiceClient, numberArr)
}

func doBidirectionalStreaming(c pb.MaxServiceClient, numberArr []int) {
	requests := make([]*pb.MaxRequest, 0)
	for _, n := range numberArr {
		requests = append(requests, &pb.MaxRequest{Number: int32(n)})
	}

	stream, err := c.FindMax(context.Background())
	if err != nil {
		log.Fatalf("error while invoking FindMax rpc: %v", err)
	}

	waitchan := make(chan struct{})
	go func(reqs []*pb.MaxRequest) {
		for _, r := range requests {
			log.Printf("sending message: %v\n", r)
			stream.Send(r)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}(requests)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving from FindMax rpc response: %v", err)
				break
			}
			log.Printf("response received: %v\n", res.GetMax())
		}
		close(waitchan)
	}()
	<-waitchan
}
