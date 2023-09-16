from pygments import lexers
from pygments.lexers._mapping import LEXERS

name = "LedgerLexer"
custom_namespace = {}
with open("docs/lexer/ledger.py", 'rb') as f:
    exec(f.read(), custom_namespace)

cls = custom_namespace[name]
LEXERS[name] = ('TODO', name, ('ledger',), ('*.ledger', '*.journal'), ('text/x-ledger',))
lexers._lexer_cache[name] = cls


