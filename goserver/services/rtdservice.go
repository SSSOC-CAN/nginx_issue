package services

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"github.com/SSSOC-CAN/nginx_issue/goserver/rpc"
	"google.golang.org/grpc"
)

type RTDService struct {
	rpc.UnimplementedRpcServiceServer
	Running 			int32
	Listeners 			int32
	DataProviderChan	chan *rpc.RealTimeData
	StateChangeChans	map[string]chan *StateChangeMsg
	ServiceBroadStates	map[string]bool
	name				string
}

func NewRTDService() *RTDService {
	return &RTDService{
		DataProviderChan: make(chan *rpc.RealTimeData),
		ServiceBroadStates: make(map[string]bool),
		StateChangeChans: make(map[string]chan *StateChangeMsg),
		name: "RTD",
	}
}

func (r *RTDService) Start() error {
	if ok := atomic.CompareAndSwapInt32(&r.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start RTD service: service already started")
	}
	return nil
}

func (r *RTDService) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&r.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop RTD service: service already stopped")
	}
	for _, channel := range r.StateChangeChans {
		close(channel)
	}
	return nil
}

func (r *RTDService) Name() string {
	return r.name
}

func (r *RTDService) RegisterDataProvider(serviceName string) {
	if _, ok := r.StateChangeChans[serviceName]; !ok {
		r.StateChangeChans[serviceName] = make(chan *StateChangeMsg)
	}
	if _, ok := r.ServiceBroadStates[serviceName]; !ok {
		r.ServiceBroadStates[serviceName] = false
	}
}

func (r *RTDService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	rpc.RegisterRpcServiceServer(grpcServer, r)
	return nil
}

func (r *RTDService) SubscribeDataStream(req *rpc.SubscribeDataRequest, updateStream rpc.RpcService_SubscribeDataStreamServer) error {
	fmt.Println("Have a new data listener.")
	_ = atomic.AddInt32(&r.Listeners, 1)
	for name, channel := range r.StateChangeChans {
		if !r.ServiceBroadStates[name] {
			channel <- &StateChangeMsg{Type: BROADCASTING, State: true, ErrMsg: nil}
			resp := <- channel
			if resp.ErrMsg != nil {
				fmt.Printf("Could not change broadcast state: %v\n", resp.ErrMsg)
				return resp.ErrMsg
			}
			r.ServiceBroadStates[name] = true
		}
	}
	for {
		select {
		case RTD := <-r.DataProviderChan:
			fmt.Println("WE GOT: ", RTD)
			if err := updateStream.Send(RTD); err != nil {
				_ = atomic.AddInt32(&r.Listeners, -1)
				if atomic.LoadInt32(&r.Listeners) == 0 {
					for name, channel := range r.StateChangeChans {
						if r.ServiceBroadStates[name] {
							channel <- &StateChangeMsg{Type: BROADCASTING, State: false, ErrMsg: nil}
							resp := <- channel
							if resp.ErrMsg != nil {
								fmt.Printf("Could not change broadcast state: %v\n", resp.ErrMsg)
								return resp.ErrMsg
							}
						}
						r.ServiceBroadStates[name] = false
					}
				}
				return err
			}
		case <-updateStream.Context().Done():
			_ = atomic.AddInt32(&r.Listeners, -1)
			if atomic.LoadInt32(&r.Listeners) == 0 {
				for name, channel := range r.StateChangeChans {
					if r.ServiceBroadStates[name] {
						channel <- &StateChangeMsg{Type: BROADCASTING, State: false, ErrMsg: nil}
						resp := <- channel
						if resp.ErrMsg != nil {
							fmt.Printf("Could not change broadcast state: %v\n", resp.ErrMsg)
							return resp.ErrMsg
						}
					}
					r.ServiceBroadStates[name] = false
				}
			}
			if errors.Is(updateStream.Context().Err(), context.Canceled) {
				fmt.Println("Listener stopped listening to stream")
				return nil
			}
			return updateStream.Context().Err()
		}
	}
}