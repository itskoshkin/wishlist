package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"

	"wishlist/internal/utils/colors"
	"wishlist/pkg/postgres"
	"wishlist/pkg/redis"
)

const (
	LogLevel       = "app.log.level"           // string ("DEBUG", "INFO", "WARN", "ERROR")
	LogFormat      = "app.log.log_format"      // string ("text" or "json")
	LogToConsole   = "app.log.log2console"     // bool
	LogToFile      = "app.log.log2file"        // bool
	LogFilePath    = "app.log.file_path"       // string (path)
	LogFileMode    = "app.log.file_mode"       // string ("append", "overwrite", "rotate")
	LogFilesFolder = "app.log.old_logs_folder" // string (path)

	ApiHost            = "app.api.host"
	ApiPort            = "app.api.port"
	ApiBasePath        = "app.api.base_path"
	GinReleaseMode     = "app.api.gin_release_mode"
	ApiShutdownTimeout = "app.api.shutdown_timeout"

	WebAppDomain = "app.webapp.domain"

	JwtIssuer           = "app.api.auth.jwt_issuer"
	JwtAudience         = "app.api.auth.jwt_audience"
	AccessTokenSecret   = "app.api.auth.access_token_secret"
	RefreshTokenSecret  = "app.api.auth.refresh_token_secret"
	AccessTokenTTL      = "app.api.auth.access_token_ttl"
	RefreshTokenTTL     = "app.api.auth.refresh_token_ttl"
	PwdResetTokenTTL    = "app.api.auth.pwd_reset_token_ttl"
	EmailVerifyTokenTTL = "app.api.auth.email_verify_token_ttl"

	DatabaseHost     = "app.database.host"
	DatabasePort     = "app.database.port"
	DatabaseUser     = "app.database.user"
	DatabasePassword = "app.database.password"
	DatabaseName     = "app.database.database_name"
	DatabaseSslMode  = "app.database.ssl_mode"

	RedisHost     = "app.redis.host"
	RedisPort     = "app.redis.port"
	RedisPassword = "app.redis.password"
	RedisDB       = "app.redis.database"

	EmailHost     = "app.email.host"
	EmailPort     = "app.email.port"
	EmailUser     = "app.email.user"
	EmailPassword = "app.email.password"
	EmailFrom     = "app.email.from"
)

func LoadConfig() {
	fmt.Print("Loading configuration...")

	viper.SetConfigFile("./config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to read configuration: %v", err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := ValidateConfigFields(); err != nil {
		fmt.Println()
		log.Fatalf("Fatal: failed to load configuration: %v", err)
	}

	fmt.Println(colors.Green("  Done."))
}

func ValidateConfigFields() error {
	var required = []string{ // Must be present and non-empty
		DatabaseHost, DatabasePort, DatabaseUser, DatabasePassword,
		ApiPort, AccessTokenSecret, RefreshTokenSecret, JwtIssuer, JwtAudience,
	}
	var dependent = map[string][]string{ // If A=true => must be non-empty B (, C...)
		LogToFile: {LogFilePath},
		EmailHost: {EmailUser, EmailPassword, EmailFrom},
	}
	var possibleValues = map[string][]string{ // If present, must be one of these values
		LogLevel:    {"DEBUG", "INFO", "WARN", "ERROR"},
		LogFormat:   {"text", "json"},
		LogFileMode: {"append", "overwrite", "rotate"},
	}
	var defaults = map[string]any{ // Will be set if not present, overwrites above required/dependent
		/* Log */ LogLevel: "INFO", LogFormat: "text", LogToConsole: true, LogToFile: true, LogFilePath: "application.log", LogFileMode: "append",
		/* Postgres */ DatabaseHost: "localhost", DatabasePort: 5432, DatabaseUser: "postgres", DatabaseName: "wishlist", DatabaseSslMode: "disable",
		/* Redis */ RedisHost: "localhost", RedisPort: 6379, RedisDB: 0,
		/* API */ ApiBasePath: "/api/v1", ApiShutdownTimeout: "5s",
		/* JWT */ AccessTokenTTL: "24h", RefreshTokenTTL: "168h" /* 7 days */, PwdResetTokenTTL: "1h", JwtIssuer: "wishlist", JwtAudience: "Wishlist API",
		/* Email */ EmailPort: "587" /* Default port */, EmailVerifyTokenTTL: "24h",
	}

	for k, v := range defaults {
		if !viper.IsSet(k) {
			viper.Set(k, v)
		}
	}

	var missing []string
	for _, key := range required {
		if !viper.IsSet(key) || strings.TrimSpace(viper.GetString(key)) == "" {
			missing = append(missing, key)
		}
	}
	if viper.GetString(LogFileMode) == "rotate" {
		if !viper.IsSet(LogFilesFolder) || strings.TrimSpace(viper.GetString(LogFilesFolder)) == "" {
			missing = append(missing, fmt.Sprintf("%s (required when %s=rotate)", LogFilesFolder, LogFileMode))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields/values in config: %s", strings.Join(missing, ", "))
	}

	for triggerKey, requiredKeys := range dependent {
		if viper.GetBool(triggerKey) || (viper.IsSet(triggerKey) && viper.GetString(triggerKey) != "") {
			for _, key := range requiredKeys {
				if strings.TrimSpace(viper.GetString(key)) == "" {
					missing = append(missing, fmt.Sprintf("%s (%s is set)", key, triggerKey))
				}
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields/values in config: %s", strings.Join(missing, ", "))
	}

	var invalid []string
	for key, allowed := range possibleValues {
		if !viper.IsSet(key) {
			continue
		}
		val := strings.TrimSpace(viper.GetString(key))
		if val == "" {
			continue
		}
		found := false
		for _, a := range allowed {
			if val == a {
				found = true
				break
			}
		}
		if !found {
			invalid = append(invalid, fmt.Sprintf("'%s' for '%s' (must be one of [%s])", val, key, strings.Join(allowed, ", ")))
		}
	}
	for _, key := range []string{ApiShutdownTimeout, AccessTokenTTL, RefreshTokenTTL, PwdResetTokenTTL, EmailVerifyTokenTTL} {
		if viper.GetDuration(key) <= 0 {
			invalid = append(invalid, fmt.Sprintf("%s (duration must be >0, got '%s')", key, viper.GetString(key)))
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid config values: %s", strings.Join(invalid, ", "))
	}

	return nil
}

func PostgresConfig() postgres.Config {
	return postgres.Config{
		Host:     viper.GetString(DatabaseHost),
		Port:     viper.GetString(DatabasePort),
		User:     viper.GetString(DatabaseUser),
		Password: viper.GetString(DatabasePassword),
		Database: viper.GetString(DatabaseName),
		SSLMode:  viper.GetString(DatabaseSslMode),
		LogLevel: viper.GetString(LogLevel),
	}
}

func RedisConfig() redis.Config {
	return redis.Config{
		Addr:     viper.GetString(RedisHost),
		Port:     viper.GetString(RedisPort),
		Password: viper.GetString(RedisPassword),
		Database: viper.GetInt(RedisDB),
	}
}
