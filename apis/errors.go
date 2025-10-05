package apis

import "errors"

// ErrModelNoPrimaryKey indicates the model schema has no primary key.
var ErrModelNoPrimaryKey = errors.New("model has no primary key")
