package fayl

import (
	"reflect"
	"regexp"
	"strings"
)

// isStruct checks if the given interface is a struct or not
func isStruct(i interface{}) bool {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference the pointer
	}
	return t.Kind() == reflect.Struct
}

// getQueryType identifies the type of SQL query
func getQueryType(sql string) string {

	// Trim leading and trailing spaces and convert to uppercase
	sql = strings.TrimSpace(sql)
	sql = strings.ToUpper(sql)

	// check if the query has returning
	if strings.Contains(sql, "RETURNING") {
		return "SELECT"
	}

	if strings.HasPrefix(sql, "WITH") {

		// Match the `WITH` clause including the CTE body and the closing parenthesis
		re := regexp.MustCompile(`(?i)WITH\s+[\s\S]*?\)`)
		// Remove everything that matches the CTE pattern
		remaining := re.ReplaceAllString(sql, "")

		return getQueryType(remaining) // Recurse to analyze the main query
	}

	switch {
	case strings.HasPrefix(sql, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(sql, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(sql, "DELETE"):
		return "DELETE"
	case strings.HasPrefix(sql, "SELECT"):
		return "SELECT"
	default:
		return "UNKNOWN"
	}
}
