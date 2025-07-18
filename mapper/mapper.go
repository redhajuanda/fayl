package mapper

import (
	"github.com/redhajuanda/fayl/vars"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// Decode decodes the input into the output
func Decode(input any, output any) error {

	cfg := &mapstructure.DecoderConfig{
		TagName: vars.TagKey,
		Result:  output,
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
