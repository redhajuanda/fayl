package mapper

import (
	"reflect"
	"time"

	"github.com/redhajuanda/fayl/vars"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// timeHook prevents time.Time from being converted to map
func timeHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		// Handle time.Time preservation in all cases
		switch {
		case from == reflect.TypeOf(time.Time{}):
			return data, nil
		case from == reflect.TypeOf(&time.Time{}):
			if ptr, ok := data.(*time.Time); ok && ptr != nil {
				return *ptr, nil
			}
		case from.Kind() == reflect.Struct &&
			to.Kind() == reflect.Map &&
			to.Key().Kind() == reflect.String:
			// Special handling for struct to map conversion
			if rv := reflect.ValueOf(data); rv.IsValid() && rv.Kind() == reflect.Struct {
				result := make(map[string]interface{})
				rt := rv.Type()

				for i := 0; i < rv.NumField(); i++ {
					field := rt.Field(i)
					fieldValue := rv.Field(i)

					if !fieldValue.CanInterface() {
						continue
					}

					// Get the tag name
					tagName := field.Tag.Get(vars.TagKey)
					if tagName == "" {
						tagName = field.Name
					}

					// If this is a time.Time field, preserve it
					if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
						result[tagName] = fieldValue.Interface()
					} else {
						// Let mapstructure handle other fields normally
						result[tagName] = fieldValue.Interface()
					}
				}
				return result, nil
			}
		}

		return data, nil
	}
}

// Decode decodes the input into the output
func Decode(input any, output any) error {

	cfg := &mapstructure.DecoderConfig{
		TagName:    vars.TagKey,
		Result:     output,
		DecodeHook: timeHook(),
	}

	// init decoder
	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return errors.Wrap(err, "cannot init decoder")
	}

	// decode
	err = dec.Decode(input)
	if err != nil {
		return errors.Wrap(err, "cannot decode dest")
	}

	return nil

}
