package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

func postClusterTokens(m *microcluster.MicroCluster, s *state.State, r *http.Request) response.Response {
	req := apiv1.TokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	var (
		token string
		err   error
	)
	if req.Worker {
		token, err = createWorkerToken(s)
	} else {
		token, err = m.NewJoinToken(req.Name)
	}
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create token: %w", err))
	}

	return response.SyncResponse(true, &apiv1.TokensResponse{EncodedToken: token})
}

func createWorkerToken(s *state.State) (string, error) {
	var token string
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetOrCreateWorkerNodeToken(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to create worker node token: %w", err)
		}
		return err
	}); err != nil {
		return "", fmt.Errorf("database transaction failed: %w", err)
	}

	remoteAddresses := s.Remotes().Addresses()
	addresses := make([]string, 0, len(remoteAddresses))
	for _, addrPort := range remoteAddresses {
		addresses = append(addresses, addrPort.String())
	}

	info := &types.InternalWorkerNodeToken{
		Token:         token,
		JoinAddresses: addresses,
	}
	token, err := info.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to encode join token: %w", err)
	}

	return token, nil
}