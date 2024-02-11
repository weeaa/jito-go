package searcher_client

import (
	"context"
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/weeaa/jito-go/proto"
	"google.golang.org/grpc/metadata"
	"time"
)

// AuthenticateAndRefresh is a function that authenticates the client and refreshes the access token.
func (c *Client) AuthenticateAndRefresh(ctx context.Context, role proto.Role) error {
	var cancel context.CancelFunc
	c.GrpcCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	respChallenge, err := c.AuthenticationService.GenerateAuthChallenge(c.GrpcCtx, &proto.GenerateAuthChallengeRequest{Role: role, Pubkey: c.Keypair.PublicKey.Bytes()})
	if err != nil {
		return err
	}

	challenge := fmt.Sprintf("%s-%s", c.Keypair.PublicKey.String(), respChallenge.GetChallenge())

	sig, err := c.generateSignature([]byte(challenge))
	if err != nil {
		return err
	}

	respToken, err := c.AuthenticationService.GenerateAuthTokens(c.GrpcCtx, &proto.GenerateAuthTokensRequest{
		Challenge:       challenge,
		SignedChallenge: sig,
		ClientPubkey:    c.Keypair.PublicKey.Bytes(),
	})
	if err != nil {
		return err
	}

	c.updateAuthorizationMetadata(respToken.AccessToken)
	c.Auth.ErrChan = make(chan error, 1)

	go func() {
		defer close(c.Auth.ErrChan)
		for {
			select {
			case <-c.GrpcCtx.Done():
				c.Auth.ErrChan <- c.GrpcCtx.Err()
			default:
				var resp *proto.RefreshAccessTokenResponse
				resp, err = c.AuthenticationService.RefreshAccessToken(c.GrpcCtx, &proto.RefreshAccessTokenRequest{
					RefreshToken: respToken.RefreshToken.Value,
				})
				if err != nil {
					c.logger.Error().Err(err).Msg("failed to refresh access token")
					continue
				}

				c.updateAuthorizationMetadata(resp.AccessToken)
				time.Sleep(time.Until(resp.AccessToken.ExpiresAtUtc.AsTime()) - 30*time.Second)
			}
		}
	}()

	return nil
}

func (c *Client) updateAuthorizationMetadata(token *proto.Token) {
	c.Auth.mu.Lock()
	defer c.Auth.mu.Unlock()

	c.GrpcCtx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token.Value))
	c.Auth.BearerToken = token.Value
	c.Auth.ExpiresAt = token.ExpiresAtUtc.Seconds
}

func (c *Client) generateSignature(challenge []byte) ([]byte, error) {
	sig, err := c.Keypair.PrivateKey.Sign(challenge)
	if err != nil {
		return nil, err
	}

	return base58.Decode(sig.String())
}
