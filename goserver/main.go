package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
	"github.com/SSSOC-CAN/nginx_issue/goserver/services"
	"google.golang.org/grpc"
)

var (
	grpcPort = 3567
)

func main() {
	//Instantiate grpc Server
	var opts []grpc.ServerOption
	var servers []services.Service
	grpcServer := grpc.NewServer(opts...)

	// Instantiate RTD service and register with gRPC server
	rtdService := services.NewRTDService()
	rtdService.RegisterWithGrpcServer(grpcServer)
	servers = append(servers, rtdService)

	// Instantiate services A and B
	serviceA := services.NewA()
	serviceA.RegisterWithRTDService(rtdService)
	servers = append(servers, serviceA)

	serviceB := services.NewB()
	serviceB.RegisterWithRTDService(rtdService)
	servers = append(servers, serviceB)

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(grpcPort))
	if err != nil {
		fmt.Printf("Could not open tcp listener on port %v: %v\n", grpcPort, err)
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		_ = grpcServer.Serve(listener)
	}()
	wg.Wait()
	fmt.Printf("gRPC listening on port %v\n", grpcPort)
	defer grpcServer.Stop()

	// start services
	for _, s := range servers {
		err = s.Start()
		if err != nil {
			fmt.Printf("Unable to start %s service: %v\n", s.Name(), err)
			return
		}
		defer s.Stop()
	}

	// Start recording and wait 5 minutes
	err = serviceA.StartRecording()
	if err != nil {
		fmt.Printf("Unablet to start recording: %v\n", err)
		return
	}
	err = serviceB.StartRecording()
	if err != nil {
		fmt.Printf("Unablet to start recording: %v\n", err)
		return
	}
	time.Sleep(5*time.Minute)
}