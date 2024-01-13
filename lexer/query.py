from pygments.lexer import RegexLexer, bygroups
from pygments.token import *

__all__ = ['QueryLexer']

class QueryLexer(RegexLexer):
     name = "query"
     aliases = ["query"]
     filenames = []
     mimetypes = ['text/x-query']

     tokens = {
          'root': [
               (r'(amount|account|total|payee|commodity|date|filename|note)', Keyword.Constant),
               (r'AND|OR|NOT', Keyword),
               (r'\[[^\]]+\]', Name.Attribute),
               (r'".*"', String),
               (r'/[^ }]+/[a-z]?', String),
               (r'[+-]?(?:[0-9,])+(\.(?:[0-9,])+)?', Number),
               (r'[A-Z]+', String),
               (r'[><=~]', Operator),
               (r'[^ }]+(/[^ }]+)+', String),
               (r'[^ }]+(:[^ }]+)+', String),
               (r' ', Text),
               (r'\n', Text),
          ]
     }
