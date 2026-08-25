package persistence

const CurrentSchemaVersion = 1

func CompatibleSchema(v int) bool { return v == CurrentSchemaVersion }
