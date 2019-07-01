package gin_helper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	uuid "github.com/gofrs/uuid"
)

func TestLookupJsonKey(t *testing.T) {
	v := gin.H{
		"a": uuid.Must(uuid.NewV4()).String(),
		"b": []interface{}{
			gin.H{
				"id": "btc",
			},
		},
	}

	data, _ := jsoniter.Marshal(v)

	keys := []string{"a", "b", "id"}

	for {
		idx, key := lookupJsonKey(data)
		if idx == -1 {
			break
		}

		if assert.NotEmpty(t, keys) && assert.Equal(t, keys[0], key) {
			keys = keys[1:]
		}

		data = data[idx:]
	}
}

func TestCamelEncoder(t *testing.T) {
	v := gin.H{
		"abc_url": uuid.Must(uuid.NewV4()).String(),
		"bcd_id": []interface{}{
			gin.H{
				"id": "btc",
			},
		},
	}

	data, _ := jsoniter.Marshal(v)
	data = TransformJsonKeys(data, ToCamelKey)

	raw := json.RawMessage{}
	assert.Nil(t, jsoniter.Unmarshal(data, &raw))

	keys := []string{"abcUrl", "bcdId", "id"}

	for {
		idx, key := lookupJsonKey(data)
		if idx == -1 {
			break
		}

		if assert.NotEmpty(t, keys) && assert.Equal(t, keys[0], key) {
			keys = keys[1:]
		}

		data = data[idx:]
	}
}

func BenchmarkCamelEncoder(b *testing.B) {
	view := map[string]interface{}{
		"asset_id":      "asset_id",
		"refresh_token": "refresh_token",
		"assets": map[string]interface{}{
			"logo":        "logo",
			"base_symbol": "btc",
		},
	}

	data, _ := jsoniter.Marshal(view)

	b.Run("camel encoder without whitewords", func(b *testing.B) {
		data := TransformJsonKeys(data, ToCamelKey)
		assert.NotEmpty(b, data)
	})

	b.Run("camel encode with whitewords", func(b *testing.B) {
		data := TransformJsonKeys(data, ToCamelKey.WithWhitelist("refresh_token"))
		assert.NotEmpty(b, data)
	})
}
