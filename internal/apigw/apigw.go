package apigw

import (
	"fmt"

	anrpc "github.com/morzhanov/go-realworld/api/grpc/analytics"
	prpc "github.com/morzhanov/go-realworld/api/grpc/pictures"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	authrpc "github.com/morzhanov/go-realworld/api/grpc/auth"
	"github.com/morzhanov/go-realworld/internal/common/sender"
	"github.com/opentracing/opentracing-go"
)

type apigw struct {
	sender sender.Sender
}

type APIGW interface {
	CheckAuth(ctx *gin.Context, transport sender.Transport, apiName string, key string, span *opentracing.Span) (res *authrpc.ValidationResponse, err error)
	Login(transport sender.Transport, input *authrpc.LoginInput, span *opentracing.Span) (res *authrpc.AuthResponse, err error)
	Signup(transport sender.Transport, input *authrpc.SignupInput, span *opentracing.Span) (res *authrpc.AuthResponse, err error)
	GetPictures(transport sender.Transport, userId string, span *opentracing.Span) (res *prpc.PicturesMessage, err error)
	GetPicture(transport sender.Transport, userId string, pictureId string, span *opentracing.Span) (res *prpc.PictureMessage, err error)
	CreatePicture(transport sender.Transport, input *prpc.CreateUserPictureRequest, span *opentracing.Span) (res *prpc.PictureMessage, err error)
	DeletePicture(transport sender.Transport, userId string, pictureId string, span *opentracing.Span) error
	GetAnalytics(transport sender.Transport, input *anrpc.LogDataRequest, span *opentracing.Span) (res *anrpc.AnalyticsEntryMessage, err error)
}

func (s *apigw) getAccessToken(ctx *gin.Context) (res string, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:getAccessToken") }()
	authorization := ctx.GetHeader("Authorization")
	if authorization == "" {
		return "", fmt.Errorf("not authorized")
	}
	return authorization[7:], nil
}

func createMetaWithUserId(userId string) sender.RequestMeta {
	return sender.RequestMeta{"urlparams": sender.UrlParams{
		"userId": userId,
	}}
}

func (s *apigw) CheckAuth(
	ctx *gin.Context,
	transport sender.Transport,
	apiName string,
	key string,
	span *opentracing.Span,
) (res *authrpc.ValidationResponse, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:CheckAuth") }()
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	var input interface{}
	var method string
	switch transport {
	case sender.RestTransport:
		api, err := s.sender.GetAPI().GetApiItem(apiName)
		if err != nil {
			return nil, err
		}
		input = &authrpc.ValidateRestRequestInput{
			Path:        api.Rest[key].Url,
			AccessToken: accessToken,
		}
		method = "validateRestRequest"
	case sender.RpcTransport:
		input = &authrpc.ValidateRpcRequestInput{
			Method:      key,
			AccessToken: accessToken,
		}
		method = "validateRpcRequest"
	case sender.EventsTransport:
		_, err := s.sender.GetAPI().GetApiItem(apiName)
		if err != nil {
			return nil, err
		}
		input = &authrpc.ValidateEventsRequestInput{
			AccessToken: accessToken,
		}
		method = "validateEventsRequest"
	default:
		return nil, fmt.Errorf("not valid transport %v", transport)
	}
	if err = s.sender.PerformRequest(transport, "auth", method, input, s.eventListener, span, nil, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *apigw) Login(
	transport sender.Transport,
	input *authrpc.LoginInput,
	span *opentracing.Span,
) (res *authrpc.AuthResponse, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:Login") }()
	res = &authrpc.AuthResponse{}
	if err := s.sender.PerformRequest(transport, "auth", "login", input, s.eventListener, span, nil, res); err != nil {
		return nil, err
	}
	return
}

func (s *apigw) Signup(
	transport sender.Transport,
	input *authrpc.SignupInput,
	span *opentracing.Span,
) (res *authrpc.AuthResponse, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:Signup") }()
	res = &authrpc.AuthResponse{}
	if err = s.sender.PerformRequest(transport, "auth", "signup", input, s.eventListener, span, nil, res); err != nil {
		return nil, err
	}
	return
}

func (s *apigw) GetPictures(
	transport sender.Transport,
	userId string,
	span *opentracing.Span,
) (res *prpc.PicturesMessage, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:GetPictures") }()
	input := prpc.GetUserPicturesRequest{UserId: userId}
	res = &prpc.PicturesMessage{}
	meta := createMetaWithUserId(userId)
	if err := s.sender.PerformRequest(transport, "pictures", "getPictures", &input, s.eventListener, span, meta, &res); err != nil {
		return nil, err
	}
	return
}

func (s *apigw) GetPicture(
	transport sender.Transport,
	userId string,
	pictureId string,
	span *opentracing.Span,
) (res *prpc.PictureMessage, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:GetPicture") }()
	input := prpc.GetUserPictureRequest{
		UserId:    userId,
		PictureId: pictureId,
	}
	meta := createMetaWithUserId(userId)
	urlparams := meta["urlparams"].(sender.UrlParams)
	urlparams["id"] = pictureId
	res = &prpc.PictureMessage{}
	if err := s.sender.PerformRequest(transport, "pictures", "getPicture", &input, s.eventListener, span, meta, &res); err != nil {
		return nil, err
	}
	return
}

func (s *apigw) CreatePicture(
	transport sender.Transport,
	input *prpc.CreateUserPictureRequest,
	span *opentracing.Span,
) (res *prpc.PictureMessage, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:CreatePicture") }()
	res = &prpc.PictureMessage{}
	meta := createMetaWithUserId(input.UserId)
	if err := s.sender.PerformRequest(transport, "pictures", "createPicture", input, s.eventListener, span, meta, &res); err != nil {
		return nil, err
	}
	return
}

func (s *apigw) DeletePicture(
	transport sender.Transport,
	userId string,
	pictureId string,
	span *opentracing.Span,
) error {
	input := prpc.DeleteUserPictureRequest{
		UserId:    userId,
		PictureId: pictureId,
	}
	meta := createMetaWithUserId(userId)
	return s.sender.PerformRequest(transport, "pictures", "deletePicture", &input, s.eventListener, span, meta, nil)
}

func (s *apigw) GetAnalytics(
	transport sender.Transport,
	input *anrpc.LogDataRequest,
	span *opentracing.Span,
) (res *anrpc.AnalyticsEntryMessage, err error) {
	defer func() { err = errors.Wrap(err, "apigwService:GetAnalytics") }()
	res = &anrpc.AnalyticsEntryMessage{}
	if err := s.sender.PerformRequest(transport, "pictures", "deletePicture", &input, s.eventListener, span, nil, &res); err != nil {
		return nil, err
	}
	return
}

func NewAPIGW() APIGW {
	return &apigw{s, el}
}
