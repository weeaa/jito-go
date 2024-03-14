package pkg

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"github.com/rs/zerolog"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"sync"
	"time"
)

type AuthenticationService struct {
	AuthService proto.AuthServiceClient
	GrpcCtx     context.Context
	KeyPair     *Keypair
	BearerToken string
	ExpiresAt   int64 // seconds
	ErrChan     chan error
	mu          sync.Mutex
	logger      zerolog.Logger
}

func NewAuthenticationService(grpcConn *grpc.ClientConn, privateKey solana.PrivateKey) *AuthenticationService {
	return &AuthenticationService{
		GrpcCtx:     context.TODO(),
		AuthService: proto.NewAuthServiceClient(grpcConn),
		KeyPair:     NewKeyPair(privateKey),
		ErrChan:     make(chan error, 1),
		mu:          sync.Mutex{},
		logger:      zerolog.New(os.Stdout).With().Timestamp().Str("service", "auth").Logger(),
	}
}

// AuthenticateAndRefresh is a function that authenticates the client and refreshes the access token.
func (as *AuthenticationService) AuthenticateAndRefresh(role proto.Role) error {
	respChallenge, err := as.AuthService.GenerateAuthChallenge(as.GrpcCtx, &proto.GenerateAuthChallengeRequest{Role: role, Pubkey: as.KeyPair.PublicKey.Bytes()})
	if err != nil {
		return err
	}

	challenge := fmt.Sprintf("%s-%s", as.KeyPair.PublicKey.String(), respChallenge.GetChallenge())

	sig, err := as.generateSignature([]byte(challenge))
	if err != nil {
		return err
	}

	respToken, err := as.AuthService.GenerateAuthTokens(as.GrpcCtx, &proto.GenerateAuthTokensRequest{
		Challenge:       challenge,
		SignedChallenge: sig,
		ClientPubkey:    as.KeyPair.PublicKey.Bytes(),
	})
	if err != nil {
		return err
	}

	as.updateAuthorizationMetadata(respToken.AccessToken)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				as.logger.Error().Msgf("recovered from panic %v", r)
				if err = as.AuthenticateAndRefresh(role); err != nil {
					as.logger.Warn().Msgf("unable to re-authenticate after panic, stopping...")
				}
			}
		}()
		for {
			select {
			case <-as.GrpcCtx.Done():
				as.ErrChan <- as.GrpcCtx.Err()
			default:
				var resp *proto.RefreshAccessTokenResponse
				resp, err = as.AuthService.RefreshAccessToken(as.GrpcCtx, &proto.RefreshAccessTokenRequest{
					RefreshToken: respToken.RefreshToken.Value,
				})
				if err != nil {
					as.logger.Error().Err(err).Msg("failed to refresh access token")
					continue
				}

				as.updateAuthorizationMetadata(resp.AccessToken)
				time.Sleep(time.Until(resp.AccessToken.ExpiresAtUtc.AsTime()) - 30*time.Second)
			}
		}
	}()

	return nil
}

// updateAuthorizationMetadata updates headers of the gRPC connection.
func (as *AuthenticationService) updateAuthorizationMetadata(token *proto.Token) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.GrpcCtx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token.Value))
	as.BearerToken = token.Value
	as.ExpiresAt = token.ExpiresAtUtc.Seconds
}

func (as *AuthenticationService) generateSignature(challenge []byte) ([]byte, error) {
	sig, err := as.KeyPair.PrivateKey.Sign(challenge)
	if err != nil {
		return nil, err
	}

	return base58.Decode(sig.String())
}
