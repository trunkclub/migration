package migration

type CustomerAccount struct {}

func (c CustomerAccount) ExtractFileName() string {
	return "members"
}

func (c CustomerAccount) PreTransformFns() []func(Record) Record {
	return []func(Record) Record {CorrectTimeStamps}
}

func (c CustomerAccount) PartitionFn() func(Record) Record {
	return PartitionByBTToken
}

func (c CustomerAccount) ImportTransformFns() []func(Record) Record {
	return []func(Record) Record {RemoveFields([]string{
		"first_name",
		"last_name",
		"email",
		"phone",
	})}
}

func (c CustomerAccount) CreateTransformFns() []func(Record) Record {
	return []func(Record) Record{}
}

func (c CustomerAccount) ImportFn(connInfo ConnectionInfo) func(Record) Result {
	stmt := NewInsertStatement(
		connInfo.ImportDB,
		"customer_accounts",
		[]string{
			"member_id",
			"braintree_token",
			"created_at",
			"updated_at",
		},
	)
	return func(record Record) Result {
		return stmt.Execute(record)
	}
}

func (c CustomerAccount) CreateFn(connInfo ConnectionInfo) func(Record) Result {
	client := connInfo.APIClient
	return func(record Record) Result {
		memberId, err := ExtractId(record, "member_id")
		if err != nil {
			return CreateResult(false, record, err, nil)
		}
		account, err := client.Finance.CreateCustomerAccount(
			memberId,
			record["first_name"].(string),
			record["last_name"].(string),
			record["email"].(string),
			record["phone"].(string),
		)
		if err != nil {
			return CreateResult(false, record, err, nil)
		}
		return CreateResult(true, record, nil, &Result{"id": account.Id})
	}
}

func (c CustomerAccount) PostProcessFn(connInfo ConnectionInfo) func(Result) {
	etlDB := connInfo.ETLDB
	stmt := NewInsertStatement(
		etlDB,
		"member_customer_accounts",
		[]string{"member_id", "customer_account_id"},
	)
	return func(result Result) {
		xlatRecord := Record{
			"member_id": result.input["member_id"],
			"customer_account_id": result.output["id"],
		}
		return stmt.Execute(xlatRecord)
	}
}
