package id

import (
	"fmt"
	"os"

	"github.com/bwmarrin/snowflake"
	"github.com/spf13/cast"

	"github.com/ilxqx/vef-framework-go/constants"
)

// DefaultSnowflakeIdGenerator is the default Snowflake ID generator instance.
var DefaultSnowflakeIdGenerator IdGenerator

// init initializes the Snowflake algorithm with custom configuration and creates the default generator.
// Configuration:
//   - Epoch: 1754582400000 (custom start time)
//   - Node bits: 6 (supports 64 nodes: 0-63)
//   - Step bits: 12 (supports 4096 IDs per millisecond per node)
func init() {
	snowflake.Epoch = 1754582400000
	snowflake.NodeBits = 6
	snowflake.StepBits = 12

	var (
		nodeId int64
		err    error
	)

	nodeIdStr := os.Getenv(constants.EnvNodeId)
	if nodeIdStr != constants.Empty {
		if nodeId, err = cast.ToInt64E(nodeIdStr); err != nil {
			panic(
				fmt.Errorf("failed to convert node id to int: %w", err),
			)
		}
	}

	if DefaultSnowflakeIdGenerator, err = NewSnowflakeIdGenerator(nodeId); err != nil {
		panic(err)
	}
}

// snowflakeIdGenerator implements IdGenerator using the Snowflake algorithm.
type snowflakeIdGenerator struct {
	node *snowflake.Node
}

// Generate creates a new Snowflake ID encoded as a Base36 string.
func (g *snowflakeIdGenerator) Generate() string {
	return g.node.Generate().Base36()
}

// NewSnowflakeIdGenerator creates a new Snowflake ID generator for the specified node.
// The nodeId must be between 0 and 63 (6-bit limit as configured in init).
// Each node in a distributed system should have a unique nodeId to ensure global uniqueness.
func NewSnowflakeIdGenerator(nodeId int64) (_ IdGenerator, err error) {
	var node *snowflake.Node
	if node, err = snowflake.NewNode(nodeId); err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	return &snowflakeIdGenerator{
		node: node,
	}, nil
}
