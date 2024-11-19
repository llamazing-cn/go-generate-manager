package pkg2

//go:generate mockgen -source=repository.go -destination=mock_repository.go
type Repository interface {
	Find(id string) (interface{}, error)
	Save(data interface{}) error
}
