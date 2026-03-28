# Gospel Library — Magazines & Other Content

*Content inventory for model experiment planning.*

---

## Magazines

| Magazine | Path | Audience | Notes |
|----------|------|----------|-------|
| Liahona | `gospel-library/eng/liahona/` | International/family | Primary Church magazine (replaced Ensign) |
| Ensign | `gospel-library/eng/ensign/` | Adult members | Historical — merged into Liahona |
| Friend | `gospel-library/eng/friend/` | Children | Monthly children's magazine |
| New Era | `gospel-library/eng/new-era/` | Youth | Merged into "For the Strength of Youth" |
| YA Weekly | `gospel-library/eng/ya-weekly/` | Young adults | Weekly content |
| FTSOY | `gospel-library/eng/ftsoy/` | Youth | For the Strength of Youth booklet |

## Music

```
gospel-library/eng/music/
├── hymns-for-home-and-church/
├── using-hymns-for-home-and-church/
├── songs-of-devotion-for-everyday-listening/
├── selections-for-christmas/
├── look-unto-christ-2025-youth-album/
├── walk-with-me-2026/
├── music-from-*-general-conference/ (2008-2025, by season)
└── ...  (13 collections total)
```

## Other

| Content | Path | Notes |
|---------|------|-------|
| Hymns | `gospel-library/eng/hymns/` | Hymnbook content |
| Broadcasts | `gospel-library/eng/broadcasts/` | Church broadcasts |
| Video | `gospel-library/eng/video/` | Video metadata/content |

## Digestion Considerations

- **Magazines** are article-based — individual articles are good summarization units
- **Music** is less relevant for text-based model experiments (lyrics could be interesting)
- **Liahona articles** overlap with conference talks but include additional editorial content
- **Lower priority** for initial model experiments — focus on scriptures, talks, and manuals first
- **Could be valuable** for testing cross-content retrieval (finding a Liahona article that discusses a specific scripture)
