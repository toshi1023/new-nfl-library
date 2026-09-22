package seed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/toshi1023/new-nfl-library/internal/domain"
)

// Laravel のシーダーと同じ件数であること。
func TestDataCounts(t *testing.T) {
	assert.Len(t, teams, 32)
	assert.Len(t, positions, 85)
	assert.Len(t, formations, 36)
	assert.Len(t, pfRelations, 374)
	assert.Len(t, foulRules, 34)
}

func TestTeams(t *testing.T) {
	// domain.AllTeams の並び（= id）と一致していること
	all := domain.AllTeams()
	require.Len(t, all, len(teams))
	for i, tm := range teams {
		assert.Equal(t, i+1, all[i].ID)
		if i < 16 {
			assert.EqualValues(t, 1, tm.Conference, "%s は AFC", tm.Name)
		} else {
			assert.EqualValues(t, 2, tm.Conference, "%s は NFC", tm.Name)
		}
		assert.EqualValues(t, (i%16)/4+1, tm.Area, tm.Name)
		require.NotNil(t, tm.ImageFile)
		require.NotNil(t, tm.BackImageFile)
	}
	assert.Equal(t, "Bills", teams[0].Name)
	assert.Equal(t, "49ers", teams[30].Name, "root の favorite_team = 31")
	assert.Equal(t, "Seahawks", teams[31].Name)
}

func TestPositions(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range positions {
		assert.False(t, seen[p.Name], "重複: %s", p.Name)
		seen[p.Name] = true
		assert.NotZero(t, p.Category)
		assert.NotZero(t, p.AbstractCategory)
	}
	assert.True(t, seen[domain.PositionNoData], "スクレイパーが空ポジションに使う No Data が必要")
	assert.True(t, seen["QB"])
}

func TestFormationsAndPfRelations(t *testing.T) {
	keys := map[string]bool{}
	for _, f := range formations {
		assert.False(t, keys[f.Key], "重複キー: %s", f.Key)
		keys[f.Key] = true
		assert.NotEmpty(t, f.Name)
	}
	assert.Equal(t, "Nickel(LCB)", formations[19].Name)
	assert.Equal(t, "Nickel(RCB)", formations[20].Name)

	for _, r := range pfRelations {
		assert.True(t, keys[r.Formation], "pf_relations が未定義のフォーメーションを参照: %s", r.Formation)
	}
}

func TestFoulRules(t *testing.T) {
	for _, r := range foulRules {
		assert.NotEmpty(t, r.EnglishName)
		assert.NotEmpty(t, r.JapaneseName)
		assert.NotEmpty(t, r.Description)
		assert.Contains(t, []int32{domain.FoulStatusCommon, domain.FoulStatusOffence, domain.FoulStatusDefence, domain.FoulStatusKick}, r.StatusType)
	}
}
