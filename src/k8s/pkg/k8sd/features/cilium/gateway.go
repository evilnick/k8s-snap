package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when gateway.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when gateway.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway returns an error if anything fails.
func ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		return fmt.Errorf("failed to install Gateway API CRDs: %w", err)
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, chartGatewayClass, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		return fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), map[string]any{"gatewayAPI": map[string]any{"enabled": gateway.GetEnabled()}})
	if err != nil {
		return fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
	}

	if !changed || !gateway.GetEnabled() {
		return nil
	}
	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply Gateway API: %w", err)
	}
	return nil
}
