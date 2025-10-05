package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/muesli/termenv"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

type queryHook struct {
	logger logPkg.Logger
	output *termenv.Output
}

func (qh *queryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (qh *queryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
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

	operationStyle := qh.output.String(fmt.Sprintf(" %-8s ", event.Operation())).Bold().Foreground(termenv.ANSIBrightBlack)

	queryStyle := qh.output.String(event.Query)
	switch event.Operation() {
	case "SELECT":
		// Green background for SELECT operations
		operationStyle = operationStyle.Background(termenv.ANSIBrightGreen)
		// Green text for SELECT queries
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightGreen)
	case "INSERT":
		// Blue background for INSERT operations
		operationStyle = operationStyle.Background(termenv.ANSIBrightBlue)
		// Blue text for INSERT queries
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightBlue)
	case "UPDATE":
		// Yellow background for UPDATE operations
		operationStyle = operationStyle.Background(termenv.ANSIBrightYellow)
		// Yellow text for UPDATE queries
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightYellow)
	case "DELETE":
		// Magenta background for DELETE operations
		operationStyle = operationStyle.Background(termenv.ANSIBrightMagenta)
		// Magenta text for DELETE queries
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightMagenta)
	default:
		// Cyan background for other operations
		operationStyle = operationStyle.Background(termenv.ANSICyan)
		// Cyan text for other queries
		queryStyle = queryStyle.Foreground(termenv.ANSICyan)
	}

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		var (
			errorMessage strings.Builder // errorMessage builds the error message string
			message      strings.Builder // message builds the final log message string
		)

		_ = errorMessage.WriteByte(constants.ByteSpace)
		_, _ = errorMessage.WriteString(event.Err.Error())
		_ = errorMessage.WriteByte(constants.ByteSpace)

		errorMessageStyle := qh.output.String(errorMessage.String()).Bold().Background(termenv.ANSIBrightRed).Foreground(termenv.ANSIBlack)

		_, _ = message.WriteString(operationStyle.String())
		_, _ = message.WriteString(elapsedStyle.String())
		_ = message.WriteByte(constants.ByteSpace)
		_, _ = message.WriteString(queryStyle.Foreground(termenv.ANSIRed).String())
		_ = message.WriteByte(constants.ByteSpace)
		_, _ = message.WriteString(errorMessageStyle.String())

		qh.logger.Error(message.String())

		return
	}

	// message builds the final log message string
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

// addQueryHook adds a query hook to the database with a custom logger.
func addQueryHook(db *bun.DB, logger logPkg.Logger) {
	db.AddQueryHook(&queryHook{
		logger: logger,
		output: termenv.DefaultOutput(),
	})
}
