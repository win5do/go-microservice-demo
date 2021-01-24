package user

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/win5do/golang-microservice-demo/docs/sample/gomock/user/mock_user"
)

func TestGetIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	index := mock_user.NewMockIndex(ctrl)
	in := "uid"
	out := "username"

	index.EXPECT().Get(in).Return(out)

	r := GetIndex(index, "uid")
	require.Equal(t, out, r)
}
