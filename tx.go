package fayl

import (
	"context"

	"github.com/redhajuanda/perkakas/logger"
)

// TxFunc represents the function signature for transaction callback.
type TxFunc func(ctx context.Context, tx *Tx) (out any, err error)

// Tx is a struct that used to run a transaction
type Tx struct {
	client *Client
	log    logger.Logger
}

// newTx returns a new transaction
func newTx(dclient *Client, log logger.Logger) *Tx {

	return &Tx{
		client: dclient,
		log:    log,
	}

}

// Run is a function to run query within the transaction
func (t *Tx) Run(runnerCode string) Runnerer {

	return newRunner(runnerParams{
		runnerCode:    runnerCode,
		client:        t.client,
		log:           t.log,
		inTransaction: true,
	})

}
