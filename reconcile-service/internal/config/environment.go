package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode"
)

const (
	configServiceBaseURL = "http://config-service"
	serviceID            = "reconcile-service"
	serviceRole          = "admin"
)

type authResponse struct {
	Token string `json:"token"`
}

func getConfigDetailsFromConfigService() (map[string]interface{}, error) {
	filePath := os.Getenv("CONFIG_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
	}
	filePath += fmt.Sprintf("/%s/config.json", serviceID)

	// --- Auth ---
	authBody, _ := json.Marshal(map[string]string{"serviceId": serviceID, "role": serviceRole})
	authResp, err := http.Post(
		fmt.Sprintf("%s/auth", configServiceBaseURL),
		"application/json",
		strings.NewReader(string(authBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with config service: %w", err)
	}
	defer authResp.Body.Close()
	if authResp.StatusCode != 200 {
		return nil, fmt.Errorf("authentication failed with status code: %d", authResp.StatusCode)
	}
	var auth authResponse
	if err := json.NewDecoder(authResp.Body).Decode(&auth); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}
	if auth.Token == "" {
		return nil, fmt.Errorf("no token received from config service")
	}

	// --- Fetch config ---
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/secure/services/%s/configs/%s", configServiceBaseURL, serviceID, filePath), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))
	configResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config from config service: %w", err)
	}
	defer configResp.Body.Close()
	if configResp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch config with status code: %d", configResp.StatusCode)
	}

	body, _ := io.ReadAll(configResp.Body)
	bodyStr := string(body)

	// Handle double-encoded JSON (body is a quoted JSON string)
	if strings.HasPrefix(bodyStr, "\"") && strings.HasSuffix(bodyStr, "\"") {
		var unquoted string
		if err := json.Unmarshal(body, &unquoted); err == nil {
			body = []byte(unquoted)
			bodyStr = unquoted
		}
	}

	// Fix malformed KAFKA_HOST array-in-string
	if strings.Contains(bodyStr, "\"KAFKA_HOST\": [\"[\"") {
		bodyStr = strings.ReplaceAll(bodyStr, "\"KAFKA_HOST\": [\"[\"", "\"KAFKA_HOST\": [\"")
		bodyStr = strings.ReplaceAll(bodyStr, "\"]\"]", "\"]")
		body = []byte(bodyStr)
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(body, &configMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config response: %w, body: %s", err, string(body))
	}
	return configMap, nil
}

// setEnvFromConfig recursively walks the config map and calls os.Setenv.
// Arrays are JSON-encoded back to string (e.g. KAFKA_HOST = '["broker:9094"]').
func setEnvFromConfig(config map[string]interface{}, prefix string) error {
	for key, value := range config {
		envKey := key
		if prefix != "" {
			envKey = fmt.Sprintf("%s_%s", prefix, key)
		}
		envKey = toEnvVarName(envKey)

		switch v := value.(type) {
		case map[string]interface{}:
			if err := setEnvFromConfig(v, envKey); err != nil {
				return err
			}
		case []interface{}:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal array for key %s: %w", envKey, err)
			}
			if err := os.Setenv(envKey, string(jsonBytes)); err != nil {
				return fmt.Errorf("failed to set environment variable %s: %w", envKey, err)
			}
		default:
			valueStr := fmt.Sprintf("%v", v)
			if err := os.Setenv(envKey, valueStr); err != nil {
				return fmt.Errorf("failed to set environment variable %s: %w", envKey, err)
			}
		}
	}
	return nil
}

// toEnvVarName converts camelCase, kebab-case, dot-notation to UPPER_SNAKE_CASE.
func toEnvVarName(s string) string {
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToUpper(r))
	}
	return result.String()
}

// LoadEnvironment is the public entry point called in main().
func LoadEnvironment() error {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		return fmt.Errorf("NODE_ENV environment variable is not set")
	}

	// Skip config service in local/test environments
	if env == "test" || env == "local" {
		fmt.Printf("Skipping config service for environment: %s\n", env)
		if os.Getenv("AWS_REGION") == "" {
			return fmt.Errorf("AWS_REGION is not set")
		}
		return nil
	}

	config, err := getConfigDetailsFromConfigService()
	if err != nil {
		return fmt.Errorf("failed to get config from config service: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no config found for config service path: %s", os.Getenv("CONFIG_PATH"))
	}
	if err := setEnvFromConfig(config, ""); err != nil {
		return fmt.Errorf("failed to set environment variables from config: %w", err)
	}
	if os.Getenv("AWS_REGION") == "" {
		return fmt.Errorf("AWS_REGION is not set")
	}
	return nil
}
