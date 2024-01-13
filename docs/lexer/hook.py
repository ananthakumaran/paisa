from pygments import lexers
from pygments.lexers._mapping import LEXERS

custom_namespace = {}

with open("docs/lexer/ledger.py", 'rb') as f:
    exec(f.read(), custom_namespace)

ledger_name = "LedgerLexer"
cls = custom_namespace[ledger_name]
LEXERS[ledger_name] = ('TODO', ledger_name, ('ledger',), ('*.ledger', '*.journal'), ('text/x-ledger',))
lexers._lexer_cache[ledger_name] = cls


with open("docs/lexer/query.py", 'rb') as f:
    exec(f.read(), custom_namespace)

query_name = "QueryLexer"
cls = custom_namespace[query_name]
LEXERS[query_name] = ('TODO', query_name, ('query',), (), ('text/x-query',))
lexers._lexer_cache[query_name] = cls


with open("docs/lexer/sheet.py", 'rb') as f:
    exec(f.read(), custom_namespace)

sheet_name = "SheetLexer"
cls = custom_namespace[sheet_name]
LEXERS[sheet_name] = ('TODO', sheet_name, ('sheet',), (), ('text/x-sheet',))
lexers._lexer_cache[sheet_name] = cls

