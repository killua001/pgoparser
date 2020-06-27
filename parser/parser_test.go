package parser

import (
	"github.com/elliotcourant/pgoparser/symbols"
	"github.com/elliotcourant/pgoparser/tokens"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		t.Run("create bad", func(t *testing.T) {
			sql := `CREATE SOMETHING`
			parsed, err := Parse(sql)
			assert.EqualError(t, err, "expected TABLE, VIEW, INDEX or SCHEMA after CREATE found SOMETHING")
			assert.Nil(t, parsed)
		})

		t.Run("create table", func(t *testing.T) {
			sql := `CREATE TABLE IF NOT EXISTS users (id BIGINT PRIMARY KEY, email TEXT UNIQUE NOT NULL);`
			parsed, err := Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, parsed)
		})

		t.Run("create table w foreign key", func(t *testing.T) {
			sql := `CREATE TABLE IF NOT EXISTS users (id BIGINT PRIMARY KEY, email TEXT UNIQUE NOT NULL, account_id BIGINT NOT NULL REFERENCES accounts (account_id));`
			parsed, err := Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, parsed)
		})
	})
}

func TestStructMatch(t *testing.T) {
	var someToken tokens.Token
	someToken = tokens.EOF{}
	assert.True(t, someToken == (tokens.EOF{}))
	assert.False(t, someToken == (symbols.Comma))
}

func BenchmarkParseElliotsParser(b *testing.B) {
	sql := `CREATE TABLE IF NOT EXISTS users (id BIGINT PRIMARY KEY, email TEXT UNIQUE NOT NULL, account_id BIGINT NOT NULL REFERENCES accounts (account_id));`

	parsed, err := Parse(sql)
	assert.NoError(b, err)
	assert.NotNil(b, parsed)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parsed, err = Parse(sql)
	}
}
