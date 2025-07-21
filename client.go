package fayl

import (
	"context"

	"github.com/redhajuanda/fayl/parser"
	"github.com/redhajuanda/perkakas/logger"

	"github.com/pkg/errors"
)

// Client is the main struct for the fayl client.
// It contains the database connection, runners, placeholder format, and logger.
// It provides methods to run queries and manage transactions.
type Client struct {
	db          *DB
	runners     map[string]string
	placeholder parser.Placeholder
	log         logger.Logger
}

// Run initializes a new Runner with the given runner code.
func (c *Client) Run(runner string) Runnerer {

	return newRunner(runnerParams{
		runnerCode:    runner,
		client:        c,
		log:           c.log,
		inTransaction: false,
	})

}

// WithTransaction initializes a new query with transaction.
// it takes a context, and callback as input.
// context is the context of the transaction.
// callback is a function that will be executed in the transaction.
// callback takes a context and tx as input.
// tx is a struct that contains the transaction configs.
func (c *Client) WithTransaction(ctx context.Context, callback TxFunc) (out any, err error) {

	// begin transaction
	c.log.WithContext(ctx).Debug("beginning transaction")
	ctx, err = c.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}

	// defer rollback or commit transaction
	// if panic occurs, rollback transaction
	// if error occurs, rollback transaction
	// if no panic or error occurs, commit transaction
	defer c.handleTransaction(ctx, &err)

	// execute callback
	c.log.WithContext(ctx).Debug("executing callback")
	out, err = callback(ctx, newTx(c, c.log))

	return

}

// handleTransaction handles the transaction logic for a given context.
// It rolls back the transaction if a panic occurs or if an error is passed as input.
// If no panic or error occurs, it commits the transaction.
// It returns an error if there is a failure in rolling back or committing the transaction.
func (c *Client) handleTransaction(ctx context.Context, errIn *error) (errOut error) {

	if p := recover(); p != nil {

		c.log.WithContext(ctx).Debug("panic occurred, rolling back transaction")

		err := c.db.Rollback(ctx)
		if err != nil {
			errOut = errors.Wrap(err, "failed to rollback transaction")
		}
		panic(p) // re-throw panic after Rollback

	} else if *errIn != nil {

		c.log.WithContext(ctx).Debug("error occurred, rolling back transaction")

		err := c.db.Rollback(ctx)
		if err != nil {
			errOut = errors.Wrap(err, "failed to rollback transaction")
		}

	} else {

		c.log.WithContext(ctx).Debug("committing transaction")

		err := c.db.Commit(ctx)
		if err != nil {
			errOut = errors.Wrap(err, "failed to commit transaction")
		}

	}
	return

}
