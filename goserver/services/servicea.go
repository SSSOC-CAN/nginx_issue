package services

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
	"github.com/SSSOC-CAN/nginx_issue/goserver/rpc"
)

type A struct {
	Running 		int32
	Recording 		int32
	Broadcasting 	int32
	QuitChan		chan struct{}
	CancelChan		chan struct{}
	StateChangeChan	chan *StateChangeMsg
	OutputChan		chan *rpc.RealTimeData
	name			string
}

func NewA() *A {
	return &A{
		QuitChan: make(chan struct{}),
		CancelChan: make(chan struct{}),
		name: "ServiceA",
	}
}

func (a *A) Start() error {
	if ok := atomic.CompareAndSwapInt32(&a.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start A: already started")
	}
	go a.ListenForSignals()
	return nil
}

func (a *A) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&a.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop A: already stopped")
	}
	close(a.CancelChan)
	close(a.QuitChan)
	return nil
}

func (a *A) Name() string {
	return a.name
}

func (a *A) StartRecording() error {
	if ok := atomic.CompareAndSwapInt32(&a.Recording, 0, 1); !ok {
		return fmt.Errorf("Could not start recording. Data recording already started")
	}
	ticker := time.NewTicker(time.Duration(10)*time.Second)
	fmt.Println("Starting data recording...")
	go func() {
		for {
			select {
			case <-ticker.C:
				err := a.record()
				if err != nil {
					fmt.Printf("Could not record data: %v\n", err)
				}
			case <-a.CancelChan:
				ticker.Stop()
				fmt.Println("Data recording stopped.")
				return
			case <-a.QuitChan:
				ticker.Stop()
				fmt.Println("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

func (a *A) record() error {
	dataField := make(map[int64]*rpc.DataField)
	for i := 0; i < 96; i++ {
		if atomic.LoadInt32(&a.Broadcasting) == 1 {
			dataField[int64(i)]= &rpc.DataField{
				Name: fmt.Sprintf("Value %v", i+1),
				Value: (rand.Float64()*5)+20,
			}
		}
		// normally there's handling for writing to a file but I'm excluding this part for simplicity
	}
	if atomic.LoadInt32(&a.Broadcasting) == 1 {
		dataFrame := &rpc.RealTimeData {
			Source: a.name,
			IsScanning: true,
			Timestamp: time.Now().UnixMilli(),
			Data: dataField,
		}
		a.OutputChan <- dataFrame
	}
	return nil
}

func (a *A) StopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&a.Recording, 1, 0); !ok {
		return fmt.Errorf("Could not stop recording. Data recording already stopped")
	}
	a.CancelChan <- struct{}{}
	return nil
}

func (a *A) ListenForSignals() {
	for {
		select {
		case msg := <-a.StateChangeChan:
			switch msg.Type {
			case BROADCASTING:
				if msg.State {
					if ok := atomic.CompareAndSwapInt32(&a.Broadcasting, 0, 1); !ok {
						a.StateChangeChan <- &StateChangeMsg{Type: BROADCASTING, State: false, ErrMsg: fmt.Errorf("Could not start broadcasting: already broadcasting")}
					} else {
						a.StateChangeChan <- &StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&a.Broadcasting, 1, 0); !ok {
						a.StateChangeChan <- &StateChangeMsg{Type: BROADCASTING, State: true, ErrMsg: fmt.Errorf("Could not stop broadcasting: already stopped")}
					} else {
						a.StateChangeChan <- &StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
			// case RECORDING:
			// 	if msg.State {
			// 		err := a.startRecording()
			// 		if err != nil {
			// 			fmt.Printf("Could not start record: %v\n", err)
			// 			a.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
			// 		} else {
			// 			fmt.Println("Started recording")
			// 			a.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: nil}
			// 		}
			// 	} else {
			// 		err := s.stopRecording()
			// 		if err != nil {
			// 			fmt.Printf("Could not stop recording: %v\n", err)
			// 			a.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: true, ErrMsg: fmt.Errorf("Could not stop recording: %v", err)}
			// 		} else {
			// 			fmt.Println("Stopped recording.")
			// 			a.StateChangeChan <- &StateChangeMsg{Type: RECORDING, State: false, ErrMsg: nil}
			// 		}
			// 	}
			}
		case <-a.QuitChan:
			return
		}
	}
}

func (a *A) RegisterWithRTDService(rtd *RTDService) {
	rtd.RegisterDataProvider(a.name)
	a.OutputChan = rtd.DataProviderChan
	a.StateChangeChan = rtd.StateChangeChans[a.name]
}