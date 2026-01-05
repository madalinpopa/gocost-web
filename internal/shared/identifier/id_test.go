package identifier

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestID_Equals(t *testing.T) {
	type fields struct {
		value uuid.UUID
	}
	type args struct {
		other ID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "returns true when IDs have same UUID value",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			args: args{
				other: ID{value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")},
			},
			want: true,
		},
		{
			name: "returns false when IDs have different UUID values",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			args: args{
				other: ID{value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")},
			},
			want: false,
		},
		{
			name: "returns true when both IDs are zero/nil UUIDs",
			fields: fields{
				value: uuid.Nil,
			},
			args: args{
				other: ID{value: uuid.Nil},
			},
			want: true,
		},
		{
			name: "returns false when one ID is nil and other is not",
			fields: fields{
				value: uuid.Nil,
			},
			args: args{
				other: ID{value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")},
			},
			want: false,
		},
		{
			name: "returns false when first ID has value and other is nil",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			args: args{
				other: ID{value: uuid.Nil},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := ID{
				value: tt.fields.value,
			}
			got := i.Equals(tt.args.other)
			assert.Equal(t, tt.want, got, "Equals() should return %v", tt.want)
		})
	}
}

func TestID_IsZero(t *testing.T) {
	type fields struct {
		value uuid.UUID
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "returns true when ID has nil UUID value",
			fields: fields{
				value: uuid.Nil,
			},
			want: true,
		},
		{
			name: "returns false when ID has valid UUID value",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			want: false,
		},
		{
			name: "returns true when ID has zero UUID value",
			fields: fields{
				value: uuid.UUID{}, // Zero value of UUID is equivalent to uuid.Nil
			},
			want: true,
		},
		{
			name: "returns false when ID has generated UUID value",
			fields: fields{
				value: uuid.Must(uuid.NewV7()),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := ID{
				value: tt.fields.value,
			}
			got := i.IsZero()
			assert.Equal(t, tt.want, got, "IsZero() should return %v", tt.want)
		})
	}
}

func TestID_String(t *testing.T) {
	type fields struct {
		value uuid.UUID
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns string representation of valid UUID",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			want: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name: "returns string representation of nil UUID",
			fields: fields{
				value: uuid.Nil,
			},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "returns string representation of zero UUID",
			fields: fields{
				value: uuid.UUID{},
			},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "returns string representation of another valid UUID",
			fields: fields{
				value: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			want: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name: "returns string representation of max UUID",
			fields: fields{
				value: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			},
			want: "ffffffff-ffff-ffff-ffff-ffffffffffff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := ID{
				value: tt.fields.value,
			}
			got := i.String()
			assert.Equal(t, tt.want, got, "String() should return %v", tt.want)
		})
	}
}

func TestID_UUID(t *testing.T) {
	type fields struct {
		value uuid.UUID
	}
	tests := []struct {
		name   string
		fields fields
		want   uuid.UUID
	}{
		{
			name: "returns underlying UUID value",
			fields: fields{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			want: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name: "returns nil UUID",
			fields: fields{
				value: uuid.Nil,
			},
			want: uuid.Nil,
		},
		{
			name: "returns zero UUID",
			fields: fields{
				value: uuid.UUID{},
			},
			want: uuid.UUID{},
		},
		{
			name: "returns another valid UUID",
			fields: fields{
				value: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			want: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "returns max UUID",
			fields: fields{
				value: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			},
			want: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := ID{
				value: tt.fields.value,
			}
			got := i.UUID()
			assert.Equal(t, tt.want, got, "UUID() should return %v", tt.want)
		})
	}
}

func TestNewID(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successfully creates new ID with valid UUID",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewID()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify that we got a valid ID
			assert.NotEqual(t, ID{}, got, "NewID() should return a non-zero ID")
			assert.NotEqual(t, uuid.Nil, got.UUID(), "NewID() should return a non-nil UUID")

			// Verify UUID is valid (not zero)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", got.String())

			// Test that multiple calls return different IDs
			got2, err2 := NewID()
			assert.NoError(t, err2)
			assert.NotEqual(t, got, got2, "NewID() should return unique IDs on each call")

			// Verify the returned ID has a proper string format
			idStr := got.String()
			assert.Len(t, idStr, 36, "UUID string should be 36 characters long")
			assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, idStr, "Should match UUID format")
		})
	}
}

func TestParseID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    ID
		wantErr bool
	}{
		{
			name: "successfully parses valid UUID string",
			args: args{
				s: "550e8400-e29b-41d4-a716-446655440000",
			},
			want: ID{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			wantErr: false,
		},
		{
			name: "successfully parses nil UUID string",
			args: args{
				s: "00000000-0000-0000-0000-000000000000",
			},
			want: ID{
				value: uuid.Nil,
			},
			wantErr: false,
		},
		{
			name: "successfully parses another valid UUID string",
			args: args{
				s: "123e4567-e89b-12d3-a456-426614174000",
			},
			want: ID{
				value: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			wantErr: false,
		},
		{
			name: "successfully parses UUID with uppercase letters",
			args: args{
				s: "550e8400-e29b-41d4-a716-446655440000",
			},
			want: ID{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			wantErr: false,
		},
		{
			name: "fails to parse invalid UUID format - too short",
			args: args{
				s: "550e8400-e29b-41d4-a716-44665544000",
			},
			want:    ID{},
			wantErr: true,
		},
		{
			name: "fails to parse invalid UUID format - too long",
			args: args{
				s: "550e8400-e29b-41d4-a716-4466554400001",
			},
			want:    ID{},
			wantErr: true,
		},
		{
			name: "fails to parse invalid characters",
			args: args{
				s: "550e8400-e29b-41d4-a716-44665544000g",
			},
			want:    ID{},
			wantErr: true,
		},
		{
			name: "fails to parse empty string",
			args: args{
				s: "",
			},
			want:    ID{},
			wantErr: true,
		},
		{
			name: "fails to parse random string",
			args: args{
				s: "not-a-uuid",
			},
			want:    ID{},
			wantErr: true,
		},
		{
			name: "fails to parse string with wrong format",
			args: args{
				s: "550e8400-e29b-41d4-a716",
			},
			want:    ID{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.args.s)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ID{}, got, "ParseID() should return zero ID on error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got, "ParseID() should return the expected ID")

				// Additional validation for successful cases
				assert.Equal(t, tt.args.s, got.String(), "Parsed ID should have the same string representation")
				assert.Equal(t, tt.want.value, got.UUID(), "Parsed ID should have the expected UUID value")
			}
		})
	}
}

func TestParseEncodedID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ID
		wantErr bool
	}{
		{
			name:  "successfully parses raw encoded string",
			input: "VQ6EAOKbQdSnFkRmVUQAAA",
			want: ID{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			wantErr: false,
		},
		{
			name:  "successfully parses padded base64 string",
			input: "VQ6EAOKbQdSnFkRmVUQAAA==",
			want: ID{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			wantErr: false,
		},
		{
			name:    "returns error when decoded bytes are not uuid length",
			input:   "AAECAwQFBgcICQoLDA0O",
			want:    ID{},
			wantErr: true,
		},
		{
			name:    "returns error for invalid base64 characters",
			input:   "not-a-valid-id===",
			want:    ID{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEncodedID(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ID{}, got, "ParseEncodedID() should return zero ID on error")
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "ParseEncodedID() should return the expected ID")
			assert.Equal(t, tt.want.value, got.UUID(), "ParseEncodedID() should return the expected UUID value")
		})
	}
}

func TestID_EncodedString(t *testing.T) {
	tests := []struct {
		name  string
		input ID
		want  string
	}{
		{
			name: "encodes regular uuid without padding",
			input: ID{
				value: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			},
			want: "VQ6EAOKbQdSnFkRmVUQAAA",
		},
		{
			name: "encodes nil uuid to all zeros",
			input: ID{
				value: uuid.Nil,
			},
			want: "AAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			name: "encodes max uuid value",
			input: ID{
				value: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			},
			want: "_____________________w",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.EncodedString()
			assert.Equal(t, tt.want, got, "EncodedString() should return the expected base64 value")
			assert.NotContains(t, got, "=", "EncodedString() should not include padding characters")
		})
	}
}
