package models

import (
	"database/sql"
	"time"
)

type JournalEntry struct {
	Account string
	Debit   string
	Credit  string
}

type TrialEntry struct {
	ID              int     `json:"id"`
	MainAccount     string  `json:"main_account"`
	SubAccount      string  `json:"sub_account"`
	AccountCategory string  `json:"account_category"`
	AccountID       string  `json:"account_id"`
	AccountName     string  `json:"account_name"`
	Debit           float64 `json:"debit"`
	Credit          float64 `json:"credit"`
}

type ChartOfAccount struct {
	MainAccountID     int            `json:"main_account_id"`
	MainAccount       string         `json:"main_account"`
	SubAccountID      int            `json:"sub_account_id"`
	SubAccount        string         `json:"sub_account"`
	AccountCategoryID sql.NullInt32  `json:"account_category_id"`
	AccountCategory   sql.NullString `json:"account_category"`
	AccountID         sql.NullInt32  `json:"account_id"`
	AccountName       sql.NullString `json:"account_name"`
}

type PaymentVoucherEntry struct {
	Account string
	Amount  string
}

type Transaction struct {
	TransactionID int     `json:"transaction_id"`
	AccountID     int     `json:"account_id"`
	AccountID2    int     `json:"account_id2"`
	AccountName   string  `json:"account_name"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
}

type LedgerEntry struct {
	Name          string  `json:"account_name"`
	TransactionID int     `json:"transaction_id"`
	PostingDate   string  `json:"posting_date"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	Remark        string  `json:"remark"`
}

type PaymentVoucherList struct {
	ID          int       `json:"id"`
	Datetime    time.Time `json:"date_time"`
	PostingDate string    `json:"posting_date"`
	FromAccount string    `json:"from_account"`
	User        string    `json:"user"`
}

type PaymentVoucherSummary struct {
	DueDate               sql.NullString          `json:"due_date"`
	CheckNumber           sql.NullString          `json:"check_number"`
	Payee                 sql.NullString          `json:"payee"`
	Remark                sql.NullString          `json:"remark"`
	Account               sql.NullString          `json:"account"`
	Datetime              sql.NullString          `json:"datetime"`
	PaymentVoucherDetails []PaymentVoucherDetails `json:"payment_voucher_details"`
}

type PaymentVoucherDetails struct {
	AccountID   int     `json:"account_id"`
	AccountName string  `json:"account_name"`
	Amount      float64 `json:"amount"`
	PostingDate string  `json:"posting_date"`
}

type JEsForAudit struct {
	Datetime      string  `json:"datetime"`
	Issuer        string  `json:"issuer"`
	TransactionID int     `json:"transaction_id"`
	Account       string  `json:"account"`
	Type          string  `json:"type"`
	PostingDate   string  `json:"posting_date"`
	Amount        float64 `json:"amount"`
	Remark        string  `json:"remark"`
}

type AccountBalanceForReports struct {
	AccountID       int     `json:"account_id"`
	MainAccount     string  `json:"main_account"`
	SubAccount      string  `json:"sub_account"`
	AccountCategory string  `json:"account_category"`
	AccountName     string  `json:"account_name"`
	Amount          float64 `json:"amount"`
}

type BalanceSheetSummary struct {
	MainAccount     string  `json:"main_account"`
	SubAccount      string  `json:"sub_account"`
	AccountCategory string  `json:"account_category"`
	Amount          float64 `json:"amount"`
}

type AccountBalanceForPNL struct {
	ID              int     `json:"id"`
	MainAccount     string  `json:"main_account"`
	SubAccount      string  `json:"sub_account"`
	AccountCategory string  `json:"account_category"`
	AccountName     string  `json:"account_name"`
	Amount          float64 `json:"amount"`
}
