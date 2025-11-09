package security

import (
	"context"

	"github.com/ilxqx/vef-framework-go/event"
)

const (
	eventTypeLogin = "vef.security.login"
)

// LoginEvent represents a user login event.
type LoginEvent struct {
	event.BaseEvent

	AuthType   string `json:"authType"`
	UserId     string `json:"userId"` // Populated on success
	Username   string `json:"username"`
	LoginIp    string `json:"loginIp"`
	UserAgent  string `json:"userAgent"`
	TraceId    string `json:"traceId"`
	IsOk       bool   `json:"isOk"`
	FailReason string `json:"failReason"` // Populated on failure
	ErrorCode  int    `json:"errorCode"`
}

// NewLoginEvent creates a new login event with the given parameters.
func NewLoginEvent(
	authType string,
	userId, username string,
	loginIp, userAgent, traceId string,
	isOk bool, failReason string, errorCode int,
) *LoginEvent {
	return &LoginEvent{
		BaseEvent:  event.NewBaseEvent(eventTypeLogin),
		AuthType:   authType,
		UserId:     userId,
		Username:   username,
		LoginIp:    loginIp,
		UserAgent:  userAgent,
		TraceId:    traceId,
		IsOk:       isOk,
		FailReason: failReason,
		ErrorCode:  errorCode,
	}
}

// SubscribeLoginEvent subscribes to login events.
// Returns an unsubscribe function that can be called to remove the subscription.
func SubscribeLoginEvent(subscriber event.Subscriber, handler func(context.Context, *LoginEvent)) event.UnsubscribeFunc {
	return subscriber.Subscribe(eventTypeLogin, func(ctx context.Context, evt event.Event) {
		if loginEvt, ok := evt.(*LoginEvent); ok {
			handler(ctx, loginEvt)
		}
	})
}
