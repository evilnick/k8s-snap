package types

type Network struct {
	Enabled          *bool   `json:"enabled,omitempty"`
	PodCIDR          *string `json:"pod-cidr,omitempty"`
	ServiceCIDR      *string `json:"service-cidr,omitempty"`
	LocalhostAddress *string `json:"localhost-address,omitempty"`
}

func (c Network) GetEnabled() bool            { return getField(c.Enabled) }
func (c Network) GetPodCIDR() string          { return getField(c.PodCIDR) }
func (c Network) GetServiceCIDR() string      { return getField(c.ServiceCIDR) }
func (c Network) GetLocalhostAddress() string { return getField(c.LocalhostAddress) }
func (c Network) Empty() bool                 { return c == Network{} }
