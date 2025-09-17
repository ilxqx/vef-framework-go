package orm

import (
	"os"
	"reflect"

	"github.com/bwmarrin/snowflake"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/spf13/cast"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

var node *snowflake.Node // node is the snowflake node for generating unique IDs

func init() {
	snowflake.Epoch = 1754582400000 // Set custom epoch for ID generation
	snowflake.NodeBits = 6          // Set node bits for distributed ID generation
	snowflake.StepBits = 12         // Set step bits for sequence generation

	var (
		nodeId int64 // nodeId is the node ID from environment variable
		err    error // err stores any initialization error
	)

	nodeIdStr := os.Getenv(constants.EnvNodeId)
	if nodeIdStr != constants.Empty {
		nodeId, err = cast.ToInt64E(nodeIdStr)
		if err != nil {
			logger.Panicf("failed to convert node id to int: %v", err)
		}
	}

	node, err = snowflake.NewNode(nodeId)
	if err != nil {
		logger.Panicf("failed to create snowflake node: %v", err)
	}
}

// GenerateId generates a new unique primary key.
func GenerateId() string {
	return node.Generate().Base36()
}

// idGenerator is a struct that implements the CreateAutoColumn interface for auto-generating an ID.
type idGenerator struct {
}

func (*idGenerator) OnCreate(_ *bun.InsertQuery, _ *schema.Table, field *schema.Field, _ any, value reflect.Value) {
	if field.IsPK && field.IndirectType.Kind() == reflect.String && value.IsZero() {
		value.SetString(GenerateId())
	}
}

func (*idGenerator) Name() string {
	return orm.ColumnId
}
