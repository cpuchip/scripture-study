"""Inspect OpenStax HTML structure to plan the converter."""
import requests
from bs4 import BeautifulSoup

url = 'https://openstax.org/books/university-physics-volume-3/pages/1-1-the-propagation-of-light'
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'}, timeout=30)
print('status', r.status_code, 'len', len(r.text))

soup = BeautifulSoup(r.text, 'lxml')

selectors = ['main', 'article', '[data-type="page"]', '.book-content', '#main-content', '#content', '#main']
for sel in selectors:
    el = soup.select_one(sel)
    if el:
        print(f'found via: {sel}  text-len={len(el.get_text())}  child-count={len(list(el.children))}')

print()
print('All <math> elements:', len(soup.select('math')))
print('All <img> elements:', len(soup.select('img')))
print('All <h1>:', [h.get_text(strip=True)[:80] for h in soup.select('h1')])
print('All <h2>:', [h.get_text(strip=True)[:80] for h in soup.select('h2')[:5]])
print('All <h3>:', [h.get_text(strip=True)[:80] for h in soup.select('h3')[:5]])
print('data-type attrs:', set(t.get('data-type', '') for t in soup.select('[data-type]'))[:20] if False else list(set(t.get('data-type', '') for t in soup.select('[data-type]')))[:20])

# Sample a math element
mathels = soup.select('math')
if mathels:
    print('\nFirst math element (raw):')
    print(str(mathels[0])[:500])

# Check if data-mathml is used
print('\nElements with data-mathml:', len(soup.select('[data-mathml]')))
print('Elements with class containing math:', len(soup.select('[class*="math"]')))
