package user

//go:generate mockgen -destination mock_user/mock_user.go github.com/win5do/golang-microservice-demo/docs/sample/gomock/user Index
type Index interface {
	Get(key string) interface{}
}

func GetIndex(in Index, key string) interface{} {
	// business logic
	r := in.Get(key)
	// handle output
	return r
}
