package gostaticstructdiff

/*
// T - NestedValue type will generate in place for sctruct types
type NestedValue struct {
	Value T
	Set   bool
}
*/
// Or
/*
// T - NestedValue type will generate in place for sctruct types
type NestedValuePtr struct {
	Value *T
	Set   bool
}
*/

/*
Value of type T will be generated in place for sctruct types
And can be simple value or pointer and another NestedValue type
*/
