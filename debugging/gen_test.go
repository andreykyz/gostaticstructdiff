package debugging

import (
	"reflect"
	"testing"
)

func TestGetTestStruct_Simple(t *testing.T) {
	type SimpleStruct struct {
		X int
		Y string
		Z bool
	}

	var result SimpleStruct
	var err error
	result, err = GetTestStruct[SimpleStruct](42)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	// Basic sanity check: struct is filled (zero values maybe? but go-fuzz-utils may fill with non-zero)
	// We can't assert specific values because they're random.
	// Just ensure no panic.
	_ = result
}

func TestGetTestStruct_Deterministic(t *testing.T) {
	type Point struct {
		X float64
		Y float64
	}

	seed := uint(12345)
	result1, err := GetTestStruct[Point](seed)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	result2, err := GetTestStruct[Point](seed)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	if result1 != result2 {
		t.Errorf("Results for same seed differ: %+v vs %+v", result1, result2)
	}
}

func TestGetTestStruct_DifferentSeedsProduceDifferent(t *testing.T) {
	type Data struct {
		A int
		B string
		C []byte
	}

	result1, err := GetTestStruct[Data](1)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	result2, err := GetTestStruct[Data](2)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	// It's possible but extremely unlikely that two different seeds produce identical structs.
	// We'll compare using reflect.DeepEqual because Data contains a slice.
	if reflect.DeepEqual(result1, result2) {
		t.Errorf("Results for different seeds are identical: %+v", result1)
	}
}

func TestGetTestStruct_ComplexType(t *testing.T) {
	type Nested struct {
		ID   int
		Name string
	}
	type Complex struct {
		Slice    []int
		Map      map[string]float64
		Ptr      *Nested
		Embedded Nested
	}

	result, err := GetTestStruct[Complex](999)
	if err != nil {
		t.Fatalf("GetTestStruct failed: %v", err)
	}
	_ = result
}

func TestGenerateTestData_Length(t *testing.T) {
	length := uint(100)
	rnd := uint(7)
	data := generateTestData(length, rnd)
	if uint(len(data)) != length {
		t.Errorf("Expected length %d, got %d", length, len(data))
	}
}

func TestGenerateTestData_Deterministic(t *testing.T) {
	data1 := generateTestData(50, 42)
	data2 := generateTestData(50, 42)
	if len(data1) != len(data2) {
		t.Fatalf("Length mismatch")
	}
	for i := range data1 {
		if data1[i] != data2[i] {
			t.Errorf("Byte at index %d differs: %d vs %d", i, data1[i], data2[i])
		}
	}
}

func TestGenerateTestData_DifferentSeeds(t *testing.T) {
	data1 := generateTestData(100, 1)
	data2 := generateTestData(100, 2)
	// Very unlikely to be identical
	identical := true
	for i := range data1 {
		if data1[i] != data2[i] {
			identical = false
			break
		}
	}
	if identical {
		t.Errorf("Byte slices for different seeds are identical")
	}
}
