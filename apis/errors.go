package apis

import "errors"

// ErrModelNoPrimaryKey indicates the model schema has no primary key.
var ErrModelNoPrimaryKey = errors.New("model has no primary key")

// ErrAuditUserCompositePK indicates the audit user model has a composite primary key which is not supported.
var ErrAuditUserCompositePK = errors.New("audit user model has composite primary key, only single primary key is supported")
