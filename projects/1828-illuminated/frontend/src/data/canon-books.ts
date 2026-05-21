// Hand-maintained canon book list. Mirrors the backend's abbr conventions
// (gospel-library/eng/scriptures/ directory naming). When the backend grows
// a /api/scripture/books endpoint, swap this for a fetch — until then,
// this is the stable, audit-friendly source for the verse-explorer canon
// selector.

export interface CanonBook {
  abbr: string       // matches scripture_books.abbr in the backend DB
  name: string       // human display name
  /** Last chapter number — used to populate the chapter selector without
   *  a round-trip. Verified against bcbooks/scriptures-json 2013 edition. */
  chapters: number
  /** Path segment used in the churchofjesuschrist.org URL. */
  urlPath: string
}

export interface CanonVolume {
  id: 'ot' | 'nt' | 'bofm' | 'dc' | 'pgp'
  label: string
  /** Volume path segment in the churchofjesuschrist.org URL. */
  urlVolume: string
  books: CanonBook[]
}

export const CANON: CanonVolume[] = [
  {
    id: 'ot',
    label: 'Old Testament',
    urlVolume: 'ot',
    books: [
      { abbr: 'gen', name: 'Genesis', chapters: 50, urlPath: 'gen' },
      { abbr: 'ex', name: 'Exodus', chapters: 40, urlPath: 'ex' },
      { abbr: 'lev', name: 'Leviticus', chapters: 27, urlPath: 'lev' },
      { abbr: 'num', name: 'Numbers', chapters: 36, urlPath: 'num' },
      { abbr: 'deut', name: 'Deuteronomy', chapters: 34, urlPath: 'deut' },
      { abbr: 'josh', name: 'Joshua', chapters: 24, urlPath: 'josh' },
      { abbr: 'judg', name: 'Judges', chapters: 21, urlPath: 'judg' },
      { abbr: 'ruth', name: 'Ruth', chapters: 4, urlPath: 'ruth' },
      { abbr: '1-sam', name: '1 Samuel', chapters: 31, urlPath: '1-sam' },
      { abbr: '2-sam', name: '2 Samuel', chapters: 24, urlPath: '2-sam' },
      { abbr: '1-kgs', name: '1 Kings', chapters: 22, urlPath: '1-kgs' },
      { abbr: '2-kgs', name: '2 Kings', chapters: 25, urlPath: '2-kgs' },
      { abbr: '1-chr', name: '1 Chronicles', chapters: 29, urlPath: '1-chr' },
      { abbr: '2-chr', name: '2 Chronicles', chapters: 36, urlPath: '2-chr' },
      { abbr: 'ezra', name: 'Ezra', chapters: 10, urlPath: 'ezra' },
      { abbr: 'neh', name: 'Nehemiah', chapters: 13, urlPath: 'neh' },
      { abbr: 'esth', name: 'Esther', chapters: 10, urlPath: 'esth' },
      { abbr: 'job', name: 'Job', chapters: 42, urlPath: 'job' },
      { abbr: 'ps', name: 'Psalms', chapters: 150, urlPath: 'ps' },
      { abbr: 'prov', name: 'Proverbs', chapters: 31, urlPath: 'prov' },
      { abbr: 'eccl', name: 'Ecclesiastes', chapters: 12, urlPath: 'eccl' },
      { abbr: 'song', name: 'Song of Solomon', chapters: 8, urlPath: 'song' },
      { abbr: 'isa', name: 'Isaiah', chapters: 66, urlPath: 'isa' },
      { abbr: 'jer', name: 'Jeremiah', chapters: 52, urlPath: 'jer' },
      { abbr: 'lam', name: 'Lamentations', chapters: 5, urlPath: 'lam' },
      { abbr: 'ezek', name: 'Ezekiel', chapters: 48, urlPath: 'ezek' },
      { abbr: 'dan', name: 'Daniel', chapters: 12, urlPath: 'dan' },
      { abbr: 'hosea', name: 'Hosea', chapters: 14, urlPath: 'hosea' },
      { abbr: 'joel', name: 'Joel', chapters: 3, urlPath: 'joel' },
      { abbr: 'amos', name: 'Amos', chapters: 9, urlPath: 'amos' },
      { abbr: 'obad', name: 'Obadiah', chapters: 1, urlPath: 'obad' },
      { abbr: 'jonah', name: 'Jonah', chapters: 4, urlPath: 'jonah' },
      { abbr: 'micah', name: 'Micah', chapters: 7, urlPath: 'micah' },
      { abbr: 'nahum', name: 'Nahum', chapters: 3, urlPath: 'nahum' },
      { abbr: 'hab', name: 'Habakkuk', chapters: 3, urlPath: 'hab' },
      { abbr: 'zeph', name: 'Zephaniah', chapters: 3, urlPath: 'zeph' },
      { abbr: 'hag', name: 'Haggai', chapters: 2, urlPath: 'hag' },
      { abbr: 'zech', name: 'Zechariah', chapters: 14, urlPath: 'zech' },
      { abbr: 'mal', name: 'Malachi', chapters: 4, urlPath: 'mal' },
    ],
  },
  {
    id: 'nt',
    label: 'New Testament',
    urlVolume: 'nt',
    books: [
      { abbr: 'matt', name: 'Matthew', chapters: 28, urlPath: 'matt' },
      { abbr: 'mark', name: 'Mark', chapters: 16, urlPath: 'mark' },
      { abbr: 'luke', name: 'Luke', chapters: 24, urlPath: 'luke' },
      { abbr: 'john', name: 'John', chapters: 21, urlPath: 'john' },
      { abbr: 'acts', name: 'Acts', chapters: 28, urlPath: 'acts' },
      { abbr: 'rom', name: 'Romans', chapters: 16, urlPath: 'rom' },
      { abbr: '1-cor', name: '1 Corinthians', chapters: 16, urlPath: '1-cor' },
      { abbr: '2-cor', name: '2 Corinthians', chapters: 13, urlPath: '2-cor' },
      { abbr: 'gal', name: 'Galatians', chapters: 6, urlPath: 'gal' },
      { abbr: 'eph', name: 'Ephesians', chapters: 6, urlPath: 'eph' },
      { abbr: 'philip', name: 'Philippians', chapters: 4, urlPath: 'philip' },
      { abbr: 'col', name: 'Colossians', chapters: 4, urlPath: 'col' },
      { abbr: '1-thes', name: '1 Thessalonians', chapters: 5, urlPath: '1-thes' },
      { abbr: '2-thes', name: '2 Thessalonians', chapters: 3, urlPath: '2-thes' },
      { abbr: '1-tim', name: '1 Timothy', chapters: 6, urlPath: '1-tim' },
      { abbr: '2-tim', name: '2 Timothy', chapters: 4, urlPath: '2-tim' },
      { abbr: 'titus', name: 'Titus', chapters: 3, urlPath: 'titus' },
      { abbr: 'philem', name: 'Philemon', chapters: 1, urlPath: 'philem' },
      { abbr: 'heb', name: 'Hebrews', chapters: 13, urlPath: 'heb' },
      { abbr: 'james', name: 'James', chapters: 5, urlPath: 'james' },
      { abbr: '1-pet', name: '1 Peter', chapters: 5, urlPath: '1-pet' },
      { abbr: '2-pet', name: '2 Peter', chapters: 3, urlPath: '2-pet' },
      { abbr: '1-jn', name: '1 John', chapters: 5, urlPath: '1-jn' },
      { abbr: '2-jn', name: '2 John', chapters: 1, urlPath: '2-jn' },
      { abbr: '3-jn', name: '3 John', chapters: 1, urlPath: '3-jn' },
      { abbr: 'jude', name: 'Jude', chapters: 1, urlPath: 'jude' },
      { abbr: 'rev', name: 'Revelation', chapters: 22, urlPath: 'rev' },
    ],
  },
  {
    id: 'bofm',
    label: 'Book of Mormon',
    urlVolume: 'bofm',
    books: [
      { abbr: '1-ne', name: '1 Nephi', chapters: 22, urlPath: '1-ne' },
      { abbr: '2-ne', name: '2 Nephi', chapters: 33, urlPath: '2-ne' },
      { abbr: 'jacob', name: 'Jacob', chapters: 7, urlPath: 'jacob' },
      { abbr: 'enos', name: 'Enos', chapters: 1, urlPath: 'enos' },
      { abbr: 'jarom', name: 'Jarom', chapters: 1, urlPath: 'jarom' },
      { abbr: 'omni', name: 'Omni', chapters: 1, urlPath: 'omni' },
      { abbr: 'w-of-m', name: 'Words of Mormon', chapters: 1, urlPath: 'w-of-m' },
      { abbr: 'mosiah', name: 'Mosiah', chapters: 29, urlPath: 'mosiah' },
      { abbr: 'alma', name: 'Alma', chapters: 63, urlPath: 'alma' },
      { abbr: 'hel', name: 'Helaman', chapters: 16, urlPath: 'hel' },
      { abbr: '3-ne', name: '3 Nephi', chapters: 30, urlPath: '3-ne' },
      { abbr: '4-ne', name: '4 Nephi', chapters: 1, urlPath: '4-ne' },
      { abbr: 'morm', name: 'Mormon', chapters: 9, urlPath: 'morm' },
      { abbr: 'ether', name: 'Ether', chapters: 15, urlPath: 'ether' },
      { abbr: 'moro', name: 'Moroni', chapters: 10, urlPath: 'moro' },
    ],
  },
  {
    id: 'dc',
    label: 'Doctrine and Covenants',
    urlVolume: 'dc-testament',
    books: [
      // D&C ships as one "book" with 138 sections (treated as chapters).
      { abbr: 'dc', name: 'Doctrine and Covenants', chapters: 138, urlPath: 'dc' },
    ],
  },
  {
    id: 'pgp',
    label: 'Pearl of Great Price',
    urlVolume: 'pgp',
    books: [
      { abbr: 'moses', name: 'Moses', chapters: 8, urlPath: 'moses' },
      { abbr: 'abr', name: 'Abraham', chapters: 5, urlPath: 'abr' },
      { abbr: 'js-m', name: 'Joseph Smith—Matthew', chapters: 1, urlPath: 'js-m' },
      { abbr: 'js-h', name: 'Joseph Smith—History', chapters: 1, urlPath: 'js-h' },
      { abbr: 'a-of-f', name: 'Articles of Faith', chapters: 1, urlPath: 'a-of-f' },
    ],
  },
]

/** Build the churchofjesuschrist.org URL for a chapter (or verse range). */
export function buildChurchUrl(
  volumeUrl: string,
  bookUrl: string,
  chapter: number,
  verseStart?: number,
  verseEnd?: number,
): string {
  let url = `https://www.churchofjesuschrist.org/study/scriptures/${volumeUrl}/${bookUrl}/${chapter}?lang=eng`
  if (verseStart) {
    const range = verseEnd && verseEnd !== verseStart ? `${verseStart}-${verseEnd}` : `${verseStart}`
    url += `&id=${range}#${verseStart}`
  }
  return url
}
