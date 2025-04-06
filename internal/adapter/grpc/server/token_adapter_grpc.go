package server

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rs/zerolog/log"

	"github.com/go-tokenization-grpc/internal/core/service"
	"github.com/go-tokenization-grpc/internal/core/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/metadata"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	token_proto_service "github.com/go-tokenization-grpc/protogen/token"
	proto "github.com/go-tokenization-grpc/protogen/token"
	//proto "github.com/eliezerraj/go-grpc-proto/protogen/token"

	go_core_observ "github.com/eliezerraj/go-core/observability"
)

var childLogger = log.With().Str("component","go-tokenization-grpc").Str("package","internal.adapter.grpc.server").Logger()
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

// About get card from token
func (a *AdapterGrpc) GetCardToken(ctx context.Context, cardTokenRequest *proto.CardTokenRequest) (*proto.ListCardTokenResponse, error) {
	childLogger.Info().Str("func","GetCardToken").Interface("cardTokenRequest", cardTokenRequest).Send()

	// Trace
	span := tracerProvider.Span(ctx, "adpater.grpc.GetCardToken")
	defer span.End()

	// get request-id
	header, _ := metadata.FromIncomingContext(ctx)
	if len(header.Get("trace-request-id")) > 0 {
		ctx = context.WithValue(ctx, "trace-request-id", header.Get("trace-request-id")[0])
	}

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
										AccountId: v.AccountId,
										Type: 		v.Type,
										Model: 		v.Model,
										Status: 	v.Status,
										TokenData: 	v.TokenData,
										Atc:		uint32(v.Atc),
										CreatedAt: 	timestamppb.New(v.CreatedAt),
										ExpiredAt: 	timestamppb.New(v.ExpiredAt), }
		res_list_card_proto = append(res_list_card_proto, &res_card_proto)
	}

	res_list_card_proto_reponse := &proto.ListCardTokenResponse{Cards: res_list_card_proto}

	return res_list_card_proto_reponse, nil
}
