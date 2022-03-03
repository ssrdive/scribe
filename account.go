package scribe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/ssrdive/mysequel"
	"github.com/ssrdive/scribe/models"
	"github.com/ssrdive/scribe/queries"
)

// AccountModel struct holds database instance
type AccountModel struct {
	DB *sql.DB
}

func validatePostingDate(postingDate string) error {
	now := time.Now()

	var m time.Month
	year, m, _ := now.Date()
	month := int(m)

	var oldestDate time.Time
	if month < 4 {
		oldestDate = time.Date(year-1, 4, 1, 0, 0, 0, 0, time.UTC)
	} else {
		oldestDate = time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC)
	}

	parsedDate, err := time.Parse("2006-01-02", postingDate)
	if err != nil {
		return errors.New("invalid posting date")
	}

	if parsedDate.Before(oldestDate) {
		return errors.New("posting date does not fall within the financial year")
	} else {
		return nil
	}
}

func CreateTransaction(tx *sql.Tx, userID, postingDate, contractID, remark string) (int64, error) {
	err := validatePostingDate(postingDate)
	if err != nil {
		return 0, err
	}

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "contract_id", "remark"},
		Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), postingDate, contractID, remark},
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	return tid, err
}

// IssueJournalEntries issues journal entries
func IssueJournalEntries(tx *sql.Tx, tid int64, journalEntries []models.JournalEntry) error {
	for _, entry := range journalEntries {
		if len(entry.Debit) != 0 {
			_, err := mysequel.Insert(mysequel.Table{
				TableName: "account_transaction",
				Columns:   []string{"transaction_id", "account_id", "type", "amount"},
				Vals:      []interface{}{tid, entry.Account, "DR", entry.Debit},
				Tx:        tx,
			})
			if err != nil {
				return err
			}
		}
		if len(entry.Credit) != 0 {
			_, err := mysequel.Insert(mysequel.Table{
				TableName: "account_transaction",
				Columns:   []string{"transaction_id", "account_id", "type", "amount"},
				Vals:      []interface{}{tid, entry.Account, "CR", entry.Credit},
				Tx:        tx,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateAccount creats an account
func (m *AccountModel) CreateAccount(rparams, oparams []string, form url.Values) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	form.Set("datetime", time.Now().Format("2006-01-02 15:04:05"))
	cid, err := mysequel.Insert(mysequel.FormTable{
		TableName: "account",
		RCols:     rparams,
		OCols:     oparams,
		Form:      form,
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	return cid, nil
}

// CreateCategory creates a category
func (m *AccountModel) CreateCategory(rparams, oparams []string, form url.Values) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	form.Set("datetime", time.Now().Format("2006-01-02 15:04:05"))
	cid, err := mysequel.Insert(mysequel.FormTable{
		TableName: "account_category",
		RCols:     rparams,
		OCols:     oparams,
		Form:      form,
		Tx:        tx,
	})
	if err != nil {
		return 0, err
	}

	return cid, nil
}

// TrialBalance returns trail balance
func (m *AccountModel) TrialBalance() ([]models.TrialEntry, error) {
	var res []models.TrialEntry
	err := mysequel.QueryToStructs(&res, m.DB, queries.TRIAL_BALANCE)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ChartOfAccounts returns chart of accounts
func (m *AccountModel) ChartOfAccounts() ([]models.ChartOfAccount, error) {
	var res []models.ChartOfAccount
	err := mysequel.QueryToStructs(&res, m.DB, queries.CHART_OF_ACCOUNTS)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// PaymentVoucher creates payment voucher
func (m *AccountModel) PaymentVoucher(userID, postingDate, fromAccountID, amount, entries, remark, dueDate, checkNumber, payee string) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	err = validatePostingDate(postingDate)
	if err != nil {
		return 0, err
	}

	var paymentVoucher []models.PaymentVoucherEntry
	json.Unmarshal([]byte(entries), &paymentVoucher)

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), postingDate, remark},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "payment_voucher",
		Columns:   []string{"transaction_id", "due_date", "check_number", "payee"},
		Vals:      []interface{}{tid, dueDate, checkNumber, payee},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "account_transaction",
		Columns:   []string{"transaction_id", "account_id", "type", "amount"},
		Vals:      []interface{}{tid, fromAccountID, "CR", amount},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, entry := range paymentVoucher {
		_, err := mysequel.Insert(mysequel.Table{
			TableName: "account_transaction",
			Columns:   []string{"transaction_id", "account_id", "type", "amount"},
			Vals:      []interface{}{tid, entry.Account, "DR", entry.Amount},
			Tx:        tx,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	return tid, nil
}

// Deposit enters bank deposits
func (m *AccountModel) Deposit(userID, postingDate, toAccountID, amount, entries, remark string) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	err = validatePostingDate(postingDate)
	if err != nil {
		return 0, err
	}

	var paymentVoucher []models.PaymentVoucherEntry
	json.Unmarshal([]byte(entries), &paymentVoucher)

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), postingDate, remark},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "deposit",
		Columns:   []string{"transaction_id"},
		Vals:      []interface{}{tid},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = mysequel.Insert(mysequel.Table{
		TableName: "account_transaction",
		Columns:   []string{"transaction_id", "account_id", "type", "amount"},
		Vals:      []interface{}{tid, toAccountID, "DR", amount},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for _, entry := range paymentVoucher {
		_, err := mysequel.Insert(mysequel.Table{
			TableName: "account_transaction",
			Columns:   []string{"transaction_id", "account_id", "type", "amount"},
			Vals:      []interface{}{tid, entry.Account, "CR", entry.Amount},
			Tx:        tx,
		})
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	return tid, nil
}

// JournalEntry issues journal entries
func (m *AccountModel) JournalEntry(userID, postingDate, remark, entries string) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	err = validatePostingDate(postingDate)
	if err != nil {
		return 0, err
	}

	var journalEntries []models.JournalEntry
	json.Unmarshal([]byte(entries), &journalEntries)

	tid, err := mysequel.Insert(mysequel.Table{
		TableName: "transaction",
		Columns:   []string{"user_id", "datetime", "posting_date", "remark"},
		Vals:      []interface{}{userID, time.Now().Format("2006-01-02 15:04:05"), postingDate, remark},
		Tx:        tx,
	})
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = IssueJournalEntries(tx, tid, journalEntries)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return tid, nil
}

// Transaction returns transaction
func (m *AccountModel) Transaction(aid int) ([]models.Transaction, error) {
	var res []models.Transaction
	err := mysequel.QueryToStructs(&res, m.DB, queries.TRANSACTION, aid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Ledger returns account ledger
func (m *AccountModel) Ledger(aid int) ([]models.LedgerEntry, error) {
	var res []models.LedgerEntry
	err := mysequel.QueryToStructs(&res, m.DB, queries.ACCOUNT_LEDGER, aid)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// PaymentVouchers returns payment vouchers
func (m *AccountModel) PaymentVouchers() ([]models.PaymentVoucherList, error) {
	var res []models.PaymentVoucherList
	err := mysequel.QueryToStructs(&res, m.DB, queries.PAYMENT_VOUCHERS)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// PaymentVoucherDetails returns payment voucher details
func (m *AccountModel) PaymentVoucherDetails(pid int) (models.PaymentVoucherSummary, error) {
	var dueDate, checkNumber, payee, remark, account, datetime sql.NullString
	err := m.DB.QueryRow(queries.PAYMENT_VOUCHER_CHECK_DETAILS, pid).Scan(&dueDate, &checkNumber, &payee, &remark, &account, &datetime)

	var vouchers []models.PaymentVoucherDetails
	err = mysequel.QueryToStructs(&vouchers, m.DB, queries.PAYMENT_VOUCHER_DETAILS, pid)
	if err != nil {
		return models.PaymentVoucherSummary{}, err
	}

	return models.PaymentVoucherSummary{DueDate: dueDate, CheckNumber: checkNumber, Payee: payee, Remark: remark, Account: account, Datetime: datetime, PaymentVoucherDetails: vouchers}, nil
}

func (m *AccountModel) JournalEntriesForAudit(date, posting_date string) ([]models.JEsForAudit, error) {
	var d, p_date sql.NullString
	if date == "" {
		d = sql.NullString{}
	} else {
		d = sql.NullString{
			Valid:  true,
			String: date,
		}
	}
	if posting_date == "" {
		p_date = sql.NullString{}
	} else {
		p_date = sql.NullString{
			Valid:  true,
			String: posting_date,
		}
	}

	var res []models.JEsForAudit
	err := mysequel.QueryToStructs(&res, m.DB, queries.JOURNAL_ENTRIES_FOR_AUDIT, d, d, p_date, p_date)
	if err != nil {
		return nil, err
	}

	return res, nil
}
