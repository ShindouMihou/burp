package burper

var OPENING_KEY = []byte{'['}
var SEPERATOR_KEY = []byte{':'}
var CLOSING_KEY = []byte{']'}
var NEWLINE_KEY = []byte{'\n'}
var AS_TOKEN = []byte{'A', 'S'}
var COMPLETE_PREFIX_KEY = []byte{'b', 'u', 'r', 'p', ':'}

var COMMENT_KEY = []byte{'#'}
var EQUALS_KEY = []byte{'='}
