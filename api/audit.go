package api

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/event"
)

const (
	eventTypeAudit = constants.VEFName + ".api.request.audit"
)

// AuditEvent represents an API audit log event.
// It captures comprehensive information about API requests including
// request details, user identity, response status, and performance metrics.
type AuditEvent struct {
	event.BaseEvent

	// API identification
	Resource string `json:"resource"` // API resource name
	Action   string `json:"action"`   // API action name
	Version  string `json:"version"`  // API version

	// User identification
	UserId    string `json:"userId"`    // Operating user id
	UserAgent string `json:"userAgent"` // User agent (User-Agent header)

	// Request information
	RequestId     string         `json:"requestId"`     // Request Id
	RequestIP     string         `json:"requestIp"`     // Request IP address
	RequestParams map[string]any `json:"requestParams"` // Request parameters
	RequestMeta   map[string]any `json:"requestMeta"`   // Request metadata

	// Response information
	ResultCode    int    `json:"resultCode"`    // Result code (business code)
	ResultMessage string `json:"resultMessage"` // Result message
	ResultData    any    `json:"resultData"`    // Result data (optional)

	// Performance metrics
	ElapsedTime int `json:"elapsedTime"` // Elapsed time in milliseconds
}

// NewAuditEvent creates a new audit event with the given parameters.
func NewAuditEvent(
	apiResource, apiAction, apiVersion string,
	userId, userAgent string,
	requestId, requestIp string,
	requestParams, requestMeta map[string]any,
	resultCode int, resultMessage string, resultData any,
	elapsedTime int,
) *AuditEvent {
	return &AuditEvent{
		BaseEvent:     event.NewBaseEvent(eventTypeAudit),
		Resource:      apiResource,
		Action:        apiAction,
		Version:       apiVersion,
		UserId:        userId,
		UserAgent:     userAgent,
		RequestId:     requestId,
		RequestIP:     requestIp,
		RequestParams: requestParams,
		RequestMeta:   requestMeta,
		ResultCode:    resultCode,
		ResultMessage: resultMessage,
		ResultData:    resultData,
		ElapsedTime:   elapsedTime,
	}
}

// SubscribeAuditEvent subscribes to audit events.
// Returns an unsubscribe function that can be called to remove the subscription.
func SubscribeAuditEvent(subscriber event.Subscriber, handler func(context.Context, *AuditEvent)) event.UnsubscribeFunc {
	return subscriber.Subscribe(eventTypeAudit, func(ctx context.Context, evt event.Event) {
		if auditEvt, ok := evt.(*AuditEvent); ok {
			handler(ctx, auditEvt)
		}
	})
}
