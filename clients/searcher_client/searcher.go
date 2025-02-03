package searcher_client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/weeaa/jito-go/pb"
	"github.com/weeaa/jito-go/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// New creates a new Searcher Client instance.
func New(
	ctx context.Context,
	blockEngineURL string,
	jitoRpcClient, rpcClient *rpc.Client,
	privateKey solana.PrivateKey,
	tlsConfig *tls.Config,
	opts ...grpc.DialOption,
) (*Client, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	chErr := make(chan error)
	conn, err := pkg.CreateAndObserveGRPCConn(ctx, chErr, blockEngineURL, opts...)
	if err != nil {
		return nil, err
	}

	searcherService := jito_pb.NewSearcherServiceClient(conn)
	authService := pkg.NewAuthenticationService(conn, privateKey)
	if err = authService.AuthenticateAndRefresh(jito_pb.Role_SEARCHER); err != nil {
		return nil, err
	}

	subBundleRes, err := searcherService.SubscribeBundleResults(authService.GrpcCtx, &jito_pb.SubscribeBundleResultsRequest{})
	if err != nil {
		return nil, err
	}

	client := Client{
		GrpcConn:                 conn,
		RpcConn:                  rpcClient,
		JitoRpcConn:              jitoRpcClient,
		SearcherService:          searcherService,
		BundleStreamSubscription: subBundleRes,
		Auth:                     authService,
		ErrChan:                  chErr,
	}

	return &client, nil
}

// NewNoAuth initializes and returns a new instance of the Searcher Client which does not require private key signing.
// Proxy feature allows you to have different Jito clients running on the same machine without hitting rate limits due to IP limits.
func NewNoAuth(ctx context.Context,
	blockEngineURL string,
	jitoRpcClient, rpcClient *rpc.Client,
	proxyURL string,
	tlsConfig *tls.Config,
	opts ...grpc.DialOption,
) (*Client, error) {
	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	if proxyURL != "" {
		dialer, err := createContextDialer(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy dialer: %w", err)
		}

		opts = append(opts,
			grpc.WithContextDialer(dialer),
			defaultKeepAlive,
		)
	}

	chErr := make(chan error)
	conn, err := pkg.CreateAndObserveGRPCConn(ctx, chErr, blockEngineURL, opts...)
	if err != nil {
		return nil, err
	}

	searcherService := jito_pb.NewSearcherServiceClient(conn)
	subBundleRes, err := searcherService.SubscribeBundleResults(ctx, &jito_pb.SubscribeBundleResultsRequest{})
	if err != nil {
		return nil, err
	}

	client := Client{
		GrpcConn:                 conn,
		RpcConn:                  rpcClient,
		JitoRpcConn:              jitoRpcClient,
		SearcherService:          searcherService,
		BundleStreamSubscription: subBundleRes,
		Auth:                     &pkg.AuthenticationService{GrpcCtx: ctx},
		ErrChan:                  chErr,
	}

	return &client, nil
}

// RotateProxy updates the client's gRPC connection to use a new proxy URL. This allows dynamic rotation of proxies to avoid rate limits.
func RotateProxy(client *Client, proxyURL string) error {
	blockEngineURL := client.GrpcConn.Target()

	dialer, err := createContextDialer(proxyURL)
	if err != nil {
		return fmt.Errorf("failed to create new proxy dialer: %w", err)
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))

	opts = append(opts,
		grpc.WithContextDialer(dialer),
		defaultKeepAlive,
	)

	if err := client.GrpcConn.Close(); err != nil {
		return fmt.Errorf("failed to close existing connection: %w", err)
	}

	chErr := make(chan error)
	ctx := client.Auth.GrpcCtx
	conn, err := pkg.CreateAndObserveGRPCConn(ctx, chErr, blockEngineURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to create new connection: %w", err)
	}

	searcherService := jito_pb.NewSearcherServiceClient(conn)
	subBundleRes, err := searcherService.SubscribeBundleResults(ctx, &jito_pb.SubscribeBundleResultsRequest{})
	if err != nil {
		return fmt.Errorf("failed to resubscribe to bundle results: %w", err)
	}

	client.GrpcConn = conn
	client.SearcherService = searcherService
	client.BundleStreamSubscription = subBundleRes
	client.ErrChan = chErr

	return nil
}

func parseProxyString(proxyStr string) (host string, port string, username string, password string, err error) {
	parts := strings.Split(proxyStr, ":")
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid proxy format, expected IP:PORT:USERNAME:PASSWORD")
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

type proxyDialer struct {
	proxyHost string
	auth      string
	timeout   time.Duration
}

func newProxyDialer(proxyStr string) (*proxyDialer, error) {
	host, port, username, password, err := parseProxyString(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy string: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	return &proxyDialer{
		proxyHost: net.JoinHostPort(host, port),
		auth:      auth,
		timeout:   30 * time.Second,
	}, nil
}

func (d *proxyDialer) dialProxy(ctx context.Context, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error

	dialer := &net.Dialer{
		Timeout:   d.timeout,
		KeepAlive: 30 * time.Second,
	}

	conn, err = dialer.DialContext(ctx, "tcp", d.proxyHost)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy %s: %w", d.proxyHost, err)
	}

	connectReq := fmt.Sprintf(
		"CONNECT %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Proxy-Authorization: Basic %s\r\n"+
			"User-Agent: jito-go/1.1\r\n"+
			"\r\n",
		addr, addr, d.auth,
	)

	if _, err = conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to write CONNECT request: %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: "CONNECT"})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		conn.Close()
		return nil, fmt.Errorf("proxy connection failed: %s", resp.Status)
	}

	return conn, nil
}

func createContextDialer(proxyStr string) (func(context.Context, string) (net.Conn, error), error) {
	pd, err := newProxyDialer(proxyStr)
	if err != nil {
		return nil, err
	}

	return pd.dialProxy, nil
}

func (c *Client) Close() error {
	close(c.ErrChan)
	defer c.Auth.GrpcCtx.Done()

	if err := c.RpcConn.Close(); err != nil {
		return err
	}

	if err := c.JitoRpcConn.Close(); err != nil {
		return err
	}

	return c.GrpcConn.Close()
}

/*
// NewMempoolStreamAccount creates a new mempool subscription on specific Solana accounts.
func (c *Client) NewMempoolStreamAccount(accounts, regions []string) (jito_pb.SearcherService_SubscribeMempoolClient, error) {
	return c.SearcherService.SubscribeMempool(c.Auth.GrpcCtx, &jito_pb.MempoolSubscription{
		Msg: &jito_pb.MempoolSubscription_WlaV0Sub{
			WlaV0Sub: &jito_pb.WriteLockedAccountSubscriptionV0{
				Accounts: accounts,
			},
		},
		Regions: regions,
	})
}
*/

/*
// NewMempoolStreamProgram creates a new mempool subscription on specific Solana programs.
func (c *Client) NewMempoolStreamProgram(programs, regions []string) (jito_pb.SearcherService_SubscribeMempoolClient, error) {
	return c.SearcherService.SubscribeMempool(c.Auth.GrpcCtx, &jito_pb.MempoolSubscription{
		Msg: &jito_pb.MempoolSubscription_ProgramV0Sub{
			ProgramV0Sub: &jito_pb.ProgramSubscriptionV0{
				Programs: programs,
			},
		},
		Regions: regions,
	})
}
*/

/*
// SubscribeAccountsMempoolTransactions subscribes to the mempool transactions of the provided accounts.
func (c *Client) SubscribeAccountsMempoolTransactions(ctx context.Context, accounts, regions []string) (<-chan *solana.Transaction, <-chan error, error) {
	sub, err := c.NewMempoolStreamAccount(accounts, regions)
	if err != nil {
		return nil, nil, err
	}

	chTx := make(chan *solana.Transaction)
	chErr := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.Auth.GrpcCtx.Done():
				return
			default:
				receipt, err := sub.Recv()
				if err != nil {
					chErr <- fmt.Errorf("SubscribeAccountsMempoolTransactions: failed to receive mempool notification: %w", err)
					continue
				}

				for _, transaction := range receipt.Transactions {
					go func(transaction *jito_pb.Packet) {
						var tx *solana.Transaction
						tx, err = pkg.ConvertProtobufPacketToTransaction(transaction)
						if err != nil {
							chErr <- fmt.Errorf("SubscribeAccountsMempoolTransactions: failed to convert protobuf packet to transaction: %w", err)
							return
						}

						chTx <- tx
					}(transaction)
				}
			}
		}
	}()

	return chTx, chErr, nil
}
*/

/*
// SubscribeProgramsMempoolTransactions subscribes to the mempool transactions of the provided programs.
func (c *Client) SubscribeProgramsMempoolTransactions(ctx context.Context, programs, regions []string) (<-chan *solana.Transaction, <-chan error, error) {
	sub, err := c.NewMempoolStreamProgram(programs, regions)
	if err != nil {
		return nil, nil, err
	}

	chTx := make(chan *solana.Transaction)
	chErr := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.Auth.GrpcCtx.Done():
				return
			default:
				var receipt *jito_pb.PendingTxNotification
				receipt, err = sub.Recv()
				if err != nil {
					chErr <- fmt.Errorf("SubscribeProgramsMempoolTransactions: failed to receive mempool notification: %w", err)
					continue
				}

				for _, transaction := range receipt.Transactions {
					go func(transaction *jito_pb.Packet) {
						var tx *solana.Transaction
						tx, err = pkg.ConvertProtobufPacketToTransaction(transaction)
						if err != nil {
							chErr <- fmt.Errorf("SubscribeProgramsMempoolTransactions: failed to convert protobuf packet to transaction: %w", err)
							return
						}

						chTx <- tx
					}(transaction)
				}
			}
		}
	}()

	return chTx, chErr, nil
}
*/

func (c *Client) GetRegions(opts ...grpc.CallOption) (*jito_pb.GetRegionsResponse, error) {
	return c.SearcherService.GetRegions(c.Auth.GrpcCtx, &jito_pb.GetRegionsRequest{}, opts...)
}

func (c *Client) GetConnectedLeaders(opts ...grpc.CallOption) (*jito_pb.ConnectedLeadersResponse, error) {
	return c.SearcherService.GetConnectedLeaders(c.Auth.GrpcCtx, &jito_pb.ConnectedLeadersRequest{}, opts...)
}

func (c *Client) GetConnectedLeadersRegioned(regions []string, opts ...grpc.CallOption) (*jito_pb.ConnectedLeadersRegionedResponse, error) {
	return c.SearcherService.GetConnectedLeadersRegioned(c.Auth.GrpcCtx, &jito_pb.ConnectedLeadersRegionedRequest{Regions: regions}, opts...)
}

// GetTipAccounts returns Jito Tip Accounts.
func (c *Client) GetTipAccounts(opts ...grpc.CallOption) (*jito_pb.GetTipAccountsResponse, error) {
	return c.SearcherService.GetTipAccounts(c.Auth.GrpcCtx, &jito_pb.GetTipAccountsRequest{}, opts...)
}

// GetRandomTipAccount returns a random Jito TipAccount.
func (c *Client) GetRandomTipAccount(opts ...grpc.CallOption) (string, error) {
	resp, err := c.GetTipAccounts(opts...)
	if err != nil {
		return "", err
	}

	return resp.Accounts[rand.Intn(len(resp.Accounts))], nil
}

func (c *Client) GetNextScheduledLeader(regions []string, opts ...grpc.CallOption) (*jito_pb.NextScheduledLeaderResponse, error) {
	return c.SearcherService.GetNextScheduledLeader(c.Auth.GrpcCtx, &jito_pb.NextScheduledLeaderRequest{Regions: regions}, opts...)
}

// NewBundleSubscriptionResults creates a new bundle subscription stream, allowing to receive information about broadcasted bundles.
func (c *Client) NewBundleSubscriptionResults(opts ...grpc.CallOption) (jito_pb.SearcherService_SubscribeBundleResultsClient, error) {
	return c.SearcherService.SubscribeBundleResults(c.Auth.GrpcCtx, &jito_pb.SubscribeBundleResultsRequest{}, opts...)
}

// SendBundle sends a bundle of transaction(s) on chain through Jito.
func (c *Client) SendBundle(transactions []*solana.Transaction, opts ...grpc.CallOption) (*jito_pb.SendBundleResponse, error) {
	bundle, err := c.AssembleBundle(transactions)
	if err != nil {
		return nil, err
	}

	return c.SearcherService.SendBundle(c.Auth.GrpcCtx, &jito_pb.SendBundleRequest{Bundle: bundle}, opts...)
}

// SpamBundle spams SendBundle (spam being the amount of bundles sent). If async is true, it will use goroutines.
func (c *Client) SpamBundle(transactions []*solana.Transaction, spam int, async bool, opts ...grpc.CallOption) ([]*jito_pb.SendBundleResponse, []error) {
	bundles := make([]*jito_pb.SendBundleResponse, spam)
	errs := make([]error, spam)
	mu := sync.Mutex{}

	f := func() {
		bundle, err := c.SendBundle(transactions, opts...)
		if err != nil {
			errs = append(errs, err)
			return
		}
		mu.Lock()
		bundles = append(bundles, bundle)
		mu.Unlock()
	}
	for i := 0; i < spam; i++ {
		if async {
			go f()
		} else {
			f()
		}
	}
	return bundles, errs
}

type SendBundleResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      int    `json:"id"`
}

// SendBundle sends a bundle through Jito API.
func SendBundle(client *http.Client, encoding Encoding, transactions []*solana.Transaction) (*SendBundleResponse, error) {
	buf := new(bytes.Buffer)

	var txns []string
	var err error
	switch encoding {
	case Base58:
		txns, err = pkg.ConvertBachTransactionsToBase58(transactions)
		if err != nil {
			return nil, err
		}
		break
	case Base64:
		txns, err = pkg.ConvertBachTransactionsToBase64(transactions)
		if err != nil {
			return nil, err
		}
		break
	default:
		return nil, errors.New("unknown encoding, expected base64 or base58")
	}

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "sendBundle",
		"params":  [][]string{txns},
	}

	if encoding == Base64 {
		payload["encoding"] = encoding
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    jitoBundleURL,
		Body:   io.NopCloser(buf),
		Header: DefaultHeader.Clone(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BroadcastBundle error: unexpected response status %s", resp.Status)
	}

	var out SendBundleResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	return &out, err
}

// SendBundleWithConfirmation sends a bundle of transactions on chain through Jito BlockEngine and waits for its confirmation.
func SendBundleWithConfirmation(ctx context.Context, client *http.Client, rpcConn *rpc.Client, encoding Encoding, transactions []*solana.Transaction) (*SendBundleResponse, error) {
	bundle, err := SendBundle(client, encoding, transactions)
	if err != nil {
		return nil, err
	}

	bundleSignatures := pkg.BatchExtractSigFromTx(transactions)
	isRPCNil(rpcConn)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(3 * time.Second)

			bundleStatuses, err := GetInflightBundleStatuses(client, []string{bundle.Result})
			if err != nil {
				return bundle, err
			}

			if err = handleBundleResult(bundleStatuses, bundle.Result); err != nil {
				if err.Error() == "pending" {
					continue
				}
				return bundle, err
			}

			var statuses *rpc.GetSignatureStatusesResult
			var start = time.Now()

			for {
				statuses, err = rpcConn.GetSignatureStatuses(ctx, false, bundleSignatures...)
				if err != nil {
					return bundle, err
				}

				ready := true
				for _, status := range statuses.Value {
					if status == nil {
						ready = false
						break
					}
				}

				if ready {
					break
				}

				if time.Since(start) > 15*time.Second {
					return bundle, errors.New("operation timed out after 15 seconds")
				} else {
					time.Sleep(1 * time.Second)
				}
			}

			for _, status := range statuses.Value {
				if status.ConfirmationStatus != rpc.ConfirmationStatusProcessed && status.ConfirmationStatus != rpc.ConfirmationStatusConfirmed {
					return bundle, errors.New("searcher service did not provide bundle status in time")
				}
			}

			return bundle, nil
		}
	}
}

// SendBundleWithConfirmation sends a bundle of transactions on chain through Jito BlockEngine and waits for its confirmation.
func (c *Client) SendBundleWithConfirmation(ctx context.Context, transactions []*solana.Transaction, opts ...grpc.CallOption) (*jito_pb.SendBundleResponse, error) {
	bundle, err := c.SendBundle(transactions, opts...)
	if err != nil {
		return nil, err
	}

	bundleSignatures := pkg.BatchExtractSigFromTx(transactions)
	isRPCNil(c.RpcConn)

	for {
		select {
		case <-c.Auth.GrpcCtx.Done():
			return nil, c.Auth.GrpcCtx.Err()
		default:
			bundleResult, err := c.BundleStreamSubscription.Recv()
			if err != nil {
				return bundle, err
			}

			if err = handleBundleResult(bundleResult, ""); err != nil {
				return bundle, err
			}

			var statuses *rpc.GetSignatureStatusesResult
			var start = time.Now()

			for {
				statuses, err = c.RpcConn.GetSignatureStatuses(ctx, false, bundleSignatures...)
				if err != nil {
					return bundle, err
				}
				ready := true

				for _, status := range statuses.Value {
					if status == nil {
						ready = false
						break
					}
				}

				if ready {
					break
				}

				if time.Since(start) > 15*time.Second {
					return bundle, errors.New("operation timed out after 15 seconds")
				} else {
					time.Sleep(1 * time.Second)
				}
			}

			for _, status := range statuses.Value {
				if status.ConfirmationStatus != rpc.ConfirmationStatusProcessed && status.ConfirmationStatus != rpc.ConfirmationStatusConfirmed {
					return bundle, errors.New("searcher service did not provide bundle status in time")
				}
			}

			return bundle, nil
		}
	}
}

// bundleID arg is solely for JSON RPC API.
func handleBundleResult[T *GetInflightBundlesStatusesResponse | *jito_pb.BundleResult](t T, bundleID string) error {
	switch bundle := any(t).(type) {
	case *jito_pb.BundleResult:
		switch bundle.Result.(type) {
		case *jito_pb.BundleResult_Accepted:
			break
		case *jito_pb.BundleResult_Rejected:
			rejected := bundle.Result.(*jito_pb.BundleResult_Rejected)
			switch rejected.Rejected.Reason.(type) {
			case *jito_pb.Rejected_SimulationFailure:
				rejection := rejected.Rejected.GetSimulationFailure()
				return NewSimulationFailureError(rejection.TxSignature, rejection.GetMsg())
			case *jito_pb.Rejected_StateAuctionBidRejected:
				rejection := rejected.Rejected.GetStateAuctionBidRejected()
				return NewStateAuctionBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
			case *jito_pb.Rejected_WinningBatchBidRejected:
				rejection := rejected.Rejected.GetWinningBatchBidRejected()
				return NewWinningBatchBidRejectedError(rejection.AuctionId, rejection.SimulatedBidLamports)
			case *jito_pb.Rejected_InternalError:
				rejection := rejected.Rejected.GetInternalError()
				return NewInternalError(rejection.Msg)
			case *jito_pb.Rejected_DroppedBundle:
				rejection := rejected.Rejected.GetDroppedBundle()
				return NewDroppedBundle(rejection.Msg)
			default:
				return nil
			}
		}
	case *GetInflightBundlesStatusesResponse: // experimental, subject to changes
		for i, value := range bundle.Result.Value {
			if value.BundleId == bundleID {
				switch value.Status {
				case "Invalid":
					return fmt.Errorf("bundle %d is invalid: %s", i, bundleID)
				case "Pending":
					return errors.New("pending")
				case "Failed":
					return fmt.Errorf("bundle %d failed to land: %s", i, bundleID)
				case "Landed":
					return nil
				default:
					return fmt.Errorf("bundle %d unknown error: %s", i, bundleID)
				}
			}
		}
	}
	return nil
}

// SimulateBundle is an RPC method that simulates a Jito bundle – exclusively available to Jito-Solana validator.
func (c *Client) SimulateBundle(ctx context.Context, bundleParams SimulateBundleParams, simulationConfigs SimulateBundleConfig) (*SimulatedBundleResponse, error) {
	encodedTxLen := len(bundleParams.EncodedTransactions)
	preExecAccCfgLen := len(simulationConfigs.PreExecutionAccountsConfigs)
	if encodedTxLen != preExecAccCfgLen {
		return nil, fmt.Errorf("pre/post execution account config length must match bundle length: encodedTxLen: %d && preExecutionAccountsConfigsLen: %d", encodedTxLen, preExecAccCfgLen)
	}

	var out SimulatedBundleResponse
	err := c.JitoRpcConn.RPCCallForInto(ctx, &out, "simulateBundle", []interface{}{bundleParams, simulationConfigs})
	return &out, err
}

// GetBundleStatuses returns the status of submitted bundle(s). This function operates similarly to the Solana RPC method getSignatureStatuses.
func (c *Client) GetBundleStatuses(ctx context.Context, bundleIDs []string) (*BundleStatusesResponse, error) {
	if len(bundleIDs) > 5 {
		return nil, fmt.Errorf("max length reached (exp 5, got %d), please use BatchGetBundleStatuses or reduce the amount of bundles", len(bundleIDs))
	}

	var params []interface{}
	for _, bundleID := range bundleIDs {
		params = append(params, bundleID)
	}

	var out BundleStatusesResponse
	err := c.JitoRpcConn.RPCCallForInto(ctx, &out, "getBundleStatuses", params)

	return &out, err
}

// BatchGetBundleStatuses returns the statuses of multiple submitted bundles by splitting the bundleIDs into groups of up to 5
// and calling GetBundleStatuses on each group.
func (c *Client) BatchGetBundleStatuses(ctx context.Context, bundleIDs ...string) ([]*BundleStatusesResponse, error) {
	if len(bundleIDs) > 5 {
		var bundles [][]string
		var out []*BundleStatusesResponse

		for _, bundleID := range bundleIDs {
			if len(bundles) == 0 || len(bundles[len(bundles)-1]) == 5 {
				bundles = append(bundles, []string{bundleID})
			} else {
				bundles[len(bundles)-1] = append(bundles[len(bundles)-1], bundleID)
			}
		}

		for _, bundle := range bundles {
			resp, err := c.GetBundleStatuses(ctx, bundle)
			if err != nil {
				return out, err
			}

			out = append(out, resp)
		}

		return out, nil
	} else {
		var out []*BundleStatusesResponse

		resp, err := c.GetBundleStatuses(ctx, bundleIDs)
		if err != nil {
			return nil, err
		}

		out = append(out, resp)

		return out, nil
	}
}

// GetBundleStatuses returns the status of submitted bundle(s). This function operates similarly to the Solana RPC method getSignatureStatuses.
func GetBundleStatuses(client *http.Client, bundleIDs []string) (*BundleStatusesResponse, error) {
	if len(bundleIDs) > 5 {
		return nil, fmt.Errorf("max length reached (exp 5, got %d), please use BatchGetBundleStatuses or reduce the amount of bundles", len(bundleIDs))
	}

	buf := new(bytes.Buffer)
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBundleStatuses",
		"params":  [][]string{bundleIDs},
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    jitoBundleURL,
		Body:   io.NopCloser(buf),
		Header: DefaultHeader.Clone(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GetBundleStatuses: client error > %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetBundleStatuses error: unexpected response status %s", resp.Status)
	}

	var out BundleStatusesResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	return &out, err
}

// BatchGetBundleStatuses returns the statuses of multiple submitted bundles by splitting the bundleIDs into groups of up to 5
// and calling GetBundleStatuses on each group.
func BatchGetBundleStatuses(client *http.Client, bundleIDs ...string) ([]*BundleStatusesResponse, error) {
	if len(bundleIDs) > 5 {
		var bundles [][]string
		var out []*BundleStatusesResponse

		for _, bundleID := range bundleIDs {
			if len(bundles) == 0 || len(bundles[len(bundles)-1]) == 5 {
				bundles = append(bundles, []string{bundleID})
			} else {
				bundles[len(bundles)-1] = append(bundles[len(bundles)-1], bundleID)
			}
		}

		for _, bundle := range bundles {
			resp, err := GetBundleStatuses(client, bundle)
			if err != nil {
				return out, err
			}

			out = append(out, resp)
		}

		return out, nil
	} else {
		var out []*BundleStatusesResponse

		resp, err := GetBundleStatuses(client, bundleIDs)
		if err != nil {
			return nil, err
		}

		out = append(out, resp)

		return out, nil
	}
}

// AssembleBundle converts an array of SOL transactions to a Jito bundle.
func (c *Client) AssembleBundle(transactions []*solana.Transaction) (*jito_pb.Bundle, error) {
	packets := make([]*jito_pb.Packet, 0, len(transactions))

	// converts an array of transactions to an array of protobuf packets
	for i, tx := range transactions {
		packet, err := pkg.ConvertTransactionToProtobufPacket(tx)
		if err != nil {
			return nil, fmt.Errorf("%d: error converting tx to jito_pb packet [%w]", i, err)
		}

		packets = append(packets, &packet)
	}

	return &jito_pb.Bundle{Packets: packets, Header: nil}, nil
}

// GetInflightBundleStatuses returns the status of submitted bundles within the last five minutes, allowing up to five bundle IDs per request.
func GetInflightBundleStatuses(client *http.Client, bundles []string) (*GetInflightBundlesStatusesResponse, error) {
	buf := new(bytes.Buffer)

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getInflightBundleStatuses",
		"params": [][]string{
			bundles,
		},
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    jitoBundleURL,
		Body:   io.NopCloser(buf),
		Header: DefaultHeader.Clone(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetInflightBundleStatuses error: unexpected response status %s", resp.Status)
	}

	var out GetInflightBundlesStatusesResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	return &out, err
}

// GetTipAccounts retrieves the tip accounts designated for tip payments for bundles.
func GetTipAccounts(client *http.Client) (*GetTipAccountsResponse, error) {
	buf := new(bytes.Buffer)

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTipAccounts",
		"params":  []string{},
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    jitoBundleURL,
		Body:   io.NopCloser(buf),
		Header: DefaultHeader.Clone(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetTipAccounts error: unexpected response status %s", resp.Status)
	}

	var out GetTipAccountsResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	return &out, err
}

// SendTransaction serves as a proxy to the Solana sendTransaction RPC method.
// It forwards the received transaction as a regular Solana transaction via the Solana RPC method and submits it as a bundle.
// Jito no longer provides a minimum tip for the bundle.
// Please note that this minimum tip might not be sufficient to get the bundle through the auction, especially during high-demand periods.
// Additionally, you need to set a priority fee and jito tip to ensure this transaction is set up correctly.
// Otherwise, if you set the query parameter bundleOnly=true, the transaction will only be sent out as a bundle and not as a regular transaction via RPC.
func SendTransaction(client *http.Client, sig string, bundleOnly bool, encoding Encoding) (*TransactionResponse, error) {
	buf := new(bytes.Buffer)

	params := []any{sig}
	if encoding == Base64 {
		params = append(params, map[string]string{
			"encoding": encoding.String(),
		})
	}

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "sendTransaction",
		"params":  params,
	}

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}

	var path = "/api/v1/transactions"
	if bundleOnly {
		path = "/api/v1/transactions?bundleOnly=true"
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "https",
			Host:   "mainnet.block-engine.jito.wtf",
			Path:   path,
		},
		Body:   io.NopCloser(buf),
		Header: DefaultHeader.Clone(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SendTransaction error: unexpected response status %s", resp.Status)
	}

	var tx TransactionResponse
	if err = json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return nil, err
	}

	tx.BundleID = resp.Header.Get("x-bundle-id")

	return &tx, nil
}

// GenerateTipInstruction is a function that generates a Solana tip instruction mandatory to broadcast a bundle to Jito.
func GenerateTipInstruction(tipAmount uint64, from, tipAccount solana.PublicKey) solana.Instruction {
	return system.NewTransferInstruction(tipAmount, from, tipAccount).Build()
}

// GenerateTipRandomAccountInstruction functions similarly to GenerateTipInstruction, but it selects a random tip account.
func (c *Client) GenerateTipRandomAccountInstruction(tipAmount uint64, from solana.PublicKey) (solana.Instruction, error) {
	tipAccount, err := c.GetRandomTipAccount()
	if err != nil {
		return nil, err
	}

	return system.NewTransferInstruction(tipAmount, from, solana.MustPublicKeyFromBase58(tipAccount)).Build(), nil
}

func isRPCNil(client *rpc.Client) {
	if client == nil {
		client = rpc.New(rpc.MainNetBeta_RPC)
	}
}

type BundleRejectionError struct {
	Message string
}

func (e BundleRejectionError) Error() string {
	return e.Message
}

func NewStateAuctionBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle lost state auction, auction: %s, tip %d lamports", auction, tip),
	}
}

func NewWinningBatchBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle won state auction but failed global auction, auction %s, tip %d lamports", auction, tip),
	}
}

func NewSimulationFailureError(tx string, message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle simulation failure on tx %s, message: %s", tx, message),
	}
}

func NewInternalError(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("internal error %s", message),
	}
}

func NewDroppedBundle(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle dropped %s", message),
	}
}

// AddProxyToHttpClient adds a proxy to an existing HTTP client.
func AddProxyToHttpClient(client *http.Client, proxy string) error {
	if client == nil {
		client = &http.Client{}
	}

	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	t, ok := transport.(*http.Transport)
	if !ok {
		return errors.New("client transport is not an *http.Transport")
	}

	newTransport := t.Clone()
	newTransport.Proxy = http.ProxyURL(proxyURL)
	client.Transport = newTransport

	return nil
}
