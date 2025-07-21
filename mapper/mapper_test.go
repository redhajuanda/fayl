package mapper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Parallel()

	t.Run("Success decoding struct", func(t *testing.T) {
		t.Parallel()

		var input struct {
			Key       string    `fayl:"key"`
			CreatedAt time.Time `fayl:"created_at"`
		}

		var output = make(map[string]interface{})

		input.Key = "value"
		input.CreatedAt = time.Now()

		err := Decode(input, &output)
		assert.NoError(t, err)
		assert.Equal(t, "value", output["key"])
		assert.Equal(t, input.CreatedAt, output["created_at"])

	})

	t.Run("Success decoding", func(t *testing.T) {
		t.Parallel()

		input := map[string]interface{}{
			"key":          "value",
			"full_address": "Jl. Kebon Sirih No. 1",
		}

		var output struct {
			Key         string `fayl:"key"`
			FullAddress string `fayl:"full_address"`
		}

		err := Decode(input, &output)
		assert.NoError(t, err)
		assert.Equal(t, "value", output.Key)
		assert.Equal(t, "Jl. Kebon Sirih No. 1", output.FullAddress)
	})

	t.Run("Failed decoding", func(t *testing.T) {
		t.Parallel()

		input := map[string]interface{}{
			"key": make(chan int),
		}

		var output struct {
			Key         string `fayl:"key"`
			FullAddress string `fayl:"full_address"`
		}

		err := Decode(input, &output)
		assert.Error(t, err)
	})
}
