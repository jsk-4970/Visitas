package repository

import (
	"os"
	"regexp"
	"sort"

	"cloud.google.com/go/spanner"
)

// IsEmulatorMode returns true if running against Spanner Emulator
func IsEmulatorMode() bool {
	return os.Getenv("SPANNER_EMULATOR_HOST") != ""
}

// ConvertQueryForDialect converts a Google SQL dialect query to PostgreSQL dialect
// when running on the emulator. Returns the original query and params when not on emulator.
//
// Google SQL: WHERE patient_id = @patientID
// PostgreSQL: WHERE patient_id = $1 (with params keyed as "p1")
func ConvertQueryForDialect(sql string, params map[string]interface{}) (string, map[string]interface{}) {
	if !IsEmulatorMode() {
		return sql, params
	}

	return convertToPostgresDialect(sql, params)
}

// convertToPostgresDialect converts @paramName syntax to $N syntax
// and remaps parameter keys to p1, p2, etc.
func convertToPostgresDialect(sql string, params map[string]interface{}) (string, map[string]interface{}) {
	if len(params) == 0 {
		return sql, params
	}

	// Find all @paramName patterns in SQL
	paramRegex := regexp.MustCompile(`@(\w+)`)
	matches := paramRegex.FindAllStringSubmatch(sql, -1)

	if len(matches) == 0 {
		return sql, params
	}

	// Get unique param names in order of first appearance
	seen := make(map[string]bool)
	var orderedParams []string
	for _, match := range matches {
		paramName := match[1]
		if !seen[paramName] {
			seen[paramName] = true
			orderedParams = append(orderedParams, paramName)
		}
	}

	// Sort for deterministic ordering (important for consistent $N assignment)
	sort.Strings(orderedParams)

	// Create mapping from original param name to position
	paramToPos := make(map[string]int)
	for i, paramName := range orderedParams {
		paramToPos[paramName] = i + 1
	}

	// Replace @paramName with $N in SQL
	convertedSQL := paramRegex.ReplaceAllStringFunc(sql, func(match string) string {
		paramName := match[1:] // Remove @ prefix
		pos := paramToPos[paramName]
		return "$" + itoa(pos)
	})

	// Convert params map keys from paramName to pN
	convertedParams := make(map[string]interface{})
	for paramName, value := range params {
		pos := paramToPos[paramName]
		if pos > 0 {
			convertedParams["p"+itoa(pos)] = value
		}
	}

	return convertedSQL, convertedParams
}

// itoa converts int to string (simple implementation to avoid strconv import)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// NewStatement creates a spanner.Statement with automatic dialect conversion
func NewStatement(sql string, params map[string]interface{}) spanner.Statement {
	convertedSQL, convertedParams := ConvertQueryForDialect(sql, params)
	return spanner.Statement{
		SQL:    convertedSQL,
		Params: convertedParams,
	}
}
