// Code generated from the Laravel seeders (database/seeders/TeamsTableSeeder.php) by a one-off script; hand edits are fine.

package seed

import (
	"github.com/toshi1023/new-nfl-library/internal/repository/sqlc"
)

// teams は teams テーブルの初期データ。順序（= id）は Laravel の TeamsTableSeeder と同じで、domain.AllTeams の ID と一致する。
// conference: 1 = AFC, 2 = NFC / area: 1 = 東, 2 = 北, 3 = 南, 4 = 西
var teams = []sqlc.InsertTeamsParams{
	// AFC 東
	{City: "Buffalo", Name: "Bills", Conference: 1, Area: 1, ImageFile: ptr("bills.gif"), BackImageFile: ptr("bills_cover.jpg")},
	{City: "New England", Name: "Patriots", Conference: 1, Area: 1, ImageFile: ptr("patriots.gif"), BackImageFile: ptr("patriots_cover.jpg")},
	{City: "Miami", Name: "Dolphins", Conference: 1, Area: 1, ImageFile: ptr("dolphins.gif"), BackImageFile: ptr("dolphins_cover.jpg")},
	{City: "New York", Name: "Jets", Conference: 1, Area: 1, ImageFile: ptr("jets.gif"), BackImageFile: ptr("jets_cover.jpg")},
	// AFC 北
	{City: "Cincinnati", Name: "Bengals", Conference: 1, Area: 2, ImageFile: ptr("bengals.gif"), BackImageFile: ptr("bengals_cover.jpg")},
	{City: "Pittsburgh", Name: "Steelers", Conference: 1, Area: 2, ImageFile: ptr("steelers.gif"), BackImageFile: ptr("steelers_cover.jpg")},
	{City: "Cleveland", Name: "Browns", Conference: 1, Area: 2, ImageFile: ptr("browns.gif"), BackImageFile: ptr("browns_cover.jpg")},
	{City: "Baltimore", Name: "Ravens", Conference: 1, Area: 2, ImageFile: ptr("ravens.gif"), BackImageFile: ptr("ravens_cover.jpg")},
	// AFC 南
	{City: "Tennessee", Name: "Titans", Conference: 1, Area: 3, ImageFile: ptr("titans.gif"), BackImageFile: ptr("titans_cover.jpg")},
	{City: "Indianapolis", Name: "Colts", Conference: 1, Area: 3, ImageFile: ptr("colts.gif"), BackImageFile: ptr("colts_cover.jpg")},
	{City: "Houston", Name: "Texans", Conference: 1, Area: 3, ImageFile: ptr("texans.gif"), BackImageFile: ptr("texans_cover.jpg")},
	{City: "Jacksonville", Name: "Jaguars", Conference: 1, Area: 3, ImageFile: ptr("jaguars.gif"), BackImageFile: ptr("jaguars_cover.jpg")},
	// AFC 西
	{City: "Kansas City", Name: "Chiefs", Conference: 1, Area: 4, ImageFile: ptr("chiefs.gif"), BackImageFile: ptr("chiefs_cover.jpg")},
	{City: "Las Vegas", Name: "Raiders", Conference: 1, Area: 4, ImageFile: ptr("raiders.gif"), BackImageFile: ptr("raiders_cover.jpg")},
	{City: "Los Angeles", Name: "Chargers", Conference: 1, Area: 4, ImageFile: ptr("chargers.gif"), BackImageFile: ptr("chargers_cover.jpg")},
	{City: "Denver", Name: "Broncos", Conference: 1, Area: 4, ImageFile: ptr("broncos.gif"), BackImageFile: ptr("broncos_cover.jpg")},
	// NFC 東
	{City: "Dallas", Name: "Cowboys", Conference: 2, Area: 1, ImageFile: ptr("cowboys.gif"), BackImageFile: ptr("cowboys_cover.jpg")},
	{City: "Philadelphia", Name: "Eagles", Conference: 2, Area: 1, ImageFile: ptr("eagles.gif"), BackImageFile: ptr("eagles_cover.jpg")},
	{City: "Washington", Name: "Commanders", Conference: 2, Area: 1, ImageFile: ptr("commanders.gif"), BackImageFile: ptr("commanders_cover.jpg")},
	{City: "New York", Name: "Giants", Conference: 2, Area: 1, ImageFile: ptr("giants.gif"), BackImageFile: ptr("giants_cover.jpg")},
	// NFC 北
	{City: "Green Bay", Name: "Packers", Conference: 2, Area: 2, ImageFile: ptr("packers.gif"), BackImageFile: ptr("packers_cover.jpg")},
	{City: "Minnesota", Name: "Vikings", Conference: 2, Area: 2, ImageFile: ptr("vikings.gif"), BackImageFile: ptr("vikings_cover.jpg")},
	{City: "Chicago", Name: "Bears", Conference: 2, Area: 2, ImageFile: ptr("bears.gif"), BackImageFile: ptr("bears_cover.jpg")},
	{City: "Detroit", Name: "Lions", Conference: 2, Area: 2, ImageFile: ptr("lions.gif"), BackImageFile: ptr("lions_cover.jpg")},
	// NFC 南
	{City: "Tampa Bay", Name: "Buccaneers", Conference: 2, Area: 3, ImageFile: ptr("buccaneers.gif"), BackImageFile: ptr("buccaneers_cover.jpg")},
	{City: "New Orleans", Name: "Saints", Conference: 2, Area: 3, ImageFile: ptr("saints.gif"), BackImageFile: ptr("saints_cover.jpg")},
	{City: "Atlanta", Name: "Falcons", Conference: 2, Area: 3, ImageFile: ptr("falcons.gif"), BackImageFile: ptr("falcons_cover.jpg")},
	{City: "Carolina", Name: "Panthers", Conference: 2, Area: 3, ImageFile: ptr("panthers.gif"), BackImageFile: ptr("panthers_cover.jpg")},
	// NFC 西
	{City: "Los Angeles", Name: "Rams", Conference: 2, Area: 4, ImageFile: ptr("rams.gif"), BackImageFile: ptr("rams_cover.jpg")},
	{City: "Arizona", Name: "Cardinals", Conference: 2, Area: 4, ImageFile: ptr("cardinals.gif"), BackImageFile: ptr("cardinals_cover.jpg")},
	{City: "San Francisco", Name: "49ers", Conference: 2, Area: 4, ImageFile: ptr("49ers.gif"), BackImageFile: ptr("49ers_cover.jpg")},
	{City: "Seattle", Name: "Seahawks", Conference: 2, Area: 4, ImageFile: ptr("seahawks.gif"), BackImageFile: ptr("seahawks_cover.jpg")},
}

func ptr[T any](v T) *T { return &v }
