package orm

import (
	"os"
	"reflect"

	"github.com/bwmarrin/snowflake"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/spf13/cast"
	"github.com/uptrace/bun"
)

// node is the snowflake node for generating unique IDs
var node *snowflake.Node

func init() {
	// Set custom epoch for ID generation
	snowflake.Epoch = 1754582400000
	// Set node bits for distributed ID generation
	snowflake.NodeBits = 6
	// Set step bits for sequence generation
	snowflake.StepBits = 12

	var (
		nodeId int64
		err    error
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

func (*idGenerator) OnCreate(_ *bun.InsertQuery, _ *orm.Table, field *orm.Field, _ any, value reflect.Value) {
	if field.IsPK && field.IndirectType.Kind() == reflect.String && value.IsZero() {
		value.SetString(GenerateId())
	}
}

func (*idGenerator) Name() string {
	return orm.ColumnId
}
