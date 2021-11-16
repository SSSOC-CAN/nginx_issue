package services

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
	"github.com/SSSOC-CAN/nginx_issue/goserver/rpc"
)

type B struct {
	Running 		int32
	Recording 		int32
	Broadcasting 	int32
	QuitChan		chan struct{}
	CancelChan		chan struct{}
	StateChangeChan	chan *StateChangeMsg
	OutputChan		chan *rpc.RealTimeData
	name			string
}

func NewB() *B {
	return &B{
		QuitChan: make(chan struct{}),
		CancelChan: make(chan struct{}),
		name: "ServiceB",
	}
}

func (b *B) Start() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start B: already started")
	}
	go b.ListenForSignals()
	return nil
}

func (b *B) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop B: already stopped")
	}
	close(b.CancelChan)
	close(b.QuitChan)
	return nil
}

func (b *B) Name() string {
	return b.name
}

func (b *B) StartRecording() error {
	if ok := atomic.CompareAndSwapInt32(&b.Recording, 0, 1); !ok {
		return fmt.Errorf("Could not start recording. Data recording already started")
	}
	ticker := time.NewTicker(time.Duration(15)*time.Second)
	fmt.Println("Starting data recording...")
	go func() {
		for {
			select {
			case <-ticker.C:
				err := b.record()
				if err != nil {
					fmt.Printf("Could not record data: %v\n", err)
				}
			case <-b.CancelChan:
				ticker.Stop()
				fmt.Println("Data recording stopped.")
				return
			case <-b.QuitChan:
				ticker.Stop()
				fmt.Println("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

func (b *B) record() error {
	dataField := make(map[int64]*rpc.DataField)
	for i := 0; i < 200; i++ {
		if atomic.LoadInt32(&b.Broadcasting) == 1 {
			dataField[int64(i)]= &rpc.DataField{
				Name: fmt.Sprintf("Value %v", i+1),
				Value: (rand.Float64()*5)+20,
			}
		}
		// normally there's handling for writing to a file but I'm excluding this part for simplicity
	}
	if atomic.LoadInt32(&b.Broadcasting) == 1 {
		dataFrame := &rpc.RealTimeData {
			Source: b.name,
			IsScanning: true,
			Timestamp: time.Now().UnixMilli(),
			Data: dataField,
		}
		b.OutputChan <- dataFrame
	}
	return nil
}

func (b *B) StopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&b.Recording, 1, 0); !ok {
		return fmt.Errorf("Could not stop recording. Data recording already stopped")
	}
	b.CancelChan <- struct{}{}
	return nil
}

func (b *B) ListenForSignals() {
	for {
		select {
		case msg := <-b.StateChangeChan:
			switch msg.Type {
			case BROADCASTING:
				if msg.State {
					if ok := atomic.CompareAndSwapInt32(&b.Broadcasting, 0, 1); !ok {
						b.StateChangeChan <- &StateChangeMsg{Type: BROADCASTING, State: false, ErrMsg: fmt.Errorf("Could not start broadcasting: already broadcasting")}
					} else {
						b.StateChangeChan <- &StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&b.Broadcasting, 1, 0); !ok {
						b.StateChangeChan <- &StateChangeMsg{Type: BROADCASTING, State: true, ErrMsg: fmt.Errorf("Could not stop broadcasting: already stopped")}
					} else {
						b.StateChangeChan <- &StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
			// case RECORDING:
			// 	if msg.State {
			// 		err := b.startRecording()
			// 		if err != nil {
			// 			fmt.Printf("Could not start record: %v\n", err)
			// 			b.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
			// 		} else {
			// 			fmt.Println("Started recording")
			// 			b.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil}
			// 		}
			// 	} else {
			// 		err := s.stopRecording()
			// 		if err != nil {
			// 			fmt.Printf("Could not stop recording: %v\n", err)
			// 			b.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: fmt.Errorf("Could not stop recording: %v", err)}
			// 		} else {
			// 			fmt.Println("Stopped recording.")
			// 			b.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: false, ErrMsg: nil}
			// 		}
			// 	}
			}
		case <-b.QuitChan:
			return
		}
	}
}

func (b *B) RegisterWithRTDService(rtd *RTDService) {
	rtd.RegisterDataProvider(b.name)
	b.OutputChan = rtd.DataProviderChan
	b.StateChangeChan = rtd.StateChangeChans[b.name]
}