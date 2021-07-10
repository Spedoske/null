package null

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// NullUint64 represents an uint64 that may be null.
// NullUInt64 implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullUint64 struct {
	Uint64 uint64
	Valid  bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullUint64) Scan(value interface{}) error {
	if value == nil {
		n.Uint64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	valueBytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal uint64 value:", value))
	}
	n.Uint64 = binary.LittleEndian.Uint64(valueBytes)
	return nil
}

// Value implements the driver Valuer interface.
func (n NullUint64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Uint64, nil
}

// Uint is an nullable uint64.
// It does not consider zero values to be null.
// It will decode to null, not zero, if null.
type Uint struct {
	NullUint64
}

// NewUint creates a new Uint
func NewUint(i uint64, valid bool) Uint {
	return Uint{
		NullUint64: NullUint64{
			Uint64: i,
			Valid:  valid,
		},
	}
}

// UintFrom creates a new Uint that will always be valid.
func UintFrom(i uint64) Uint {
	return NewUint(i, true)
}

// UintFromPtr creates a new Uint that be null if i is nil.
func UintFromPtr(i *uint64) Uint {
	if i == nil {
		return NewUint(0, false)
	}
	return NewUint(*i, true)
}

// ValueOrZero returns the inner value if valid, otherwise zero.
func (i Uint) ValueOrZero() uint64 {
	if !i.Valid {
		return 0
	}
	return i.Uint64
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number, string, and null input.
// 0 will not be considered a null Uint.
func (i *Uint) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		i.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &i.Uint64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need uint or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to uint: %w", err)
			}
			i.Uint64 = n
			i.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	i.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Uint if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (i *Uint) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		i.Valid = false
		return nil
	}
	var err error
	i.Uint64, err = strconv.ParseUint(string(text), 10, 64)
	if err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	i.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Uint is null.
func (i Uint) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatUint(i.Uint64, 10)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Uint is null.
func (i Uint) MarshalText() ([]byte, error) {
	if !i.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatUint(i.Uint64, 10)), nil
}

// SetValid changes this Uint's value and also sets it to be non-null.
func (i *Uint) SetValid(n uint64) {
	i.Uint64 = n
	i.Valid = true
}

// Ptr returns a pointer to this Uint's value, or a nil pointer if this Uint is null.
func (i Uint) Ptr() *uint64 {
	if !i.Valid {
		return nil
	}
	return &i.Uint64
}

// IsZero returns true for invalid Uints, for future omitempty support (Go 1.4?)
// A non-null Uint with a 0 value will not be considered zero.
func (i Uint) IsZero() bool {
	return !i.Valid
}

// Equal returns true if both uints have the same value or are both null.
func (i Uint) Equal(other Uint) bool {
	return i.Valid == other.Valid && (!i.Valid || i.Uint64 == other.Uint64)
}
