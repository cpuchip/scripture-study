"""Batch convert OpenStax pages most relevant to the light/QM/observer study.

University Physics Volume 3:
  - Chapter 1: The Nature of Light (sections 1-introduction through 1-7)
  - Chapter 6: Photons and Matter Waves (6-introduction through 6-6)
  - Chapter 7: Quantum Mechanics (7-introduction through 7-6)
"""

from __future__ import annotations
import sys
import time
from pathlib import Path

# Add scripts dir to path so we can import the converter
sys.path.insert(0, str(Path(__file__).resolve().parent))
from importlib import import_module
mod = import_module('openstax-to-md')
convert_page = mod.convert_page

BOOK = 'university-physics-volume-3'

PAGES = [
    # Chapter 1 - The Nature of Light
    '1-introduction',
    '1-1-the-propagation-of-light',
    '1-2-the-law-of-reflection',
    '1-3-refraction',
    '1-4-total-internal-reflection',
    '1-5-dispersion',
    '1-6-huygenss-principle',
    '1-7-polarization',
    # Chapter 6 - Photons and Matter Waves
    '6-introduction',
    '6-1-blackbody-radiation',
    '6-2-photoelectric-effect',
    '6-3-the-compton-effect',
    '6-4-bohrs-model-of-the-hydrogen-atom',
    '6-5-de-broglies-matter-waves',
    '6-6-wave-particle-duality',
    # Chapter 7 - Quantum Mechanics
    '7-introduction',
    '7-1-wave-functions',
    '7-2-the-heisenberg-uncertainty-principle',
    '7-3-the-schrodinger-equation',
    '7-4-the-quantum-particle-in-a-box',
    '7-5-the-quantum-harmonic-oscillator',
    '7-6-the-quantum-tunneling-of-particles-through-potential-barriers',
]

succeeded = 0
failed = []

for slug in PAGES:
    try:
        convert_page(BOOK, slug)
        succeeded += 1
    except Exception as e:
        print(f'FAILED {slug}: {e}')
        failed.append((slug, str(e)))
    time.sleep(0.5)  # be polite

print(f'\nDone. {succeeded}/{len(PAGES)} succeeded.')
if failed:
    print('Failures:')
    for s, e in failed:
        print(f'  {s}: {e}')
