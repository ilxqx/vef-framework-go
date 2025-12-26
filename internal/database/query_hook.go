package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/muesli/termenv"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/database/sqlguard"
	"github.com/ilxqx/vef-framework-go/log"
)

// whitespaceRegex matches consecutive whitespace characters (spaces, tabs, newlines).
var whitespaceRegex = regexp.MustCompile(`\s+`)

// guardErrorStashKey is the stash key for storing guard errors.
const guardErrorStashKey = "__sqlguard_error"

type queryHook struct {
	logger   log.Logger
	output   *termenv.Output
	sqlGuard *sqlguard.Guard
}

func (qh *queryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	if qh.sqlGuard == nil || sqlguard.IsWhitelisted(ctx) {
		return ctx
	}

	if err := qh.sqlGuard.Check(event.Query); err != nil {
		if event.Stash == nil {
			event.Stash = make(map[any]any)
		}

		event.Stash[guardErrorStashKey] = err

		cancelCtx, cancel := context.WithCancelCause(ctx)
		cancel(err)

		return cancelCtx
	}

	return ctx
}

func (qh *queryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	var guardErr error
	if event.Stash != nil {
		if err, ok := event.Stash[guardErrorStashKey].(error); ok {
			guardErr = err
		}
	}

	elapsed := time.Since(event.StartTime).Milliseconds()
	elapsedStyle := qh.output.String(fmt.Sprintf("%6d ms", elapsed))

	switch {
	case elapsed >= 1000:
		// Red for slow queries (>=1s)
		elapsedStyle = elapsedStyle.Bold().Foreground(termenv.ANSIRed)
	case elapsed >= 500:
		// Yellow for medium queries (>=500ms)
		elapsedStyle = elapsedStyle.Bold().Foreground(termenv.ANSIYellow)
	case elapsed >= 200:
		// Blue for moderate queries (>=200ms)
		elapsedStyle = elapsedStyle.Foreground(termenv.ANSIBlue)
	default:
		// Green for fast queries (<200ms)
		elapsedStyle = elapsedStyle.Foreground(termenv.ANSIGreen)
	}

	operationStyle := qh.output.String(fmt.Sprintf(" %-8s", event.Operation())).Bold()

	// Normalize SQL: collapse multiple whitespace (including newlines) into single spaces
	// This ensures consistent coloring across the entire query string
	normalizedQuery := strings.TrimSpace(whitespaceRegex.ReplaceAllString(event.Query, constants.Space))
	// Use muted gray for SQL to reduce visual noise and keep focus on operation type and timing
	queryStyle := qh.output.String(normalizedQuery).Foreground(termenv.ANSIBrightBlack)

	// Color operation type by category (foreground only, no background for cleaner look)
	switch event.Operation() {
	case "SELECT":
		operationStyle = operationStyle.Foreground(termenv.ANSIGreen)
	case "INSERT":
		operationStyle = operationStyle.Foreground(termenv.ANSIBlue)
	case "UPDATE":
		operationStyle = operationStyle.Foreground(termenv.ANSIYellow)
	case "DELETE":
		operationStyle = operationStyle.Foreground(termenv.ANSIMagenta)
	default:
		operationStyle = operationStyle.Foreground(termenv.ANSICyan)
	}

	// Use guard error if present, otherwise use the event error
	displayErr := event.Err
	if guardErr != nil {
		displayErr = guardErr
	}

	if displayErr != nil && !errors.Is(displayErr, sql.ErrNoRows) {
		var message strings.Builder

		errorMessageStyle := qh.output.String(displayErr.Error()).Foreground(termenv.ANSIRed)

		_, _ = message.WriteString(operationStyle.String())
		_, _ = message.WriteString(elapsedStyle.String())
		_ = message.WriteByte(constants.ByteSpace)
		_, _ = message.WriteString(queryStyle.String())
		_ = message.WriteByte(constants.ByteSpace)
		_, _ = message.WriteString(errorMessageStyle.String())

		qh.logger.Error(message.String())

		return
	}

	var message strings.Builder

	_, _ = message.WriteString(operationStyle.String())
	_, _ = message.WriteString(elapsedStyle.String())
	_ = message.WriteByte(constants.ByteSpace)
	_, _ = message.WriteString(queryStyle.String())

	if elapsed >= 500 {
		qh.logger.Warn(message.String())
	} else {
		qh.logger.Info(message.String())
	}
}

func addQueryHook(db *bun.DB, logger log.Logger, guardConfig *sqlguard.Config) {
	var guard *sqlguard.Guard
	if guardConfig != nil && guardConfig.Enabled {
		guard = sqlguard.NewGuard(logger)
	}

	db.AddQueryHook(&queryHook{
		logger:   logger,
		output:   termenv.DefaultOutput(),
		sqlGuard: guard,
	})
}
