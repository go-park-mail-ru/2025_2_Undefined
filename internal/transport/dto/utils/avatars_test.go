package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringMapToPointerMap_Success(t *testing.T) {
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result := StringMapToPointerMap(input)

	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.NotNil(t, result["key1"])
	assert.Equal(t, "value1", *result["key1"])
	assert.NotNil(t, result["key2"])
	assert.Equal(t, "value2", *result["key2"])
}

func TestStringMapToPointerMap_EmptyString(t *testing.T) {
	input := map[string]string{
		"key1": "",
		"key2": "value2",
	}

	result := StringMapToPointerMap(input)

	assert.NotNil(t, result)
	assert.Nil(t, result["key1"])
	assert.NotNil(t, result["key2"])
	assert.Equal(t, "value2", *result["key2"])
}

func TestStringMapToPointerMap_Nil(t *testing.T) {
	result := StringMapToPointerMap(nil)
	assert.Nil(t, result)
}

func TestStringMapToPointerMap_Empty(t *testing.T) {
	input := map[string]string{}
	result := StringMapToPointerMap(input)

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestPointerMapToStringMap_Success(t *testing.T) {
	val1 := "value1"
	val2 := "value2"
	input := map[string]*string{
		"key1": &val1,
		"key2": &val2,
	}

	result := PointerMapToStringMap(input)

	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
}

func TestPointerMapToStringMap_NilPointer(t *testing.T) {
	val1 := "value1"
	input := map[string]*string{
		"key1": nil,
		"key2": &val1,
	}

	result := PointerMapToStringMap(input)

	assert.NotNil(t, result)
	assert.Equal(t, "", result["key1"])
	assert.Equal(t, "value1", result["key2"])
}

func TestPointerMapToStringMap_Nil(t *testing.T) {
	result := PointerMapToStringMap(nil)
	assert.Nil(t, result)
}

func TestPointerMapToStringMap_Empty(t *testing.T) {
	input := map[string]*string{}
	result := PointerMapToStringMap(input)

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}
