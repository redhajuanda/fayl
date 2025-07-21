package fayl

import (
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"
)

type rower interface {
	ColumnTypes() ([]*sql.ColumnType, error)
	Scan(dest ...any) error
	MapScan(dest map[string]any) error
}

// MapScan is like sqlx.Rows.Scan, but instead of a slice of pointers, it takes a map of pointers.
// MapScan maps column names to dest[i] via the same mechanism that Scans uses, so if rows.Scan would
// scan into dest, then rows.MapScan will map into the map.
func MapScan(r rower, dest map[string]any) error {

	// Use MapScan to scan the row into the map
	if err := r.MapScan(dest); err != nil {
		return errors.Wrap(err, "failed to scan map")
	}

	// Serialize the map
	if err := serializeMap(dest); err != nil {
		return err
	}

	return nil

}

// serializeMap converts []byte to map[string]any or []map[string]any or float64
func serializeMap(mapValue map[string]any) error {

	for k, v := range mapValue {
		switch v := v.(type) {
		case []byte:

			if isMap, replacedMap := isMap(v); isMap {
				mapValue[k] = replacedMap
			} else if isSliceOfMap, replacedSlicedOfMap := isSliceOfMap(v); isSliceOfMap {
				mapValue[k] = replacedSlicedOfMap
			} else {
				mapValue[k] = string(v)
			}
		}
	}

	return nil

}

// isMap checks whether the value is a map.
func isMap(v []byte) (bool, map[string]any) {

	var tempMap map[string]any
	err := json.Unmarshal(v, &tempMap)
	if err == nil {
		return true, tempMap
	}
	return false, nil

}

// isSliceOfMap checks whether the value is a slice of map.
func isSliceOfMap(v []byte) (bool, []map[string]any) {

	var tempSliceOfMap []map[string]any
	err := json.Unmarshal(v, &tempSliceOfMap)
	if err == nil {
		return true, tempSliceOfMap
	}
	return false, nil

}
