package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"github.com/muesli/termenv"
	"github.com/uptrace/bun"
)

type queryHook struct {
	logger logPkg.Logger   // logger is the logger instance for query logging
	output *termenv.Output // output is the terminal output for colored formatting
}

func (qh *queryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (qh *queryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	elapsed := time.Since(event.StartTime).Milliseconds()            // elapsed is the query execution time in milliseconds
	elapsedStyle := qh.output.String(fmt.Sprintf("%6d ms", elapsed)) // elapsedStyle formats the elapsed time

	switch {
	case elapsed >= 1000:
		elapsedStyle = elapsedStyle.Bold().Foreground(termenv.ANSIRed) // Red for slow queries (>=1s)
	case elapsed >= 500:
		elapsedStyle = elapsedStyle.Bold().Foreground(termenv.ANSIYellow) // Yellow for medium queries (>=500ms)
	case elapsed >= 200:
		elapsedStyle = elapsedStyle.Foreground(termenv.ANSIBlue) // Blue for moderate queries (>=200ms)
	default:
		elapsedStyle = elapsedStyle.Foreground(termenv.ANSIGreen) // Green for fast queries (<200ms)
	}

	operationStyle := qh.output.String(fmt.Sprintf(" %-8s ", event.Operation())).Bold().Foreground(termenv.ANSIBrightBlack) // operationStyle formats the SQL operation
	queryStyle := qh.output.String(event.Query)                                                                             // queryStyle formats the SQL query
	switch event.Operation() {
	case "SELECT":
		operationStyle = operationStyle.Background(termenv.ANSIBrightGreen) // Green background for SELECT operations
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightGreen)         // Green text for SELECT queries
	case "INSERT":
		operationStyle = operationStyle.Background(termenv.ANSIBrightBlue) // Blue background for INSERT operations
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightBlue)         // Blue text for INSERT queries
	case "UPDATE":
		operationStyle = operationStyle.Background(termenv.ANSIBrightYellow) // Yellow background for UPDATE operations
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightYellow)         // Yellow text for UPDATE queries
	case "DELETE":
		operationStyle = operationStyle.Background(termenv.ANSIBrightMagenta) // Magenta background for DELETE operations
		queryStyle = queryStyle.Foreground(termenv.ANSIBrightMagenta)         // Magenta text for DELETE queries
	default:
		operationStyle = operationStyle.Background(termenv.ANSICyan) // Cyan background for other operations
		queryStyle = queryStyle.Foreground(termenv.ANSICyan)         // Cyan text for other queries
	}

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		var (
			errorMessage strings.Builder // errorMessage builds the error message string
			message      strings.Builder // message builds the final log message string
		)

		_ = errorMessage.WriteByte(' ')
		_, _ = errorMessage.WriteString(event.Err.Error())
		_ = errorMessage.WriteByte(' ')

		errorMessageStyle := qh.output.String(errorMessage.String()).Bold().Background(termenv.ANSIBrightRed).Foreground(termenv.ANSIBlack) // errorMessageStyle formats error with red background

		_, _ = message.WriteString(operationStyle.String())
		_, _ = message.WriteString(elapsedStyle.String())
		_ = message.WriteByte(' ')
		_, _ = message.WriteString(queryStyle.Foreground(termenv.ANSIRed).String())
		_ = message.WriteByte(' ')
		_, _ = message.WriteString(errorMessageStyle.String())

		qh.logger.Error(message.String())
		return
	}

	var message strings.Builder // message builds the final log message string
	_, _ = message.WriteString(operationStyle.String())
	_, _ = message.WriteString(elapsedStyle.String())
	_ = message.WriteByte(' ')
	_, _ = message.WriteString(queryStyle.String())

	if elapsed >= 200 {
		qh.logger.Info(message.String())
	} else {
		qh.logger.Debug(message.String())
	}
}

// addQueryHook adds a query hook to the database.
func addQueryHook(db *bun.DB) {
	db.AddQueryHook(&queryHook{
		logger: log.Named("sql"),        // logger is named "sql" for SQL query logging
		output: termenv.DefaultOutput(), // output uses default terminal output for coloring
	})
}
