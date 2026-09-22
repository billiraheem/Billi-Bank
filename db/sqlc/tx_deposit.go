package db

import "context"

type DepositTxParams struct {
	AccountID int64 `json:"account_id"`
	Amount    int64 `json:"amount"`
}

type DepositTxResult struct {
	Account Account `json:"account"`
	Entry   Entry   `json:"entry"`
}

// DepositTx performs a money deposit to one acct.
// It creates an entry record, and update acct's balance within a single DB transaction
func (store *SQLStore) DepositTx(ctx context.Context, args DepositTxParams) (DepositTxResult, error) {
	var result DepositTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Account, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: args.AccountID,
			Amount: args.Amount,
		})
		if err != nil {
			return err
		}

		result.Entry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: args.AccountID,
			Amount: args.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
