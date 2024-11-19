package pkg1

//go:generate mockgen -source=service.go -destination=mock_service.go
type Service interface {
	Start() error
	Stop() error
}
