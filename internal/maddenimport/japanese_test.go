package maddenimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// db/data/madden の全14シーズン分の実データで実際に観測された全ヘッダー（370件）を
// 1件も取り逃さず翻訳できることを保証する。
func TestJapaneseLabel_CoversAllObservedHeaders(t *testing.T) {
	for _, h := range allObservedMaddenHeaders {
		t.Run(h, func(t *testing.T) {
			label, ok := JapaneseLabel(h)
			require.True(t, ok, "未翻訳のヘッダー: %q", h)
			assert.NotEmpty(t, label)
		})
	}
}

func TestJapaneseLabel(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"plain", "Speed", "スピード"},
		{"all caps no space (2025)", "SPEED", "スピード"},
		{"Rating suffix (2021)", "SpeedRating", "スピード"},
		{"nested value (2024)", "stats/speed/value", "スピード"},
		{"nested diff (2024)", "stats/speed/diff", "スピード（変化量）"},
		{"typo: Finnesse Moves", "Finnesse Moves", "フィネスムーブ"},
		{"typo: FINESSEE MOVE", "FINESSEE MOVE", "フィネスムーブ"},
		{"typo: Stength", "Stength", "パワー"},
		{"Block/Blocking wording: Impact Block", "Impact Block", "インパクトブロック"},
		{"Block/Blocking wording: Impact Blocking", "Impact Blocking", "インパクトブロック"},
		{"Madden spec change: Pass Block Footwork", "Pass Block Footwork", "パスブロックフィネス"},
		{"Madden spec change: Pass Block Strength", "Pass Block Strength", "パスブロックパワー"},
		{"short name: Catch", "Catch", "キャッチ"},
		{"short name: Deep Accuracy", "Deep Accuracy", "スローアキュラシー（ディープ）"},
		{"abbreviation: THROW ACCURACY MED", "THROW ACCURACY MED", "スローアキュラシー（ミドル）"},
		{"metadata", "College", "大学"},
		{"metadata variant", "PLYR_BIRTHDATE", "生年月日"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := JapaneseLabel(tt.raw)
			require.True(t, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJapaneseLabel_Unknown(t *testing.T) {
	_, ok := JapaneseLabel("___NOT_A_REAL_HEADER___")
	assert.False(t, ok)
}
