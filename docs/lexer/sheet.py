from pygments.lexer import RegexLexer, bygroups
from pygments.token import *

__all__ = ['SheetLexer']

class SheetLexer(RegexLexer):
     name = "sheet"
     aliases = ["sheet"]
     filenames = []
     mimetypes = ['text/x-sheet']

     tokens = {
          'root': [
               (r';(.*?)$', Comment.Single),
               (r'//(.*?)$', Comment.Single),
               (r'([a-z_]+)(\()', bygroups(Name.Function, Text)),
               (r'AND|OR', Keyword),
               (r'#.+', Generic.Heading),
               (r'[*+/^=-]', Operator),
               (r'{', Operator, 'query'),
               (r'[+-]?(?:[0-9,])+(\.(?:[0-9,])+)?(%)?', Number),
          ],
          'query': [
               (r'}', Operator, '#pop'),
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
