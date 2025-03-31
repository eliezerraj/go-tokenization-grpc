package grpc

import (
	"fmt"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"

	"github.com/go-tokenization-grpc/internal/core/service"
	"github.com/go-tokenization-grpc/internal/core/model"
	"github.com/go-tokenization-grpc/internal/core/erro"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	token_proto_service "github.com/go-tokenization-grpc/protogen/token"
	proto "github.com/go-tokenization-grpc/protogen/token"
	//proto "github.com/eliezerraj/go-grpc-proto/protogen/token"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.adapter.grpc").Logger()
var tracerProvider go_core_observ.TracerProvider

type AdapterGrpc struct{
	appServer 		*model.AppServer
	workerService 	*service.WorkerService
	token_proto_service.UnimplementedTokenServiceServer
}

// About create new adapter
func NewAdapterGrpc(appServer *model.AppServer, workerService *service.WorkerService) *AdapterGrpc {
	childLogger.Info().Str("func","NewAdapterGrpc").Send()

	return &AdapterGrpc{
		appServer: appServer,
		workerService: workerService,
	}
}

// About get pod data
func (a *AdapterGrpc) GetPod(ctx context.Context, podRequest *proto.PodRequest) (*proto.PodResponse, error) {
	childLogger.Info().Str("func","GetPodInfo").Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.GetPod")
	defer span.End()

	pod := proto.Pod{	IpAddress: 	a.appServer.InfoPod.IPAddress,
						PodName: a.appServer.InfoPod.PodName,
						AvailabilityZone: a.appServer.InfoPod.AvailabilityZone,
						Host: a.appServer.Server.Port,
						Version: a.appServer.InfoPod.ApiVersion,
					}

	res_pod := &proto.PodResponse {
		Pod: &pod,
	}
	
	return res_pod, nil
}

// About create token from data
func (a *AdapterGrpc) CreateCardToken(ctx context.Context, cardTokenRequest *proto.CardTokenRequest) (*proto.CardTokenResponse, error) {
	childLogger.Info().Str("func","CreateCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("cardTokenRequest", cardTokenRequest).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.CreateCardToken")
	defer span.End()

	card := model.Card{	ID: int(cardTokenRequest.Card.Id),
						CardNumber: cardTokenRequest.Card.CardNumber,
						Status: cardTokenRequest.Card.Status,
						}

	res_card, err := a.workerService.CreateCardToken(ctx, card)
	if (err != nil) {
		s := status.New(codes.InvalidArgument, err.Error())
		s, _ = s.WithDetails(&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "ID:CardNumber",
					Description: "ID or CardNumber informad is invalid",
				},
			},
		})
		return nil, s.Err()
	}	

	res_card_proto := proto.Card{ 	Id: uint32(res_card.ID),
									CardNumber: res_card.CardNumber,
							 		TokenData: res_card.TokenData,
									CreatedAt: timestamppb.New(res_card.CreatedAt),
									ExpiredAt: timestamppb.New(res_card.ExpiredAt), }

	card_proto_reponse := &proto.CardTokenResponse { Card: &res_card_proto }
	
	return card_proto_reponse, nil
}

// About get card from token
func (a *AdapterGrpc) GetCardToken(ctx context.Context, cardTokenRequest *proto.CardTokenRequest) (*proto.ListCardTokenResponse, error) {
	childLogger.Info().Str("func","GetCardToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("cardTokenRequest", cardTokenRequest).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.GetCardToken")
	defer span.End()

	// Prepare
	card := model.Card{	TokenData: cardTokenRequest.Card.TokenData }

	// Call service
	res_list_card, err := a.workerService.GetCardToken(ctx, card)
	if (err != nil) {
		s := status.New(codes.Internal, err.Error())
		s, _ = s.WithDetails(&errdetails.ErrorInfo{
			Domain: "database",
			Reason: "database unreacheable",
		})
		return nil, s.Err()
	}	

	// Prepare the proto response
	res_list_card_proto := []*proto.Card{}
	for _, v := range *res_list_card {
		res_card_proto := proto.Card{ 	Id: uint32(v.ID),
										CardNumber: v.CardNumber,
										Status: v.Status,
										TokenData: v.TokenData,
										CreatedAt: timestamppb.New(v.CreatedAt),
										ExpiredAt: timestamppb.New(v.ExpiredAt), }
		res_list_card_proto = append(res_list_card_proto, &res_card_proto)
	}

	res_list_card_proto_reponse := &proto.ListCardTokenResponse{Cards: res_list_card_proto}

	return res_list_card_proto_reponse, nil
}

// About get card from token
func (a *AdapterGrpc) AddPaymentToken(ctx context.Context, paymentRequest *proto.PaymentTokenRequest) (*proto.PaymentTokenResponse, error) {
	childLogger.Info().Str("func","AddPaymentToken").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("paymentRequest", paymentRequest).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.AddPaymentToken")
	defer span.End()

	// Prepare
	payment := model.Payment{ TokenData: paymentRequest.Payment.TokenData,
							  Terminal: paymentRequest.Payment.Terminal,	
							  Currency: paymentRequest.Payment.Currency,
							  Amount: paymentRequest.Payment.Amount,
							  CardType: paymentRequest.Payment.CardType,
							  Mcc: paymentRequest.Payment.Mcc,								
							}

	// Call service
	res_payment, err := a.workerService.AddPaymentToken(ctx, payment)
	if (err != nil) {
		switch err {
		case erro.ErrCardTypeInvalid:
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "cart.type",
						Description: fmt.Sprintf("card type (%v) informed not valid", paymentRequest.Payment.CardType),
					},
				},
			})
			return nil, s.Err()		
		case erro.ErrNotFound:
			s := status.New(codes.InvalidArgument, err.Error())
			s, _ = s.WithDetails(&errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "token/terminal",
						Description: fmt.Sprintf("token (%v) or terminal (%v) informed not found", paymentRequest.Payment.TokenData, paymentRequest.Payment.Terminal),
					},
				},
			})
			return nil, s.Err()
		default:
			s := status.New(codes.Internal, err.Error())
			s, _ = s.WithDetails(&errdetails.ErrorInfo{
				Domain: "service",
				Reason: "service payment unreacheable",
			})
			return nil, s.Err()
		}
	}	

	res_payment_proto_response := &proto.PaymentTokenResponse {
		Payment: &proto.Payment{	TokenData: res_payment.TokenData,
									CardType:  res_payment.CardType,
									Status:  res_payment.Status,
									Currency:  res_payment.Currency,
									Amount:  res_payment.Amount,
									CardModel:  res_payment.CardModel,
									Mcc: res_payment.Mcc,
									Terminal: res_payment.Terminal,
									PaymentAt: timestamppb.New(res_payment.PaymentAt),
							},	
	}

	return res_payment_proto_response, nil
}
