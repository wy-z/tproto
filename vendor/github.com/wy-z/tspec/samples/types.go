package samples

import "time"

// BasicTypes defines basic types
type BasicTypes struct {
	BoolField       bool       `json:"bool_field"`
	UintField       uint       `json:"uint_field"`
	Uint8Field      uint8      `json:"uint8_field"`
	Uint16Field     uint16     `json:"uint16_field"`
	Uint32Field     uint32     `json:"uint32_field"`
	Uint64Field     uint64     `json:"uint64_field"`
	IntField        int        `json:"int_field"`
	Int8Field       int8       `json:"int8_field"`
	Int16Field      int16      `json:"int16_field"`
	Int32Field      int32      `json:"int32_field"`
	Int64Field      int64      `json:"int64_field"`
	UintptrField    uintptr    `json:"uintptr_field"`
	Float32Field    float32    `json:"float32_field"`
	Float64Field    float64    `json:"float64_field"`
	StringField     string     `json:"string_field"`
	Complex64Field  complex64  `json:"complex64_field"`
	Complex128Field complex128 `json:"complex128_field"`
	ByteField       byte       `json:"byte_field"`
	RuneField       rune       `json:"rune_field"`
	TimeField       time.Time  `json:"time_field"`
}

// NormalStruct defines normal struct
type NormalStruct struct {
	BasicTypes *BasicTypes `json:"basic_types"`
	Number     int         `json:"number"`
	Create     time.Time   `json:"create"`
}

// StructWithNoExportField defines struct with no export field
type StructWithNoExportField struct {
	number int
	Create time.Time `json:"-"`
}

// StructWithAnonymousField defines struct with anonymous field
type StructWithAnonymousField struct {
	AnonymousStruct *struct {
		StringField string `json:"string_field"`
		BoolField   bool   `json:"bool_field"`
	} `json:"anonymous_struct"`
	AnonymousMap map[string]*struct {
		StringField string `json:"string_field"`
		BoolField   bool   `json:"bool_field"`
	} `json:"anonymous_map"`
	AnonymousArray []*struct {
		StringField string `json:"string_field"`
		BoolField   bool   `json:"bool_field"`
	} `json:"anonymous_array"`
}

// StructWithCircularReference defines struct with circular reference
type StructWithCircularReference struct {
	CircularReference *StructWithCircularReference `json:"circular_reference"`
}

// StructWithInheritance defines struct with inheritance
type StructWithInheritance struct {
	*NormalStruct `json:"normal_struct"`
	*StructWithCircularReference
}

// IntType defines int type
type IntType int

// MapType defines map type
type MapType map[string]*IntType

// InvalidMap defines map with invalid key type
type InvalidMap map[int]*IntType

// ArrayType defines array type
type ArrayType []*IntType
