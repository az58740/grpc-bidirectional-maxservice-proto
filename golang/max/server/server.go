package server

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	pb "github.com/az58740/grpc-bidirectional-maxservice-proto/golang/max"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	pb.UnimplementedMaxServiceServer
}

func NewApp() *App {
	return &App{}
}
func (a *App) start() {
	serverport := 50053
	serverAdd := fmt.Sprintf("0.0.0.0:%d", serverport)
	fmt.Println("starting max gRPC server at ", serverAdd)

	lis, err := net.Listen("tcp", serverAdd)
	if err != nil {
		log.Fatalf("failed to listen:%v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	pb.RegisterMaxServiceServer(s, a)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to strart server : %v", err)
	}
}
func (a *App) FindMax(stream pb.MaxService_FindMaxServer) error {
	max := int32(math.MinInt32)
	responsechan := make(chan int32, 1)
	go func() {
		for {
			time.Sleep(2000 * time.Millisecond)
			err := stream.Send(&pb.MaxResponse{Max: max})
			if err != nil {
				log.Fatalf("Error while sending to client: %v", err)
				return
			}
		}
	}()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
			return err
		}
		num := req.GetNumber()
		if max < num {
			max = num
		}
		// Send the updated max to the goroutine through the channel
		responsechan <- max
	}
}
