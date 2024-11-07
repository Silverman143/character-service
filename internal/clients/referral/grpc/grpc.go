package referralgrpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Silverman143/character-service/internal/config"
	referralv1 "github.com/Silverman143/protos_chadnaldo/gen/go/referral"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api referralv1.ReferralClient
	log *slog.Logger
}

func New (ctx context.Context, log *slog.Logger, cfg *config.Client) (*Client, error){
	const op = "clients.user.grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(cfg.RetriesCount)),
		grpcretry.WithPerRetryTimeout(cfg.Timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	con, err := grpc.DialContext(	ctx, 
									cfg.Addr, 
									grpc.WithTransportCredentials(	insecure.NewCredentials()), 
									grpc.WithChainUnaryInterceptor(	grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
									grpcretry.UnaryClientInterceptor(retryOpts...)),
								)

	if err != nil{
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	return &Client{
		api: referralv1.NewReferralClient(con),
		log: log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
    return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
        l.Log(ctx, slog.Level(lvl), msg, fields...)
    })
}

func (c *Client) GetReferralsAmount(ctx context.Context, userID int64) (int, error) {
    const op = "clients.user.grpc.GetCoinsAmount"
    
    resp, err := c.api.GetReferralsCount(ctx, &referralv1.GetReferralsCountRequest{UserId: userID})
    if err != nil {
        return 0, fmt.Errorf("%s: failed to get user balance: %w", op, err)
    }
    
    return int(resp.Count), nil
}
