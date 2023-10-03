from pygments.lexer import RegexLexer, bygroups
from pygments.token import *

__all__ = ['LedgerLexer']

class LedgerLexer(RegexLexer):
     name = "ledger"
     aliases = ["ledger"]
     filenames = ['*.ledger', '*.journal']
     mimetypes = ['text/x-ledger']

     tokens = {
          'root': [
               (r'=.*$', Text),
               (r'^(include)( )(.+)$',
                bygroups(Keyword, Text, Keyword.Namespace)),
               (r'^(\d{4}/\d{2}/\d{2})(/[*][^*]+[*]/)(\s*)(.+)(/[*][^*]+[*]/)$',
                bygroups(Name.Attribute, Comment.Multiline, Text, Generic.Heading, Comment.Multiline)),
               (r';(.*?)$', Comment.Single),
               (r'/(\\\n)?[*](.|\n)*?[*](\\\n)?/', Comment.Multiline),
               (r'^(\d{4}/\d{2}/\d{2})(\s*)(.+)$',
                bygroups(Name.Attribute, Text, Generic.Heading)),
               (r'\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}', Name.Attribute),
               (r'\d{4}/\d{2}/\d{2}', Name.Attribute),
               (r'[@]', Operator),
               (r'[+-]?(?:[0-9,])+(\.(?:[0-9,])+)?', Number),
               (r'[$€]', Keyword),
               (r'include', Keyword),
               (r'\b[A-Z_]+\b', Keyword),
               (r'^\s{4}[^\][(); \t\n]((?!\s{2})[^/\][();\t\n])*', String),
               (r'^[^\][(); \t\n]((?!\s{2})[^/\][();\t\n])*$', String),
          ]
     }
