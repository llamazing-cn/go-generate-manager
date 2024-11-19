package simple

//go:generate mockgen -source=simple.go -destination=mock_simple.go
type Simple interface {
	DoSomething() error
	GetValue() string
}

//go:generate protoc --go_out=. simple.proto
type protoMessage struct {
	value string
}
