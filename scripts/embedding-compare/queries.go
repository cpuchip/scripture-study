package main

// testQueries returns all test queries organized by category.
// 80 queries across 8 categories to stress-test embedding quality differences.
func testQueries() []Query {
	return []Query{
		// ================================================================
		// FACTUAL/SPECIFIC — clear "right answers" in the text
		// ================================================================
		{Text: "Lehi's dream of the tree of life", Category: "factual"},
		{Text: "brass plates of Laban", Category: "factual"},
		{Text: "Nephi breaks his bow", Category: "factual"},
		{Text: "Liahona compass", Category: "factual"},
		{Text: "ship building in Bountiful", Category: "factual"},
		{Text: "Laman and Lemuel murmur against their father", Category: "factual"},
		{Text: "Zoram the servant of Laban", Category: "factual"},
		{Text: "Ishmael and his family join Lehi", Category: "factual"},
		{Text: "Nephi slays Laban with his own sword", Category: "factual"},
		{Text: "Lehi finds the Liahona outside his tent", Category: "factual"},

		// ================================================================
		// THEMATIC/CONCEPTUAL — tests semantic understanding
		// ================================================================
		{Text: "faith and obedience to God", Category: "thematic"},
		{Text: "God's love for his children", Category: "thematic"},
		{Text: "following the prophet", Category: "thematic"},
		{Text: "the power of the word of God", Category: "thematic"},
		{Text: "family conflict and forgiveness", Category: "thematic"},
		{Text: "trusting God in the wilderness", Category: "thematic"},
		{Text: "the consequences of rebellion", Category: "thematic"},
		{Text: "courage in the face of opposition", Category: "thematic"},
		{Text: "God prepares a way for the righteous", Category: "thematic"},
		{Text: "personal revelation and spiritual guidance", Category: "thematic"},

		// ================================================================
		// CHRISTOLOGICAL — tests scriptural depth
		// ================================================================
		{Text: "types and shadows of Christ", Category: "christological"},
		{Text: "the Messiah will redeem his people", Category: "christological"},
		{Text: "Lamb of God", Category: "christological"},
		{Text: "baptism and remission of sins", Category: "christological"},
		{Text: "the tree of life as God's love", Category: "christological"},
		{Text: "the condescension of God", Category: "christological"},
		{Text: "a Savior born of a virgin", Category: "christological"},
		{Text: "the ministry and crucifixion of Christ", Category: "christological"},
		{Text: "the twelve apostles of the Lamb", Category: "christological"},
		{Text: "the rod of iron as the word of God leading to Christ", Category: "christological"},

		// ================================================================
		// CROSS-REFERENCE — referencing concepts from other scripture
		// ================================================================
		{Text: "Isaiah's prophecy of the last days", Category: "cross-reference"},
		{Text: "the scattering and gathering of Israel", Category: "cross-reference"},
		{Text: "the plan of salvation", Category: "cross-reference"},
		{Text: "priesthood authority", Category: "cross-reference"},
		{Text: "the Holy Ghost as a guide", Category: "cross-reference"},
		{Text: "Moses and the exodus parallel", Category: "cross-reference"},
		{Text: "the covenant of Abraham", Category: "cross-reference"},
		{Text: "olive tree allegory and the house of Israel", Category: "cross-reference"},
		{Text: "great and abominable church", Category: "cross-reference"},
		{Text: "plain and precious things removed from the Bible", Category: "cross-reference"},

		// ================================================================
		// KJV PHRASING — archaic language the models might handle differently
		// ================================================================
		{Text: "I will go and do the things which the Lord hath commanded", Category: "kjv-phrasing"},
		{Text: "the mists of darkness", Category: "kjv-phrasing"},
		{Text: "great and spacious building", Category: "kjv-phrasing"},
		{Text: "rod of iron", Category: "kjv-phrasing"},
		{Text: "river of filthy water", Category: "kjv-phrasing"},
		{Text: "exceedingly sorrowful because of the hardness of their hearts", Category: "kjv-phrasing"},
		{Text: "the Spirit of the Lord constraineth me", Category: "kjv-phrasing"},
		{Text: "large and spacious field", Category: "kjv-phrasing"},
		{Text: "the tender mercies of the Lord", Category: "kjv-phrasing"},
		{Text: "whither shall I go that I may find ore to molten", Category: "kjv-phrasing"},

		// ================================================================
		// ABSTRACT/MODERN — same concepts in modern phrasing (perturbation pairs with KJV)
		// ================================================================
		{Text: "obedience to God's commandments despite difficulty", Category: "modern-phrasing"},
		{Text: "spiritual blindness and confusion", Category: "modern-phrasing"},
		{Text: "worldly pride and vanity", Category: "modern-phrasing"},
		{Text: "holding fast to scripture", Category: "modern-phrasing"},
		{Text: "sin and spiritual pollution", Category: "modern-phrasing"},
		{Text: "grief over unbelieving family members", Category: "modern-phrasing"},
		{Text: "feeling prompted by the Spirit to act", Category: "modern-phrasing"},
		{Text: "an open vision of a large meadow", Category: "modern-phrasing"},
		{Text: "God's grace and compassion for the faithful", Category: "modern-phrasing"},
		{Text: "finding raw materials to build tools", Category: "modern-phrasing"},

		// ================================================================
		// SHORT/AMBIGUOUS — minimal context, tests disambiguation
		// ================================================================
		{Text: "dream", Category: "short"},
		{Text: "sword", Category: "short"},
		{Text: "wilderness", Category: "short"},
		{Text: "plates", Category: "short"},
		{Text: "ship", Category: "short"},
		{Text: "tree", Category: "short"},
		{Text: "angel", Category: "short"},
		{Text: "fire", Category: "short"},
		{Text: "iron", Category: "short"},
		{Text: "Jerusalem", Category: "short"},

		// ================================================================
		// MULTI-HOP — require connecting multiple concepts
		// ================================================================
		{Text: "why Nephi had to kill Laban to get the plates for his family's spiritual survival", Category: "multi-hop"},
		{Text: "how Lehi's dream connects the tree of life to baptism through the river", Category: "multi-hop"},
		{Text: "the pattern of murmuring leading to rebellion leading to divine correction", Category: "multi-hop"},
		{Text: "Nephi's vision expanding Lehi's dream with Christological interpretation", Category: "multi-hop"},
		{Text: "why the brass plates were necessary for preserving language and prophecy", Category: "multi-hop"},
		{Text: "how the Liahona worked by faith like prayer", Category: "multi-hop"},
		{Text: "Laman and Lemuel seeing an angel and still not believing", Category: "multi-hop"},
		{Text: "Isaiah chapters quoted by Nephi about the last days restoration", Category: "multi-hop"},
		{Text: "the great and spacious building as the world's opposition to God's covenant people", Category: "multi-hop"},
		{Text: "Nephi's psalm of trust despite weakness and enemies", Category: "multi-hop"},
	}
}

// GroundTruth maps queries to expected top results (by chapter number).
// Used to compute precision — did the model find the *right* chapter?
type GroundTruth struct {
	Query           string
	ExpectedSummary []int // chapter numbers expected in top-3 summaries
}

// groundTruths returns queries with known correct answers.
// These are chapters where the content unambiguously lives.
func groundTruths() []GroundTruth {
	return []GroundTruth{
		// Factual — clear chapter locations
		{"Lehi's dream of the tree of life", []int{8, 15}},
		{"brass plates of Laban", []int{3, 4, 5}},
		{"Nephi breaks his bow", []int{16}},
		{"Liahona compass", []int{16, 18}},
		{"ship building in Bountiful", []int{17, 18}},
		{"Nephi slays Laban with his own sword", []int{4}},
		{"Ishmael and his family join Lehi", []int{7}},
		{"Zoram the servant of Laban", []int{4}},
		{"Lehi finds the Liahona outside his tent", []int{16}},
		{"Laman and Lemuel murmur against their father", []int{2, 3, 17}},

		// Christological — specific visions
		{"the condescension of God", []int{11}},
		{"Lamb of God", []int{11, 13, 14}},
		{"the twelve apostles of the Lamb", []int{11, 12}},
		{"great and abominable church", []int{13, 14}},
		{"plain and precious things removed from the Bible", []int{13}},

		// KJV — exact phrases from specific chapters
		{"I will go and do the things which the Lord hath commanded", []int{3}},
		{"great and spacious building", []int{8, 11, 12}},
		{"rod of iron", []int{8, 11, 15}},
		{"the tender mercies of the Lord", []int{1}},
		{"the Spirit of the Lord constraineth me", []int{4}},

		// Multi-hop
		{"Nephi's psalm of trust despite weakness and enemies", []int{4, 17}},
		{"Isaiah chapters quoted by Nephi about the last days restoration", []int{20, 21, 22}},
	}
}
