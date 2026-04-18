package debugging

import (
	go_fuzz_utils "github.com/trailofbits/go-fuzz-utils"
)

func GetTestStruct[T any](rnd uint) (T, error) {
	b := generateTestData(0x1000, rnd)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	if err != nil {
		return *new(T), err
	}

	// Create a test structure and fill it.
	var n1 T
	err = tp.SetParamsBiasesCommon(0, 0)
	if err != nil {
		return *new(T), err
	}
	tp.SetParamsFillUnexportedFields(false)
	err = tp.Fill(&n1)
	if err != nil {
		return *new(T), err
	}
	return n1, nil
}

func generateTestData(length, rnd uint) []byte {
	// Create our test data
	b := make([]byte, length)
	for i := 0; i < len(b); i++ {
		b[i] = 65 + (123 - 65) - byte((uint(i)+rnd)%(123-65))
	}

	return b
}
