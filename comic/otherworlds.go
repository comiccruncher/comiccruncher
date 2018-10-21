package comic

import "strings"

// universeDefinition is a definition for main and alternate universes.
type universeDefinition string

var  (
	// dcAltUniverses defines the alternate universes for DC.
	// Unfortunately have to define all possible alternate universes versus just Earth-0, etc.,
	// since there's no indicator for main sources.
	dcAltUniverses = []universeDefinition{
		"Animated",
		"New Frontier",
		"Kingdom Come",
		"(2%)",
		"Beyond",
		"Bombshells",
		"Earth -44",
		"Earth-16", // The Just
		"The Just",
		"Earth-14", // Assassins
		"Assassins",
		"Earth-15", // Legacy Earth
		"Legacy Earth",
		"Earth-17", // Dystopia
		"Dystopia",
		"Earth-18",
		"Justice Riders",
		"Earth-19", // Gotham By Gaslight
		"Gaslight",
		"Earth One",
		"Earth-1", // All Earth-1xx
		"Earth-2", // All Earth-2xx
		"Earth-3", // All Earth-3xx
		"All-Star",
		"Thrillkiller",
		"All-Star",
		"Earth-4",
		"Liberty Files",
		"Metal Men",
		"Earth-5",
		"Utopia",
		"Earth-9",
		"Tangent",
		"Earth-A",
		"Lawless League",
		"Reverse Gender",
		"Lawless League",
		"Injustice",
		"JL 3000",
		"Just Imagine",
		"Li'l Gotham",
		"New Order",
		"One Million",
		"Pax",
		"Polyester Age",
		"Smallville",
		"Titans",
		"non-continuity",
		"Batman '66",
		"Arkhamverse",
		"Digital Justice",
		"First Wave",
		"Gotham City Garage",
		"3000",
		"White Knight",
		"Century",
		"Beware",
		"Amalgam",
		"Ame-Comi",
		"Arrowverse",
		"Arrow/Flash",
		"cartoon",
		"Legion - 5 Years",
		"DKR",
		"Future's End",
		"Generation Lost",
		"Matrix",
		"Pocket",
		"Golden Age",
		"'77",
		"Superman Red",
		"Superman Blue",
		"Superboy",
	}
	// marvelAltUniverses defines the alternate universes of the MU.
	// Unfortunately have to define all possible alternate universes versus just 616
	// since there's no indicator for 616 sources.
	marvelAltUniverses = []universeDefinition{
		"earth-",
		"2020",
		"2099",
		"26th Century",
		"Agent of Hydra",
		"100th Anniversary",
		"A-Babies",
		"(Marvel)(Adventures)",
		"Animated",
		"Cancerverse",
		"Earth X",
		"Bullet Points",
		"Undead",
		"Knowhere",
		"Mangaverse",
		"Mini Marvels",
		"Movies",
		"Noir",
		"Last Gun on Earth",
		"Super Hero Squad",
		"Next Avengers",
		"Timeslip",
		"Ultimate",
		"MC2",
		"Mutant X",
		"20xx",
		"(Marvel)(Spider-Gwen)",
		"Battle of the Atom",
		"E is for Extincti",
		"Exiles",
		"Egyptia",
		"(Marvel)(Forward)",
		"Old Man Logan",
		"Old Woman Laura",
		"X-Campus",
		"(Secret War)(Limbo)",
		"Age Of Apocalypse",
		"Age of Ultron",
		"Apes",
		"Age of X",
		"venomverse",
		"India",
		"Mitey 'vengers",
		"2211",
		"X-Men The End",
		"1872",
		"1602",
		"Inferno",
		"Mutopia",
		"2055",
		"Babies",
		"Mojoverse",
		"Killville",
		"What If",
		"Spider-Verse",
		"Shadow-X",
		"Days Of",
		"Omega World",
		"Killiseum",
		"a-force",
		"1,000,000 B.C.",
		"Guardians 3000",
		"Armor Wars",
		"Secret Wars",
		"Future Imperfect",
		"Dystopia",
		"Children's Crusade Future",
		"X-Tinction Agenda",
		"Renew Your Vow",
		"Hex-men",
		"scorched earth",
		"(Thors)",
		"(Heroes Reborn)",
		"Years of Future",
		"newuniversal",
		"spider-island",
		"Contest of Champions",
		"(Deadpool ",
		"5 Ronin",
		"Ghost Racers",
		"Spirit of Vengeance",
		"x-men forever",
		"(Red Skull)",
		"(Robot)",
		"cartoon",
		"(Spider-Woman)",
		"Attilan Rising",
		"PS4",
		"Hail Hydra",
		"Spider-Island",
		"Zombies",
	}
	// marvelDisabledUniverses defines the universes that should be disabled for character sources.
	marvelDisabledUniverses = []universeDefinition{
		"A.I.vengers",
		"imposter",
		"impostor",
		"Kree Avengers",
		"mutate",
		"clone",
	}
	// dcDisabledUniverses defines the sources that should be disabled for DC characters.
	dcDisabledUniverses = []universeDefinition {
		"clone",
		"fake",
		"robot",
	}
)
// pgSearchString returns a string suitable for a postgres array.
func pgSearchString(ud []universeDefinition) string {
	str := ""
	for idx := range ud {
		// escape single `'` to `''` so it works with postgres.
		str += "'%" + strings.Replace(string(ud[idx]), "'", "''", -1) + "%'"
		// if not last one or it doesn't have only one item
		if len(ud)-1 != idx {
			// append a comma
			str += ","
		}
	}
	return str
}
