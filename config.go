package migration

type Configuration struct {
	ApplicationName      string `toml:app-name`
	ExtractFileDirectory string `toml:extract-file-directory`
	ETLDatabaseURL       string `toml:etl-database-url`
	ImportDatabaseURL    string `toml:import-database-url`
	APIURIPrefix         string `toml:api-uri-prefix`
}
