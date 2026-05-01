"""Investigate the math duplication issue."""
import sys
sys.stdout.reconfigure(encoding='utf-8')
import requests
from bs4 import BeautifulSoup

url = 'https://openstax.org/books/university-physics-volume-3/pages/1-1-the-propagation-of-light'
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'}, timeout=30)
r.encoding = 'utf-8'  # force UTF-8 in case server didn't set it
print('apparent encoding:', r.apparent_encoding)
print('declared encoding:', r.encoding)

soup = BeautifulSoup(r.text, 'lxml')

# Find the eq for c = 2.99...
for eq in soup.select('[data-type="equation"]')[:3]:
    print('\n=== equation ===')
    print(str(eq)[:1500])
