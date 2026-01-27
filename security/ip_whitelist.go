package security

import (
	"net"
	"strings"

	"github.com/ilxqx/go-collections"

	"github.com/ilxqx/vef-framework-go/constants"
)

// IPWhitelistValidator validates IP addresses against a whitelist.
// It supports both individual IP addresses and CIDR notation.
type IPWhitelistValidator struct {
	// networks contains parsed CIDR networks for range matching.
	networks []*net.IPNet
	// ips contains individual IP addresses for O(1) exact matching.
	ips collections.Set[string]
	// isEmpty indicates whether the whitelist is empty (allow all).
	isEmpty bool
}

// NewIPWhitelistValidator creates a new IP whitelist validator from a comma-separated string.
// Supports individual IP addresses (e.g., "192.168.1.1") and CIDR notation (e.g., "192.168.1.0/24").
// An empty whitelist means all IPs are allowed.
func NewIPWhitelistValidator(whitelist string) *IPWhitelistValidator {
	validator := &IPWhitelistValidator{
		networks: make([]*net.IPNet, 0),
		ips:      collections.NewHashSet[string](),
	}

	whitelist = strings.TrimSpace(whitelist)
	if whitelist == constants.Empty {
		validator.isEmpty = true

		return validator
	}

	for entry := range strings.SplitSeq(whitelist, constants.Comma) {
		entry = strings.TrimSpace(entry)
		if entry == constants.Empty {
			continue
		}

		// Handle CIDR notation
		if strings.Contains(entry, constants.Slash) {
			if _, network, err := net.ParseCIDR(entry); err != nil {
				logger.Warnf("Failed to parse CIDR %s: %v", entry, err)
			} else {
				validator.networks = append(validator.networks, network)
			}

			continue
		}

		// Handle individual IP address
		if ip := net.ParseIP(entry); ip != nil {
			validator.ips.Add(ip.String())
		}
	}

	validator.isEmpty = len(validator.networks) == 0 && validator.ips.IsEmpty()

	return validator
}

// IsAllowed checks if the given IP address is in the whitelist.
func (v *IPWhitelistValidator) IsAllowed(ipStr string) bool {
	if v.isEmpty {
		return true
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if v.ips.Contains(ip.String()) {
		return true
	}

	for _, network := range v.networks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// IsEmpty returns true if the whitelist is empty (no restrictions).
func (v *IPWhitelistValidator) IsEmpty() bool {
	return v.isEmpty
}
