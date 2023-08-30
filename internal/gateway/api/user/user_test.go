package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"toktik/internal/gateway/pkg/apiutil"
	"toktik/internal/user/kitex_gen/user"
	mock_userservice "toktik/pkg/test/mock/user"
)

func newMockUserClient(t *testing.T) *mock_userservice.MockClient {
	ctl := gomock.NewController(t)
	return mock_userservice.NewMockClient(ctl)
}

func TestUserAPI_Login(t *testing.T) {
	t.Run("rpc error", func(t *testing.T) {
		mockUserClient := newMockUserClient(t)
		api := &UserAPI{
			userClient: mockUserClient,
		}

		mockUserClient.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc error")).AnyTimes()

		ctx := app.NewContext(16)
		ctx.Request.SetQueryString("username=test_user&password=123456")
		api.Login(context.Background(), ctx)

		assert.Equal(t, http.StatusInternalServerError, ctx.Response.StatusCode())
		payload := LoginResp{}
		assert.NoError(t, json.Unmarshal(ctx.Response.Body(), &payload))
		assert.Equal(t, apiutil.StatusFailed, payload.StatusCode)
		assert.Equal(t, "rpc error", payload.StatusMsg)
	})

	t.Run("login failed", func(t *testing.T) {
		mockUserClient := newMockUserClient(t)
		api := &UserAPI{
			userClient: mockUserClient,
		}

		mockUserClient.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&user.LoginRes{
			Status: user.Status_ERROR,
			ErrMsg: "password unmatch",
		}, nil).AnyTimes()

		ctx := app.NewContext(16)
		ctx.Request.SetQueryString("username=test_user&password=123456")
		api.Login(context.Background(), ctx)

		assert.Equal(t, http.StatusBadRequest, ctx.Response.StatusCode())
		payload := LoginResp{}
		assert.NoError(t, json.Unmarshal(ctx.Response.Body(), &payload))
		assert.Equal(t, apiutil.StatusFailed, payload.StatusCode)
		assert.Equal(t, "password unmatch", payload.StatusMsg)
	})

	t.Run("login success", func(t *testing.T) {
		mockUserClient := newMockUserClient(t)
		api := &UserAPI{
			userClient: mockUserClient,
		}

		mockUserClient.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&user.LoginRes{
			Status: user.Status_OK,
			UserId: 10,
		}, nil).AnyTimes()

		ctx := app.NewContext(16)
		ctx.Request.SetQueryString("username=test_user&password=123456")
		api.Login(context.Background(), ctx)

		assert.Equal(t, http.StatusOK, ctx.Response.StatusCode())
		payload := LoginResp{}
		assert.NoError(t, json.Unmarshal(ctx.Response.Body(), &payload))
		assert.Equal(t, apiutil.StatusOK, payload.StatusCode)
		assert.Equal(t, int64(10), payload.UserId)
		assert.NotEmpty(t, payload.Token)
	})
}