package migration

func CorrectTimeStamps(record Record) Record {
	switch {
	case record["created_at"] == "" && record["updated_at"] != "":
		record["created_at"] = record["updated_at"]
		return record
	case record["updated_at"] == "" && record["created_at"] != "":
		record["updated_at"] = record["created_at"]
		return record
	}
	return record
}

func RenameFields(fieldMap map[string]string) func(Record) Record {
	return func(record Record) Record {
		for oldField, newField := range fieldMap {
			record[newField] = record[oldField]
			delete(record, oldField)
		}
		return record
	}
}

func RemoveFields(fields []string) func(Record) Record {
	return func(record Record) Record {
		for _, field := range fields {
			delete(record, field)
		}
		return record
	}
}

func PartitionByBTToken(record Record) int {
	btToken, ok := record["braintree_token"]
	if !ok {
		return PROCESS_RECORD_AS_IMPORT
	}

	if btToken != "" {
		return PROCESS_RECORD_AS_IMPORT
	} else {
		return PROCESS_RECORD_AS_CREATE
	}
}
