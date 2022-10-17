package queries

const TrialBalance = `
	SELECT * FROM ((SELECT A.id, MA.name AS main_account, SA.name AS sub_account, AC.name AS account_category, A.name AS account_name, COALESCE(AT.debit-AT.credit, 0) AS debit, 0 AS credit
	FROM account A
	LEFT JOIN (
		SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit
		FROM account_transaction AT
		LEFT JOIN transaction T ON AT.transaction_id = T.id
		WHERE T.posting_date <= '2022-10-17'
		GROUP BY AT.account_id
	) AT ON AT.account_id = A.id
	LEFT JOIN account_category AC on A.account_category_id = AC.id
	LEFT JOIN sub_account SA on AC.sub_account_id = SA.id
	LEFT JOIN main_account MA ON SA.main_account_id = MA.id
	WHERE MA.name IN ('Assets', 'Expenses', 'Cost of Sales')
	ORDER BY main_account, sub_account, account_category ASC, debit DESC)
	UNION
	(SELECT A.id, MA.name AS main_account, SA.name AS sub_account, AC.name AS account_category, A.name AS account_name, 0 AS debit, COALESCE(AT.credit-AT.debit, 0) AS credit
	FROM account A
	LEFT JOIN (
		SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit
		FROM account_transaction AT
		LEFT JOIN transaction T ON AT.transaction_id = T.id
		WHERE T.posting_date <= '2022-10-17'
		GROUP BY AT.account_id
	) AT ON AT.account_id = A.id
	LEFT JOIN account_category AC on A.account_category_id = AC.id
	LEFT JOIN sub_account SA on AC.sub_account_id = SA.id
	LEFT JOIN main_account MA ON SA.main_account_id = MA.id
	WHERE MA.name IN ('Equity', 'Liabilities', 'Revenue', 'Other Revenue')
	ORDER BY main_account, sub_account, account_category, credit ASC)) T ORDER BY FIELD(main_account, "Assets", "Liabilities", "Equity", "Expenses", "Revenue", "Other Revenue"), sub_account, account_category
`

const ChartOfAccounts = `
	SELECT MA.account_id AS main_account_id, MA.name AS main_account, SA.account_id AS sub_account_id, SA.name AS sub_account, AC.account_id AS account_category_id, AC.name AS account_category, A.account_id, A.name AS account_name
	FROM account A
	RIGHT JOIN account_category AC ON AC.id = A.account_category_id
	RIGHT JOIN sub_account SA ON SA.id = AC.sub_account_id
	RIGHT JOIN main_account MA ON MA.id = SA.main_account_id
`

const Transaction = `
	SELECT AT.transaction_id, A.account_id, A.id AS account_id2, A.name AS account_name, AT.type, AT.amount
	FROM account_transaction AT
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE AT.transaction_id = ?
`

const AccountLedger = `
	SELECT A.name, AT.transaction_id, DATE_FORMAT(T.posting_date, '%Y-%m-%d') as posting_date, AT.amount, AT.type, T.remark
	FROM account_transaction AT
	LEFT JOIN account A ON A.id = AT.account_id
	LEFT JOIN transaction T ON T.id = AT.transaction_id
	WHERE AT.account_id = ?
`

const PaymentVouchers = `
	SELECT PV.id, T.datetime, T.posting_date, A.name AS from_account, U.name AS user
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'CR'
	LEFT JOIN account A ON A.id = AT.account_id
	LEFT JOIN user U ON T.user_id = U.id
	ORDER BY T.datetime DESC
`

const PaymentVoucherCheckDetails = `
	SELECT PV.due_date, PV.check_number, PV.payee, T.remark, A.name AS account_name, T.datetime
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'CR'
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE PV.id = ?
`

const PaymentVoucherDetails = `
	SELECT A.account_id, A.name AS account_name, AT.amount, DATE(T.posting_date) as posting_date
	FROM payment_voucher PV
	LEFT JOIN transaction T ON T.id = PV.transaction_id
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id AND AT.type = 'DR'
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE PV.id = ?
`

const JournalEntriesForAudit = `
	SELECT T.datetime, U.name AS issuer, AT.transaction_id, A.name AS account, AT.type, T.posting_date, AT.amount,  T.remark
	FROM transaction T
	LEFT JOIN account_transaction AT ON AT.transaction_id = T.id
	LEFT JOIN user U ON T.user_id = U.id
	LEFT JOIN account A ON A.id = AT.account_id
	WHERE (? IS NULL OR DATE(T.datetime) = ?) AND (? IS NULL OR T.posting_date = ?)
	ORDER BY T.datetime, AT.transaction_id, AT.type DESC, AT.amount ASC
`

const AccountBalancesForReporting = `
	SELECT A.id, MA.name as main_account, SA.name as sub_account, AC.name as account_category, A.name, COALESCE(AT.debit-AT.credit, 0) AS balance 
	FROM account A 
	LEFT JOIN ( SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit FROM (SELECT AT.* FROM account_transaction AT LEFT JOIN transaction T ON T.id = AT.transaction_id WHERE T.posting_date <= ?) AT GROUP BY AT.account_id ) AT ON AT.account_id = A.id 
	LEFT JOIN account_category AC ON AC.id = A.account_category_id 
	LEFT JOIN sub_account SA ON SA.id = AC.sub_account_id 
	LEFT JOIN main_account MA ON MA.id = SA.main_account_id
	WHERE COALESCE(AT.debit-AT.credit, 0) != 0
	ORDER BY FIELD(main_account, "Assets", "Liabilities", "Equity", "Expenses", "Revenue", "Other Revenue"), sub_account, account_category, name, balance DESC
`

const BalanceSheetSummary = `
	SELECT main_account, sub_account, account_category, SUM(balance) AS balance FROM (SELECT A.id, MA.name as main_account, SA.name as sub_account, AC.name as account_category, A.account_id, A.name, COALESCE(AT.debit-AT.credit, 0) AS balance 
	FROM account A 
	LEFT JOIN ( SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit FROM (SELECT AT.* FROM account_transaction AT LEFT JOIN transaction T ON T.id = AT.transaction_id WHERE T.posting_date <= ?) AT GROUP BY AT.account_id ) AT ON AT.account_id = A.id 
	LEFT JOIN account_category AC ON AC.id = A.account_category_id 
	LEFT JOIN sub_account SA ON SA.id = AC.sub_account_id 
	LEFT JOIN main_account MA ON MA.id = SA.main_account_id
	WHERE COALESCE(AT.debit-AT.credit, 0) != 0
	ORDER BY main_account, sub_account, account_category, name, balance DESC) AR
	GROUP BY main_account, sub_account, account_category
	ORDER BY FIELD(main_account, "Assets", "Liabilities", "Equity", "Expenses", "Revenue", "Other Revenue"), sub_account, balance DESC
`

const AccountSummariesForPnl = `
	SELECT id, main_account, sub_account, account_category, name, balance FROM (SELECT A.id, MA.name as main_account, SA.name as sub_account, AC.name as account_category, A.account_id, A.name, COALESCE(AT.debit-AT.credit, 0) AS balance 
	FROM account A 
	LEFT JOIN ( SELECT AT.account_id, SUM(CASE WHEN AT.type = "DR" THEN AT.amount ELSE 0 END) AS debit, SUM(CASE WHEN AT.type = "CR" THEN AT.amount ELSE 0 END) AS credit FROM (SELECT AT.* FROM account_transaction AT LEFT JOIN transaction T ON T.id = AT.transaction_id WHERE T.posting_date BETWEEN ? AND ?) AT GROUP BY AT.account_id ) AT ON AT.account_id = A.id 
	LEFT JOIN account_category AC ON AC.id = A.account_category_id 
	LEFT JOIN sub_account SA ON SA.id = AC.sub_account_id 
	LEFT JOIN main_account MA ON MA.id = SA.main_account_id
	WHERE COALESCE(AT.debit-AT.credit, 0) != 0
	ORDER BY main_account, sub_account, account_category, name, balance DESC) AR
	WHERE main_account IN ("Expenses", "Revenue", "Other Revenue")
	ORDER BY FIELD(main_account, "Expenses", "Revenue", "Other Revenue"), sub_account, account_category, ABS(balance) DESC
`
