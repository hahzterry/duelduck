package apikey

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super-secret-key"

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		projectID uuid.UUID
		secret    string
	}{
		{
			name:      "generates key with correct prefix",
			projectID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			secret:    testSecret,
		},
		{
			name:      "nil uuid produces valid key",
			projectID: uuid.Nil,
			secret:    testSecret,
		},
		{
			name:      "different secrets produce different keys",
			projectID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			secret:    "other-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := Generate(tt.projectID, tt.secret)

			parts := strings.SplitN(key, ".", 3)
			require.Len(t, parts, 3, "key must have exactly 3 dot-separated parts")
			assert.Equal(t, "v1", parts[0])
			assert.NotEmpty(t, parts[1])
			assert.NotEmpty(t, parts[2])
		})
	}
}

func TestGenerateThenParse_RoundTrip(t *testing.T) {
	ids := []uuid.UUID{
		uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		uuid.New(),
	}

	for _, id := range ids {
		t.Run(id.String(), func(t *testing.T) {
			key := Generate(id, testSecret)
			got, ok := Parse(key, testSecret)
			require.True(t, ok)
			assert.Equal(t, id, got)
		})
	}
}

func TestParse(t *testing.T) {
	validID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	validKey := Generate(validID, testSecret)

	tests := []struct {
		name   string
		key    string
		secret string
		wantID uuid.UUID
		wantOK bool
	}{
		{
			name:   "valid key and secret",
			key:    validKey,
			secret: testSecret,
			wantID: validID,
			wantOK: true,
		},
		{
			name:   "wrong secret",
			key:    validKey,
			secret: "wrong-secret",
			wantOK: false,
		},
		{
			name:   "empty key",
			key:    "",
			secret: testSecret,
			wantOK: false,
		},
		{
			name:   "only two parts",
			key:    "v1.payload",
			secret: testSecret,
			wantOK: false,
		},
		{
			name:   "wrong version prefix",
			key:    "v2." + strings.Join(strings.SplitN(validKey, ".", 3)[1:], "."),
			secret: testSecret,
			wantOK: false,
		},
		{
			name:   "tampered payload",
			key:    "v1.AAAAAAAAAAAAAAAAAAAAAA." + strings.SplitN(validKey, ".", 3)[2],
			secret: testSecret,
			wantOK: false,
		},
		{
			name:   "tampered signature",
			key:    strings.SplitN(validKey, ".", 3)[0] + "." + strings.SplitN(validKey, ".", 3)[1] + ".invalidsig",
			secret: testSecret,
			wantOK: false,
		},
		{
			name:   "payload too short to be a uuid",
			key:    "v1.dG9vc2hvcnQ.invalidsig",
			secret: testSecret,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, ok := Parse(tt.key, tt.secret)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantID, gotID)
			} else {
				assert.Equal(t, uuid.UUID{}, gotID)
			}
		})
	}
}

func TestGenerateDifferentSecretsProduceDifferentKeys(t *testing.T) {
	id := uuid.New()
	key1 := Generate(id, "secret-a")
	key2 := Generate(id, "secret-b")
	assert.NotEqual(t, key1, key2)
}
