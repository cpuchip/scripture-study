"""Check raw bytes for encoding."""
import sys
sys.stdout.reconfigure(encoding='utf-8')
import requests

url = 'https://openstax.org/books/university-physics-volume-3/pages/1-1-the-propagation-of-light'
r = requests.get(url, headers={'User-Agent': 'Mozilla/5.0'}, timeout=30)
print('content-type header:', r.headers.get('content-type'))
print('apparent encoding:', r.apparent_encoding)
# Find the bytes for the multiplication-sign equation area
b = r.content
# Find "2.99792458" which is in the speed of light eq
i = b.find(b'2.99792458')
if i >= 0:
    print(f'\nfound at byte offset {i}')
    print('next 200 bytes (raw):', b[i:i+200])
    print('\nas utf-8:', b[i:i+200].decode('utf-8', errors='replace'))
    print('\nas latin-1:', b[i:i+200].decode('latin-1', errors='replace'))
