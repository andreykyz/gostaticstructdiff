package debugging_test

import (
	"reflect"
	"testing"

	"github.com/andreykyz/gostaticstructdiff/debugging"
	"github.com/andreykyz/gostaticstructdiff/examples"
)

func TestGetTestStruct_ComplexStruct(t *testing.T) {
	result, err := debugging.GetTestStruct[examples.ComplexStruct](42)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	// Ensure the result is not a zero value for some fields? Not required.
	// Just ensure no panic.
	_ = result
}

func TestGetTestStruct_ComplexStruct_Deterministic(t *testing.T) {
	seed := uint(12345)
	result1, err := debugging.GetTestStruct[examples.ComplexStruct](seed)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	result2, err := debugging.GetTestStruct[examples.ComplexStruct](seed)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	if !reflect.DeepEqual(result1, result2) {
		t.Errorf("Results for same seed differ:\n%+v\nvs\n%+v", result1, result2)
	}
}

func TestGetTestStruct_ComplexStruct_DifferentSeeds(t *testing.T) {
	result1, err := debugging.GetTestStruct[examples.ComplexStruct](1)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	result2, err := debugging.GetTestStruct[examples.ComplexStruct](2)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	// It's possible but extremely unlikely that two different seeds produce identical structs.
	if reflect.DeepEqual(result1, result2) {
		t.Errorf("Results for different seeds are identical: %+v", result1)
	}
}
