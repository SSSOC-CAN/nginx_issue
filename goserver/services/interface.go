package services

type Service interface {
	Start() error
	Stop() error
	Name() string
}

type StateType int32

const (
	BROADCASTING StateType = 0
	RECORDING StateType = 1
)

type StateChangeMsg struct {
	Type	StateType
	State 	bool
	ErrMsg	error
}