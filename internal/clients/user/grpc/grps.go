package usergrpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Silverman143/character-service/internal/config"
	userv1 "github.com/Silverman143/protos_chadnaldo/gen/go/user"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api userv1.UserClient
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
		api: userv1.NewUserClient(con),
		log: log,
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
    return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
        l.Log(ctx, slog.Level(lvl), msg, fields...)
    })
}

func (c *Client) GetCoinsAmount(ctx context.Context, userID int64) (int64, error) {
    const op = "clients.user.grpc.GetCoinsAmount"
    
    resp, err := c.api.GetUserCoinsBalance(ctx, &userv1.GetUserCoinsBalanceRequest{UserId: userID})
    if err != nil {
        return 0, fmt.Errorf("%s: failed to get user balance: %w", op, err)
    }
    
    return resp.Coins, nil
}

func (c *Client) InitiatePayment(ctx context.Context, userID int64, amount int64, paymentID string) error {
    const op = "clients.user.grpc.InitiatePayment"
    
    _, err := c.api.InitiatePayment(ctx, &userv1.InitiatePaymentRequest{UserId: userID, Amount: amount, TransactionId: paymentID})
    if err != nil {
        return fmt.Errorf("%s: failed ti init payment: %w", op, err)
    }
    
    return nil
}

func (c *Client) FinalizePayment(ctx context.Context, paymentID string, success bool) error{
    const op = "clients.user.grpc.FinalizePaymentfunc"
    
    _, err := c.api.FinalizePayment(ctx, &userv1.FinalizePaymentRequest{PaymentId: paymentID, Complete: success})
    if err != nil {
        return fmt.Errorf("%s: failed ti init payment: %w", op, err)
    }
    
    return nil
}