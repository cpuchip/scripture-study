from mathml_to_latex.converter import MathMLToLaTeX

c = MathMLToLaTeX()

samples = {
    'frac':       '<math xmlns="http://www.w3.org/1998/Math/MathML"><mfrac><mi>x</mi><mn>2</mn></mfrac></math>',
    'semantics':  '<math display="inline"><semantics><mrow><mi>A</mi><mtext>prime</mtext></mrow></semantics></math>',
    'pythag':     '<math><msup><mi>c</mi><mn>2</mn></msup><mo>=</mo><msup><mi>a</mi><mn>2</mn></msup><mo>+</mo><msup><mi>b</mi><mn>2</mn></msup></math>',
    'integral':   '<math><msubsup><mo>&#x222B;</mo><mn>0</mn><mi>&#x221E;</mi></msubsup><mi>f</mi><mo>(</mo><mi>x</mi><mo>)</mo><mi>dx</mi></math>',
    'einstein':   '<math><mi>E</mi><mo>=</mo><mi>m</mi><msup><mi>c</mi><mn>2</mn></msup></math>',
    'schrod':     '<math><mi>i</mi><mi>&#x210F;</mi><mfrac><mrow><mo>&#x2202;</mo><mi>&#x03A8;</mi></mrow><mrow><mo>&#x2202;</mo><mi>t</mi></mrow></mfrac><mo>=</mo><mover accent="true"><mi>H</mi><mo>^</mo></mover><mi>&#x03A8;</mi></math>',
}

for name, mml in samples.items():
    try:
        result = c.convert(mml)
        print(f'{name:12s} -> {result!r}')
    except Exception as e:
        print(f'{name:12s} FAILED: {type(e).__name__}: {e}')
