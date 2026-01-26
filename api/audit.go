package api

import (
	"context"

	"github.com/ilxqx/vef-framework-go/event"
)

const (
	eventTypeAudit = "vef.api.request.audit"
)

// AuditEvent represents an Api request audit log event.
type AuditEvent struct {
	event.BaseEvent

	// Api identification
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Version  string `json:"version"`

	// User identification
	UserID    string `json:"userId"`
	UserAgent string `json:"userAgent"`

	// Request information
	RequestID     string         `json:"requestId"`
	RequestIP     string         `json:"requestIp"`
	RequestParams map[string]any `json:"requestParams"`
	RequestMeta   map[string]any `json:"requestMeta"`

	// Response information
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	ResultData    any    `json:"resultData"`

	// Performance metrics
	ElapsedTime int64 `json:"elapsedTime"` // Elapsed time in milliseconds
}

// NewAuditEvent creates a new audit event with the given parameters.
func NewAuditEvent(
	apiResource, apiAction, apiVersion string,
	userID, userAgent string,
	requestID, requestIP string,
	requestParams, requestMeta map[string]any,
	resultCode int, resultMessage string, resultData any,
	elapsedTime int64,
) *AuditEvent {
	return &AuditEvent{
		BaseEvent:     event.NewBaseEvent(eventTypeAudit),
		Resource:      apiResource,
		Action:        apiAction,
		Version:       apiVersion,
		UserID:        userID,
		UserAgent:     userAgent,
		RequestID:     requestID,
		RequestIP:     requestIP,
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
